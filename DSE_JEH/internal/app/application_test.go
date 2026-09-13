package app

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	dsejehv1 "tramuthus/dse-jeh-transsat-1/gen/dse_jeh/v1"
	"tramuthus/dse-jeh-transsat-1/internal/config"
)

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

func testBar(sequence uint64) *dsejehv1.BarEvent {
	return &dsejehv1.BarEvent{
		EventId: fmt.Sprintf("AAPL-%d", sequence), EntityId: "AAPL",
		High: 101 + float64(sequence), Low: 99 + float64(sequence),
		Provenance: &dsejehv1.SourceProvenance{EntitySequenceScope: "run|AAPL", EntitySequence: sequence},
	}
}
