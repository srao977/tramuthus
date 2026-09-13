package services

import (
	"context"
	"testing"
	"time"

	dsejehv1 "tramuthus/dse-jeh-transsat-1/gen/dse_jeh/v1"
	"tramuthus/dse-jeh-transsat-1/internal/evidence"
	runtimeapp "tramuthus/dse-jeh-transsat-1/internal/runtime"
)

func TestProductionEligibilityRuleOneBoundary(t *testing.T) {
	runtimeID := "runtime-test"
	state := runtimeapp.NewState(runtimeID, dsejehv1.RuntimeMode_RUNTIME_MODE_OFFLINE, &dsejehv1.SourceSubscriptionEvidence{}, time.Now())
	service, err := NewProductionEligibility(state, evidence.NewBus(), runtimeID)
	if err != nil {
		t.Fatal(err)
	}

	initializing, err := service.EvaluateProductionEligibility(context.Background(), eligibilityRequest(63, dsejehv1.PhaseStatus_PHASE_STATUS_INITIALIZING, false))
	if err != nil {
		t.Fatal(err)
	}
	if got := initializing.GetProductionEligibilityEvidence().GetOutcome(); got != dsejehv1.ProductionEligibilityOutcome_PRODUCTION_ELIGIBILITY_OUTCOME_INITIALIZING {
		t.Fatalf("count 63 outcome = %s", got)
	}
	if initializing.GetRuleEvaluationEvidence().GetRawBooleanResult() {
		t.Fatal("count 63 raw rule result must be false")
	}

	eligible, err := service.EvaluateProductionEligibility(context.Background(), eligibilityRequest(64, dsejehv1.PhaseStatus_PHASE_STATUS_OBSERVABLE, true))
	if err != nil {
		t.Fatal(err)
	}
	want := dsejehv1.ProductionEligibilityOutcome_PRODUCTION_ELIGIBILITY_OUTCOME_PRODUCTION_ELIGIBLE
	if got := eligible.GetProductionEligibilityEvidence().GetOutcome(); got != want {
		t.Fatalf("count 64 outcome = %s; want %s", got, want)
	}
	ruleEvidence := eligible.GetRuleEvaluationEvidence()
	if !ruleEvidence.GetRawBooleanResult() || ruleEvidence.GetProductionEligibilityOutcome().GetOutcome() != want {
		t.Fatalf("count 64 rule evidence = %+v", ruleEvidence)
	}
	if state.Snapshot().GetActivity().GetBarsPhaseEligible() != 1 {
		t.Fatalf("eligible counter = %d", state.Snapshot().GetActivity().GetBarsPhaseEligible())
	}
}

func eligibilityRequest(count uint64, phaseStatus dsejehv1.PhaseStatus, phasePresent bool) *dsejehv1.EvaluateProductionEligibilityRequest {
	phase := &dsejehv1.PhaseEvidence{
		EvidenceId: "phase", EntityId: "AAPL", EntitySequence: count, Status: phaseStatus,
	}
	if phasePresent {
		value := 0.0
		phase.PhaseDegrees = &value
	}
	return &dsejehv1.EvaluateProductionEligibilityRequest{
		RuntimeId: "runtime-test", PhaseEvidence: phase,
		AnalyticalStateEvidence: &dsejehv1.AnalyticalStateEvidence{
			EvidenceId: "analytical", ContiguousValidBarCount: count,
			SequenceIntegrity:    dsejehv1.SequenceIntegrityStatus_SEQUENCE_INTEGRITY_STATUS_VALID,
			PriorStateEvidenceId: "prior",
		},
	}
}
