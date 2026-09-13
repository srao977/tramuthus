package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	dsejehv1 "tramuthus/dse-jeh-transsat-1/gen/dse_jeh/v1"
	"tramuthus/dse-jeh-transsat-1/internal/admission"
	"tramuthus/dse-jeh-transsat-1/internal/analytical"
	"tramuthus/dse-jeh-transsat-1/internal/evidence"
	runtimeapp "tramuthus/dse-jeh-transsat-1/internal/runtime"
)

type RuntimeOperations struct {
	dsejehv1.UnimplementedRuntimeOperationsServiceServer
	state *runtimeapp.State
}

func NewRuntimeOperations(state *runtimeapp.State) *RuntimeOperations {
	return &RuntimeOperations{state: state}
}

func (service *RuntimeOperations) GetRuntimeStatus(context.Context, *dsejehv1.GetRuntimeStatusRequest) (*dsejehv1.GetRuntimeStatusResponse, error) {
	return &dsejehv1.GetRuntimeStatusResponse{RuntimeStatus: service.state.Snapshot()}, nil
}

type RuntimeEvidence struct {
	dsejehv1.UnimplementedRuntimeEvidenceServiceServer
	bus *evidence.Bus
}

func NewRuntimeEvidence(bus *evidence.Bus) *RuntimeEvidence {
	return &RuntimeEvidence{bus: bus}
}

func (service *RuntimeEvidence) SubscribeRuntimeEvidence(_ *dsejehv1.SubscribeRuntimeEvidenceRequest, stream grpc.ServerStreamingServer[dsejehv1.SubscribeRuntimeEvidenceResponse]) error {
	channel, unsubscribe := service.bus.Subscribe()
	defer unsubscribe()
	for {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		case envelope := <-channel:
			if err := stream.Send(&dsejehv1.SubscribeRuntimeEvidenceResponse{Evidence: envelope}); err != nil {
				return err
			}
		}
	}
}

type BarReception struct {
	dsejehv1.UnimplementedBarReceptionServiceServer
	state     *runtimeapp.State
	bus       *evidence.Bus
	runtimeID string
}

func NewBarReception(state *runtimeapp.State, bus *evidence.Bus, runtimeID string) *BarReception {
	return &BarReception{state: state, bus: bus, runtimeID: runtimeID}
}

func (service *BarReception) ReceiveBar(_ context.Context, request *dsejehv1.ReceiveBarRequest) (*dsejehv1.ReceiveBarResponse, error) {
	event := request.GetBarEvent()
	if event == nil || event.GetEventId() == "" || event.GetEntityId() == "" || event.GetProvenance() == nil {
		return nil, status.Error(codes.InvalidArgument, "bar_event with event_id, entity_id, and provenance is required")
	}
	reception := &dsejehv1.BarReceptionEvidence{
		ReceptionId: evidence.ID("reception", service.runtimeID, event.GetEventId()), RuntimeId: service.runtimeID,
		BarEventId: event.GetEventId(), EntityId: event.GetEntityId(),
		SourceProvenance: proto.Clone(event.GetProvenance()).(*dsejehv1.SourceProvenance), ReceivedUnixMs: time.Now().UnixMilli(),
	}
	service.state.Receive(event)
	service.bus.Publish(service.runtimeID, &dsejehv1.RuntimeEvidenceEnvelope{Evidence: &dsejehv1.RuntimeEvidenceEnvelope_BarReception{BarReception: reception}})
	return &dsejehv1.ReceiveBarResponse{ReceptionEvidence: reception}, nil
}

type BarAdmission struct {
	dsejehv1.UnimplementedBarAdmissionServiceServer
	admitter  *admission.Admitter
	state     *runtimeapp.State
	bus       *evidence.Bus
	runtimeID string
}

func NewBarAdmission(state *runtimeapp.State, bus *evidence.Bus, runtimeID string) *BarAdmission {
	return &BarAdmission{admitter: admission.New(), state: state, bus: bus, runtimeID: runtimeID}
}

