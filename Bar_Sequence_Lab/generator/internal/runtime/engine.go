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
	acceptedBySymbol := make(map[string]uint64, len(e.cfg.SubscribeSymbols))
	for _, symbol := range e.cfg.SubscribeSymbols {
		acceptedBySymbol[symbol] = 0
	}
	stopReason := "SOURCE_ENDED"
	dispatching := true
	for dispatching {
		select {
		case <-ctx.Done():
			stopReason = "SIGNAL_OR_CONTEXT"
			srcCancel()
			dispatching = false
		case <-deadline:
			stopReason = "DURATION_LIMIT"
			if e.cfg.Duration == 2*time.Hour {
				stopReason = "TWO_HOUR_LIMIT"
			}
			log.Printf("duration reached; stop_reason=%s; shutting down", stopReason)
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
			if _, configured := acceptedBySymbol[obs.Symbol]; configured {
				acceptedBySymbol[obs.Symbol]++
			}
			if e.cfg.TargetBarsPerSymbol > 0 && allSymbolsReached(acceptedBySymbol, uint64(e.cfg.TargetBarsPerSymbol)) {
				stopReason = fmt.Sprintf("ALL_SYMBOLS_REACHED_%d", e.cfg.TargetBarsPerSymbol)
				log.Printf("target_bars_per_symbol=%d reached; stop_reason=%s; shutting down", e.cfg.TargetBarsPerSymbol, stopReason)
				srcCancel()
				dispatching = false
				break
			}
			if e.cfg.MaxBars > 0 && int(acceptedTotal) >= e.cfg.MaxBars {
				stopReason = "MAX_BARS_REACHED"
				log.Printf("max_bars=%d reached; stop_reason=%s; shutting down", e.cfg.MaxBars, stopReason)
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
	log.Printf("SHUTDOWN stop_reason=%s", stopReason)
	if e.mongo != nil {
		_ = e.mongo.Close(context.Background())
		e.mongo = nil
	}
	if srcErr != nil && srcCtx.Err() == nil {
		return srcErr
	}
	return nil
}

func allSymbolsReached(counts map[string]uint64, target uint64) bool {
	if target == 0 || len(counts) == 0 {
		return false
	}
	for _, count := range counts {
		if count < target {
			return false
		}
	}
	return true
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
	minimum, maximum, at63, at72, at90, at120 := readinessSummary(e.cfg.SubscribeSymbols, latest)
	log.Printf("PHASE_READINESS symbols=%d min=%d max=%d at_least_63=%d at_least_72=%d at_least_90=%d at_least_120=%d",
		len(e.cfg.SubscribeSymbols), minimum, maximum, at63, at72, at90, at120)
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

func readinessSummary(symbols []string, counts map[string]uint64) (minimum, maximum uint64, at63, at72, at90, at120 int) {
	if len(symbols) == 0 {
		return 0, 0, 0, 0, 0, 0
	}
	minimum = ^uint64(0)
	for _, symbol := range symbols {
		count := counts[symbol]
		minimum = min(minimum, count)
		maximum = max(maximum, count)
		if count >= 63 {
			at63++
		}
		if count >= 72 {
			at72++
		}
		if count >= 90 {
			at90++
		}
		if count >= 120 {
			at120++
		}
	}
	return minimum, maximum, at63, at72, at90, at120
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
