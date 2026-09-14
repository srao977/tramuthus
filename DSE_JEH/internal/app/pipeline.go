package app

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	dsejehv1 "tramuthus/dse-jeh-transsat-1/gen/dse_jeh/v1"
	"tramuthus/dse-jeh-transsat-1/internal/evidence"
	"tramuthus/dse-jeh-transsat-1/internal/execution"
	runtimeapp "tramuthus/dse-jeh-transsat-1/internal/runtime"
	"tramuthus/dse-jeh-transsat-1/internal/services"
)

type Pipeline struct {
	runtimeID        string
	reception        *services.BarReception
	admission        *services.BarAdmission
	analytical       *services.AnalyticalState
	phase            *services.JehPhase
	eligibility      *services.ProductionEligibility
	state            *runtimeapp.State
	bus              *evidence.Bus
	terminal         *runtimeapp.Terminal
	phaseWriter      *evidence.PhaseWriter
	dynamic          *services.DynamicExecution
	executor         execution.MockExecutor
	traceWriter      execution.TraceWriter
	parameters       execution.RunParameters
	pipelineRunID    string
	runType          execution.RunType
	collectionRunID  string
	entities         map[string]*pipelineEntity
	reservoir        *execution.CapitalReservoir
	now              func() time.Time
	reservoirStarted bool
	reservoirEnded   bool
}

type pipelineEntity struct {
	priorPhase             *dsejehv1.PhaseEvidence
	state                  dsejehv1.DynamicExecutionState
	initialized            bool
	account                *execution.Account
	hopOnCount             uint64
	holdCount              uint64
	hopOffCount            uint64
	safetyLiquidationCount uint64
}

