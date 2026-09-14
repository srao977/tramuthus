package services

import (
	"context"
	"math"
	"testing"

	dsejehv1 "tramuthus/dse-jeh-transsat-1/gen/dse_jeh/v1"
)

func TestCalculatePhaseTransition(t *testing.T) {
	tests := []struct {
		name          string
		prior         float64
		current       float64
		wantDelta     float64
		wantDirection dsejehv1.PhaseTransitionDirection
	}{
		{name: "ordinary forward", prior: 10, current: 20, wantDelta: 10, wantDirection: dsejehv1.PhaseTransitionDirection_PHASE_TRANSITION_DIRECTION_FORWARD},
		{name: "ordinary reverse", prior: 20, current: 10, wantDelta: -10, wantDirection: dsejehv1.PhaseTransitionDirection_PHASE_TRANSITION_DIRECTION_REVERSE},
		{name: "positive wrap", prior: 350, current: 10, wantDelta: 20, wantDirection: dsejehv1.PhaseTransitionDirection_PHASE_TRANSITION_DIRECTION_FORWARD},
		{name: "negative wrap", prior: 10, current: 350, wantDelta: -20, wantDirection: dsejehv1.PhaseTransitionDirection_PHASE_TRANSITION_DIRECTION_REVERSE},
		{name: "zero displacement", prior: 45, current: 45, wantDelta: 0, wantDirection: dsejehv1.PhaseTransitionDirection_PHASE_TRANSITION_DIRECTION_STATIONARY},
		{name: "exact positive tie", prior: 0, current: 180, wantDelta: 180, wantDirection: dsejehv1.PhaseTransitionDirection_PHASE_TRANSITION_DIRECTION_FORWARD},
		{name: "reverse-form exact tie", prior: 180, current: 0, wantDelta: 180, wantDirection: dsejehv1.PhaseTransitionDirection_PHASE_TRANSITION_DIRECTION_FORWARD},
		{name: "normalize above 360", prior: 370, current: 20, wantDelta: 10, wantDirection: dsejehv1.PhaseTransitionDirection_PHASE_TRANSITION_DIRECTION_FORWARD},
		{name: "normalize negative", prior: -10, current: 10, wantDelta: 20, wantDirection: dsejehv1.PhaseTransitionDirection_PHASE_TRANSITION_DIRECTION_FORWARD},
	}

	service := NewDynamicExecution()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response, err := service.CalculatePhaseTransition(context.Background(), phaseTransitionRequest(test.prior, test.current))
			if err != nil {
				t.Fatal(err)
			}
			transition := response.GetPhaseTransitionState()
			if transition.GetStatus() != dsejehv1.PhaseTransitionStatus_PHASE_TRANSITION_STATUS_VALID {
				t.Fatalf("status = %s; want VALID", transition.GetStatus())
			}
			if got := transition.GetSignedCircularDisplacementDegrees(); got != test.wantDelta {
				t.Fatalf("delta = %v; want %v", got, test.wantDelta)
			}
			if got := transition.GetMagnitudeDegrees(); got != math.Abs(test.wantDelta) {
				t.Fatalf("magnitude = %v; want %v", got, math.Abs(test.wantDelta))
			}
			if got := transition.GetPhaseVelocityDegreesPerBar(); got != test.wantDelta {
				t.Fatalf("velocity = %v; want %v", got, test.wantDelta)
			}
			if got := transition.GetDirection(); got != test.wantDirection {
				t.Fatalf("direction = %s; want %s", got, test.wantDirection)
			}
			if transition.GetPriorPhaseEvidenceId() != "phase-prior" || transition.GetCurrentPhaseEvidenceId() != "phase-current" || transition.GetProductionEligibilityEvidenceId() != "eligibility-current" {
				t.Fatalf("causal references not preserved: %+v", transition)
			}
			if transition.GetAlgorithmIdentity().GetAlgorithmId() != "V1_SHORTEST_SIGNED_CIRCULAR_DISPLACEMENT" {
				t.Fatalf("algorithm identity = %+v", transition.GetAlgorithmIdentity())
			}
		})
	}
}

func TestCalculatePhaseTransitionRejectsNonFinitePhase(t *testing.T) {
	response, err := NewDynamicExecution().CalculatePhaseTransition(context.Background(), phaseTransitionRequest(math.NaN(), 10))
	if err != nil {
		t.Fatal(err)
	}
	transition := response.GetPhaseTransitionState()
	if transition.GetStatus() != dsejehv1.PhaseTransitionStatus_PHASE_TRANSITION_STATUS_INVALID {
		t.Fatalf("status = %s; want INVALID", transition.GetStatus())
	}
	if transition.SignedCircularDisplacementDegrees != nil || transition.MagnitudeDegrees != nil || transition.PhaseVelocityDegreesPerBar != nil {
		t.Fatalf("invalid transition must omit numeric results: %+v", transition)
	}
}

func phaseTransitionRequest(priorDegrees, currentDegrees float64) *dsejehv1.CalculatePhaseTransitionRequest {
	return &dsejehv1.CalculatePhaseTransitionRequest{
		RuntimeId: "runtime-test",
		PriorPhaseEvidence: &dsejehv1.PhaseEvidence{
			EvidenceId: "phase-prior", EntityId: "AAPL", EntitySequence: 63,
			Status: dsejehv1.PhaseStatus_PHASE_STATUS_INITIALIZING, PhaseDegrees: &priorDegrees,
		},
		CurrentPhaseEvidence: &dsejehv1.PhaseEvidence{
			EvidenceId: "phase-current", EntityId: "AAPL", EntitySequence: 64,
			Status: dsejehv1.PhaseStatus_PHASE_STATUS_OBSERVABLE, PhaseDegrees: &currentDegrees,
		},
		ProductionEligibilityEvidence: &dsejehv1.ProductionEligibilityEvidence{
			EvidenceId: "eligibility-current", PhaseEvidenceId: "phase-current",
			Outcome: dsejehv1.ProductionEligibilityOutcome_PRODUCTION_ELIGIBILITY_OUTCOME_PRODUCTION_ELIGIBLE,
		},
		AlgorithmIdentity: &dsejehv1.AlgorithmIdentity{
			AlgorithmId: "V1_SHORTEST_SIGNED_CIRCULAR_DISPLACEMENT", AlgorithmVersion: "V1",
		},
	}
}
