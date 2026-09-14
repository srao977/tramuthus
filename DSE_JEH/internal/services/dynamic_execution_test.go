package services

import (
	"context"
	"testing"

	dsejehv1 "tramuthus/dse-jeh-transsat-1/gen/dse_jeh/v1"
)

func TestEvaluateDynamicExecutionBoundarySemantics(t *testing.T) {
	tests := []struct {
		name             string
		prior            float64
		current          float64
		priorState       dsejehv1.DynamicExecutionState
		wantBoundaries   []float64
		wantEvaluations  []dsejehv1.BoundaryPolicyOutcome
		wantDirection    dsejehv1.PhaseTransitionDirection
		wantOutcome      dsejehv1.BoundaryPolicyOutcome
		wantCurrentState dsejehv1.DynamicExecutionState
	}{
		{name: "no crossing", prior: 10, current: 20, priorState: dsejehv1.DynamicExecutionState_DYNAMIC_EXECUTION_STATE_HOLD_AND_TRAIL, wantEvaluations: []dsejehv1.BoundaryPolicyOutcome{dsejehv1.BoundaryPolicyOutcome_BOUNDARY_POLICY_OUTCOME_NO_TRANSITION}, wantOutcome: dsejehv1.BoundaryPolicyOutcome_BOUNDARY_POLICY_OUTCOME_NO_TRANSITION, wantCurrentState: dsejehv1.DynamicExecutionState_DYNAMIC_EXECUTION_STATE_HOLD_AND_TRAIL},
		{name: "forward single boundary", prior: 260, current: 280, priorState: dsejehv1.DynamicExecutionState_DYNAMIC_EXECUTION_STATE_DISREGARD, wantBoundaries: []float64{270}, wantEvaluations: []dsejehv1.BoundaryPolicyOutcome{dsejehv1.BoundaryPolicyOutcome_BOUNDARY_POLICY_OUTCOME_HOP_ON_ALLOCATE}, wantDirection: dsejehv1.PhaseTransitionDirection_PHASE_TRANSITION_DIRECTION_FORWARD, wantOutcome: dsejehv1.BoundaryPolicyOutcome_BOUNDARY_POLICY_OUTCOME_HOP_ON_ALLOCATE, wantCurrentState: dsejehv1.DynamicExecutionState_DYNAMIC_EXECUTION_STATE_ALLOCATE},
		{name: "reverse single boundary is fact only", prior: 280, current: 260, priorState: dsejehv1.DynamicExecutionState_DYNAMIC_EXECUTION_STATE_ALLOCATE, wantBoundaries: []float64{270}, wantEvaluations: []dsejehv1.BoundaryPolicyOutcome{dsejehv1.BoundaryPolicyOutcome_BOUNDARY_POLICY_OUTCOME_NO_TRANSITION}, wantDirection: dsejehv1.PhaseTransitionDirection_PHASE_TRANSITION_DIRECTION_REVERSE, wantOutcome: dsejehv1.BoundaryPolicyOutcome_BOUNDARY_POLICY_OUTCOME_NO_TRANSITION, wantCurrentState: dsejehv1.DynamicExecutionState_DYNAMIC_EXECUTION_STATE_ALLOCATE},
		{name: "forward wrap", prior: 350, current: 10, priorState: dsejehv1.DynamicExecutionState_DYNAMIC_EXECUTION_STATE_ALLOCATE, wantBoundaries: []float64{0}, wantEvaluations: []dsejehv1.BoundaryPolicyOutcome{dsejehv1.BoundaryPolicyOutcome_BOUNDARY_POLICY_OUTCOME_ENTER_HOLD_AND_TRAIL}, wantDirection: dsejehv1.PhaseTransitionDirection_PHASE_TRANSITION_DIRECTION_FORWARD, wantOutcome: dsejehv1.BoundaryPolicyOutcome_BOUNDARY_POLICY_OUTCOME_ENTER_HOLD_AND_TRAIL, wantCurrentState: dsejehv1.DynamicExecutionState_DYNAMIC_EXECUTION_STATE_HOLD_AND_TRAIL},
		{name: "reverse wrap is fact only", prior: 10, current: 350, priorState: dsejehv1.DynamicExecutionState_DYNAMIC_EXECUTION_STATE_HOLD_AND_TRAIL, wantBoundaries: []float64{0}, wantEvaluations: []dsejehv1.BoundaryPolicyOutcome{dsejehv1.BoundaryPolicyOutcome_BOUNDARY_POLICY_OUTCOME_NO_TRANSITION}, wantDirection: dsejehv1.PhaseTransitionDirection_PHASE_TRANSITION_DIRECTION_REVERSE, wantOutcome: dsejehv1.BoundaryPolicyOutcome_BOUNDARY_POLICY_OUTCOME_NO_TRANSITION, wantCurrentState: dsejehv1.DynamicExecutionState_DYNAMIC_EXECUTION_STATE_HOLD_AND_TRAIL},
		{name: "start boundary excluded", prior: 270, current: 280, priorState: dsejehv1.DynamicExecutionState_DYNAMIC_EXECUTION_STATE_ALLOCATE, wantEvaluations: []dsejehv1.BoundaryPolicyOutcome{dsejehv1.BoundaryPolicyOutcome_BOUNDARY_POLICY_OUTCOME_NO_TRANSITION}, wantOutcome: dsejehv1.BoundaryPolicyOutcome_BOUNDARY_POLICY_OUTCOME_NO_TRANSITION, wantCurrentState: dsejehv1.DynamicExecutionState_DYNAMIC_EXECUTION_STATE_ALLOCATE},
		{name: "end boundary included", prior: 260, current: 270, priorState: dsejehv1.DynamicExecutionState_DYNAMIC_EXECUTION_STATE_DISREGARD, wantBoundaries: []float64{270}, wantEvaluations: []dsejehv1.BoundaryPolicyOutcome{dsejehv1.BoundaryPolicyOutcome_BOUNDARY_POLICY_OUTCOME_HOP_ON_ALLOCATE}, wantDirection: dsejehv1.PhaseTransitionDirection_PHASE_TRANSITION_DIRECTION_FORWARD, wantOutcome: dsejehv1.BoundaryPolicyOutcome_BOUNDARY_POLICY_OUTCOME_HOP_ON_ALLOCATE, wantCurrentState: dsejehv1.DynamicExecutionState_DYNAMIC_EXECUTION_STATE_ALLOCATE},
		{name: "stationary", prior: 45, current: 45, priorState: dsejehv1.DynamicExecutionState_DYNAMIC_EXECUTION_STATE_HOLD_AND_TRAIL, wantEvaluations: []dsejehv1.BoundaryPolicyOutcome{dsejehv1.BoundaryPolicyOutcome_BOUNDARY_POLICY_OUTCOME_NO_TRANSITION}, wantOutcome: dsejehv1.BoundaryPolicyOutcome_BOUNDARY_POLICY_OUTCOME_NO_TRANSITION, wantCurrentState: dsejehv1.DynamicExecutionState_DYNAMIC_EXECUTION_STATE_HOLD_AND_TRAIL},
		{name: "multiple boundaries preserve order", prior: 200, current: 20, priorState: dsejehv1.DynamicExecutionState_DYNAMIC_EXECUTION_STATE_DISREGARD, wantBoundaries: []float64{270, 0}, wantEvaluations: []dsejehv1.BoundaryPolicyOutcome{dsejehv1.BoundaryPolicyOutcome_BOUNDARY_POLICY_OUTCOME_HOP_ON_ALLOCATE, dsejehv1.BoundaryPolicyOutcome_BOUNDARY_POLICY_OUTCOME_ENTER_HOLD_AND_TRAIL}, wantDirection: dsejehv1.PhaseTransitionDirection_PHASE_TRANSITION_DIRECTION_FORWARD, wantOutcome: dsejehv1.BoundaryPolicyOutcome_BOUNDARY_POLICY_OUTCOME_ENTER_HOLD_AND_TRAIL, wantCurrentState: dsejehv1.DynamicExecutionState_DYNAMIC_EXECUTION_STATE_HOLD_AND_TRAIL},
		{name: "exact positive tie", prior: 0, current: 180, priorState: dsejehv1.DynamicExecutionState_DYNAMIC_EXECUTION_STATE_HOLD_AND_TRAIL, wantBoundaries: []float64{90, 180}, wantEvaluations: []dsejehv1.BoundaryPolicyOutcome{dsejehv1.BoundaryPolicyOutcome_BOUNDARY_POLICY_OUTCOME_HOP_OFF_LIQUIDATE, dsejehv1.BoundaryPolicyOutcome_BOUNDARY_POLICY_OUTCOME_ENTER_DISREGARD}, wantDirection: dsejehv1.PhaseTransitionDirection_PHASE_TRANSITION_DIRECTION_FORWARD, wantOutcome: dsejehv1.BoundaryPolicyOutcome_BOUNDARY_POLICY_OUTCOME_ENTER_DISREGARD, wantCurrentState: dsejehv1.DynamicExecutionState_DYNAMIC_EXECUTION_STATE_DISREGARD},
	}

	service := NewDynamicExecution()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := evaluateTransition(t, service, test.prior, test.current, test.priorState)
			result := response.GetDynamicExecutionResult()
			if result.GetBoundaryPolicyOutcome() != test.wantOutcome {
				t.Fatalf("outcome = %s; want %s", result.GetBoundaryPolicyOutcome(), test.wantOutcome)
			}
			if result.GetCurrentState() != test.wantCurrentState {
				t.Fatalf("state = %s; want %s", result.GetCurrentState(), test.wantCurrentState)
			}
			if len(result.GetBoundaryFacts()) != len(test.wantBoundaries) {
				t.Fatalf("boundary count = %d; want %d", len(result.GetBoundaryFacts()), len(test.wantBoundaries))
			}
			for index, wantBoundary := range test.wantBoundaries {
				fact := result.GetBoundaryFacts()[index]
				if fact.GetBoundaryDegrees() != wantBoundary || fact.GetDirection() != test.wantDirection || fact.GetEncounterOrder() != uint32(index+1) {
					t.Fatalf("fact[%d] = %+v; want boundary=%v direction=%s order=%d", index, fact, wantBoundary, test.wantDirection, index+1)
				}
			}
			evaluations := response.GetBoundaryPolicyRuleEvaluationEvidences()
			if len(evaluations) != len(test.wantEvaluations) {
				t.Fatalf("evaluation count = %d; want %d", len(evaluations), len(test.wantEvaluations))
			}
			for index, wantEvaluation := range test.wantEvaluations {
				if got := evaluations[index].GetBoundaryPolicyOutcome().GetOutcome(); got != wantEvaluation {
					t.Fatalf("evaluation[%d] = %s; want %s", index, got, wantEvaluation)
				}
			}
			if result.GetExecutionInstruction() != nil {
				t.Fatal("EvaluateDynamicExecution must not generate downstream execution instructions in this slice")
			}
		})
	}
}

