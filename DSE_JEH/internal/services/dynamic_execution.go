package services

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	dsejehv1 "tramuthus/dse-jeh-transsat-1/gen/dse_jeh/v1"
	"tramuthus/dse-jeh-transsat-1/internal/evidence"
)

var authoritativeBoundaryDegrees = [...]float64{0, 90, 180, 270}

type DynamicExecution struct {
	dsejehv1.UnimplementedDynamicExecutionServiceServer
	boundaryPolicyProgram    *vm.Program
	boundaryPolicyCompileErr error
}

func NewDynamicExecution() *DynamicExecution {
	program, err := expr.Compile("crossed && forward", expr.AsBool())
	return &DynamicExecution{boundaryPolicyProgram: program, boundaryPolicyCompileErr: err}
}

func (service *DynamicExecution) EvaluateDynamicExecution(_ context.Context, request *dsejehv1.EvaluateDynamicExecutionRequest) (*dsejehv1.EvaluateDynamicExecutionResponse, error) {
	if service.boundaryPolicyCompileErr != nil {
		return nil, status.Errorf(codes.Internal, "compile boundary policy: %v", service.boundaryPolicyCompileErr)
	}
	if request == nil || request.GetCurrentPhaseEvidence() == nil || request.GetPhaseTransitionState() == nil || request.GetBoundaryPolicyRuleIdentity() == nil {
		return nil, status.Error(codes.InvalidArgument, "current_phase_evidence, phase_transition_state, and boundary_policy_rule_identity are required")
	}
	current := request.GetCurrentPhaseEvidence()
	transition := request.GetPhaseTransitionState()
	if !isPersistentDynamicExecutionState(request.GetPriorState()) {
		return nil, status.Error(codes.InvalidArgument, "prior_state must be one of the four persistent Dynamic Execution states")
	}
	if transition.GetStatus() != dsejehv1.PhaseTransitionStatus_PHASE_TRANSITION_STATUS_VALID || transition.SignedCircularDisplacementDegrees == nil {
		return nil, status.Error(codes.FailedPrecondition, "phase_transition_state must contain a valid signed displacement")
	}
	if current.PhaseDegrees == nil || math.IsNaN(current.GetPhaseDegrees()) || math.IsInf(current.GetPhaseDegrees(), 0) {
		return nil, status.Error(codes.InvalidArgument, "current phase must be finite")
	}
	if transition.GetEntityId() == "" || transition.GetEntityId() != current.GetEntityId() {
		return nil, status.Error(codes.InvalidArgument, "phase transition and current phase must identify the same entity")
	}
	if transition.GetCurrentPhaseEvidenceId() != "" && transition.GetCurrentPhaseEvidenceId() != current.GetEvidenceId() {
		return nil, status.Error(codes.InvalidArgument, "phase transition must reference the current phase evidence")
	}

	facts, err := deriveBoundaryFacts(current.GetPhaseDegrees(), transition)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if len(request.GetBoundaryFacts()) > 0 && !equalBoundaryFacts(request.GetBoundaryFacts(), facts) {
		return nil, status.Error(codes.InvalidArgument, "provided boundary_facts do not match deterministic derivation")
	}

	produced := time.Now().UnixMilli()
	currentState := request.GetPriorState()
	finalOutcome := dsejehv1.BoundaryPolicyOutcome_BOUNDARY_POLICY_OUTCOME_NO_TRANSITION
	evaluations := make([]*dsejehv1.RuleEvaluationEvidence, 0, max(1, len(facts)))
	if len(facts) == 0 {
		evaluation, evaluationErr := service.evaluateBoundaryPolicy(request, nil, currentState, produced)
		if evaluationErr != nil {
			return nil, evaluationErr
		}
		evaluations = append(evaluations, evaluation)
	} else {
		for _, fact := range facts {
			evaluation, evaluationErr := service.evaluateBoundaryPolicy(request, fact, currentState, produced)
			if evaluationErr != nil {
				return nil, evaluationErr
			}
			typedOutcome := evaluation.GetBoundaryPolicyOutcome()
			finalOutcome = typedOutcome.GetOutcome()
			currentState = typedOutcome.GetResultingState()
			evaluations = append(evaluations, evaluation)
		}
	}

	resultStatus := dsejehv1.DynamicExecutionStateStatus_DYNAMIC_EXECUTION_STATE_STATUS_PERSISTED
	if currentState != request.GetPriorState() {
		resultStatus = dsejehv1.DynamicExecutionStateStatus_DYNAMIC_EXECUTION_STATE_STATUS_TRANSITIONED
	}
	resultID := evidence.ID("dynamic-execution", transition.GetTransitionStateId(), request.GetPriorState().String(), currentState.String())
	action := stateActionOutcome(resultID, currentState, produced)
	lastEvaluation := evaluations[len(evaluations)-1]
	result := &dsejehv1.DynamicExecutionResult{
		ResultId: resultID, RuntimeId: request.GetRuntimeId(), EntityId: transition.GetEntityId(), EntitySequence: transition.GetEntitySequence(),
		Status: resultStatus, PreviousState: request.GetPriorState(), CurrentState: currentState, BoundaryPolicyOutcome: finalOutcome,
		PhaseTransitionStateId: transition.GetTransitionStateId(), StateActionOutcome: action,
		RuleEvaluationEvidenceId: lastEvaluation.GetEvidenceId(), PolicyIdentity: request.GetContext().GetPolicyIdentity(),
		Reason: "ordered deterministic boundary facts evaluated by governed forward policy", ProducedUnixMs: produced,
		BoundaryFacts: cloneBoundaryFacts(facts),
	}
	return &dsejehv1.EvaluateDynamicExecutionResponse{
		DynamicExecutionResult: result, BoundaryPolicyRuleEvaluationEvidence: lastEvaluation,
		BoundaryPolicyRuleEvaluationEvidences: evaluations,
	}, nil
}