func (pipeline *Pipeline) Process(ctx context.Context, event *dsejehv1.BarEvent) error {
	receptionResponse, err := pipeline.reception.ReceiveBar(ctx, &dsejehv1.ReceiveBarRequest{RuntimeId: pipeline.runtimeID, BarEvent: event})
	if err != nil {
		return pipeline.fail(event, nil, err)
	}
	sequence := event.GetProvenance().GetEntitySequence()
	pipeline.terminal.Event("BAR RX", "%s | seq=%d | event=%s", event.GetEntityId(), sequence, event.GetEventId())

	admissionResponse, err := pipeline.admission.AdmitBar(ctx, &dsejehv1.AdmitBarRequest{RuntimeId: pipeline.runtimeID, BarEvent: event, ReceptionEvidence: receptionResponse.GetReceptionEvidence()})
	if err != nil {
		return pipeline.fail(event, receptionResponse.GetReceptionEvidence(), err)
	}
	admissionEvidence := admissionResponse.GetAdmissionEvidence()
	pipeline.terminal.Event("DEP-01", "%s | seq=%d | %s", event.GetEntityId(), sequence, admissionEvidence.GetStatus())
	if admissionEvidence.GetStatus() != dsejehv1.BarAdmissionStatus_BAR_ADMISSION_STATUS_ADMITTED {
		pipeline.terminalOutcome(event, receptionResponse.GetReceptionEvidence(), admissionEvidence, nil, nil, nil, dsejehv1.BarProcessingOutcomeType_BAR_PROCESSING_OUTCOME_TYPE_ADMISSION_REJECTED, admissionEvidence.GetReason())
		return nil
	}

	analyticalResponse, err := pipeline.analytical.UpdateAnalyticalState(ctx, &dsejehv1.UpdateAnalyticalStateRequest{RuntimeId: pipeline.runtimeID, BarEvent: event, AdmissionEvidence: admissionEvidence})
	if err != nil {
		return pipeline.fail(event, receptionResponse.GetReceptionEvidence(), err)
	}
	analyticalEvidence := analyticalResponse.GetAnalyticalStateEvidence()
	pipeline.terminal.Event("DEP-02", "%s | seq=%d | contiguous=%d | integrity=%s", event.GetEntityId(), sequence, analyticalEvidence.GetContiguousValidBarCount(), analyticalEvidence.GetSequenceIntegrity())

	phaseResponse, err := pipeline.phase.UpdateJehPhase(ctx, &dsejehv1.UpdateJehPhaseRequest{RuntimeId: pipeline.runtimeID, BarEvent: event, AdmissionEvidence: admissionEvidence, AnalyticalStateEvidence: analyticalEvidence})
	if err != nil {
		return pipeline.fail(event, receptionResponse.GetReceptionEvidence(), err)
	}
	phaseEvidence := phaseResponse.GetPhaseEvidence()
	entity := pipeline.entity(event.GetEntityId())
	if pipeline.reservoir != nil {
		pipeline.reservoir.Mark(event.GetEntityId(), event.GetClose())
	}
	priorPhase := entity.priorPhase
	if phaseEvidence.PhaseDegrees != nil {
		entity.priorPhase = phaseEvidence
	}
	if pipeline.phaseWriter != nil {
		if err := pipeline.phaseWriter.Write(phaseEvidence); err != nil {
			return pipeline.fail(event, receptionResponse.GetReceptionEvidence(), err)
		}
	}
	if phaseEvidence.PhaseDegrees == nil {
		pipeline.terminal.Event("JEH", "%s | seq=%d | %s", event.GetEntityId(), sequence, phaseEvidence.GetStatus())
	} else {
		pipeline.terminal.Event("JEH", "%s | seq=%d | %s | phase=%.12f", event.GetEntityId(), sequence, phaseEvidence.GetStatus(), phaseEvidence.GetPhaseDegrees())
	}

	eligibilityResponse, err := pipeline.eligibility.EvaluateProductionEligibility(ctx, &dsejehv1.EvaluateProductionEligibilityRequest{RuntimeId: pipeline.runtimeID, PhaseEvidence: phaseEvidence, AnalyticalStateEvidence: analyticalEvidence})
	if err != nil {
		return pipeline.fail(event, receptionResponse.GetReceptionEvidence(), err)
	}
	eligibilityEvidence := eligibilityResponse.GetProductionEligibilityEvidence()
	pipeline.terminal.Event("ELIGIBILITY", "%s | seq=%d | %s", event.GetEntityId(), sequence, eligibilityEvidence.GetOutcome())
	if eligibilityEvidence.GetOutcome() == dsejehv1.ProductionEligibilityOutcome_PRODUCTION_ELIGIBILITY_OUTCOME_PRODUCTION_ELIGIBLE {
		if entity.initialized && (priorPhase == nil || priorPhase.PhaseDegrees == nil) {
			return pipeline.fail(event, receptionResponse.GetReceptionEvidence(), fmt.Errorf("eligible phase has no retained predecessor phase for %s sequence %d", event.GetEntityId(), sequence))
		}
		if err := pipeline.processDynamic(ctx, event, priorPhase, phaseEvidence, eligibilityEvidence, entity); err != nil {
			return pipeline.fail(event, receptionResponse.GetReceptionEvidence(), err)
		}
		pipeline.terminalOutcome(event, receptionResponse.GetReceptionEvidence(), admissionEvidence, analyticalEvidence, phaseEvidence, eligibilityEvidence, dsejehv1.BarProcessingOutcomeType_BAR_PROCESSING_OUTCOME_TYPE_FOUR_STATE_RESULT, "complete four-state Dynamic Execution evaluation")
	} else if eligibilityEvidence.GetOutcome() == dsejehv1.ProductionEligibilityOutcome_PRODUCTION_ELIGIBILITY_OUTCOME_INITIALIZING {
		pipeline.terminalOutcome(event, receptionResponse.GetReceptionEvidence(), admissionEvidence, analyticalEvidence, phaseEvidence, eligibilityEvidence, dsejehv1.BarProcessingOutcomeType_BAR_PROCESSING_OUTCOME_TYPE_INITIALIZING, "production eligibility initialization")
	} else {
		pipeline.terminalOutcome(event, receptionResponse.GetReceptionEvidence(), admissionEvidence, analyticalEvidence, phaseEvidence, eligibilityEvidence, dsejehv1.BarProcessingOutcomeType_BAR_PROCESSING_OUTCOME_TYPE_PRODUCTION_ELIGIBILITY_BLOCKED, eligibilityEvidence.GetReason())
	}
	return nil
}

