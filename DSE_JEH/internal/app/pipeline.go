package app

import (
	"context"
	"fmt"
	"time"

	dsejehv1 "tramuthus/dse-jeh-transsat-1/gen/dse_jeh/v1"
	"tramuthus/dse-jeh-transsat-1/internal/evidence"
	runtimeapp "tramuthus/dse-jeh-transsat-1/internal/runtime"
	"tramuthus/dse-jeh-transsat-1/internal/services"
)

type Pipeline struct {
	runtimeID   string
	reception   *services.BarReception
	admission   *services.BarAdmission
	analytical  *services.AnalyticalState
	phase       *services.JehPhase
	eligibility *services.ProductionEligibility
	state       *runtimeapp.State
	bus         *evidence.Bus
	terminal    *runtimeapp.Terminal
	phaseWriter *evidence.PhaseWriter
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
		pipeline.terminal.Event("PIPELINE", "%s | seq=%d | production eligible | DEP-05 awaiting approved motion rule", event.GetEntityId(), sequence)
		pipeline.terminalOutcome(event, receptionResponse.GetReceptionEvidence(), admissionEvidence, analyticalEvidence, phaseEvidence, eligibilityEvidence, dsejehv1.BarProcessingOutcomeType_BAR_PROCESSING_OUTCOME_TYPE_PHASE_MOTION_UNAVAILABLE, "DEP-05 awaiting approved motion mathematics and bar64/bar65 decision")
	} else if eligibilityEvidence.GetOutcome() == dsejehv1.ProductionEligibilityOutcome_PRODUCTION_ELIGIBILITY_OUTCOME_INITIALIZING {
		pipeline.terminalOutcome(event, receptionResponse.GetReceptionEvidence(), admissionEvidence, analyticalEvidence, phaseEvidence, eligibilityEvidence, dsejehv1.BarProcessingOutcomeType_BAR_PROCESSING_OUTCOME_TYPE_INITIALIZING, "production eligibility initialization")
	} else {
		pipeline.terminalOutcome(event, receptionResponse.GetReceptionEvidence(), admissionEvidence, analyticalEvidence, phaseEvidence, eligibilityEvidence, dsejehv1.BarProcessingOutcomeType_BAR_PROCESSING_OUTCOME_TYPE_PRODUCTION_ELIGIBILITY_BLOCKED, eligibilityEvidence.GetReason())
	}
	return nil
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