func TestEvaluateDynamicExecutionRecrossesOnConsecutiveBars(t *testing.T) {
	service := NewDynamicExecution()
	forward := evaluateTransition(t, service, 179.5, 180.5, dsejehv1.DynamicExecutionState_DYNAMIC_EXECUTION_STATE_LIQUIDATE)
	reverse := evaluateTransition(t, service, 180.5, 179.5, dsejehv1.DynamicExecutionState_DYNAMIC_EXECUTION_STATE_DISREGARD)

	forwardFact := forward.GetDynamicExecutionResult().GetBoundaryFacts()[0]
	reverseFact := reverse.GetDynamicExecutionResult().GetBoundaryFacts()[0]
	if forwardFact.GetBoundaryDegrees() != 180 || forwardFact.GetDirection() != dsejehv1.PhaseTransitionDirection_PHASE_TRANSITION_DIRECTION_FORWARD {
		t.Fatalf("forward recross fact = %+v", forwardFact)
	}
	if reverseFact.GetBoundaryDegrees() != 180 || reverseFact.GetDirection() != dsejehv1.PhaseTransitionDirection_PHASE_TRANSITION_DIRECTION_REVERSE {
		t.Fatalf("reverse recross fact = %+v", reverseFact)
	}
	if reverse.GetDynamicExecutionResult().GetBoundaryPolicyOutcome() != dsejehv1.BoundaryPolicyOutcome_BOUNDARY_POLICY_OUTCOME_NO_TRANSITION {
		t.Fatal("reverse recross must not fire forward policy")
	}
}

func evaluateTransition(t *testing.T, service *DynamicExecution, prior, current float64, priorState dsejehv1.DynamicExecutionState) *dsejehv1.EvaluateDynamicExecutionResponse {
	t.Helper()
	transitionResponse, err := service.CalculatePhaseTransition(context.Background(), phaseTransitionRequest(prior, current))
	if err != nil {
		t.Fatal(err)
	}
	request := phaseTransitionRequest(prior, current)
	response, err := service.EvaluateDynamicExecution(context.Background(), &dsejehv1.EvaluateDynamicExecutionRequest{
		RuntimeId: "runtime-test", CurrentPhaseEvidence: request.GetCurrentPhaseEvidence(),
		PhaseTransitionState: transitionResponse.GetPhaseTransitionState(), PriorState: priorState,
		Context:                    &dsejehv1.DynamicExecutionContext{ContextId: "context-test", PolicyIdentity: "FOUR_STATE_FORWARD_V1"},
		BoundaryPolicyRuleIdentity: &dsejehv1.RuleIdentity{RuleId: "FOUR_STATE_FORWARD_BOUNDARY_POLICY", RuleVersion: "V1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	return response
}
