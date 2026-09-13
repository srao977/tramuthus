package analytical

import (
	"strconv"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"

	dsejehv1 "tramuthus/dse-jeh-transsat-1/gen/dse_jeh/v1"
	"tramuthus/dse-jeh-transsat-1/internal/evidence"
	"tramuthus/dse-jeh-transsat-1/internal/jeh"
)

const ConfigurationID = "JEH_PHASE_V0_1_MEDIAN_PRICE_LOOKBACK_63"
const ProductionEligibilityBar = jeh.Lookback + 1

type entityState struct {
	mu                       sync.Mutex
	solver                   *jeh.Solver
	contiguousValidBars      uint64
	lastAnalyticalEvidenceID string
}

type Coordinator struct {
	mu       sync.Mutex
	entities map[string]*entityState
}

func NewCoordinator() *Coordinator {
	return &Coordinator{entities: make(map[string]*entityState)}
}

func (coordinator *Coordinator) Process(event *dsejehv1.BarEvent, admission *dsejehv1.BarAdmissionEvidence) *dsejehv1.PhaseEvidence {
	_, phaseEvidence := coordinator.ProcessDetailed("", event, admission)
	return phaseEvidence
}

func (coordinator *Coordinator) ProcessDetailed(runtimeID string, event *dsejehv1.BarEvent, admission *dsejehv1.BarAdmissionEvidence) (*dsejehv1.AnalyticalStateEvidence, *dsejehv1.PhaseEvidence) {
	if admission.GetStatus() != dsejehv1.BarAdmissionStatus_BAR_ADMISSION_STATUS_ADMITTED {
		return nil, nil
	}
	scope := event.GetProvenance().GetEntitySequenceScope()
	coordinator.mu.Lock()
	state := coordinator.entities[scope]
	if state == nil {
		state = &entityState{solver: jeh.NewSolver()}
		coordinator.entities[scope] = state
	}
	coordinator.mu.Unlock()

	state.mu.Lock()
	result, err := state.solver.Update(event.GetHigh(), event.GetLow())
	if err == nil {
		state.contiguousValidBars++
	}
	contiguousValidBars := state.contiguousValidBars
	priorAnalyticalEvidenceID := state.lastAnalyticalEvidenceID
	sequence := event.GetProvenance().GetEntitySequence()
	analyticalEvidence := &dsejehv1.AnalyticalStateEvidence{
		EvidenceId:              evidence.ID("analytical-state", admission.GetAdmissionId(), ConfigurationID),
		RuntimeId:               runtimeID,
		EntityId:                event.GetEntityId(),
		EntitySequence:          sequence,
		ContiguousValidBarCount: contiguousValidBars,
		SequenceIntegrity:       dsejehv1.SequenceIntegrityStatus_SEQUENCE_INTEGRITY_STATUS_VALID,
		PriorStateEvidenceId:    priorAnalyticalEvidenceID,
		AdmissionId:             admission.GetAdmissionId(),
		SolverConfigurationId:   ConfigurationID,
		ProducedUnixMs:          time.Now().UnixMilli(),
	}
	state.lastAnalyticalEvidenceID = analyticalEvidence.GetEvidenceId()
	state.mu.Unlock()

	status := dsejehv1.PhaseStatus_PHASE_STATUS_INITIALIZING
	var phaseDegrees *float64
	if err != nil {
		status = dsejehv1.PhaseStatus_PHASE_STATUS_INVALID
	} else if contiguousValidBars >= ProductionEligibilityBar && result.PhaseAngle != nil {
		status = dsejehv1.PhaseStatus_PHASE_STATUS_OBSERVABLE
		phaseDegrees = result.PhaseAngle
	}
	phaseEvidence := &dsejehv1.PhaseEvidence{
		EvidenceId:                evidence.ID("phase", admission.GetAdmissionId(), ConfigurationID),
		BarEventId:                event.GetEventId(),
		AdmissionId:               admission.GetAdmissionId(),
		EntityId:                  event.GetEntityId(),
		EntitySequence:            sequence,
		Status:                    status,
		PhaseDegrees:              phaseDegrees,
		SolverIdentity:            Identity(),
		SourceProvenance:          proto.Clone(event.GetProvenance()).(*dsejehv1.SourceProvenance),
		AnalyticalStateEvidenceId: analyticalEvidence.GetEvidenceId(),
		ProducedUnixMs:            time.Now().UnixMilli(),
	}
	return analyticalEvidence, phaseEvidence
}

func Identity() *dsejehv1.SolverIdentity {
	return &dsejehv1.SolverIdentity{
		SolverName:                 jeh.SolverName,
		SolverVersion:              jeh.SolverVersion,
		AlgorithmReference:         jeh.Algorithm,
		InputSeriesType:            jeh.InputSeriesType,
		InitializationObservations: jeh.Lookback,
		ConfigurationId:            ConfigurationID,
		ImplementationId:           "DSE_JEH_TRANSSAT_1_PHASE_1",
	}
}

func Scope(runID, entityID string) string {
	return runID + "|" + entityID
}

func SequenceString(sequence uint64) string {
	return strconv.FormatUint(sequence, 10)
}
