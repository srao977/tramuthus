package runtime

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"bar_sequence_lab/generator/internal/alpaca"
	"bar_sequence_lab/generator/internal/config"
	"bar_sequence_lab/generator/internal/persistence"
	"bar_sequence_lab/generator/internal/pipeline"
	"bar_sequence_lab/generator/internal/sequence"
	"bar_sequence_lab/generator/internal/types"
)

var runSeq atomic.Uint64

type Engine struct {
	cfg          config.Config
	runID        string
	seq          *sequence.Authority
	parts        map[string]*pipeline.Partition
	mongo        *persistence.Store
	unknown      atomic.Uint64
	droppedValid atomic.Uint64
}

func NewEngine(cfg config.Config) (*Engine, error) {
	runID := fmt.Sprintf("%s-%d", time.Now().UTC().Format("20060102T150405Z"), runSeq.Add(1))
	seq := sequence.New(runID)
	var store *persistence.Store
	if cfg.MongoEnabled {
		s, err := persistence.OpenMongo(context.Background(), cfg)
		if err != nil {
			return nil, err
		}
		store = s
	}
	parts := map[string]*pipeline.Partition{}
	for _, g := range cfg.SelectedGroups {
		w, err := persistence.Open(cfg.DataDir, runID, g)
		if err != nil {
			if store != nil {
				_ = store.Close(context.Background())
			}
			return nil, err
		}
		var mw *persistence.MongoWriter
		if store != nil {
			mw = store.Writer(g)
		}
		parts[g] = pipeline.NewPartition(g, cfg, seq, w, mw)
	}
	return &Engine{cfg: cfg, runID: runID, seq: seq, parts: parts, mongo: store}, nil
}

func (e *Engine) Close() {
	for _, p := range e.parts {
		_ = p.CloseWriter()
	}
	if e.mongo != nil {
		_ = e.mongo.Close(context.Background())
	}
}

func (e *Engine) Run(ctx context.Context, source LiveSource) error {
	for _, line := range config.StartupLines(e.cfg) {
		log.Printf("STARTUP %s", line)
	}
	log.Printf("STARTUP collection_run_id=%s", e.runID)
	log.Printf("STARTUP persistence_paths:")
	for _, g := range e.cfg.SelectedGroups {
		log.Printf("STARTUP   %s -> %s", g, e.parts[g].Path())
	}
	if e.mongo != nil {
		log.Printf("STARTUP mongo host=%s db=%s collection=%s status=connected", e.mongo.Host(), e.mongo.Database(), e.mongo.Collection())
	} else {
		log.Printf("STARTUP mongo status=disabled")
	}

	var wg sync.WaitGroup
	partCtx, partCancel := context.WithCancel(context.Background())
	defer partCancel()
	for _, p := range e.parts {
		p.Start(partCtx, &wg)
	}

	ingest := make(chan types.Observation, e.cfg.BufferCapacity)
	srcCtx, srcCancel := context.WithCancel(ctx)
	defer srcCancel()

	var srcErr error
	var srcWG sync.WaitGroup
	srcWG.Add(1)
	go func() {
		defer srcWG.Done()
		srcErr = source.Run(srcCtx, e.cfg.SubscribeSymbols, ingest)
		close(ingest)
	}()

	metrics := time.NewTicker(e.cfg.MetricsInterval)
	defer metrics.Stop()

	var deadline <-chan time.Time
	if e.cfg.Duration > 0 {
		t := time.NewTimer(e.cfg.Duration)
		defer t.Stop()
		deadline = t.C
	}

	var acceptedTotal uint64
	dispatching := true
	for dispatching {
		select {
		case <-ctx.Done():
			srcCancel()
			dispatching = false
		case <-deadline:
			log.Printf("duration reached; shutting down")
			srcCancel()
			dispatching = false
		case <-metrics.C:
			e.printMetrics(source)
		case obs, ok := <-ingest:
			if !ok {
				dispatching = false
				break
			}
			if err := e.dispatch(srcCtx, obs); err != nil {
				log.Printf("CRITICAL %v", err)
				e.droppedValid.Add(1)
				srcCancel()
				dispatching = false
			}
			acceptedTotal++
			if e.cfg.MaxBars > 0 && int(acceptedTotal) >= e.cfg.MaxBars {
				log.Printf("max_bars=%d reached; shutting down", e.cfg.MaxBars)
				srcCancel()
				dispatching = false
			}
		}
	}

	srcCancel()
	srcWG.Wait()
	for obs := range ingest {
		if err := e.dispatch(context.Background(), obs); err != nil {
			log.Printf("CRITICAL %v", err)
			e.droppedValid.Add(1)
		}
	}
	for _, p := range e.parts {
		p.CloseIngress()
	}
	wg.Wait()
	partCancel()
	for _, g := range e.cfg.SelectedGroups {
		_ = e.parts[g].CloseWriter()
	}
	e.printShutdown(source)
	if e.mongo != nil {
		_ = e.mongo.Close(context.Background())
		e.mongo = nil
	}
	if srcErr != nil && srcCtx.Err() == nil {
		return srcErr
	}
	return nil
}

