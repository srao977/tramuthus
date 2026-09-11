// Package processing converts an ordered raw prefix into lean derived phase evidence.
// It never sorts, fills, repeats, interpolates, or renumbers source observations.
package processing

import (
	"fmt"
	"time"

	"bar_sequence_lab/phase_angle_series_generator/internal/phase"
)

// RawBar contains only fields required to calculate and join phase evidence.
type RawBar struct {
	CollectionRunID     string  `bson:"collection_run_id"`
	PartitionID         string  `bson:"partition_id"`
	Symbol              string  `bson:"symbol"`
	GeneratorSequenceNo int64   `bson:"generator_sequence_no"`
	High                float64 `bson:"high"`
	Low                 float64 `bson:"low"`
}

// DerivedRecord is the lean Mongo phase document. Nil PhaseAngleDegrees is
// intentionally encoded as BSON null while the solver is initializing.
type DerivedRecord struct {
	CollectionRunID     string    `bson:"collection_run_id"`
	PartitionID         string    `bson:"partition_id"`
	Symbol              string    `bson:"symbol"`
	GeneratorSequenceNo int64     `bson:"generator_sequence_no"`
	SeriesSize          int       `bson:"series_size"`
	SolverName          string    `bson:"solver_name"`
	SolverVersion       string    `bson:"solver_version"`
	InputSeriesType     string    `bson:"input_series_type"`
	PhaseAngleDegrees   *float64  `bson:"phase_angle_degrees"`
	PhaseObservable     bool      `bson:"phase_observable"`
	ValidityState       string    `bson:"validity_state"`
	AnalysisCreatedAt   time.Time `bson:"analysis_created_at"`
}

// Identity returns the complete deterministic analytical identity filter.
func (record DerivedRecord) Identity() map[string]any {
	return map[string]any{
		"collection_run_id":     record.CollectionRunID,
		"symbol":                record.Symbol,
		"generator_sequence_no": record.GeneratorSequenceNo,
		"series_size":           record.SeriesSize,
		"solver_name":           record.SolverName,
		"solver_version":        record.SolverVersion,
	}
}

// SelectPrefix returns at most seriesSize existing ordered bars. It validates
// strict ascending order but deliberately permits sequence-number gaps.
func SelectPrefix(bars []RawBar, seriesSize int) ([]RawBar, error) {
	if seriesSize <= 0 {
		return nil, fmt.Errorf("series size must be greater than zero")
	}
	for i := 1; i < len(bars); i++ {
		if bars[i].GeneratorSequenceNo <= bars[i-1].GeneratorSequenceNo {
			return nil, fmt.Errorf("raw bars are not in strict sequence order at index %d", i)
		}
	}
	count := min(seriesSize, len(bars))
	return append([]RawBar(nil), bars[:count]...), nil
}

// Analyze supplies exactly the selected prefix to a fresh solver invocation.
func Analyze(bars []RawBar, seriesSize int, createdAt time.Time) ([]DerivedRecord, error) {
	selected, err := SelectPrefix(bars, seriesSize)
	if err != nil {
		return nil, err
	}
	prices := make([]float64, len(selected))
	for i, bar := range selected {
		prices[i], err = phase.MedianPrice(bar.High, bar.Low)
		if err != nil {
			return nil, fmt.Errorf("sequence %d: %w", bar.GeneratorSequenceNo, err)
		}
	}
	results, err := phase.DominantCyclePhase(prices)
	if err != nil {
		return nil, err
	}
	records := make([]DerivedRecord, len(selected))
	for i, bar := range selected {
		records[i] = DerivedRecord{
			CollectionRunID: bar.CollectionRunID, PartitionID: bar.PartitionID,
			Symbol: bar.Symbol, GeneratorSequenceNo: bar.GeneratorSequenceNo,
			SeriesSize: seriesSize, SolverName: phase.SolverName,
			SolverVersion: phase.SolverVersion, InputSeriesType: phase.InputSeriesType,
			PhaseAngleDegrees: results[i].PhaseAngle, PhaseObservable: results[i].Observable,
			ValidityState: results[i].State, AnalysisCreatedAt: createdAt,
		}
	}
	return records, nil
}