func (pipeline *Pipeline) processDynamic(ctx context.Context, event *dsejehv1.BarEvent, priorPhase, currentPhase *dsejehv1.PhaseEvidence, eligibility *dsejehv1.ProductionEligibilityEvidence, entity *pipelineEntity) error {
	if !entity.initialized {
		entity.state = stateForPhase(currentPhase.GetPhaseDegrees())
		entity.initialized = true
		return nil
	}
	transitionResponse, err := pipeline.dynamic.CalculatePhaseTransition(ctx, &dsejehv1.CalculatePhaseTransitionRequest{
		RuntimeId: pipeline.runtimeID, PriorPhaseEvidence: priorPhase, CurrentPhaseEvidence: currentPhase,
		ProductionEligibilityEvidence: eligibility,
		AlgorithmIdentity:             &dsejehv1.AlgorithmIdentity{AlgorithmId: "V1_SHORTEST_SIGNED_CIRCULAR_DISPLACEMENT", AlgorithmVersion: "V1", ConfigurationId: pipeline.pipelineRunID},
	})
	if err != nil {
		return fmt.Errorf("calculate phase transition: %w", err)
	}
	transition := transitionResponse.GetPhaseTransitionState()
	evaluationResponse, err := pipeline.dynamic.EvaluateDynamicExecution(ctx, &dsejehv1.EvaluateDynamicExecutionRequest{
		RuntimeId: pipeline.runtimeID, CurrentPhaseEvidence: currentPhase, PhaseTransitionState: transition, PriorState: entity.state,
		Context:                    &dsejehv1.DynamicExecutionContext{ContextId: evidence.ID("dynamic-context", pipeline.pipelineRunID, event.GetEventId()), PolicyIdentity: "FOUR_STATE_FORWARD_V1"},
		BoundaryPolicyRuleIdentity: &dsejehv1.RuleIdentity{RuleId: "FOUR_STATE_FORWARD_BOUNDARY_POLICY", RuleVersion: "V1", OwningComponent: dsejehv1.RuleOwningComponent_RULE_OWNING_COMPONENT_DYNAMIC_EXECUTION},
	})
	if err != nil {
		return fmt.Errorf("evaluate dynamic execution: %w", err)
	}
	result := evaluationResponse.GetDynamicExecutionResult()
	entity.state = result.GetCurrentState()
	pipeline.publishDynamic(transition, result)

	if err := pipeline.trace(ctx, event, 1, "PIPELINE_CONTEXT", "", map[string]any{
		"bar_event_id": event.GetEventId(), "admission_id": currentPhase.GetAdmissionId(), "phase_evidence_id": currentPhase.GetEvidenceId(), "production_eligibility_evidence_id": eligibility.GetEvidenceId(),
	}, map[string]any{"phase_degrees": currentPhase.GetPhaseDegrees(), "close": event.GetClose()}, nil); err != nil {
		return err
	}
	if err := pipeline.trace(ctx, event, 2, "PHASE_TRANSITION", "", map[string]any{"prior_phase_evidence_id": priorPhase.GetEvidenceId(), "current_phase_evidence_id": currentPhase.GetEvidenceId()}, map[string]any{
		"phase_transition_state_id": transition.GetTransitionStateId(), "signed_circular_displacement_degrees": transition.GetSignedCircularDisplacementDegrees(), "phase_velocity_degrees_per_bar": transition.GetPhaseVelocityDegreesPerBar(), "direction": transition.GetDirection().String(),
	}, nil); err != nil {
		return err
	}
	boundaryValues := make([]map[string]any, 0, len(result.GetBoundaryFacts()))
	for _, fact := range result.GetBoundaryFacts() {
		boundaryValues = append(boundaryValues, map[string]any{"degrees": fact.GetBoundaryDegrees(), "direction": fact.GetDirection().String(), "encounter_order": fact.GetEncounterOrder()})
	}
	if err := pipeline.trace(ctx, event, 3, "BOUNDARY_FACTS", "", map[string]any{"phase_transition_state_id": transition.GetTransitionStateId()}, map[string]any{"facts": boundaryValues}, nil); err != nil {
		return err
	}
	if err := pipeline.trace(ctx, event, 4, "DYNAMIC_EXECUTION", result.GetBoundaryPolicyOutcome().String(), map[string]any{"prior_state": result.GetPreviousState().String()}, map[string]any{
		"dynamic_execution_result_id": result.GetResultId(), "current_state": result.GetCurrentState().String(), "boundary_policy_outcome": result.GetBoundaryPolicyOutcome().String(), "state_action_outcome_id": result.GetStateActionOutcome().GetActionOutcomeId(),
	}, nil); err != nil {
		return err
	}

	actionTaken := false
	for _, evaluation := range evaluationResponse.GetBoundaryPolicyRuleEvaluationEvidences() {
		outcome := evaluation.GetBoundaryPolicyOutcome().GetOutcome()
		switch outcome {
		case dsejehv1.BoundaryPolicyOutcome_BOUNDARY_POLICY_OUTCOME_HOP_ON_ALLOCATE:
			actionTaken = true
			entity.hopOnCount++
			if err := pipeline.allocate(ctx, event, result, entity); err != nil {
				return err
			}
		case dsejehv1.BoundaryPolicyOutcome_BOUNDARY_POLICY_OUTCOME_ENTER_HOLD_AND_TRAIL:
			actionTaken = true
			if err := pipeline.holdOrSafetyLiquidate(ctx, event, result, entity); err != nil {
				return err
			}
		case dsejehv1.BoundaryPolicyOutcome_BOUNDARY_POLICY_OUTCOME_HOP_OFF_LIQUIDATE:
			actionTaken = true
			if entity.account.ActiveQuantity > 0 {
				entity.hopOffCount++
				if err := pipeline.liquidate(ctx, event, result, entity, "HOP_OFF_LIQUIDATE"); err != nil {
					return err
				}
			}
		}
	}
	if !actionTaken && entity.state == dsejehv1.DynamicExecutionState_DYNAMIC_EXECUTION_STATE_HOLD_AND_TRAIL {
		if err := pipeline.holdOrSafetyLiquidate(ctx, event, result, entity); err != nil {
			return err
		}
	}
	currentCapital, err := entity.account.CurrentCapital(event.GetClose())
	if err != nil {
		return err
	}
	return pipeline.trace(ctx, event, 8, "CAPITAL_RESULT", "CAPITAL", map[string]any{"dynamic_execution_result_id": result.GetResultId()}, map[string]any{}, pipeline.capitalValues(entity, currentCapital))
}

