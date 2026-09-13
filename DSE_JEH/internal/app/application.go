package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"

	dsejehv1 "tramuthus/dse-jeh-transsat-1/gen/dse_jeh/v1"
	"tramuthus/dse-jeh-transsat-1/internal/config"
	"tramuthus/dse-jeh-transsat-1/internal/evidence"
	runtimeapp "tramuthus/dse-jeh-transsat-1/internal/runtime"
	"tramuthus/dse-jeh-transsat-1/internal/services"
)

var ExternallyRegisteredServices = []string{
	"dsejeh.v1.RuntimeOperationsService",
	"dsejeh.v1.RuntimeEvidenceService",
}

var InternalServiceStatus = map[string]string{
	"BarReceptionService": "READY", "BarAdmissionService": "READY", "AnalyticalStateService": "READY",
	"JehPhaseService": "READY", "ProductionEligibilityService": "READY", "RuleRegistryService": "READY",
	"PhaseMotionService": "BLOCKED", "BoundaryCrossoverService": "BLOCKED", "StrategyRegionService": "NOT_STARTED",
	"UniverseStateService": "BLOCKED", "CandidateRankingService": "BLOCKED", "StrategyDecisionService": "BLOCKED",
	"ExecutionIntentService": "BLOCKED", "ExecutorService": "NOT_STARTED",
}

type Options struct {
	Source Source
	Output io.Writer
	Now    func() time.Time
}

type Application struct {
	cfg         config.Config
	source      Source
	state       *runtimeapp.State
	bus         *evidence.Bus
	terminal    *runtimeapp.Terminal
	server      *grpc.Server
	listener    net.Listener
	pipeline    *Pipeline
	phaseWriter *evidence.PhaseWriter
	runtimeID   string
	stopOnce    sync.Once
}

func New(cfg config.Config, options Options) (*Application, error) {
	now := options.Now
	if now == nil {
		now = time.Now
	}
	output := options.Output
	if output == nil {
		output = io.Discard
	}
	started := now()
	runtimeID := evidence.ID("runtime", started.UTC().Format(time.RFC3339Nano))
	sourceID := cfg.GRPCAddress
	endpoint := cfg.GRPCAddress
	if cfg.Mode == dsejehv1.RuntimeMode_RUNTIME_MODE_OFFLINE {
		sourceID = cfg.CollectionRunID
		endpoint = ""
	}
	sourceStatus := &dsejehv1.SourceSubscriptionEvidence{
		EvidenceId: evidence.ID("source-status", runtimeID), RuntimeId: runtimeID, RuntimeMode: cfg.Mode,
		ConnectionStatus: dsejehv1.SourceConnectionStatus_SOURCE_CONNECTION_STATUS_CONNECTING,
		SourceId:         sourceID, Endpoint: endpoint, CollectionRunId: cfg.CollectionRunID, ProducedUnixMs: started.UnixMilli(),
	}
	state := runtimeapp.NewState(runtimeID, cfg.Mode, sourceStatus, started)
	bus := evidence.NewBus()
	terminal := runtimeapp.NewTerminal(output)
	analyticalService := services.NewAnalyticalState(bus, runtimeID)
	eligibilityService, err := services.NewProductionEligibility(state, bus, runtimeID)
	if err != nil {
		return nil, err
	}
	var phaseWriter *evidence.PhaseWriter
	if cfg.OutputPath != "" {
		phaseWriter, err = evidence.NewPhaseWriter(cfg.OutputPath)
		if err != nil {
			return nil, err
		}
	}
	application := &Application{cfg: cfg, source: options.Source, state: state, bus: bus, terminal: terminal, phaseWriter: phaseWriter, runtimeID: runtimeID}
	if application.source == nil {
		application.source = productionSource(cfg)
	}
	application.pipeline = &Pipeline{
		runtimeID: runtimeID, reception: services.NewBarReception(state, bus, runtimeID),
		admission: services.NewBarAdmission(state, bus, runtimeID), analytical: analyticalService,
		phase: services.NewJehPhase(analyticalService, state, bus, runtimeID), eligibility: eligibilityService,
		state: state, bus: bus, terminal: terminal, phaseWriter: phaseWriter,
	}
	application.server = grpc.NewServer()
	dsejehv1.RegisterRuntimeOperationsServiceServer(application.server, services.NewRuntimeOperations(state))
	dsejehv1.RegisterRuntimeEvidenceServiceServer(application.server, services.NewRuntimeEvidence(bus))
	return application, nil
}