func (service *BarAdmission) AdmitBar(_ context.Context, request *dsejehv1.AdmitBarRequest) (*dsejehv1.AdmitBarResponse, error) {
	if request.GetBarEvent() == nil || request.GetReceptionEvidence() == nil {
		return nil, status.Error(codes.InvalidArgument, "bar_event and reception_evidence are required")
	}
	result := service.admitter.Admit(request.GetBarEvent())
	result.ReceptionEvidenceId = request.GetReceptionEvidence().GetReceptionId()
	result.ProducedUnixMs = time.Now().UnixMilli()
	result.Findings = admissionFindings(result.GetStatus())
	service.state.Admission(result.GetStatus())
	service.bus.Publish(service.runtimeID, &dsejehv1.RuntimeEvidenceEnvelope{Evidence: &dsejehv1.RuntimeEvidenceEnvelope_BarAdmission{BarAdmission: result}})
	return &dsejehv1.AdmitBarResponse{AdmissionEvidence: result}, nil
}

func admissionFindings(value dsejehv1.BarAdmissionStatus) []dsejehv1.BarAdmissionFinding {
	switch value {
	case dsejehv1.BarAdmissionStatus_BAR_ADMISSION_STATUS_DUPLICATE:
		return []dsejehv1.BarAdmissionFinding{dsejehv1.BarAdmissionFinding_BAR_ADMISSION_FINDING_DUPLICATE}
	case dsejehv1.BarAdmissionStatus_BAR_ADMISSION_STATUS_CONFLICT:
		return []dsejehv1.BarAdmissionFinding{dsejehv1.BarAdmissionFinding_BAR_ADMISSION_FINDING_CONFLICT}
	case dsejehv1.BarAdmissionStatus_BAR_ADMISSION_STATUS_GAP:
		return []dsejehv1.BarAdmissionFinding{dsejehv1.BarAdmissionFinding_BAR_ADMISSION_FINDING_MISSING_PREDECESSOR}
	case dsejehv1.BarAdmissionStatus_BAR_ADMISSION_STATUS_OUT_OF_ORDER:
		return []dsejehv1.BarAdmissionFinding{dsejehv1.BarAdmissionFinding_BAR_ADMISSION_FINDING_OUT_OF_ORDER}
	case dsejehv1.BarAdmissionStatus_BAR_ADMISSION_STATUS_INVALID:
		return []dsejehv1.BarAdmissionFinding{dsejehv1.BarAdmissionFinding_BAR_ADMISSION_FINDING_MALFORMED}
	default:
		return nil
	}
}

type AnalyticalState struct {
	dsejehv1.UnimplementedAnalyticalStateServiceServer
	coordinator *analytical.Coordinator
	bus         *evidence.Bus
	runtimeID   string
	mu          sync.Mutex
	phases      map[string]*dsejehv1.PhaseEvidence
}

func NewAnalyticalState(bus *evidence.Bus, runtimeID string) *AnalyticalState {
	return &AnalyticalState{coordinator: analytical.NewCoordinator(), bus: bus, runtimeID: runtimeID, phases: make(map[string]*dsejehv1.PhaseEvidence)}
}

func (service *AnalyticalState) UpdateAnalyticalState(_ context.Context, request *dsejehv1.UpdateAnalyticalStateRequest) (*dsejehv1.UpdateAnalyticalStateResponse, error) {
	if request.GetAdmissionEvidence().GetStatus() != dsejehv1.BarAdmissionStatus_BAR_ADMISSION_STATUS_ADMITTED {
		return nil, status.Error(codes.FailedPrecondition, "only an admitted bar may advance analytical state")
	}
	analyticalEvidence, phaseEvidence := service.coordinator.ProcessDetailed(service.runtimeID, request.GetBarEvent(), request.GetAdmissionEvidence())
	if analyticalEvidence == nil || phaseEvidence == nil {
		return nil, status.Error(codes.Internal, "analytical coordinator produced no evidence")
	}
	service.mu.Lock()
	service.phases[analyticalEvidence.GetEvidenceId()] = phaseEvidence
	service.mu.Unlock()
	service.bus.Publish(service.runtimeID, &dsejehv1.RuntimeEvidenceEnvelope{Evidence: &dsejehv1.RuntimeEvidenceEnvelope_AnalyticalState{AnalyticalState: analyticalEvidence}})
	return &dsejehv1.UpdateAnalyticalStateResponse{AnalyticalStateEvidence: analyticalEvidence}, nil
}

