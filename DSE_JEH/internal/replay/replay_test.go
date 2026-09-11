// Package replay tests reproducibility through the same strategy path used by proving execution.
// Inputs are admitted phase records; outputs are complete strategy/decision results.
// Test data spans independent entities, initialization, wrap, and actual zone transitions.
// Equality here validates deterministic machinery, not profitability or phase-velocity science.
package replay

import (
	"reflect"
	"testing"

	"tramuthus/dse-jeh/internal/evidence"
)

func phase(symbol string, sequence int64, angle *float64, observable bool) evidence.PhaseEvidence {
	validity, text := "INITIALIZING", ""
	if observable {
		validity, text = "OBSERVABLE", "0"
	}
	return evidence.PhaseEvidence{
		SourceRecordNumber: int(sequence) + 1, CollectionRunID: "run", PartitionID: "A",
		Symbol: symbol, GeneratorSequence: sequence, SeriesSize: 120,
		SolverName: "EHLERS_DOMINANT_CYCLE_PHASE", SolverVersion: "V0.1",
		InputSeriesType: "MEDIAN_PRICE", PhaseAngleDegrees: angle, PhaseAngleText: text,
		PhaseObservable: observable, ValidityState: validity,
	}
}

func value(number float64) *float64 { return &number }

func TestRunIsDeterministicThroughSamePath(t *testing.T) {
	inputs := []evidence.PhaseEvidence{
		phase("AAPL", 1, nil, false),
		phase("AAPL", 64, value(359), true),
		phase("AAPL", 65, value(1), true),
		phase("MSFT", 64, value(100), true),
	}
	first, err := Run(inputs)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Run(inputs)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("replay diverged:\nfirst=%+v\nsecond=%+v", first, second)
	}
	if len(first.FinalStates) != 2 {
		t.Fatalf("cross-entity processing lost state: %+v", first.FinalStates)
	}
}