func deriveBoundaryFacts(currentDegrees float64, transition *dsejehv1.PhaseTransitionState) ([]*dsejehv1.BoundaryFact, error) {
	delta := transition.GetSignedCircularDisplacementDegrees()
	if math.IsNaN(delta) || math.IsInf(delta, 0) || delta <= -180 || delta > 180 {
		return nil, fmt.Errorf("signed displacement must be finite and in (-180,180]")
	}
	if delta == 0 {
		if transition.GetDirection() != dsejehv1.PhaseTransitionDirection_PHASE_TRANSITION_DIRECTION_STATIONARY {
			return nil, fmt.Errorf("zero displacement must be STATIONARY")
		}
		return nil, nil
	}
	direction := dsejehv1.PhaseTransitionDirection_PHASE_TRANSITION_DIRECTION_FORWARD
	if delta < 0 {
		direction = dsejehv1.PhaseTransitionDirection_PHASE_TRANSITION_DIRECTION_REVERSE
	}
	if transition.GetDirection() != direction {
		return nil, fmt.Errorf("transition direction does not match signed displacement")
	}

	start := normalizePhase(normalizePhase(currentDegrees) - delta)
	type encounteredBoundary struct {
		degrees  float64
		distance float64
	}
	encountered := make([]encounteredBoundary, 0, 2)
	for _, boundary := range authoritativeBoundaryDegrees {
		distance := normalizePhase(boundary - start)
		limit := delta
		if delta < 0 {
			distance = normalizePhase(start - boundary)
			limit = -delta
		}
		if distance > 0 && distance <= limit {
			encountered = append(encountered, encounteredBoundary{degrees: boundary, distance: distance})
		}
	}
	sort.Slice(encountered, func(left, right int) bool { return encountered[left].distance < encountered[right].distance })
	facts := make([]*dsejehv1.BoundaryFact, 0, len(encountered))
	for index, boundary := range encountered {
		facts = append(facts, &dsejehv1.BoundaryFact{
			BoundaryDegrees: boundary.degrees, Crossed: true, Direction: direction, EncounterOrder: uint32(index + 1),
		})
	}
	return facts, nil
}