func (service *AnalyticalState) phase(analyticalEvidenceID string) *dsejehv1.PhaseEvidence {
	service.mu.Lock()
	defer service.mu.Unlock()
	phaseEvidence := service.phases[analyticalEvidenceID]
	delete(service.phases, analyticalEvidenceID)
	return phaseEvidence
}

type JehPhase struct {
	dsejehv1.UnimplementedJehPhaseServiceServer
	analytical *AnalyticalState
	state      *runtimeapp.State
	bus        *evidence.Bus
	runtimeID  string
}

func NewJehPhase(analyticalService *AnalyticalState, state *runtimeapp.State, bus *evidence.Bus, runtimeID string) *JehPhase {
	return &JehPhase{analytical: analyticalService, state: state, bus: bus, runtimeID: runtimeID}
}

func (service *JehPhase) UpdateJehPhase(_ context.Context, request *dsejehv1.UpdateJehPhaseRequest) (*dsejehv1.UpdateJehPhaseResponse, error) {
	phaseEvidence := service.analytical.phase(request.GetAnalyticalStateEvidence().GetEvidenceId())
	if phaseEvidence == nil {
		return nil, status.Error(codes.FailedPrecondition, "analytical state has no pending JEH result")
	}
	service.state.Phase(phaseEvidence.GetStatus())
	service.bus.Publish(service.runtimeID, &dsejehv1.RuntimeEvidenceEnvelope{Evidence: &dsejehv1.RuntimeEvidenceEnvelope_Phase{Phase: phaseEvidence}})
	return &dsejehv1.UpdateJehPhaseResponse{PhaseEvidence: phaseEvidence}, nil
}

type ProductionEligibility struct {
	dsejehv1.UnimplementedProductionEligibilityServiceServer
	program   *vm.Program
	rule      *dsejehv1.RuleIdentity
	ruleSet   *dsejehv1.RuleSetIdentity
	state     *runtimeapp.State
	bus       *evidence.Bus
	runtimeID string
}

func NewProductionEligibility(state *runtimeapp.State, bus *evidence.Bus, runtimeID string) (*ProductionEligibility, error) {
	program, err := expr.Compile("phase_observable && phase_value_present && contiguous_valid_bar_count >= 64 && sequence_valid && current_bar_valid && prior_analytical_state_valid", expr.AsBool())
	if err != nil {
		return nil, fmt.Errorf("compile Rule #1: %w", err)
	}
	return &ProductionEligibility{
		program: program, state: state, bus: bus, runtimeID: runtimeID,
		rule:    &dsejehv1.RuleIdentity{RuleId: "RULE_1_63_CONTIGUOUS_BAR_ELIGIBILITY", RuleVersion: "V0.1", RuleName: "63-Contiguous-Bar Eligibility", Purpose: "gate DEP-05 after the current phase and 63 causal predecessors are valid", OwningComponent: dsejehv1.RuleOwningComponent_RULE_OWNING_COMPONENT_PRODUCTION_ELIGIBILITY, ExpectedResultType: dsejehv1.RuleRawResultType_RULE_RAW_RESULT_TYPE_BOOLEAN},
		ruleSet: &dsejehv1.RuleSetIdentity{RuleSetId: "DSE_JEH_RULE_SET", RuleSetVersion: "V0.1", ConfigurationId: analytical.ConfigurationID},
	}, nil
}

