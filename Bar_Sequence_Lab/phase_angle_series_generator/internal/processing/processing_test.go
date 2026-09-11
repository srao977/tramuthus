// Package processing tests exact prefix selection and analytical identities.
package processing

import (
	"testing"
	"time"
)

func bars(count int) []RawBar {
	result := make([]RawBar, count)
	for i := range result {
		result[i] = RawBar{CollectionRunID: "run", PartitionID: "A", Symbol: "TEST", GeneratorSequenceNo: int64(i + 1), High: float64(i + 2), Low: float64(i)}
	}
	return result
}

func TestSeriesSizeSixUsesOnlyPrefix(t *testing.T) {
	records, err := Analyze(bars(12), 6, time.Unix(1, 0))
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 6 || records[5].GeneratorSequenceNo != 6 {
		t.Fatalf("got %d records ending at %d; want six ending at six", len(records), records[len(records)-1].GeneratorSequenceNo)
	}
	for _, record := range records {
		if record.PhaseAngleDegrees != nil || record.ValidityState != "INITIALIZING" {
			t.Fatalf("six-bar result must remain initializing: %+v", record)
		}
	}
}

func TestShortAndGappedSequencesAreNotFabricated(t *testing.T) {
	input := bars(3)
	input[2].GeneratorSequenceNo = 7
	records, err := Analyze(input, 6, time.Unix(1, 0))
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 3 || records[2].GeneratorSequenceNo != 7 {
		t.Fatalf("records were fabricated or renumbered: %+v", records)
	}
}

func TestIdentitySeparatesSizeAndSolverVersion(t *testing.T) {
	base := DerivedRecord{CollectionRunID: "run", Symbol: "TEST", GeneratorSequenceNo: 6, SeriesSize: 6, SolverName: "solver", SolverVersion: "V0.1"}
	same := base.Identity()
	if same["series_size"] != 6 || same["solver_version"] != "V0.1" {
		t.Fatalf("identity omitted experiment dimensions: %+v", same)
	}
	otherSize := base
	otherSize.SeriesSize = 12
	otherVersion := base
	otherVersion.SolverVersion = "V0.2"
	if otherSize.Identity()["series_size"] == same["series_size"] || otherVersion.Identity()["solver_version"] == same["solver_version"] {
		t.Fatal("different sizes and solver versions must produce different identities")
	}
}
