package phasecompare

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
)

type Report struct {
	Tolerance          float64 `json:"tolerance"`
	ReferenceRows      int     `json:"reference_rows"`
	MatchedRows        int     `json:"matched_rows"`
	MissingRuntimeRows int     `json:"missing_runtime_rows"`
	StatusMismatches   int     `json:"status_mismatches"`
	ObservableCompared int     `json:"observable_compared"`
	NumericMismatches  int     `json:"numeric_mismatches"`
	MaxDifference      float64 `json:"max_difference"`
}

type runtimeRow struct {
	EntityID       string   `json:"entity_id"`
	EntitySequence string   `json:"entity_sequence"`
	Status         string   `json:"status"`
	PhaseDegrees   *float64 `json:"phase_degrees"`
	Provenance     struct {
		CollectionRunID string `json:"collection_run_id"`
	} `json:"source_provenance"`
}

func Compare(referencePath, runtimePath string, tolerance float64) (Report, error) {
	runtimeRows, err := readRuntime(runtimePath)
	if err != nil {
		return Report{}, err
	}
	contents, err := os.ReadFile(referencePath)
	if err != nil {
		return Report{}, fmt.Errorf("open reference CSV: %w", err)
	}
	contents = bytes.TrimPrefix(contents, []byte{0xef, 0xbb, 0xbf})
	reader := csv.NewReader(bytes.NewReader(contents))
	rows, err := reader.ReadAll()
	if err != nil {
		return Report{}, fmt.Errorf("read reference CSV: %w", err)
	}
	if len(rows) == 0 {
		return Report{}, fmt.Errorf("reference CSV is empty")
	}
	columns := make(map[string]int, len(rows[0]))
	for index, name := range rows[0] {
		columns[name] = index
	}
	report := Report{Tolerance: tolerance, ReferenceRows: len(rows) - 1}
	for _, row := range rows[1:] {
		sequence := row[columns["generator_sequence_no"]]
		key := identity(row[columns["collection_run_id"]], row[columns["symbol"]], sequence)
		actual, ok := runtimeRows[key]
		if !ok {
			report.MissingRuntimeRows++
			continue
		}
		report.MatchedRows++
		referenceObservable, err := strconv.ParseBool(row[columns["phase_observable"]])
		if err != nil {
			return Report{}, fmt.Errorf("parse phase_observable for %s: %w", key, err)
		}
		actualObservable := actual.Status == "PHASE_STATUS_OBSERVABLE"
		if referenceObservable != actualObservable {
			report.StatusMismatches++
			continue
		}
		if !referenceObservable {
			continue
		}
		referencePhase, err := strconv.ParseFloat(row[columns["phase_angle_degrees"]], 64)
		if err != nil {
			return Report{}, fmt.Errorf("parse phase for %s: %w", key, err)
		}
		if actual.PhaseDegrees == nil {
			report.StatusMismatches++
			continue
		}
		report.ObservableCompared++
		difference := math.Abs(referencePhase - *actual.PhaseDegrees)
		if difference > report.MaxDifference {
			report.MaxDifference = difference
		}
		if difference > tolerance {
			report.NumericMismatches++
		}
	}
	return report, nil
}

func WriteReport(path string, report Report) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create comparison report directory: %w", err)
	}
	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal comparison report: %w", err)
	}
	encoded = append(encoded, '\n')
	if err := os.WriteFile(path, encoded, 0o644); err != nil {
		return fmt.Errorf("write comparison report: %w", err)
	}
	return nil
}

func readRuntime(path string) (map[string]runtimeRow, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open runtime evidence: %w", err)
	}
	defer file.Close()
	rows := make(map[string]runtimeRow)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var row runtimeRow
		if err := json.Unmarshal(scanner.Bytes(), &row); err != nil {
			return nil, fmt.Errorf("decode runtime evidence: %w", err)
		}
		rows[identity(row.Provenance.CollectionRunID, row.EntityID, row.EntitySequence)] = row
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan runtime evidence: %w", err)
	}
	return rows, nil
}

func identity(runID, entityID, sequence string) string {
	return runID + "\x00" + entityID + "\x00" + sequence
}