func (application *Application) Run(ctx context.Context) error {
	listener, err := net.Listen("tcp", application.cfg.ServerAddress)
	if err != nil {
		application.state.Transition(dsejehv1.RuntimeLifecycleStatus_RUNTIME_LIFECYCLE_STATUS_FAILED, err.Error())
		return fmt.Errorf("listen on %s: %w", application.cfg.ServerAddress, err)
	}
	application.listener = listener
	application.header(listener.Addr().String())
	serveErrors := make(chan error, 1)
	go func() { serveErrors <- application.server.Serve(listener) }()
	application.terminal.Event("GRPC", "listening %s", listener.Addr())
	for _, serviceName := range ExternallyRegisteredServices {
		application.terminal.Event("GRPC", "service registered: %s", serviceName)
	}
	for serviceName, serviceStatus := range InternalServiceStatus {
		application.terminal.Event("SERVICE", "internal %s | %s", serviceName, serviceStatus)
	}
	application.state.Source(sourceReadyStatus(application.cfg.Mode), "source adapter ready")
	application.state.Transition(dsejehv1.RuntimeLifecycleStatus_RUNTIME_LIFECYCLE_STATUS_RUNNING, "runtime ready")
	application.publishStatus()
	application.terminal.Event("RUNTIME", "RUNNING")

	sourceErrors := make(chan error, 1)
	go func() {
		sourceErrors <- application.source.Run(ctx, func(event *dsejehv1.BarEvent) error {
			return application.pipeline.Process(ctx, event)
		})
	}()
	heartbeat := time.NewTicker(application.cfg.StatusInterval)
	defer heartbeat.Stop()
	for {
		select {
		case <-ctx.Done():
			application.shutdown()
			return nil
		case err := <-serveErrors:
			if err != nil && !errors.Is(err, grpc.ErrServerStopped) {
				application.state.Transition(dsejehv1.RuntimeLifecycleStatus_RUNTIME_LIFECYCLE_STATUS_FAILED, err.Error())
				return fmt.Errorf("serve gRPC: %w", err)
			}
		case err := <-sourceErrors:
			if err != nil && !errors.Is(err, context.Canceled) {
				application.state.Source(dsejehv1.SourceConnectionStatus_SOURCE_CONNECTION_STATUS_FAILED, err.Error())
				application.state.Transition(dsejehv1.RuntimeLifecycleStatus_RUNTIME_LIFECYCLE_STATUS_DEGRADED, err.Error())
				application.terminal.Event("ERROR", "SOURCE | %s", err)
			} else {
				application.state.Source(dsejehv1.SourceConnectionStatus_SOURCE_CONNECTION_STATUS_COMPLETED, "bounded source completed")
				application.terminal.Event("SOURCE", "COMPLETED | bars_received=%d", application.state.Snapshot().GetActivity().GetBarsReceived())
			}
			sourceErrors = nil
			if !application.cfg.RemainRunning {
				application.shutdown()
				return err
			}
		case <-heartbeat.C:
			application.terminal.Status(application.state)
			application.publishStatus()
		}
	}
}

func (application *Application) Address() string {
	if application.listener == nil {
		return ""
	}
	return application.listener.Addr().String()
}

func (application *Application) Snapshot() *dsejehv1.RuntimeStatusEvidence {
	return application.state.Snapshot()
}

func (application *Application) Process(ctx context.Context, event *dsejehv1.BarEvent) error {
	return application.pipeline.Process(ctx, event)
}

func (application *Application) shutdown() {
	application.stopOnce.Do(func() {
		application.state.Transition(dsejehv1.RuntimeLifecycleStatus_RUNTIME_LIFECYCLE_STATUS_STOPPING, "shutdown requested")
		application.terminal.Event("RUNTIME", "STOPPING")
		application.state.Transition(dsejehv1.RuntimeLifecycleStatus_RUNTIME_LIFECYCLE_STATUS_DRAINING, "draining accepted work")
		application.terminal.Event("PIPELINE", "draining")
		application.server.GracefulStop()
		application.terminal.Event("GRPC", "stopped")
		if application.phaseWriter != nil {
			if err := application.phaseWriter.Close(); err != nil {
				application.terminal.Event("ERROR", "phase evidence close: %s", err)
			} else {
				application.terminal.Event("EVIDENCE", "phase digest=%s", application.phaseWriter.Digest())
			}
		}
		application.state.Transition(dsejehv1.RuntimeLifecycleStatus_RUNTIME_LIFECYCLE_STATUS_STOPPED, "shutdown complete")
		application.publishStatus()
		application.terminal.Event("RUNTIME", "STOPPED")
	})
}

func (application *Application) publishStatus() {
	snapshot := application.state.Snapshot()
	application.bus.Publish(application.runtimeID, &dsejehv1.RuntimeEvidenceEnvelope{Evidence: &dsejehv1.RuntimeEvidenceEnvelope_RuntimeStatus{RuntimeStatus: snapshot}})
}

func (application *Application) header(address string) {
	application.terminal.Event("RUNTIME", "DSE_JEH_TransSat_1 | id=%s | mode=%s", application.runtimeID, application.cfg.Mode)
	application.terminal.Event("CONFIG", "gRPC=%s | source=%s", address, application.cfg.GRPCAddress)
}

func sourceReadyStatus(mode dsejehv1.RuntimeMode) dsejehv1.SourceConnectionStatus {
	if mode == dsejehv1.RuntimeMode_RUNTIME_MODE_ONLINE {
		return dsejehv1.SourceConnectionStatus_SOURCE_CONNECTION_STATUS_SUBSCRIBED
	}
	return dsejehv1.SourceConnectionStatus_SOURCE_CONNECTION_STATUS_LISTENING
}
