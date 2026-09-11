// Package persistence writes deterministic local evidence for the CSV proving stage.
// Inputs are completed proving results and metadata; outputs are JSONL evidence files.
// File output receives already-derived evidence and owns no scientific parameters.
// It cannot classify phase, advance strategy state, rank candidates, or change lineage.
package persistence

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// RunMetadata identifies one proving or replay execution and its deterministic outcome.
type RunMetadata struct {
	RunID               string `json:"run_id"`
	Mode                string `json:"mode"`
	SourceCSV           string `json:"source_csv"`
	CollectionRunID     string `json:"collection_run_id"`
	SeriesSize          int    `json:"series_size"`
	SolverName          string `json:"solver_name"`
	SolverVersion       string `json:"solver_version"`
	InputSeriesType     string `json:"input_series_type"`
	StrategyVersion     string `json:"strategy_version"`
	RankingMode         string `json:"ranking_mode"`
	SourceCSVSHA256     string `json:"source_csv_sha256"`
	InputCount          int    `json:"input_count"`
	StrategyEventCount  int    `json:"strategy_event_count"`
	DecisionEventCount  int    `json:"decision_event_count"`
	DuplicateInputCount int    `json:"duplicate_input_count"`
	OutputDigest        string `json:"output_digest"`
	ReplayDigest        string `json:"replay_digest"`
	ReplayMatched       bool   `json:"replay_matched"`
}

// WriteJSON writes one deterministic indented JSON document.
func WriteJSON(path string, value any) error {
	encoded, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode %s: %w", filepath.Base(path), err)
	}
	encoded = append(encoded, '\n')
	return writeFile(path, encoded)
}

// WriteJSONL writes one compact JSON object per line in the supplied order.
func WriteJSONL[T any](path string, values []T) error {
	file, err := createFile(path)
	if err != nil {
		return err
	}
	succeeded := false
	defer func() {
		_ = file.Close()
		if !succeeded {
			_ = os.Remove(path)
		}
	}()
	encoder := json.NewEncoder(file)
	encoder.SetEscapeHTML(false)
	for index, value := range values {
		if err := encoder.Encode(value); err != nil {
			return fmt.Errorf("encode %s record %d: %w", filepath.Base(path), index+1, err)
		}
	}
	if err := file.Sync(); err != nil {
		return fmt.Errorf("sync %s: %w", path, err)
	}
	succeeded = true
	return nil
}

func createFile(path string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create output directory: %w", err)
	}
	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("create %s: %w", path, err)
	}
	return file, nil
}

func writeFile(path string, contents []byte) error {
	file, err := createFile(path)
	if err != nil {
		return err
	}
	if _, err := file.Write(contents); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return fmt.Errorf("write %s: %w", path, err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close %s: %w", path, err)
	}
	return nil
}