func (service *DynamicExecution) evaluateBoundaryPolicy(request *dsejehv1.EvaluateDynamicExecutionRequest, fact *dsejehv1.BoundaryFact, currentState dsejehv1.DynamicExecutionState, produced int64) (*dsejehv1.RuleEvaluationEvidence, error) {
	crossed := fact != nil && fact.GetCrossed()
	forward := fact != nil && fact.GetDirection() == dsejehv1.PhaseTransitionDirection_PHASE_TRANSITION_DIRECTION_FORWARD
	raw, err := expr.Run(service.boundaryPolicyProgram, map[string]any{"crossed": crossed, "forward": forward})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "evaluate boundary policy: %v", err)
	}
	fired := raw.(bool)
	outcome := dsejehv1.BoundaryPolicyOutcome_BOUNDARY_POLICY_OUTCOME_NO_TRANSITION
	resultingState := currentState
	if fired {
		outcome, resultingState = forwardBoundaryPolicy(fact.GetBoundaryDegrees(), currentState)
	}
	evaluationStatus := dsejehv1.RuleEvaluationStatus_RULE_EVALUATION_STATUS_NO_ACTION
	if fired {
		evaluationStatus = dsejehv1.RuleEvaluationStatus_RULE_EVALUATION_STATUS_PASSED
	}
	causalIdentity := request.GetPhaseTransitionState().GetTransitionStateId()
	encounterOrder := uint32(0)
	if fact != nil {
		encounterOrder = fact.GetEncounterOrder()
		causalIdentity = evidence.ID(causalIdentity, fmt.Sprintf("boundary-%g", fact.GetBoundaryDegrees()), fact.GetDirection().String(), fmt.Sprint(encounterOrder))
	}
	typedOutcome := &dsejehv1.BoundaryPolicyRuleOutcome{
		Outcome: outcome, ResultingState: resultingState,
		TransitionInput: proto.Clone(request.GetPhaseTransitionState()).(*dsejehv1.PhaseTransitionState),
	}
	return &dsejehv1.RuleEvaluationEvidence{
		EvidenceId: evidence.ID("boundary-policy-evaluation", request.GetBoundaryPolicyRuleIdentity().GetRuleId(), causalIdentity),
		RuntimeId:  request.GetRuntimeId(), RuleIdentity: proto.Clone(request.GetBoundaryPolicyRuleIdentity()).(*dsejehv1.RuleIdentity),
		CausalInputIdentity: causalIdentity, EvaluationSequence: request.GetPhaseTransitionState().GetEntitySequence(), EvaluationStatus: evaluationStatus,
		RawResult:    &dsejehv1.RuleEvaluationEvidence_RawBooleanResult{RawBooleanResult: fired},
		TypedOutcome: &dsejehv1.RuleEvaluationEvidence_BoundaryPolicyOutcome{BoundaryPolicyOutcome: typedOutcome}, ProducedUnixMs: produced,
	}, nil
}

