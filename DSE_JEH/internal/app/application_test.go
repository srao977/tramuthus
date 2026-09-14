package app

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	dsejehv1 "tramuthus/dse-jeh-transsat-1/gen/dse_jeh/v1"
	"tramuthus/dse-jeh-transsat-1/internal/config"
	"tramuthus/dse-jeh-transsat-1/internal/evidence"
	"tramuthus/dse-jeh-transsat-1/internal/execution"
	"tramuthus/dse-jeh-transsat-1/internal/services"
)

type recordingTraceWriter struct {
	records []execution.TraceRecord
	ids     []bson.ObjectID
}

func (writer *recordingTraceWriter) Write(_ context.Context, record execution.TraceRecord) (bson.ObjectID, error) {
	writer.records = append(writer.records, record)
	insertedID := bson.NewObjectID()
	writer.ids = append(writer.ids, insertedID)
	return insertedID, nil
}

func (*recordingTraceWriter) Close(context.Context) error { return nil }

type recordingReservoirWriter struct {
	events []execution.CapitalReservoirEvent
}

func (writer *recordingReservoirWriter) Write(_ context.Context, event execution.CapitalReservoirEvent) error {
	writer.events = append(writer.events, event)
	return nil
}

func (*recordingReservoirWriter) Close(context.Context) error { return nil }

type idleSource struct {
	started chan struct{}
}

func (source *idleSource) Run(ctx context.Context, _ func(*dsejehv1.BarEvent) error) error {
	close(source.started)
	<-ctx.Done()
	return ctx.Err()
}

