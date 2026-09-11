// Package persistence tests deterministic local evidence serialization for CSV proving.
// Inputs are fixed metadata and evidence slices; outputs are local JSON and JSONL files.
// Test paths are temporary and no database, network, or scientific parameters are involved.
// Byte equality validates reproducible file output, not strategy profitability or phase science.
package persistence

import (
	"os"
	"path/filepath"
	"testing"

	"tramuthus/dse-jeh/internal/evidence"
)

func TestLocalEvidenceOutputIsByteStable(t *testing.T) {
	firstDirectory := filepath.Join(t.TempDir(), "first")
	secondDirectory := filepath.Join(t.TempDir(), "second")
	inputs := []evidence.PhaseEvidence{{
		SourceRecordNumber: 65, CollectionRunID: "run", PartitionID: "A", Symbol: "GOOGL",
		GeneratorSequence: 64, SeriesSize: 120, SolverName: "EHLERS_DOMINANT_CYCLE_PHASE",
		SolverVersion: "V0.1", InputSeriesType: "MEDIAN_PRICE", PhaseAngleText: "327.67589648095657",
		PhaseObservable: true, ValidityState: "OBSERVABLE",
	}}
	metadata := RunMetadata{RunID: "csv-prove-digest", Mode: "CSV_PROVE_AND_REPLAY", OutputDigest: "digest", ReplayDigest: "digest", ReplayMatched: true}

	for _, directory := range []string{firstDirectory, secondDirectory} {
		if err := WriteJSONL(filepath.Join(directory, "inputs.jsonl"), inputs); err != nil {
			t.Fatal(err)
		}
		if err := WriteJSON(filepath.Join(directory, "metadata.json"), metadata); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"inputs.jsonl", "metadata.json"} {
		first, err := os.ReadFile(filepath.Join(firstDirectory, name))
		if err != nil {
			t.Fatal(err)
		}
		second, err := os.ReadFile(filepath.Join(secondDirectory, name))
		if err != nil {
			t.Fatal(err)
		}
		if string(first) != string(second) {
			t.Fatalf("%s output differs between identical runs", name)
		}
	}
}
