// Package strategy tests candidate interpretation independently from phase generation.
// Inputs are synthetic admitted phase records; outputs are deterministic state/events.
// Test parameters cover exact boundaries, circular wrap, validity, duplicates, and order.
// These tests do not validate profitability or define unresolved phase velocity.
package strategy

import (
	"math"
	"testing"

	"tramuthus/dse-jeh/internal/evidence"
)

func phaseInput(symbol string, sequence int64, phase *float64, observable bool) evidence.PhaseEvidence {
	validity := "INITIALIZING"
	text := ""
	if observable {
		validity = "OBSERVABLE"
		text = "0"
		if phase != nil {
			text = "value"
		}
	}
	return evidence.PhaseEvidence{
		CollectionRunID: "run", PartitionID: "A", Symbol: symbol, GeneratorSequence: sequence,
		SeriesSize: 120, SolverName: "EHLERS_DOMINANT_CYCLE_PHASE", SolverVersion: "V0.1",
		InputSeriesType: "MEDIAN_PRICE", PhaseAngleDegrees: phase, PhaseAngleText: text,
		PhaseObservable: observable, ValidityState: validity,
	}
}

func pointer(value float64) *float64 { return &value }

func TestClassifyExactHalfOpenBoundaries(t *testing.T) {
	tests := []struct {
		phase float64
		want  evidence.Zone
	}{
		{0, evidence.ZoneMomentumHold}, {math.Nextafter(90, 0), evidence.ZoneMomentumHold},
		{90, evidence.ZoneHopOff}, {math.Nextafter(180, 0), evidence.ZoneHopOff},
		{180, evidence.ZoneDisregard}, {math.Nextafter(270, 0), evidence.ZoneDisregard},
		{270, evidence.ZoneHopOn}, {math.Nextafter(360, 0), evidence.ZoneHopOn},
	}
	for _, test := range tests {
		got, err := Classify(test.phase)
		if err != nil || got != test.want {
			t.Fatalf("phase %v: got %q, %v; want %q", test.phase, got, err, test.want)
		}
	}
}

func TestInitializingAndLegitimateZero(t *testing.T) {
	engine := NewEngine()
	if err := engine.Apply(phaseInput("AAPL", 1, nil, false)); err != nil {
		t.Fatal(err)
	}
	if got := engine.Result(nil); len(got.StrategyEvents) != 0 || len(got.FinalStates) != 0 {
		t.Fatalf("initializing input entered strategy: %+v", got)
	}
	if err := engine.Apply(phaseInput("AAPL", 2, pointer(0), true)); err != nil {
		t.Fatal(err)
	}
	got := engine.Result(nil)
	state := got.FinalStates["AAPL"]
	if state.CurrentPhaseDegrees != 0 || state.HasPreviousPhase {
		t.Fatalf("zero or prior state corrupted: %+v", state)
	}
	if len(got.StrategyEvents) != 2 || got.StrategyEvents[0].Type != evidence.EventPhaseBecameObservable {
		t.Fatalf("unexpected first observable events: %+v", got.StrategyEvents)
	}
}

func TestSameZoneDoesNotTransitionAndWrapDoes(t *testing.T) {
	engine := NewEngine()
	for _, input := range []evidence.PhaseEvidence{
		phaseInput("AAPL", 1, pointer(300), true),
		phaseInput("AAPL", 2, pointer(359), true),
		phaseInput("AAPL", 3, pointer(1), true),
	} {
		if err := engine.Apply(input); err != nil {
			t.Fatal(err)
		}
	}
	result := engine.Result(nil)
	var exits, entries int
	for _, event := range result.StrategyEvents {
		if event.Type == evidence.EventZoneExited {
			exits++
		}
		if event.Type == evidence.EventZoneEntered {
			entries++
		}
	}
	if exits != 1 || entries != 2 {
		t.Fatalf("same-zone/wrap transitions: exits=%d entries=%d", exits, entries)
	}
	state := result.FinalStates["AAPL"]
	if !state.HasPreviousPhase || state.PreviousPhaseDegrees != 359 || state.CurrentPhaseDegrees != 1 {
		t.Fatalf("wrap state corrupted: %+v", state)
	}
}

func TestEntityOrderDuplicateAndNoCrossSymbolSynchronization(t *testing.T) {
	engine := NewEngine()
	a := phaseInput("AAPL", 64, pointer(275), true)
	b := phaseInput("MSFT", 64, pointer(100), true)
	if err := engine.Apply(a); err != nil {
		t.Fatal(err)
	}
	if err := engine.Apply(b); err != nil {
		t.Fatal(err)
	}
	if err := engine.Apply(a); err != nil {
		t.Fatal(err)
	}
	result := engine.Result(nil)
	if len(result.FinalStates) != 2 || result.DuplicateInputCount != 1 {
		t.Fatalf("independent/duplicate state: %+v", result)
	}
	if err := engine.Apply(phaseInput("AAPL", 63, pointer(280), true)); err == nil {
		t.Fatal("expected per-symbol ordering error")
	}
}