func TestPersistentHostObservationAndOneBarPipeline(t *testing.T) {
	source := &idleSource{started: make(chan struct{})}
	var output bytes.Buffer
	application, err := New(config.Config{
		Mode: dsejehv1.RuntimeMode_RUNTIME_MODE_ONLINE, GRPCAddress: "upstream:50051",
		ServerAddress: "127.0.0.1:0", StatusInterval: time.Hour, RemainRunning: true,
	}, Options{Source: source, Output: &output})
	if err != nil {
		t.Fatal(err)
	}
	services := application.server.GetServiceInfo()
	if len(services) != 2 || services["dsejeh.v1.RuntimeOperationsService"].Methods == nil || services["dsejeh.v1.RuntimeEvidenceService"].Methods == nil {
		t.Fatalf("externally registered services = %v", services)
	}
	for serviceName := range services {
		if serviceName != "dsejeh.v1.RuntimeOperationsService" && serviceName != "dsejeh.v1.RuntimeEvidenceService" {
			t.Fatalf("internal service exposed over gRPC: %s", serviceName)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- application.Run(ctx) }()
	select {
	case <-source.started:
	case <-time.After(2 * time.Second):
		t.Fatal("application did not start source")
	}

	connection, err := grpc.NewClient(application.Address(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	operations := dsejehv1.NewRuntimeOperationsServiceClient(connection)
	callContext, callCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer callCancel()
	initial, err := operations.GetRuntimeStatus(callContext, &dsejehv1.GetRuntimeStatusRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if initial.GetRuntimeStatus().GetLifecycleStatus() != dsejehv1.RuntimeLifecycleStatus_RUNTIME_LIFECYCLE_STATUS_RUNNING || initial.GetRuntimeStatus().GetActivity().GetBarsReceived() != 0 {
		t.Fatalf("initial runtime status = %+v", initial.GetRuntimeStatus())
	}

	stream, err := dsejehv1.NewRuntimeEvidenceServiceClient(connection).SubscribeRuntimeEvidence(callContext, &dsejehv1.SubscribeRuntimeEvidenceRequest{})
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(20 * time.Millisecond)
	event := testBar(1)
	if err := application.Process(callContext, event); err != nil {
		t.Fatal(err)
	}
	if response, err := stream.Recv(); err != nil || response.GetEvidence() == nil {
		t.Fatalf("runtime evidence response = %v, %v", response, err)
	}

	observed, err := operations.GetRuntimeStatus(callContext, &dsejehv1.GetRuntimeStatusRequest{})
	if err != nil {
		t.Fatal(err)
	}
	activity := observed.GetRuntimeStatus().GetActivity()
	if activity.GetBarsReceived() != 1 || activity.GetBarsAdmitted() != 1 || activity.GetBarsInitializing() != 1 {
		t.Fatalf("one-bar activity = %+v", activity)
	}
	localActivity := application.Snapshot().GetActivity()
	if localActivity.GetBarsReceived() != activity.GetBarsReceived() || localActivity.GetBarsAdmitted() != activity.GetBarsAdmitted() || localActivity.GetBarsInitializing() != activity.GetBarsInitializing() {
		t.Fatalf("RPC activity %+v differs from local activity %+v", activity, localActivity)
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("application did not stop gracefully")
	}
	if application.Snapshot().GetLifecycleStatus() != dsejehv1.RuntimeLifecycleStatus_RUNTIME_LIFECYCLE_STATUS_STOPPED {
		t.Fatalf("final lifecycle = %s", application.Snapshot().GetLifecycleStatus())
	}
	if !bytes.Contains(output.Bytes(), []byte("STOPPING")) || !bytes.Contains(output.Bytes(), []byte("STOPPED")) {
		t.Fatalf("shutdown output missing lifecycle transitions:\n%s", output.String())
	}
}

func TestInitialPhaseClassificationCannotProduceStateAction(t *testing.T) {
	tests := []struct {
		name         string
		priorPhase   float64
		currentPhase float64
		wantState    dsejehv1.DynamicExecutionState
	}{
		{name: "allocate quadrant", priorPhase: 260, currentPhase: 280, wantState: dsejehv1.DynamicExecutionState_DYNAMIC_EXECUTION_STATE_ALLOCATE},
		{name: "hold quadrant", priorPhase: 350, currentPhase: 10, wantState: dsejehv1.DynamicExecutionState_DYNAMIC_EXECUTION_STATE_HOLD_AND_TRAIL},
		{name: "liquidate quadrant", priorPhase: 80, currentPhase: 100, wantState: dsejehv1.DynamicExecutionState_DYNAMIC_EXECUTION_STATE_LIQUIDATE},
		{name: "disregard quadrant", priorPhase: 170, currentPhase: 190, wantState: dsejehv1.DynamicExecutionState_DYNAMIC_EXECUTION_STATE_DISREGARD},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			account, err := execution.NewAccount(100_000)
			if err != nil {
				t.Fatal(err)
			}
			traceWriter := &recordingTraceWriter{}
			pipeline := &Pipeline{
				runtimeID: "runtime-test", dynamic: services.NewDynamicExecution(), bus: evidence.NewBus(),
				traceWriter: traceWriter, parameters: execution.RunParameters{StartingCapital: 100_000, AllocationPct: 1},
				pipelineRunID: "pipeline-test", collectionRunID: "source-test",
			}
			entity := &pipelineEntity{account: account}
			current := phaseEvidence("current", 65, test.currentPhase)
			eligibility := &dsejehv1.ProductionEligibilityEvidence{
				EvidenceId: "eligible", PhaseEvidenceId: current.GetEvidenceId(),
				Outcome: dsejehv1.ProductionEligibilityOutcome_PRODUCTION_ELIGIBILITY_OUTCOME_PRODUCTION_ELIGIBLE,
			}
			event := testBar(65)
			event.Close = 100

			if err := pipeline.processDynamic(context.Background(), event, nil, current, eligibility, entity); err != nil {
				t.Fatal(err)
			}
			if !entity.initialized || entity.state != test.wantState {
				t.Fatalf("classification initialized=%v state=%s, want %s", entity.initialized, entity.state, test.wantState)
			}
			if entity.hopOnCount != 0 || entity.holdCount != 0 || entity.hopOffCount != 0 || entity.safetyLiquidationCount != 0 {
				t.Fatalf("initial classification produced action counts: %+v", entity)
			}
			if entity.account.ActiveQuantity != 0 || entity.account.AvailableCapital != 100_000 || len(traceWriter.records) != 0 {
				t.Fatalf("initial classification mutated execution state: account=%+v traces=%d", entity.account, len(traceWriter.records))
			}
		})
	}
}

func TestTraceKeepsRunMetadataAndRiskIndependent(t *testing.T) {
	traceWriter := &recordingTraceWriter{}
	pipeline := &Pipeline{
		traceWriter: traceWriter, pipelineRunID: "pipeline-a", runType: execution.RunTypeA,
		collectionRunID: "source-a", parameters: execution.RunParameters{StartingCapital: 100_000, AllocationPct: 1, Risk: 0},
	}
	event := testBar(65)
	if err := pipeline.trace(context.Background(), event, 1, "CAPITAL_RESULT", "", nil, nil, pipeline.capitalValues(&pipelineEntity{account: mustAccount(t, 100_000)}, 100_000)); err != nil {
		t.Fatal(err)
	}
	if len(traceWriter.records) != 1 {
		t.Fatalf("trace count = %d, want 1", len(traceWriter.records))
	}
	record := traceWriter.records[0]
	if record.PipelineRunID != "pipeline-a" || record.RunType != execution.RunTypeA || record.CollectionRunID != "source-a" {
		t.Fatalf("trace metadata = %+v", record)
	}
	if record.CapitalAllocation["risk_r"] != float64(0) {
		t.Fatalf("independent risk_r = %#v, want 0", record.CapitalAllocation["risk_r"])
	}
}

func TestConfirmedExecutionPublishesReservoirEventWithTraceObjectID(t *testing.T) {
	traceWriter := &recordingTraceWriter{}
	reservoirWriter := &recordingReservoirWriter{}
	reservoir, err := execution.NewCapitalReservoir("pipeline-a", "source-a", execution.RunTypeA, 30, 100_000, reservoirWriter)
	if err != nil {
		t.Fatal(err)
	}
	fixedNow := time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC)
	bus := evidence.NewBus()
	events, unsubscribe := bus.Subscribe()
	defer unsubscribe()
	pipeline := &Pipeline{
		runtimeID: "runtime-a", bus: bus, traceWriter: traceWriter,
		parameters:    execution.RunParameters{StartingCapital: 100_000, AllocationPct: 1, Risk: 0},
		pipelineRunID: "pipeline-a", runType: execution.RunTypeA, collectionRunID: "source-a",
		reservoir: reservoir, now: func() time.Time { return fixedNow },
	}
	if err := pipeline.StartReservoir(context.Background()); err != nil {
		t.Fatal(err)
	}
	account := mustAccount(t, 100_000)
	entity := &pipelineEntity{account: account}
	bar := testBar(65)
	bar.Close = 10
	result := &dsejehv1.DynamicExecutionResult{
		ResultId:           "result-a",
		StateActionOutcome: &dsejehv1.StateActionOutcome{ActionOutcomeId: "outcome-a"},
	}
	if err := pipeline.execute(context.Background(), bar, result, entity, dsejehv1.GovernedExecutionAction_GOVERNED_EXECUTION_ACTION_ALLOCATE, 10, "HOP_ON_ALLOCATE"); err != nil {
		t.Fatal(err)
	}
	if len(reservoirWriter.events) != 2 {
		t.Fatalf("reservoir events = %d, want RUN_START and OUTFLOW", len(reservoirWriter.events))
	}
	outflow := reservoirWriter.events[1]
	if outflow.StageEmitValueID == nil || len(traceWriter.ids) != 3 || *outflow.StageEmitValueID != traceWriter.ids[2] {
		t.Fatalf("stage_emit_value_id = %v, trace IDs = %v", outflow.StageEmitValueID, traceWriter.ids)
	}
	if outflow.Cause != execution.CapitalReservoirCauseHopOn || outflow.SignedFlowAmount == nil || *outflow.SignedFlowAmount != -100 {
		t.Fatalf("OUTFLOW = %+v", outflow)
	}
	foundLive := false
	for index := 0; index < 4; index++ {
		select {
		case envelope := <-events:
			live := envelope.GetCapitalReservoirEvent()
			if live != nil && live.GetEventSequence() == outflow.EventSequence {
				foundLive = live.GetStageEmitValueId() == outflow.StageEmitValueID.Hex() && live.GetCause() == dsejehv1.CapitalReservoirCause_CAPITAL_RESERVOIR_CAUSE_HOP_ON
			}
		default:
		}
	}
	if !foundLive {
		t.Fatal("matching CapitalReservoirEvent was not published on the runtime evidence bus")
	}
}

func mustAccount(t *testing.T, startingCapital float64) *execution.Account {
	t.Helper()
	account, err := execution.NewAccount(startingCapital)
	if err != nil {
		t.Fatal(err)
	}
	return account
}

func phaseEvidence(id string, sequence uint64, phase float64) *dsejehv1.PhaseEvidence {
	return &dsejehv1.PhaseEvidence{EvidenceId: id, EntityId: "AAPL", EntitySequence: sequence, PhaseDegrees: &phase}
}

func testBar(sequence uint64) *dsejehv1.BarEvent {
	return &dsejehv1.BarEvent{
		EventId: fmt.Sprintf("AAPL-%d", sequence), EntityId: "AAPL",
		High: 101 + float64(sequence), Low: 99 + float64(sequence),
		Provenance: &dsejehv1.SourceProvenance{EntitySequenceScope: "run|AAPL", EntitySequence: sequence},
	}
}
