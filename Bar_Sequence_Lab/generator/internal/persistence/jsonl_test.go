package persistence

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"bar_sequence_lab/generator/internal/types"
)

func TestWriteBatchOneObjectPerLine(t *testing.T) {
	dir := t.TempDir()
	w, err := Open(dir, "run1", "A")
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	now := time.Now().UTC()
	obs := []types.Observation{
		{CollectionRunID: "run1", PartitionID: "A", Symbol: "AAPL", GeneratorSequenceNo: 1, SourceEventTime: now, ReceivedTime: now, Interval: "1Min", Open: 1, High: 1, Low: 1, Close: 1, Volume: 1, SourceID: "ALPACA_TEST", AlpacaMessageType: "b"},
		{CollectionRunID: "run1", PartitionID: "A", Symbol: "AAPL", GeneratorSequenceNo: 2, SourceEventTime: now.Add(time.Minute), ReceivedTime: now, Interval: "1Min", Open: 1, High: 1, Low: 1, Close: 1, Volume: 1, SourceID: "ALPACA_TEST", AlpacaMessageType: "u"},
	}
	if err := w.WriteBatch(obs); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "run1", "partition_A.jsonl")
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	n := 0
	for sc.Scan() {
		n++
		var got types.Observation
		if err := json.Unmarshal(sc.Bytes(), &got); err != nil {
			t.Fatal(err)
		}
		if got.PersistedTime.IsZero() || got.GeneratorSequenceNo == 0 {
			t.Fatalf("%+v", got)
		}
	}
	if n != 2 {
		t.Fatalf("lines=%d", n)
	}
}

func TestInactivePartitionFileNotCreated(t *testing.T) {
	dir := t.TempDir()
	w, err := Open(dir, "run1", "A")
	if err != nil {
		t.Fatal(err)
	}
	w.Close()
	if _, err := os.Stat(filepath.Join(dir, "run1", "partition_B.jsonl")); !os.IsNotExist(err) {
		t.Fatalf("B file should not exist: %v", err)
	}
}