func (pipeline *Pipeline) allocate(ctx context.Context, event *dsejehv1.BarEvent, result *dsejehv1.DynamicExecutionResult, entity *pipelineEntity) error {
	if entity.account.ActiveQuantity != 0 {
		return pipeline.trace(ctx, event, 5, "STATE_ACTION", "ALLOCATE_SKIPPED_ACTIVE_POSITION", nil, map[string]any{"active_quantity": entity.account.ActiveQuantity}, nil)
	}
	allocationBasis := pipeline.parameters.StartingCapital
	if pipeline.reservoir != nil {
		allocationBasis = pipeline.reservoir.AllocationCapacity(allocationBasis)
	}
	allocation, err := execution.SizeAllocation(allocationBasis, pipeline.parameters.AllocationPct, event.GetClose())
	if err != nil {
		return err
	}
	if allocation.Quantity == 0 {
		return pipeline.trace(ctx, event, 5, "STATE_ACTION", "ALLOCATE_BLOCKED_ZERO_QUANTITY", nil, map[string]any{"reference_price": event.GetClose()}, pipeline.capitalValues(entity, entity.account.AvailableCapital))
	}
	return pipeline.execute(ctx, event, result, entity, dsejehv1.GovernedExecutionAction_GOVERNED_EXECUTION_ACTION_ALLOCATE, allocation.Quantity, "HOP_ON_ALLOCATE")
}

func (pipeline *Pipeline) holdOrSafetyLiquidate(ctx context.Context, event *dsejehv1.BarEvent, result *dsejehv1.DynamicExecutionResult, entity *pipelineEntity) error {
	if entity.account.ActiveQuantity == 0 {
		return nil
	}
	peak, stop, err := execution.TrailingPrices(entity.account.PeakPrice, event.GetClose(), pipeline.parameters.Risk)
	if err != nil {
		return err
	}
	entity.account.PeakPrice = peak
	if event.GetClose() <= stop {
		entity.safetyLiquidationCount++
		return pipeline.liquidate(ctx, event, result, entity, "SAFETY_LIQUIDATION")
	}
	entity.holdCount++
	currentCapital, err := entity.account.CurrentCapital(event.GetClose())
	if err != nil {
		return err
	}
	return pipeline.trace(ctx, event, 5, "STATE_ACTION", "HOLD", map[string]any{"dynamic_execution_result_id": result.GetResultId()}, map[string]any{"price": event.GetClose(), "peak_price": peak, "stop_price": stop}, pipeline.capitalValues(entity, currentCapital))
}

