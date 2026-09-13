package runtime

import (
	"fmt"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"

	dsejehv1 "tramuthus/dse-jeh-transsat-1/gen/dse_jeh/v1"
	"tramuthus/dse-jeh-transsat-1/internal/evidence"
)

type State struct {
	mu            sync.RWMutex
	status        *dsejehv1.RuntimeStatusEvidence
	lastBar       string
	lastBarUnixMs int64
	processingErr uint64
}

func NewState(runtimeID string, mode dsejehv1.RuntimeMode, source *dsejehv1.SourceSubscriptionEvidence, started time.Time) *State {
	return &State{status: &dsejehv1.RuntimeStatusEvidence{
		EvidenceId: evidence.ID("runtime-status", runtimeID, fmt.Sprint(started.UnixMilli())),
		RuntimeIdentity: &dsejehv1.RuntimeIdentity{
			RuntimeId: runtimeID, ApplicationName: "DSE_JEH_TransSat_1", ApplicationVersion: "V0.2",
			BuildId: "local-build", InstanceId: runtimeID, StartedUnixMs: started.UnixMilli(),
		},
		RuntimeMode: mode, LifecycleStatus: dsejehv1.RuntimeLifecycleStatus_RUNTIME_LIFECYCLE_STATUS_STARTING,
		HealthStatus: dsejehv1.RuntimeHealthStatus_RUNTIME_HEALTH_STATUS_HEALTHY, ProcessLive: true,
		SourceStatus: source, Activity: &dsejehv1.RuntimeActivitySnapshot{
			SnapshotId: evidence.ID("runtime-activity", runtimeID), RuntimeId: runtimeID,
			WindowStartUnixMs: started.UnixMilli(), ProducedUnixMs: started.UnixMilli(),
		}, ProducedUnixMs: started.UnixMilli(),
	}}
}

func (state *State) Snapshot() *dsejehv1.RuntimeStatusEvidence {
	state.mu.RLock()
	defer state.mu.RUnlock()
	snapshot := proto.Clone(state.status).(*dsejehv1.RuntimeStatusEvidence)
	snapshot.ProducedUnixMs = time.Now().UnixMilli()
	snapshot.Activity.ProducedUnixMs = snapshot.ProducedUnixMs
	return snapshot
}

func (state *State) Transition(lifecycle dsejehv1.RuntimeLifecycleStatus, reason string) {
	state.mu.Lock()
	defer state.mu.Unlock()
	state.status.LifecycleStatus = lifecycle
	state.status.Reason = reason
	state.status.ProcessLive = lifecycle != dsejehv1.RuntimeLifecycleStatus_RUNTIME_LIFECYCLE_STATUS_STOPPED && lifecycle != dsejehv1.RuntimeLifecycleStatus_RUNTIME_LIFECYCLE_STATUS_FAILED
}

func (state *State) Source(status dsejehv1.SourceConnectionStatus, reason string) {
	state.mu.Lock()
	defer state.mu.Unlock()
	state.status.SourceStatus.ConnectionStatus = status
	state.status.SourceStatus.Reason = reason
	state.status.SourceStatus.ProducedUnixMs = time.Now().UnixMilli()
}

func (state *State) Receive(event *dsejehv1.BarEvent) {
	state.mu.Lock()
	defer state.mu.Unlock()
	state.status.Activity.BarsReceived++
	state.lastBar = fmt.Sprintf("%s#%d", event.GetEntityId(), event.GetProvenance().GetEntitySequence())
	state.lastBarUnixMs = time.Now().UnixMilli()
}

func (state *State) Admission(status dsejehv1.BarAdmissionStatus) {
	state.mu.Lock()
	defer state.mu.Unlock()
	if status == dsejehv1.BarAdmissionStatus_BAR_ADMISSION_STATUS_ADMITTED {
		state.status.Activity.BarsAdmitted++
	} else {
		state.status.Activity.BarsRejected++
	}
}

func (state *State) Phase(status dsejehv1.PhaseStatus) {
	state.mu.Lock()
	defer state.mu.Unlock()
	if status == dsejehv1.PhaseStatus_PHASE_STATUS_INITIALIZING {
		state.status.Activity.BarsInitializing++
	}
}

func (state *State) Eligibility(outcome dsejehv1.ProductionEligibilityOutcome) {
	state.mu.Lock()
	defer state.mu.Unlock()
	if outcome == dsejehv1.ProductionEligibilityOutcome_PRODUCTION_ELIGIBILITY_OUTCOME_PRODUCTION_ELIGIBLE {
		state.status.Activity.BarsPhaseEligible++
	}
}

func (state *State) ProcessingError() {
	state.mu.Lock()
	defer state.mu.Unlock()
	state.processingErr++
}

func (state *State) LastBar() (string, int64) {
	state.mu.RLock()
	defer state.mu.RUnlock()
	return state.lastBar, state.lastBarUnixMs
}

func (state *State) ProcessingErrors() uint64 {
	state.mu.RLock()
	defer state.mu.RUnlock()
	return state.processingErr
}
