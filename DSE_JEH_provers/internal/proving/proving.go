// Package proving composes CSV admission, strategy processing, and deterministic replay.
// Inputs are an explicit CSV path and experiment identity; outputs are compared run evidence.
// Configuration selects only persisted phase evidence and does not configure phase science.
// This package never calculates Ehlers phase, velocity, profitability, orders, or P&L.
package proving

import (
	"fmt"
	"sort"

	"tramuthus/dse-jeh/internal/evidence"
	"tramuthus/dse-jeh/internal/inputcsv"
	"tramuthus/dse-jeh/internal/replay"
	"tramuthus/dse-jeh/internal/strategy"
)

// Result contains one authoritative proving result and its independently repeated replay digest.
type Result struct {
	Evidence     evidence.RunResult
	Digest       string
	ReplayDigest string
	ReplayMatch  bool
}

// Run loads one experiment and executes the same strategy path twice for comparison.
func Run(csvPath string, experiment inputcsv.Experiment) (Result, error) {
	inputs, err := inputcsv.Load(csvPath, experiment)
	if err != nil {
		return Result{}, err
	}
	first, err := replay.Run(inputs)
	if err != nil {
		return Result{}, err
	}
	second, err := replay.Run(inputs)
	if err != nil {
		return Result{}, fmt.Errorf("repeat replay: %w", err)
	}
	firstDigest, err := replay.Digest(first)
	if err != nil {
		return Result{}, err
	}
	secondDigest, err := replay.Digest(second)
	if err != nil {
		return Result{}, err
	}
	if firstDigest != secondDigest {
		return Result{}, fmt.Errorf("deterministic replay mismatch: %s != %s", firstDigest, secondDigest)
	}
	return Result{Evidence: first, Digest: firstDigest, ReplayDigest: secondDigest, ReplayMatch: true}, nil
}

// Summary describes proving evidence without assigning scientific ranking or profitability.
type Summary struct {
	InputCount              int
	SymbolCount             int
	ObservableCount         int
	InitializingCount       int
	ZoneCounts              map[evidence.Zone]int
	ZoneTransitionCount     int
	HopOnCandidateEvents    int
	HopOffCandidateEvents   int
	DecisionEvents          int
	HopOnCandidateSet       []string
	InitializingOnlySymbols []string
}

// Summarize counts admitted evidence and emitted candidate interpretation.
func Summarize(result evidence.RunResult) Summary {
	symbols := make(map[string]struct{})
	observableSymbols := make(map[string]struct{})
	summary := Summary{InputCount: len(result.Inputs), ZoneCounts: make(map[evidence.Zone]int), DecisionEvents: len(result.DecisionEvents), HopOnCandidateSet: append([]string(nil), result.HopOnCandidates...)}
	for _, input := range result.Inputs {
		symbols[input.Symbol] = struct{}{}
		if input.PhaseObservable {
			summary.ObservableCount++
			observableSymbols[input.Symbol] = struct{}{}
			if input.PhaseAngleDegrees != nil {
				if zone, err := strategy.Classify(*input.PhaseAngleDegrees); err == nil {
					summary.ZoneCounts[zone]++
				}
			}
		} else {
			summary.InitializingCount++
		}
	}
	for _, event := range result.StrategyEvents {
		switch event.Type {
		case evidence.EventZoneEntered:
			if event.PreviousZone != nil {
				summary.ZoneTransitionCount++
			}
		case evidence.EventHopOnCandidate:
			summary.HopOnCandidateEvents++
		case evidence.EventHopOffCandidate:
			summary.HopOffCandidateEvents++
		}
	}
	summary.SymbolCount = len(symbols)
	for symbol := range symbols {
		if _, observable := observableSymbols[symbol]; !observable {
			summary.InitializingOnlySymbols = append(summary.InitializingOnlySymbols, symbol)
		}
	}
	sort.Strings(summary.InitializingOnlySymbols)
	return summary
}
