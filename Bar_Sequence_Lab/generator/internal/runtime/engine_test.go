package runtime

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"bar_sequence_lab/generator/internal/alpaca"
	"bar_sequence_lab/generator/internal/config"
	"bar_sequence_lab/generator/internal/types"
)

type fakeSource struct {
	bars []types.Observation
}

func (f fakeSource) Run(ctx context.Context, symbols []string, out chan<- types.Observation) error {
	for _, b := range f.bars {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case out <- b:
		}
	}
	return nil
}

func (f fakeSource) Health() alpaca.Health {
	return alpaca.Health{State: "healthy"}
}

func testCfg(t *testing.T, groups []string, mapping map[string]string, subscribe []string) config.Config {
	t.Helper()
	return config.Config{
		Feed:              "test",
		StreamURL:         config.TestStreamURL,
		APIKey:            "k",
		APISecret:         "s",
		DataDir:           t.TempDir(),
		SelectedGroups:    groups,
		SymbolToPartition: mapping,
		SubscribeSymbols:  subscribe,
		BufferCapacity:    32,
		FlushCount:        2,
		FlushInterval:     20 * time.Millisecond,
		MetricsInterval:   time.Hour,
	}
}

func obs(symbol string, seqHint time.Time) types.Observation {
	return types.Observation{
		Symbol:              symbol,
		SourceEventTime:     seqHint,
		SourceTimestampText: seqHint.Format(time.RFC3339Nano),
		ReceivedTime:        time.Now().UTC(),
		Interval:            "1Min",
		Open:                1, High: 1, Low: 1, Close: 1,
		Volume:            1,
		SourceID:          "ALPACA_TEST",
		AlpacaMessageType: "b",
		PayloadHash:       types.HashPayload(symbol, seqHint.String()),
	}
}

func TestEnginePersistsSelectedPartitionsOnly(t *testing.T) {
	t0 := time.Date(2026, 9, 10, 14, 0, 0, 0, time.UTC)
	cfg := testCfg(t, []string{"A", "C"}, map[string]string{"AAPL": "A", "NVDA": "C"}, []string{"AAPL", "NVDA"})
	eng, err := NewEngine(cfg)
	if err != nil {
		t.Fatal(err)
	}
	src := fakeSource{bars: []types.Observation{
		obs("AAPL", t0),
		obs("NVDA", t0),
		obs("AAPL", t0.Add(time.Minute)),
		obs("MSFT", t0),
	}}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := eng.Run(ctx, src); err != nil {
		t.Fatal(err)
	}
	aPath := filepath.Join(cfg.DataDir, eng.runID, "partition_A.jsonl")
	cPath := filepath.Join(cfg.DataDir, eng.runID, "partition_C.jsonl")
	bPath := filepath.Join(cfg.DataDir, eng.runID, "partition_B.jsonl")
	if _, err := os.Stat(bPath); !os.IsNotExist(err) {
		t.Fatal("inactive B jsonl must not exist")
	}
	if n := countLines(t, aPath); n != 2 {
		t.Fatalf("A lines=%d", n)
	}
	if n := countLines(t, cPath); n != 1 {
		t.Fatalf("C lines=%d", n)
	}
}

func TestIndependentPersistenceProgress(t *testing.T) {
	t0 := time.Now().UTC()
	cfg := testCfg(t, []string{"A", "B"}, map[string]string{"AAPL": "A", "MSFT": "B"}, []string{"AAPL", "MSFT"})
	eng, err := NewEngine(cfg)
	if err != nil {
		t.Fatal(err)
	}
	var bars []types.Observation
	for i := 0; i < 5; i++ {
		bars = append(bars, obs("AAPL", t0.Add(time.Duration(i)*time.Minute)))
		bars = append(bars, obs("MSFT", t0.Add(time.Duration(i)*time.Minute)))
	}
	if err := eng.Run(context.Background(), fakeSource{bars: bars}); err != nil {
		t.Fatal(err)
	}
	a := countLines(t, filepath.Join(cfg.DataDir, eng.runID, "partition_A.jsonl"))
	b := countLines(t, filepath.Join(cfg.DataDir, eng.runID, "partition_B.jsonl"))
	if a != 5 || b != 5 {
		t.Fatalf("independent persist A=%d B=%d", a, b)
	}
}

func TestRestartNewRunID(t *testing.T) {
	cfg := testCfg(t, []string{"A"}, map[string]string{"AAPL": "A"}, []string{"AAPL"})
	e1, err := NewEngine(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer e1.Close()
	e2, err := NewEngine(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer e2.Close()
	if e1.runID == e2.runID {
		t.Fatalf("expected new run id, both %s", e1.runID)
	}
}

func TestConcurrentPartitionWriters(t *testing.T) {
	t0 := time.Now().UTC()
	cfg := testCfg(t, []string{"A", "B", "C"}, map[string]string{"AAPL": "A", "MSFT": "B", "NVDA": "C"}, []string{"AAPL", "MSFT", "NVDA"})
	eng, err := NewEngine(cfg)
	if err != nil {
		t.Fatal(err)
	}
	var bars []types.Observation
	for i := 0; i < 8; i++ {
		bars = append(bars,
			obs("AAPL", t0.Add(time.Duration(i)*time.Second)),
			obs("MSFT", t0.Add(time.Duration(i)*time.Second)),
			obs("NVDA", t0.Add(time.Duration(i)*time.Second)),
		)
	}
	if err := eng.Run(context.Background(), fakeSource{bars: bars}); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	wg.Add(3)
	counts := make(chan int, 3)
	for _, g := range []string{"A", "B", "C"} {
		g := g
		go func() {
			defer wg.Done()
			counts <- countLines(t, filepath.Join(cfg.DataDir, eng.runID, "partition_"+g+".jsonl"))
		}()
	}
	wg.Wait()
	close(counts)
	for n := range counts {
		if n != 8 {
			t.Fatalf("got %d", n)
		}
	}
}

func countLines(t *testing.T, path string) int {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	n := 0
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		n++
		var obs types.Observation
		if err := json.Unmarshal(sc.Bytes(), &obs); err != nil {
			t.Fatal(err)
		}
		if obs.GeneratorSequenceNo == 0 || obs.CollectionRunID == "" {
			t.Fatalf("bad record %+v", obs)
		}
	}
	return n
}
