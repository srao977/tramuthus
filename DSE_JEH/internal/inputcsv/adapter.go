// Package inputcsv admits authoritative phase evidence from a bounded CSV fixture.
// Inputs are a file plus explicit experiment identity; output is ordered per-entity evidence.
// The adapter validates columns, precision text, validity, and source lineage without
// calculating, smoothing, interpolating, synchronizing, or otherwise altering phase.
package inputcsv

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"tramuthus/dse-jeh/internal/evidence"
)

type Experiment struct {
	CollectionRunID string
	SeriesSize      int
	SolverName      string
	SolverVersion   string
	InputSeriesType string
}

var requiredColumns = []string{
	"collection_run_id", "partition_id", "symbol", "generator_sequence_no",
	"series_size", "solver_name", "solver_version", "input_series_type",
	"phase_angle_degrees", "phase_observable", "validity_state", "analysis_created_at",
}

// Load reads and validates exactly one requested experiment, then sorts by symbol
// and entity-local generator_sequence_no. It rejects duplicate source identities.
func Load(path string, experiment Experiment) ([]evidence.PhaseEvidence, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open phase CSV: %w", err)
	}
	defer file.Close()

	buffered := bufio.NewReader(file)
	if prefix, peekErr := buffered.Peek(3); peekErr == nil && string(prefix) == "\xef\xbb\xbf" {
		_, _ = buffered.Discard(3)
	}
	reader := csv.NewReader(buffered)
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("read CSV header: %w", err)
	}
	columns := make(map[string]int, len(header))
	for index, name := range header {
		columns[strings.TrimSpace(name)] = index
	}
	for _, name := range requiredColumns {
		if _, exists := columns[name]; !exists {
			return nil, fmt.Errorf("required CSV column %q is missing", name)
		}
	}

	var admitted []evidence.PhaseEvidence
	for recordNumber := 2; ; recordNumber++ {
		record, readErr := reader.Read()
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return nil, fmt.Errorf("read CSV record %d: %w", recordNumber, readErr)
		}
		value := func(name string) string { return strings.TrimSpace(record[columns[name]]) }
		seriesSize, parseErr := strconv.Atoi(value("series_size"))
		if parseErr != nil {
			return nil, fmt.Errorf("record %d series_size: %w", recordNumber, parseErr)
		}
		if value("collection_run_id") != experiment.CollectionRunID || seriesSize != experiment.SeriesSize ||
			value("solver_name") != experiment.SolverName || value("solver_version") != experiment.SolverVersion ||
			value("input_series_type") != experiment.InputSeriesType {
			continue
		}
		row, parseErr := parseRecord(recordNumber, value, seriesSize)
		if parseErr != nil {
			return nil, parseErr
		}
		admitted = append(admitted, row)
	}
	if len(admitted) == 0 {
		return nil, fmt.Errorf("requested experiment has no CSV records")
	}
	sort.SliceStable(admitted, func(left, right int) bool {
		if admitted[left].Symbol != admitted[right].Symbol {
			return admitted[left].Symbol < admitted[right].Symbol
		}
		return admitted[left].GeneratorSequence < admitted[right].GeneratorSequence
	})
	seen := make(map[string]struct{}, len(admitted))
	for _, row := range admitted {
		key := row.CollectionRunID + "|" + row.Symbol + "|" + strconv.FormatInt(row.GeneratorSequence, 10) + "|" + strconv.Itoa(row.SeriesSize) + "|" + row.SolverName + "|" + row.SolverVersion
		if _, duplicate := seen[key]; duplicate {
			return nil, fmt.Errorf("duplicate phase input identity at %s sequence %d", row.Symbol, row.GeneratorSequence)
		}
		seen[key] = struct{}{}
	}
	return admitted, nil
}

// parseRecord validates one selected CSV row without transforming its phase value or validity.
func parseRecord(recordNumber int, value func(string) string, seriesSize int) (evidence.PhaseEvidence, error) {
	sequence, err := strconv.ParseInt(value("generator_sequence_no"), 10, 64)
	if err != nil || sequence <= 0 {
		return evidence.PhaseEvidence{}, fmt.Errorf("record %d has invalid generator_sequence_no", recordNumber)
	}
	observable, err := strconv.ParseBool(value("phase_observable"))
	if err != nil {
		return evidence.PhaseEvidence{}, fmt.Errorf("record %d phase_observable: %w", recordNumber, err)
	}
	createdAt, err := time.Parse(time.RFC3339Nano, value("analysis_created_at"))
	if err != nil {
		return evidence.PhaseEvidence{}, fmt.Errorf("record %d analysis_created_at: %w", recordNumber, err)
	}
	phaseText := value("phase_angle_degrees")
	var phase *float64
	if observable {
		parsed, parseErr := strconv.ParseFloat(phaseText, 64)
		if parseErr != nil || parsed < 0 || parsed >= 360 {
			return evidence.PhaseEvidence{}, fmt.Errorf("record %d has invalid observable phase %q", recordNumber, phaseText)
		}
		if value("validity_state") != "OBSERVABLE" {
			return evidence.PhaseEvidence{}, fmt.Errorf("record %d observable phase has validity %q", recordNumber, value("validity_state"))
		}
		phase = &parsed
	} else if phaseText != "" || value("validity_state") != "INITIALIZING" {
		return evidence.PhaseEvidence{}, fmt.Errorf("record %d initializing evidence must have blank phase and INITIALIZING validity", recordNumber)
	}
	if value("partition_id") == "" || value("symbol") == "" || value("solver_name") == "" || value("solver_version") == "" {
		return evidence.PhaseEvidence{}, fmt.Errorf("record %d has incomplete lineage", recordNumber)
	}
	return evidence.PhaseEvidence{
		SourceRecordNumber: recordNumber, CollectionRunID: value("collection_run_id"),
		PartitionID: value("partition_id"), Symbol: value("symbol"), GeneratorSequence: sequence,
		SeriesSize: seriesSize, SolverName: value("solver_name"), SolverVersion: value("solver_version"),
		InputSeriesType: value("input_series_type"), PhaseAngleDegrees: phase, PhaseAngleText: phaseText,
		PhaseObservable: observable, ValidityState: value("validity_state"), AnalysisCreatedAt: createdAt,
	}, nil
}
