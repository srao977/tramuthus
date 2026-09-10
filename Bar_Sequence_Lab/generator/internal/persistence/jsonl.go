package persistence

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"bar_sequence_lab/generator/internal/types"
)

type Stats struct {
	Partition        string
	Path             string
	Persisted        uint64
	Batches          uint64
	FailedBatches    uint64
	LastBatchSize    int
	LastWriteLatency time.Duration
	Pending          int
}

type Writer struct {
	partition string
	path      string

	mu    sync.Mutex
	file  *os.File
	buf   *bufio.Writer
	stats Stats
}

func Open(dataDir, runID, partition string) (*Writer, error) {
	dir := filepath.Join(dataDir, runID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("mkdir jsonl dir: %w", err)
	}
	path := filepath.Join(dir, fmt.Sprintf("partition_%s.jsonl", partition))
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open jsonl: %w", err)
	}
	return &Writer{
		partition: partition,
		path:      path,
		file:      f,
		buf:       bufio.NewWriterSize(f, 64*1024),
		stats: Stats{
			Partition: partition,
			Path:      path,
		},
	}, nil
}

func (w *Writer) WriteBatch(obs []types.Observation) error {
	if len(obs) == 0 {
		return nil
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	start := time.Now()
	now := start.UTC()
	for i := range obs {
		obs[i].PersistedTime = now
		line, err := json.Marshal(obs[i])
		if err != nil {
			w.stats.FailedBatches++
			return fmt.Errorf("marshal jsonl: %w", err)
		}
		if _, err := w.buf.Write(line); err != nil {
			w.stats.FailedBatches++
			return fmt.Errorf("write jsonl: %w", err)
		}
		if err := w.buf.WriteByte('\n'); err != nil {
			w.stats.FailedBatches++
			return err
		}
	}
	if err := w.buf.Flush(); err != nil {
		w.stats.FailedBatches++
		return fmt.Errorf("flush jsonl: %w", err)
	}
	if err := w.file.Sync(); err != nil {
		w.stats.FailedBatches++
		return fmt.Errorf("sync jsonl: %w", err)
	}
	w.stats.Persisted += uint64(len(obs))
	w.stats.Batches++
	w.stats.LastBatchSize = len(obs)
	w.stats.LastWriteLatency = time.Since(start)
	return nil
}

func (w *Writer) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.buf != nil {
		_ = w.buf.Flush()
	}
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

func (w *Writer) Snapshot(pending int) Stats {
	w.mu.Lock()
	defer w.mu.Unlock()
	s := w.stats
	s.Pending = pending
	return s
}

func (w *Writer) Path() string {
	return w.path
}
