package analytical

import (
	"strconv"
	"sync"

	"google.golang.org/protobuf/proto"

	dsejehv1 "tramuthus/dse-jeh-transsat-1/gen/dse_jeh/v1"
	"tramuthus/dse-jeh-transsat-1/internal/evidence"
	"tramuthus/dse-jeh-transsat-1/internal/jeh"
)

const ConfigurationID = "JEH_PHASE_V0_1_MEDIAN_PRICE_LOOKBACK_63"
const ProductionEligibilityBar = jeh.Lookback + 1

type entityState struct {
	mu                  sync.Mutex
	solver              *jeh.Solver
	contiguousValidBars uint64
}

type Coordinator struct {
	mu       sync.Mutex
	entities map[string]*entityState
}

func NewCoordinator() *Coordinator {
	return &Coordinator{entities: make(map[string]*entityState)}
}

func (coordinator *Coordinator) Process(event *dsejehv1.BarEvent, admission *dsejehv1.BarAdmissionEvidence) *dsejehv1.PhaseEvidence {
	if admission.GetStatus() != dsejehv1.BarAdmissionStatus_BAR_ADMISSION_STATUS_ADMITTED {
		return nil
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
	state.mu.Unlock()

	status := dsejehv1.PhaseStatus_PHASE_STATUS_INITIALIZING
	var phaseDegrees *float64
	if err != nil {
		status = dsejehv1.PhaseStatus_PHASE_STATUS_INVALID
	} else if contiguousValidBars >= ProductionEligibilityBar && result.PhaseAngle != nil {
		status = dsejehv1.PhaseStatus_PHASE_STATUS_OBSERVABLE
		phaseDegrees = result.PhaseAngle
	}
	sequence := event.GetProvenance().GetEntitySequence()
	return &dsejehv1.PhaseEvidence{
		EvidenceId:       evidence.ID("phase", admission.GetAdmissionId(), ConfigurationID),
		BarEventId:       event.GetEventId(),
		AdmissionId:      admission.GetAdmissionId(),
		EntityId:         event.GetEntityId(),
		EntitySequence:   sequence,
		Status:           status,
		PhaseDegrees:     phaseDegrees,
		SolverIdentity:   Identity(),
		SourceProvenance: proto.Clone(event.GetProvenance()).(*dsejehv1.SourceProvenance),
	}
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
