package phasecompare

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCompareAcceptsBOMAndMatchesObservablePhase(t *testing.T) {
	directory := t.TempDir()
	reference := filepath.Join(directory, "reference.csv")
	runtime := filepath.Join(directory, "runtime.jsonl")
	csv := "\xef\xbb\xbf\"collection_run_id\",\"symbol\",\"generator_sequence_no\",\"phase_angle_degrees\",\"phase_observable\"\n\"run\",\"AAPL\",\"64\",\"12.5\",\"True\"\n"
	jsonl := "{\"entity_id\":\"AAPL\",\"entity_sequence\":\"64\",\"status\":\"PHASE_STATUS_OBSERVABLE\",\"phase_degrees\":12.5,\"source_provenance\":{\"collection_run_id\":\"run\"}}\n"
	if err := os.WriteFile(reference, []byte(csv), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(runtime, []byte(jsonl), 0o644); err != nil {
		t.Fatal(err)
	}
	report, err := Compare(reference, runtime, 1e-9)
	if err != nil {
		t.Fatal(err)
	}
	if report.MatchedRows != 1 || report.ObservableCompared != 1 || report.NumericMismatches != 0 {
		t.Fatalf("report = %+v", report)
	}
}