func (service *ProductionEligibility) EvaluateProductionEligibility(_ context.Context, request *dsejehv1.EvaluateProductionEligibilityRequest) (*dsejehv1.EvaluateProductionEligibilityResponse, error) {
	phaseEvidence := request.GetPhaseEvidence()
	analyticalEvidence := request.GetAnalyticalStateEvidence()
	if phaseEvidence == nil || analyticalEvidence == nil {
		return nil, status.Error(codes.InvalidArgument, "phase_evidence and analytical_state_evidence are required")
	}
	contextValue := request.GetContext()
	if contextValue == nil {
		contextValue = &dsejehv1.ProductionEligibilityContext{
			ContextId: evidence.ID("eligibility-context", phaseEvidence.GetEvidenceId()), PhaseEvidenceId: phaseEvidence.GetEvidenceId(),
			AnalyticalStateEvidenceId: analyticalEvidence.GetEvidenceId(), PhaseStatus: phaseEvidence.GetStatus(),
			PhaseValuePresent: phaseEvidence.PhaseDegrees != nil, ContiguousValidBarCount: analyticalEvidence.GetContiguousValidBarCount(),
			SequenceIntegrity: analyticalEvidence.GetSequenceIntegrity(), CurrentBarValid: phaseEvidence.GetStatus() != dsejehv1.PhaseStatus_PHASE_STATUS_INVALID,
			PriorAnalyticalStateValid: analyticalEvidence.GetContiguousValidBarCount() == 1 || analyticalEvidence.GetPriorStateEvidenceId() != "",
		}
	}
	environment := map[string]any{
		"phase_observable":    contextValue.GetPhaseStatus() == dsejehv1.PhaseStatus_PHASE_STATUS_OBSERVABLE,
		"phase_value_present": contextValue.GetPhaseValuePresent(), "contiguous_valid_bar_count": contextValue.GetContiguousValidBarCount(),
		"sequence_valid":    contextValue.GetSequenceIntegrity() == dsejehv1.SequenceIntegrityStatus_SEQUENCE_INTEGRITY_STATUS_VALID,
		"current_bar_valid": contextValue.GetCurrentBarValid(), "prior_analytical_state_valid": contextValue.GetPriorAnalyticalStateValid(),
	}
	raw, err := expr.Run(service.program, environment)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "evaluate Rule #1: %v", err)
	}
	passed := raw.(bool)
	outcome := eligibilityOutcome(contextValue, passed)
	produced := time.Now().UnixMilli()
	ruleEvidence := &dsejehv1.RuleEvaluationEvidence{
		EvidenceId: evidence.ID("rule-evaluation", service.rule.GetRuleId(), phaseEvidence.GetEvidenceId()), RuntimeId: service.runtimeID,
		RuleIdentity: proto.Clone(service.rule).(*dsejehv1.RuleIdentity), RuleSetIdentity: proto.Clone(service.ruleSet).(*dsejehv1.RuleSetIdentity),
		CausalInputIdentity: phaseEvidence.GetEvidenceId(), EvaluationSequence: phaseEvidence.GetEntitySequence(),
		EvaluationStatus: evaluationStatus(passed), RawResult: &dsejehv1.RuleEvaluationEvidence_RawBooleanResult{RawBooleanResult: passed},
		TypedOutcome: &dsejehv1.RuleEvaluationEvidence_ProductionEligibilityOutcome{ProductionEligibilityOutcome: &dsejehv1.ProductionEligibilityRuleOutcome{Outcome: outcome}}, ProducedUnixMs: produced,
	}
	eligibilityEvidence := &dsejehv1.ProductionEligibilityEvidence{
		EvidenceId: evidence.ID("production-eligibility", phaseEvidence.GetEvidenceId()), RuntimeId: service.runtimeID,
		EntityId: phaseEvidence.GetEntityId(), EntitySequence: phaseEvidence.GetEntitySequence(), PhaseEvidenceId: phaseEvidence.GetEvidenceId(),
		ContextId: contextValue.GetContextId(), Outcome: outcome, ContiguousValidBarCount: contextValue.GetContiguousValidBarCount(),
		SequenceIntegrity: contextValue.GetSequenceIntegrity(), RuleEvaluationEvidenceId: ruleEvidence.GetEvidenceId(),
		AnalyticalStateEvidenceId: analyticalEvidence.GetEvidenceId(), ProducedUnixMs: produced,
	}
	service.state.Eligibility(outcome)
	service.bus.Publish(service.runtimeID, &dsejehv1.RuntimeEvidenceEnvelope{Evidence: &dsejehv1.RuntimeEvidenceEnvelope_RuleEvaluation{RuleEvaluation: ruleEvidence}})
	service.bus.Publish(service.runtimeID, &dsejehv1.RuntimeEvidenceEnvelope{Evidence: &dsejehv1.RuntimeEvidenceEnvelope_ProductionEligibility{ProductionEligibility: eligibilityEvidence}})
	return &dsejehv1.EvaluateProductionEligibilityResponse{ProductionEligibilityEvidence: eligibilityEvidence, RuleEvaluationEvidence: ruleEvidence}, nil
}

