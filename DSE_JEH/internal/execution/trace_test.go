package execution

import (
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestRunTypeAndRiskAreIndependent(t *testing.T) {
	for _, test := range []struct {
		name    string
		runType RunType
		risk    float64
		wantErr bool
	}{
		{name: "run C risk one", runType: "RUN_C", risk: 1},
		{name: "run D repeated fractional risk", runType: "RUN_D", risk: 0.20},
		{name: "run E repeated fractional risk", runType: "RUN_E", risk: 0.20},
		{name: "arbitrary governance label", runType: "RESERVOIR_VALIDATION_001", risk: 0.35},
		{name: "run type does not constrain risk", runType: RunTypeA, risk: 1},
		{name: "risk does not determine run type", runType: "PAPER_PRECHECK_002", risk: 0},
		{name: "run A historical endpoint", runType: RunTypeA, risk: 0},
		{name: "run B historical endpoint", runType: RunTypeB, risk: 1},
		{name: "risk below domain", runType: "RUN_F", risk: -0.01, wantErr: true},
		{name: "risk above domain", runType: "RUN_G", risk: 1.01, wantErr: true},
		{name: "empty run type", runType: "", risk: 0.20, wantErr: true},
		{name: "invalid lowercase run type", runType: "run_h", risk: 0.20, wantErr: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateRunType(test.runType)
			if err == nil {
				err = (RunParameters{StartingCapital: 1, AllocationPct: 1, Risk: test.risk}).Validate()
			}
			if (err != nil) != test.wantErr {
				t.Fatalf("independent validation of run_type=%q risk=%v error = %v, wantErr %v", test.runType, test.risk, err, test.wantErr)
			}
		})
	}
}

func TestTraceRecordPersistsRunTypeAtTopLevel(t *testing.T) {
	encoded, err := bson.Marshal(TraceRecord{PipelineRunID: "pipeline-a", RunType: RunTypeA})
	if err != nil {
		t.Fatal(err)
	}
	var document bson.M
	if err := bson.Unmarshal(encoded, &document); err != nil {
		t.Fatal(err)
	}
	if document["run_type"] != "RUN_A" {
		t.Fatalf("top-level run_type = %#v, want RUN_A", document["run_type"])
	}
	if _, exists := document["stage_run_type"]; exists {
		t.Fatal("run_type must not be represented as a stage field")
	}
}