func (e *Engine) dispatch(ctx context.Context, obs types.Observation) error {
	partID, ok := e.cfg.SymbolToPartition[obs.Symbol]
	if !ok {
		e.unknown.Add(1)
		log.Printf("unknown symbol %s not in selected groups; rejected", obs.Symbol)
		return nil
	}
	p := e.parts[partID]
	return p.Accept(ctx, obs)
}

func (e *Engine) printMetrics(source LiveSource) {
	log.Printf("METRICS collection_run_id=%s dropped_valid=%d unknown_symbol=%d", e.runID, e.droppedValid.Load(), e.unknown.Load())
	h := source.Health()
	log.Printf("ALPACA state=%s last_error=%q heartbeat_misses=%d last_pong_rtt=%s", h.State, h.LastError, h.HeartbeatMisses, h.LastPongRTT)
	reg, dup, latest := e.seq.Stats()
	log.Printf("SEQUENCE source_time_regressions=%d duplicate_arrivals=%d latest=%v", reg, dup, latest)
	for _, g := range e.cfg.SelectedGroups {
		p := e.parts[g]
		acc, per, crit := p.Counts()
		log.Printf("PARTITION %s accepted=%d jsonl_persisted=%d mongo_persisted=%d critical=%d", g, acc, per, p.MongoCounts(), crit)
		for _, snap := range p.Snapshot() {
			log.Printf("BUFFER %s", pipeline.FormatBuffer(snap))
		}
		ws := p.WriterSnapshot()
		log.Printf("JSONL %s path=%s persisted=%d batches=%d failed_batches=%d last_batch=%d last_latency=%s pending=%d",
			g, ws.Path, ws.Persisted, ws.Batches, ws.FailedBatches, ws.LastBatchSize, ws.LastWriteLatency, ws.Pending)
		ms := p.MongoSnapshot()
		log.Printf("MONGO %s attempted=%d persisted=%d failed=%d pending=%d", g, ms.Attempted, ms.Persisted, ms.Failed, ms.Pending)
	}
}

func (e *Engine) printShutdown(source LiveSource) {
	log.Printf("SHUTDOWN collection_run_id=%s", e.runID)
	e.printMetrics(source)
	var unflushedJSONL uint64
	var unflushedMongo uint64
	var mongoFailed uint64
	for _, g := range e.cfg.SelectedGroups {
		ws := e.parts[g].WriterSnapshot()
		ms := e.parts[g].MongoSnapshot()
		unflushedJSONL += uint64(ws.Pending)
		unflushedMongo += uint64(ms.Pending)
		mongoFailed += ms.Failed
		log.Printf("SHUTDOWN jsonl_%s persisted=%d pending=%d file=%s", g, ws.Persisted, ws.Pending, ws.Path)
		log.Printf("SHUTDOWN mongo_%s attempted=%d persisted=%d failed=%d pending=%d", g, ms.Attempted, ms.Persisted, ms.Failed, ms.Pending)
	}
	log.Printf("SHUTDOWN dropped_valid_count=%d (target 0)", e.droppedValid.Load())
	log.Printf("SHUTDOWN jsonl_unflushed_total=%d mongo_unflushed_total=%d mongo_failed_total=%d", unflushedJSONL, unflushedMongo, mongoFailed)
}

type LiveSource interface {
	Run(ctx context.Context, symbols []string, out chan<- types.Observation) error
	Health() alpaca.Health
}

func NewAlpacaSource(cfg config.Config) *alpaca.Stream {
	return alpaca.NewStream(cfg.StreamURL, cfg.Feed, alpaca.Credentials{
		Key:    cfg.APIKey,
		Secret: cfg.APISecret,
	})
}