func forwardBoundaryPolicy(boundary float64, currentState dsejehv1.DynamicExecutionState) (dsejehv1.BoundaryPolicyOutcome, dsejehv1.DynamicExecutionState) {
	switch boundary {
	case 270:
		return dsejehv1.BoundaryPolicyOutcome_BOUNDARY_POLICY_OUTCOME_HOP_ON_ALLOCATE, dsejehv1.DynamicExecutionState_DYNAMIC_EXECUTION_STATE_ALLOCATE
	case 0:
		return dsejehv1.BoundaryPolicyOutcome_BOUNDARY_POLICY_OUTCOME_ENTER_HOLD_AND_TRAIL, dsejehv1.DynamicExecutionState_DYNAMIC_EXECUTION_STATE_HOLD_AND_TRAIL
	case 90:
		return dsejehv1.BoundaryPolicyOutcome_BOUNDARY_POLICY_OUTCOME_HOP_OFF_LIQUIDATE, dsejehv1.DynamicExecutionState_DYNAMIC_EXECUTION_STATE_LIQUIDATE
	case 180:
		return dsejehv1.BoundaryPolicyOutcome_BOUNDARY_POLICY_OUTCOME_ENTER_DISREGARD, dsejehv1.DynamicExecutionState_DYNAMIC_EXECUTION_STATE_DISREGARD
	default:
		return dsejehv1.BoundaryPolicyOutcome_BOUNDARY_POLICY_OUTCOME_NO_TRANSITION, currentState
	}
}

func stateActionOutcome(resultID string, state dsejehv1.DynamicExecutionState, produced int64) *dsejehv1.StateActionOutcome {
	actionType := dsejehv1.StateActionType_STATE_ACTION_TYPE_NO_ALLOCATION_ACTION
	actionStatus := dsejehv1.StateActionStatus_STATE_ACTION_STATUS_NO_ACTION
	reason := "DISREGARD requires no allocation action"
	switch state {
	case dsejehv1.DynamicExecutionState_DYNAMIC_EXECUTION_STATE_ALLOCATE:
		actionType, actionStatus, reason = dsejehv1.StateActionType_STATE_ACTION_TYPE_RANK_BY_PHASE_VELOCITY, dsejehv1.StateActionStatus_STATE_ACTION_STATUS_BLOCKED, "ALLOCATE ranking algorithm is not implemented"
	case dsejehv1.DynamicExecutionState_DYNAMIC_EXECUTION_STATE_HOLD_AND_TRAIL:
		actionType, actionStatus, reason = dsejehv1.StateActionType_STATE_ACTION_TYPE_TRAIL_STOPS_DYNAMICALLY, dsejehv1.StateActionStatus_STATE_ACTION_STATUS_BLOCKED, "HOLD_AND_TRAIL algorithm is not implemented"
	case dsejehv1.DynamicExecutionState_DYNAMIC_EXECUTION_STATE_LIQUIDATE:
		actionType, actionStatus, reason = dsejehv1.StateActionType_STATE_ACTION_TYPE_REALLOCATE_FREED_CAPITAL, dsejehv1.StateActionStatus_STATE_ACTION_STATUS_BLOCKED, "LIQUIDATE reallocation algorithm is not implemented"
	}
	return &dsejehv1.StateActionOutcome{
		ActionOutcomeId: evidence.ID("state-action", resultID, state.String()), State: state, ActionType: actionType,
		Status: actionStatus, Reason: reason, ProducedUnixMs: produced,
	}
}

func isPersistentDynamicExecutionState(state dsejehv1.DynamicExecutionState) bool {
	return state >= dsejehv1.DynamicExecutionState_DYNAMIC_EXECUTION_STATE_DISREGARD && state <= dsejehv1.DynamicExecutionState_DYNAMIC_EXECUTION_STATE_LIQUIDATE
}

func equalBoundaryFacts(left, right []*dsejehv1.BoundaryFact) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index].GetBoundaryDegrees() != right[index].GetBoundaryDegrees() || left[index].GetCrossed() != right[index].GetCrossed() || left[index].GetDirection() != right[index].GetDirection() || left[index].GetEncounterOrder() != right[index].GetEncounterOrder() {
			return false
		}
	}
	return true
}

func cloneBoundaryFacts(facts []*dsejehv1.BoundaryFact) []*dsejehv1.BoundaryFact {
	clones := make([]*dsejehv1.BoundaryFact, 0, len(facts))
	for _, fact := range facts {
		clones = append(clones, proto.Clone(fact).(*dsejehv1.BoundaryFact))
	}
	return clones
}