func (pipeline *Pipeline) liquidate(ctx context.Context, event *dsejehv1.BarEvent, result *dsejehv1.DynamicExecutionResult, entity *pipelineEntity, cause string) error {
	if entity.account.ActiveQuantity == 0 {
		return nil
	}
	return pipeline.execute(ctx, event, result, entity, dsejehv1.GovernedExecutionAction_GOVERNED_EXECUTION_ACTION_LIQUIDATE, entity.account.ActiveQuantity, cause)
}

func (pipeline *Pipeline) execute(ctx context.Context, event *dsejehv1.BarEvent, result *dsejehv1.DynamicExecutionResult, entity *pipelineEntity, action dsejehv1.GovernedExecutionAction, quantity uint64, cause string) error {
	instructionID := evidence.ID("governed-execution-instruction", pipeline.pipelineRunID, event.GetEventId(), action.String(), cause)
	instruction := &dsejehv1.GovernedExecutionInstruction{
		InstructionId: instructionID, RuntimeId: pipeline.runtimeID, EntityId: event.GetEntityId(), RequestedAction: action,
		Status:                   dsejehv1.GovernedExecutionInstructionStatus_GOVERNED_EXECUTION_INSTRUCTION_STATUS_ELIGIBLE,
		DynamicExecutionResultId: result.GetResultId(), StateActionOutcomeId: result.GetStateActionOutcome().GetActionOutcomeId(),
		RequestedQuantity: &dsejehv1.DecimalValue{Value: strconv.FormatUint(quantity, 10), Unit: "SHARES"},
		ConfigurationId:   pipeline.pipelineRunID, PolicyIdentity: "V1_WHOLE_SHARE_LOCAL_PAPER", IdempotencyKey: instructionID,
		CorrelationId: evidence.ID("execution-correlation", result.GetResultId(), cause), RuleEvaluationEvidenceId: result.GetRuleEvaluationEvidenceId(), ProducedUnixMs: time.Now().UnixMilli(),
	}
	result.ExecutionInstruction = instruction
	pipeline.bus.Publish(pipeline.runtimeID, &dsejehv1.RuntimeEvidenceEnvelope{Evidence: &dsejehv1.RuntimeEvidenceEnvelope_GovernedExecutionInstruction{GovernedExecutionInstruction: instruction}})
	if err := pipeline.trace(ctx, event, 5, "STATE_ACTION", cause, map[string]any{"dynamic_execution_result_id": result.GetResultId()}, map[string]any{"requested_action": action.String(), "requested_quantity": quantity}, nil); err != nil {
		return err
	}
	if err := pipeline.trace(ctx, event, 6, "GOVERNED_EXECUTION_INSTRUCTION", cause, map[string]any{"dynamic_execution_result_id": result.GetResultId()}, map[string]any{"instruction_id": instructionID, "action": action.String(), "quantity": quantity}, nil); err != nil {
		return err
	}
	executionEvent, err := pipeline.executor.Fill(instruction, event.GetClose())
	if err != nil {
		return err
	}
	if err := entity.account.ApplyConfirmed(instruction, executionEvent); err != nil {
		return err
	}
	pipeline.bus.Publish(pipeline.runtimeID, &dsejehv1.RuntimeEvidenceEnvelope{Evidence: &dsejehv1.RuntimeEvidenceEnvelope_ExecutionEvent{ExecutionEvent: executionEvent}})
	currentCapital, err := entity.account.CurrentCapital(event.GetClose())
	if err != nil {
		return err
	}
	stageEmitValueID, err := pipeline.traceWithID(ctx, event, 7, "EXECUTION_EVENT", cause, map[string]any{"instruction_id": instructionID}, map[string]any{"execution_event_id": executionEvent.GetEventId(), "status": executionEvent.GetStatus().String(), "fill_price": event.GetClose(), "filled_quantity": quantity}, pipeline.capitalValues(entity, currentCapital))
	if err != nil {
		return err
	}
	if pipeline.reservoir == nil {
		return nil
	}
	reservoirCause, err := capitalReservoirCause(cause)
	if err != nil {
		return err
	}
	reservoirEvent, err := pipeline.reservoir.RecordExecution(ctx, instruction, executionEvent, reservoirCause, stageEmitValueID, time.UnixMilli(event.GetIntervalStartUnixMs()), pipeline.now().UTC())
	if err != nil {
		return err
	}
	pipeline.publishReservoir(reservoirEvent)
	return nil
}

