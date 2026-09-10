package pipeline

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"bar_sequence_lab/generator/internal/buffer"
	"bar_sequence_lab/generator/internal/config"
	"bar_sequence_lab/generator/internal/persistence"
	"bar_sequence_lab/generator/internal/sequence"
	"bar_sequence_lab/generator/internal/types"
)

type Partition struct {
	ID      string
	ingress *buffer.Stage
	ready   *buffer.Stage
	persist *buffer.Stage
	writer  *persistence.Writer
	mongo   *persistence.MongoWriter
	seq     *sequence.Authority
	cfg     config.Config

	accepted       atomic.Uint64
	persisted      atomic.Uint64
	mongoPersisted atomic.Uint64
	critical       atomic.Uint64
}

func NewPartition(id string, cfg config.Config, seq *sequence.Authority, writer *persistence.Writer, mongo *persistence.MongoWriter) *Partition {
	return &Partition{
		ID:      id,
		ingress: buffer.NewStage(id+"1", cfg.BufferCapacity),
		ready:   buffer.NewStage(id+"2", cfg.BufferCapacity),
		persist: buffer.NewStage(id+"3", cfg.BufferCapacity),
		writer:  writer,
		mongo:   mongo,
		seq:     seq,
		cfg:     cfg,
	}
}

func (p *Partition) Start(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(3)
	go func() {
		defer wg.Done()
		p.sequenceLoop(ctx)
	}()
	go func() {
		defer wg.Done()
		p.transferLoop(ctx)
	}()
	go func() {
		defer wg.Done()
		p.persistLoop()
	}()
}

func (p *Partition) Accept(ctx context.Context, obs types.Observation) error {
	obs.PartitionID = p.ID
	if err := p.ingress.Push(ctx, obs); err != nil {
		p.critical.Add(1)
		return err
	}
	return nil
}

func (p *Partition) sequenceLoop(ctx context.Context) {
	defer p.ready.Close()
	for {
		obs, ok, err := p.ingress.Pop(ctx)
		if err != nil || !ok {
			return
		}
		numbered := p.seq.Accept(obs)
		p.accepted.Add(1)
		if err := p.ready.Push(context.Background(), numbered); err != nil {
			p.critical.Add(1)
			log.Printf("CRITICAL %v", err)
			return
		}
	}
}

func (p *Partition) transferLoop(ctx context.Context) {
	defer p.persist.Close()
	for {
		obs, ok, err := p.ready.Pop(ctx)
		if err != nil || !ok {
			return
		}
		if err := p.persist.Push(context.Background(), obs); err != nil {
			p.critical.Add(1)
			log.Printf("CRITICAL %v", err)
			return
		}
	}
}

func (p *Partition) persistLoop() {
	ticker := time.NewTicker(p.cfg.FlushInterval)
	defer ticker.Stop()
	batch := make([]types.Observation, 0, p.cfg.FlushCount)
	flush := func() (stop bool) {
		if len(batch) == 0 {
			return false
		}
		for {
			if err := p.writer.WriteBatch(batch); err != nil {
				log.Printf("jsonl write failed partition=%s: %v; retrying (no silent drop)", p.ID, err)
				time.Sleep(50 * time.Millisecond)
				continue
			}
			p.persisted.Add(uint64(len(batch)))
			break
		}
		if p.mongo != nil {
			for {
				if err := p.mongo.WriteBatch(context.Background(), batch); err != nil {
					if persistence.IsDuplicateKey(err) {
						p.critical.Add(1)
						log.Printf("CRITICAL mongo duplicate-key partition=%s: %v (not retrying; indexes not modified)", p.ID, err)
						return true
					}
					log.Printf("mongo write failed partition=%s: %v; retrying (no silent drop)", p.ID, err)
					time.Sleep(50 * time.Millisecond)
					continue
				}
				p.mongoPersisted.Add(uint64(len(batch)))
				break
			}
		}
		batch = batch[:0]
		return false
	}
	for {
		select {
		case <-ticker.C:
			if flush() {
				return
			}
		case obs, ok := <-p.persist.C():
			if !ok {
				_ = flush()
				return
			}
			batch = append(batch, obs)
			if len(batch) >= p.cfg.FlushCount {
				if flush() {
					return
				}
			}
		}
	}
}

func (p *Partition) CloseIngress() {
	p.ingress.Close()
}

func (p *Partition) Snapshot() map[string]buffer.Snapshot {
	return map[string]buffer.Snapshot{
		p.ID + "1": p.ingress.Snapshot(),
		p.ID + "2": p.ready.Snapshot(),
		p.ID + "3": p.persist.Snapshot(),
	}
}

func (p *Partition) Counts() (accepted, persisted, critical uint64) {
	return p.accepted.Load(), p.persisted.Load(), p.critical.Load()
}

func (p *Partition) MongoCounts() uint64 {
	return p.mongoPersisted.Load()
}

func (p *Partition) MongoSnapshot() persistence.MongoStats {
	pending := p.persist.Snapshot().Depth
	if p.mongo == nil {
		return persistence.MongoStats{Partition: p.ID, Pending: pending}
	}
	return p.mongo.Snapshot(pending)
}

func (p *Partition) WriterSnapshot() persistence.Stats {
	pending := p.persist.Snapshot().Depth
	return p.writer.Snapshot(pending)
}

func (p *Partition) CloseWriter() error {
	return p.writer.Close()
}

func (p *Partition) Path() string {
	return p.writer.Path()
}

func FormatBuffer(s buffer.Snapshot) string {
	pct := 0.0
	if s.Capacity > 0 {
		pct = 100 * float64(s.Depth) / float64(s.Capacity)
	}
	return fmt.Sprintf("%s depth=%d/%d (%.0f%%) high=%d oldest=%s", s.Name, s.Depth, s.Capacity, pct, s.HighWater, s.OldestAge.Truncate(time.Millisecond))
}