func eligibilityOutcome(contextValue *dsejehv1.ProductionEligibilityContext, passed bool) dsejehv1.ProductionEligibilityOutcome {
	if passed {
		return dsejehv1.ProductionEligibilityOutcome_PRODUCTION_ELIGIBILITY_OUTCOME_PRODUCTION_ELIGIBLE
	}
	if contextValue.GetSequenceIntegrity() != dsejehv1.SequenceIntegrityStatus_SEQUENCE_INTEGRITY_STATUS_VALID {
		return dsejehv1.ProductionEligibilityOutcome_PRODUCTION_ELIGIBILITY_OUTCOME_BLOCKED_CONTINUITY
	}
	if contextValue.GetPhaseStatus() == dsejehv1.PhaseStatus_PHASE_STATUS_INVALID {
		return dsejehv1.ProductionEligibilityOutcome_PRODUCTION_ELIGIBILITY_OUTCOME_BLOCKED_INVALID_PHASE
	}
	return dsejehv1.ProductionEligibilityOutcome_PRODUCTION_ELIGIBILITY_OUTCOME_INITIALIZING
}

func evaluationStatus(passed bool) dsejehv1.RuleEvaluationStatus {
	if passed {
		return dsejehv1.RuleEvaluationStatus_RULE_EVALUATION_STATUS_PASSED
	}
	return dsejehv1.RuleEvaluationStatus_RULE_EVALUATION_STATUS_BLOCKED
}

func (service *ProductionEligibility) RuleSetDefinition() *dsejehv1.RuleSetDefinition {
	return &dsejehv1.RuleSetDefinition{RuleSetIdentity: proto.Clone(service.ruleSet).(*dsejehv1.RuleSetIdentity), Rules: []*dsejehv1.RuleDefinition{{RuleIdentity: proto.Clone(service.rule).(*dsejehv1.RuleIdentity), RuleSetIdentity: proto.Clone(service.ruleSet).(*dsejehv1.RuleSetIdentity), ExprExpression: "phase_observable && phase_value_present && contiguous_valid_bar_count >= 64 && sequence_valid && current_bar_valid && prior_analytical_state_valid"}}}
}

type RuleRegistry struct {
	dsejehv1.UnimplementedRuleRegistryServiceServer
	active *dsejehv1.RuleSetDefinition
}

func NewRuleRegistry(active *dsejehv1.RuleSetDefinition) *RuleRegistry {
	return &RuleRegistry{active: active}
}

func (service *RuleRegistry) ValidateRuleSet(_ context.Context, request *dsejehv1.ValidateRuleSetRequest) (*dsejehv1.ValidateRuleSetResponse, error) {
	definition := request.GetRuleSetDefinition()
	if definition == nil || definition.GetRuleSetIdentity().GetRuleSetId() == "" {
		return &dsejehv1.ValidateRuleSetResponse{Valid: false}, nil
	}
	for _, rule := range definition.GetRules() {
		if rule.GetRuleIdentity().GetRuleId() == "" || rule.GetExprExpression() == "" {
			return &dsejehv1.ValidateRuleSetResponse{Valid: false, RuleSetIdentity: definition.GetRuleSetIdentity()}, nil
		}
		if _, err := expr.Compile(rule.GetExprExpression()); err != nil {
			return &dsejehv1.ValidateRuleSetResponse{Valid: false, RuleSetIdentity: definition.GetRuleSetIdentity()}, nil
		}
	}
	return &dsejehv1.ValidateRuleSetResponse{Valid: true, RuleSetIdentity: definition.GetRuleSetIdentity()}, nil
}

func (service *RuleRegistry) GetActiveRuleSet(context.Context, *dsejehv1.GetActiveRuleSetRequest) (*dsejehv1.GetActiveRuleSetResponse, error) {
	return &dsejehv1.GetActiveRuleSetResponse{RuleSetDefinition: proto.Clone(service.active).(*dsejehv1.RuleSetDefinition)}, nil
}
