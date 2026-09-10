package pipeline

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"bar_sequence_lab/generator/internal/config"
	"bar_sequence_lab/generator/internal/persistence"
	"bar_sequence_lab/generator/internal/sequence"
	"bar_sequence_lab/generator/internal/types"
)

func TestPartitionAcceptedArrivalPersisted(t *testing.T) {
	dir := t.TempDir()
	w, err := persistence.Open(dir, "run1", "A")
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{BufferCapacity: 16, FlushCount: 10, FlushInterval: 15 * time.Millisecond}
	p := NewPartition("A", cfg, sequence.New("run1"), w, nil)
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	p.Start(ctx, &wg)
	t0 := time.Date(2026, 9, 10, 15, 1, 0, 0, time.UTC)
	t1 := t0.Add(-time.Minute)
	if err := p.Accept(ctx, types.Observation{Symbol: "AAPL", SourceEventTime: t0, SourceTimestampText: "t0", ReceivedTime: time.Now().UTC(), Open: 1, High: 1, Low: 1, Close: 1, Volume: 1, Interval: "1Min", AlpacaMessageType: "b", PayloadHash: "1"}); err != nil {
		t.Fatal(err)
	}
	if err := p.Accept(ctx, types.Observation{Symbol: "AAPL", SourceEventTime: t1, SourceTimestampText: "t1", ReceivedTime: time.Now().UTC(), Open: 1, High: 1, Low: 1, Close: 1, Volume: 1, Interval: "1Min", AlpacaMessageType: "b", PayloadHash: "2"}); err != nil {
		t.Fatal(err)
	}
	p.CloseIngress()
	wg.Wait()
	cancel()
	_ = p.CloseWriter()

	f, err := os.Open(filepath.Join(dir, "run1", "partition_A.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var seqs []uint64
	var flags []bool
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		var obs types.Observation
		if err := json.Unmarshal(sc.Bytes(), &obs); err != nil {
			t.Fatal(err)
		}
		seqs = append(seqs, obs.GeneratorSequenceNo)
		flags = append(flags, obs.SourceTimeRegression)
	}
	if len(seqs) != 2 || seqs[0] != 1 || seqs[1] != 2 {
		t.Fatalf("seqs=%v", seqs)
	}
	if flags[1] != true {
		t.Fatal("expected source_time_regression on second accepted bar")
	}
}