func (pipeline *Pipeline) StartReservoir(ctx context.Context) error {
	if pipeline.reservoir == nil || pipeline.reservoirStarted {
		return nil
	}
	event, err := pipeline.reservoir.Start(ctx, pipeline.now().UTC())
	if err != nil {
		return err
	}
	pipeline.reservoirStarted = true
	pipeline.publishReservoir(event)
	return nil
}

func (pipeline *Pipeline) EndReservoir(ctx context.Context) error {
	if pipeline.reservoir == nil || !pipeline.reservoirStarted || pipeline.reservoirEnded {
		return nil
	}
	event, err := pipeline.reservoir.End(ctx, pipeline.now().UTC())
	if err != nil {
		return err
	}
	pipeline.reservoirEnded = true
	pipeline.publishReservoir(event)
	return nil
}

func (pipeline *Pipeline) publishReservoir(event execution.CapitalReservoirEvent) {
	pipeline.bus.Publish(pipeline.runtimeID, &dsejehv1.RuntimeEvidenceEnvelope{
		Evidence: &dsejehv1.RuntimeEvidenceEnvelope_CapitalReservoirEvent{CapitalReservoirEvent: event.Proto()},
	})
}

func capitalReservoirCause(cause string) (execution.CapitalReservoirCause, error) {
	switch cause {
	case "HOP_ON_ALLOCATE":
		return execution.CapitalReservoirCauseHopOn, nil
	case "HOP_OFF_LIQUIDATE":
		return execution.CapitalReservoirCauseHopOff, nil
	case "SAFETY_LIQUIDATION":
		return execution.CapitalReservoirCauseSafetyLiquidation, nil
	default:
		return "", fmt.Errorf("unsupported capital reservoir cause %q", cause)
	}
}

func (pipeline *Pipeline) publishDynamic(transition *dsejehv1.PhaseTransitionState, result *dsejehv1.DynamicExecutionResult) {
	pipeline.bus.Publish(pipeline.runtimeID, &dsejehv1.RuntimeEvidenceEnvelope{Evidence: &dsejehv1.RuntimeEvidenceEnvelope_PhaseTransitionState{PhaseTransitionState: transition}})
	pipeline.bus.Publish(pipeline.runtimeID, &dsejehv1.RuntimeEvidenceEnvelope{Evidence: &dsejehv1.RuntimeEvidenceEnvelope_DynamicExecutionResult{DynamicExecutionResult: result}})
	pipeline.bus.Publish(pipeline.runtimeID, &dsejehv1.RuntimeEvidenceEnvelope{Evidence: &dsejehv1.RuntimeEvidenceEnvelope_StateActionOutcome{StateActionOutcome: result.GetStateActionOutcome()}})
}

func (pipeline *Pipeline) trace(ctx context.Context, event *dsejehv1.BarEvent, stageOrder int32, stage, emit string, input, output, capital map[string]any) error {
	_, err := pipeline.traceWithID(ctx, event, stageOrder, stage, emit, input, output, capital)
	return err
}

func (pipeline *Pipeline) traceWithID(ctx context.Context, event *dsejehv1.BarEvent, stageOrder int32, stage, emit string, input, output, capital map[string]any) (bson.ObjectID, error) {
	if pipeline.traceWriter == nil {
		return bson.NilObjectID, nil
	}
	var emitType *string
	if emit != "" {
		emitType = &emit
	}
	eventTime := time.UnixMilli(event.GetIntervalStartUnixMs())
	return pipeline.traceWriter.Write(ctx, execution.TraceRecord{
		PipelineRunID: pipeline.pipelineRunID, RunType: pipeline.runType, CollectionRunID: pipeline.collectionRunID, Symbol: event.GetEntityId(), SequenceNo: int64(event.GetProvenance().GetEntitySequence()),
		Stage: stage, StageOrder: stageOrder, EventTime: eventTime, EmitType: emitType, Input: input, Output: output, CapitalAllocation: capital,
	})
}

