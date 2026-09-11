// Package inputcsv tests the bounded CSV proving adapter independently of strategy logic.
// Inputs are temporary CSV fixtures; outputs are admitted authoritative phase evidence.
// Test configuration selects one explicit run, size, solver, version, and input type.
// These tests prohibit phase calculation, null substitution, experiment mixing, and cross-entity synchronization.
package inputcsv

import (
	"os"
	"path/filepath"
	"testing"
)

const csvHeader = "collection_run_id,partition_id,symbol,generator_sequence_no,series_size,solver_name,solver_version,input_series_type,phase_angle_degrees,phase_observable,validity_state,analysis_created_at\n"

func targetExperiment() Experiment {
	return Experiment{
		CollectionRunID: "20260911T161623Z-1",
		SeriesSize:      120,
		SolverName:      "EHLERS_DOMINANT_CYCLE_PHASE",
		SolverVersion:   "V0.1",
		InputSeriesType: "MEDIAN_PRICE",
	}
}

func writeFixture(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "phase.csv")
	if err := os.WriteFile(path, []byte(csvHeader+body), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeBOMFixture(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "phase-bom.csv")
	if err := os.WriteFile(path, append([]byte{0xef, 0xbb, 0xbf}, []byte(csvHeader+body)...), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadFiltersExperimentPreservesPrecisionAndOrdersPerSymbol(t *testing.T) {
	path := writeFixture(t,
		"20260911T161623Z-1,A,GOOGL,65,120,EHLERS_DOMINANT_CYCLE_PHASE,V0.1,MEDIAN_PRICE,338.9899202628311,true,OBSERVABLE,2026-09-11T19:00:00Z\n"+
			"20260911T161623Z-1,A,GOOGL,64,120,EHLERS_DOMINANT_CYCLE_PHASE,V0.1,MEDIAN_PRICE,327.67589648095657,true,OBSERVABLE,2026-09-11T19:00:00Z\n"+
			"20260911T161623Z-1,B,DIA,1,120,EHLERS_DOMINANT_CYCLE_PHASE,V0.1,MEDIAN_PRICE,,false,INITIALIZING,2026-09-11T19:00:00Z\n"+
			"20260911T161623Z-1,A,GOOGL,1,6,EHLERS_DOMINANT_CYCLE_PHASE,V0.1,MEDIAN_PRICE,,false,INITIALIZING,2026-09-11T19:00:00Z\n")

	rows, err := Load(path, targetExperiment())
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 3 {
		t.Fatalf("admitted %d records, want 3", len(rows))
	}
	if rows[0].Symbol != "DIA" || rows[0].PhaseAngleDegrees != nil || rows[0].ValidityState != "INITIALIZING" {
		t.Fatalf("initializing evidence changed: %+v", rows[0])
	}
	if rows[1].Symbol != "GOOGL" || rows[1].GeneratorSequence != 64 || rows[2].GeneratorSequence != 65 {
		t.Fatalf("entity ordering changed: %+v", rows)
	}
	if rows[1].PhaseAngleText != "327.67589648095657" || rows[1].PhaseAngleDegrees == nil || *rows[1].PhaseAngleDegrees != 327.67589648095657 {
		t.Fatalf("phase precision changed: %+v", rows[1])
	}
}

func TestLoadRejectsInvalidInitializationAndDuplicates(t *testing.T) {
	t.Run("initializing phase must remain blank", func(t *testing.T) {
		path := writeFixture(t, "20260911T161623Z-1,B,DIA,1,120,EHLERS_DOMINANT_CYCLE_PHASE,V0.1,MEDIAN_PRICE,0,false,INITIALIZING,2026-09-11T19:00:00Z\n")
		if _, err := Load(path, targetExperiment()); err == nil {
			t.Fatal("expected invalid initialization error")
		}
	})
	t.Run("duplicate experiment identity", func(t *testing.T) {
		row := "20260911T161623Z-1,A,GOOGL,64,120,EHLERS_DOMINANT_CYCLE_PHASE,V0.1,MEDIAN_PRICE,327.67589648095657,true,OBSERVABLE,2026-09-11T19:00:00Z\n"
		if _, err := Load(writeFixture(t, row+row), targetExperiment()); err == nil {
			t.Fatal("expected duplicate identity error")
		}
	})
}

func TestInitializingOnlyEntitiesNeverAcquirePhase(t *testing.T) {
	path := writeFixture(t,
		"20260911T161623Z-1,B,DIA,1,120,EHLERS_DOMINANT_CYCLE_PHASE,V0.1,MEDIAN_PRICE,,false,INITIALIZING,2026-09-11T19:00:00Z\n"+
			"20260911T161623Z-1,B,VXX,1,120,EHLERS_DOMINANT_CYCLE_PHASE,V0.1,MEDIAN_PRICE,,false,INITIALIZING,2026-09-11T19:00:00Z\n")
	rows, err := Load(path, targetExperiment())
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		if row.PhaseObservable || row.PhaseAngleDegrees != nil {
			t.Fatalf("%s acquired fabricated phase: %+v", row.Symbol, row)
		}
	}
}

func TestLoadAcceptsPowerShellUTF8BOM(t *testing.T) {
	path := writeBOMFixture(t, "20260911T161623Z-1,A,GOOGL,64,120,EHLERS_DOMINANT_CYCLE_PHASE,V0.1,MEDIAN_PRICE,327.67589648095657,true,OBSERVABLE,2026-09-11T19:00:00Z\n")
	rows, err := Load(path, targetExperiment())
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].PhaseAngleText != "327.67589648095657" {
		t.Fatalf("BOM fixture changed evidence: %+v", rows)
	}
}
