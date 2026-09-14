package services

import (
	"context"
	"math"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	dsejehv1 "tramuthus/dse-jeh-transsat-1/gen/dse_jeh/v1"
	"tramuthus/dse-jeh-transsat-1/internal/evidence"
)

func (service *DynamicExecution) CalculatePhaseTransition(_ context.Context, request *dsejehv1.CalculatePhaseTransitionRequest) (*dsejehv1.CalculatePhaseTransitionResponse, error) {
	if request == nil || request.GetPriorPhaseEvidence() == nil || request.GetCurrentPhaseEvidence() == nil || request.GetProductionEligibilityEvidence() == nil {
		return nil, status.Error(codes.InvalidArgument, "prior_phase_evidence, current_phase_evidence, and production_eligibility_evidence are required")
	}

	prior := request.GetPriorPhaseEvidence()
	current := request.GetCurrentPhaseEvidence()
	eligibility := request.GetProductionEligibilityEvidence()
	if prior.PhaseDegrees == nil || current.PhaseDegrees == nil {
		return nil, status.Error(codes.InvalidArgument, "prior and current phase values are required")
	}
	if prior.GetEntityId() == "" || prior.GetEntityId() != current.GetEntityId() {
		return nil, status.Error(codes.InvalidArgument, "prior and current phase evidence must identify the same entity")
	}
	if eligibility.GetOutcome() != dsejehv1.ProductionEligibilityOutcome_PRODUCTION_ELIGIBILITY_OUTCOME_PRODUCTION_ELIGIBLE {
		return nil, status.Error(codes.FailedPrecondition, "current phase must be production eligible")
	}
	if eligibility.GetPhaseEvidenceId() != "" && eligibility.GetPhaseEvidenceId() != current.GetEvidenceId() {
		return nil, status.Error(codes.InvalidArgument, "production eligibility evidence must reference the current phase")
	}

	transition := newPhaseTransitionState(request)
	priorDegrees := prior.GetPhaseDegrees()
	currentDegrees := current.GetPhaseDegrees()
	if math.IsNaN(priorDegrees) || math.IsInf(priorDegrees, 0) || math.IsNaN(currentDegrees) || math.IsInf(currentDegrees, 0) {
		transition.Status = dsejehv1.PhaseTransitionStatus_PHASE_TRANSITION_STATUS_INVALID
		transition.Direction = dsejehv1.PhaseTransitionDirection_PHASE_TRANSITION_DIRECTION_UNSPECIFIED
		transition.Reason = "prior and current phase values must be finite"
		return &dsejehv1.CalculatePhaseTransitionResponse{PhaseTransitionState: transition}, nil
	}

	delta := shortestCircularDisplacement(normalizePhase(priorDegrees), normalizePhase(currentDegrees))
	magnitude := math.Abs(delta)
	transition.Status = dsejehv1.PhaseTransitionStatus_PHASE_TRANSITION_STATUS_VALID
	transition.Direction = phaseTransitionDirection(delta)
	transition.SignedCircularDisplacementDegrees = &delta
	transition.MagnitudeDegrees = &magnitude
	transition.PhaseVelocityDegreesPerBar = proto.Float64(delta)
	return &dsejehv1.CalculatePhaseTransitionResponse{PhaseTransitionState: transition}, nil
}

func newPhaseTransitionState(request *dsejehv1.CalculatePhaseTransitionRequest) *dsejehv1.PhaseTransitionState {
	prior := request.GetPriorPhaseEvidence()
	current := request.GetCurrentPhaseEvidence()
	eligibility := request.GetProductionEligibilityEvidence()
	algorithmIdentity := request.GetAlgorithmIdentity()
	transitionIDParts := []string{"phase-transition", request.GetRuntimeId(), prior.GetEvidenceId(), current.GetEvidenceId(), eligibility.GetEvidenceId()}
	if algorithmIdentity != nil {
		transitionIDParts = append(transitionIDParts, algorithmIdentity.GetAlgorithmId(), algorithmIdentity.GetAlgorithmVersion(), algorithmIdentity.GetConfigurationId())
	}
	transition := &dsejehv1.PhaseTransitionState{
		TransitionStateId:               evidence.ID(transitionIDParts...),
		RuntimeId:                       request.GetRuntimeId(),
		EntityId:                        current.GetEntityId(),
		EntitySequence:                  current.GetEntitySequence(),
		PriorPhaseEvidenceId:            prior.GetEvidenceId(),
		CurrentPhaseEvidenceId:          current.GetEvidenceId(),
		ProductionEligibilityEvidenceId: eligibility.GetEvidenceId(),
		ProducedUnixMs:                  time.Now().UnixMilli(),
	}
	if algorithmIdentity != nil {
		transition.AlgorithmIdentity = proto.Clone(algorithmIdentity).(*dsejehv1.AlgorithmIdentity)
	}
	return transition
}

func normalizePhase(value float64) float64 {
	normalized := math.Mod(value, 360)
	if normalized < 0 {
		normalized += 360
	}
	if normalized == 0 {
		return 0
	}
	return normalized
}

func shortestCircularDisplacement(priorDegrees, currentDegrees float64) float64 {
	delta := currentDegrees - priorDegrees
	if delta <= -180 {
		delta += 360
	} else if delta > 180 {
		delta -= 360
	}
	if delta == 0 {
		return 0
	}
	return delta
}

func phaseTransitionDirection(delta float64) dsejehv1.PhaseTransitionDirection {
	switch {
	case delta > 0:
		return dsejehv1.PhaseTransitionDirection_PHASE_TRANSITION_DIRECTION_FORWARD
	case delta < 0:
		return dsejehv1.PhaseTransitionDirection_PHASE_TRANSITION_DIRECTION_REVERSE
	default:
		return dsejehv1.PhaseTransitionDirection_PHASE_TRANSITION_DIRECTION_STATIONARY
	}
}