func (pipeline *Pipeline) capitalValues(entity *pipelineEntity, currentCapital float64) map[string]any {
	return map[string]any{
		"starting_capital": entity.account.StartingCapital, "available_capital": entity.account.AvailableCapital,
		"cash_remaining": entity.account.CashRemaining, "active_quantity": entity.account.ActiveQuantity,
		"entry_price": entity.account.EntryPrice, "peak_price": entity.account.PeakPrice, "current_capital": currentCapital,
		"realized_pnl": entity.account.CumulativeRealized, "return_pct": ((currentCapital - entity.account.StartingCapital) / entity.account.StartingCapital) * 100,
		"allocation_pct": pipeline.parameters.AllocationPct, "risk_r": pipeline.parameters.Risk,
	}
}

func (pipeline *Pipeline) entity(entityID string) *pipelineEntity {
	if existing := pipeline.entities[entityID]; existing != nil {
		return existing
	}
	account, err := execution.NewAccount(pipeline.parameters.StartingCapital)
	if err != nil {
		panic(err)
	}
	created := &pipelineEntity{account: account}
	pipeline.entities[entityID] = created
	return created
}

func stateForPhase(phase float64) dsejehv1.DynamicExecutionState {
	switch {
	case phase < 90:
		return dsejehv1.DynamicExecutionState_DYNAMIC_EXECUTION_STATE_HOLD_AND_TRAIL
	case phase < 180:
		return dsejehv1.DynamicExecutionState_DYNAMIC_EXECUTION_STATE_LIQUIDATE
	case phase < 270:
		return dsejehv1.DynamicExecutionState_DYNAMIC_EXECUTION_STATE_DISREGARD
	default:
		return dsejehv1.DynamicExecutionState_DYNAMIC_EXECUTION_STATE_ALLOCATE
	}
}

func (pipeline *Pipeline) terminalOutcome(event *dsejehv1.BarEvent, reception *dsejehv1.BarReceptionEvidence, admission *dsejehv1.BarAdmissionEvidence, analytical *dsejehv1.AnalyticalStateEvidence, phase *dsejehv1.PhaseEvidence, eligibility *dsejehv1.ProductionEligibilityEvidence, outcome dsejehv1.BarProcessingOutcomeType, reason string) {
	evidenceValue := &dsejehv1.BarProcessingOutcomeEvidence{
		EvidenceId: evidence.ID("bar-outcome", reception.GetReceptionId()), RuntimeId: pipeline.runtimeID,
		ReceptionId: reception.GetReceptionId(), BarEventId: event.GetEventId(), EntityId: event.GetEntityId(),
		EntitySequence: event.GetProvenance().GetEntitySequence(), OutcomeType: outcome, Reason: reason, ProducedUnixMs: time.Now().UnixMilli(),
	}
	if admission != nil {
		evidenceValue.AdmissionEvidenceId = admission.GetAdmissionId()
	}
	if analytical != nil {
		evidenceValue.AnalyticalStateEvidenceId = analytical.GetEvidenceId()
	}
	if phase != nil {
		evidenceValue.PhaseEvidenceId = phase.GetEvidenceId()
	}
	if eligibility != nil {
		evidenceValue.ProductionEligibilityEvidenceId = eligibility.GetEvidenceId()
	}
	pipeline.bus.Publish(pipeline.runtimeID, &dsejehv1.RuntimeEvidenceEnvelope{Evidence: &dsejehv1.RuntimeEvidenceEnvelope_BarProcessingOutcome{BarProcessingOutcome: evidenceValue}})
	pipeline.terminal.Event("BAR TERMINAL", "%s | seq=%d | %s", event.GetEntityId(), event.GetProvenance().GetEntitySequence(), outcome)
}

func (pipeline *Pipeline) fail(event *dsejehv1.BarEvent, reception *dsejehv1.BarReceptionEvidence, cause error) error {
	pipeline.state.ProcessingError()
	pipeline.terminal.Event("ERROR", "%s", cause)
	if event != nil && reception != nil {
		pipeline.terminalOutcome(event, reception, nil, nil, nil, nil, dsejehv1.BarProcessingOutcomeType_BAR_PROCESSING_OUTCOME_TYPE_PROCESSING_ERROR, cause.Error())
	}
	return fmt.Errorf("process bar: %w", cause)
}
