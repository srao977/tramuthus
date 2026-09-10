package sequence

import (
	"testing"
	"time"

	"bar_sequence_lab/generator/internal/types"
)

func bar(symbol string, t time.Time, hash string) types.Observation {
	return types.Observation{
		Symbol:              symbol,
		SourceEventTime:     t,
		SourceTimestampText: t.Format(time.RFC3339),
		ReceivedTime:        time.Now().UTC(),
		PayloadHash:         hash,
		AlpacaMessageType:   "b",
	}
}

func TestAcceptedArrivalNotSourceSorted(t *testing.T) {
	a := New("run1")
	t1 := time.Date(2026, 9, 10, 12, 1, 0, 0, time.UTC)
	t0 := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	first := a.Accept(bar("AAPL", t1, "h1"))
	second := a.Accept(bar("AAPL", t0, "h2"))
	if first.GeneratorSequenceNo != 1 || second.GeneratorSequenceNo != 2 {
		t.Fatalf("seq %d %d", first.GeneratorSequenceNo, second.GeneratorSequenceNo)
	}
	if second.SourceTimeRegression != true {
		t.Fatal("expected source_time_regression flag")
	}
	if second.SourceEventTime.Equal(first.SourceEventTime) {
		t.Fatal("source times should remain distinct")
	}
}

func TestIndependentSymbols(t *testing.T) {
	a := New("run1")
	now := time.Now().UTC()
	aapl := a.Accept(bar("AAPL", now, "a"))
	msft := a.Accept(bar("MSFT", now, "m"))
	if aapl.GeneratorSequenceNo != 1 || msft.GeneratorSequenceNo != 1 {
		t.Fatalf("expected independent sequences, got %d %d", aapl.GeneratorSequenceNo, msft.GeneratorSequenceNo)
	}
}

func TestDuplicatePreserved(t *testing.T) {
	a := New("run1")
	now := time.Now().UTC()
	first := a.Accept(bar("AAPL", now, "same"))
	second := a.Accept(bar("AAPL", now, "same"))
	if first.GeneratorSequenceNo != 1 || second.GeneratorSequenceNo != 2 {
		t.Fatalf("duplicate must still receive next sequence, got %d %d", first.GeneratorSequenceNo, second.GeneratorSequenceNo)
	}
	if !second.DuplicateArrival {
		t.Fatal("duplicate evidence flag missing")
	}
}
