package execution

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	dsejehv1 "tramuthus/dse-jeh-transsat-1/gen/dse_jeh/v1"
	"tramuthus/dse-jeh-transsat-1/internal/evidence"
)

const GovernedReservoirSymbolCount int32 = 30

type CapitalReservoirEventType string

const (
	CapitalReservoirEventRunStart CapitalReservoirEventType = "RUN_START"
	CapitalReservoirEventOutflow  CapitalReservoirEventType = "OUTFLOW"
	CapitalReservoirEventInflow   CapitalReservoirEventType = "INFLOW"
	CapitalReservoirEventRunEnd   CapitalReservoirEventType = "RUN_END"
)

type CapitalReservoirCause string

const (
	CapitalReservoirCauseHopOn             CapitalReservoirCause = "HOP_ON"
	CapitalReservoirCauseHopOff            CapitalReservoirCause = "HOP_OFF"
	CapitalReservoirCauseSafetyLiquidation CapitalReservoirCause = "SAFETY_LIQUIDATION"
)

type CapitalReservoirEvent struct {
	PipelineRunID           string                    `bson:"pipeline_run_id"`
	CollectionRunID         string                    `bson:"collection_run_id"`
	RunType                 RunType                   `bson:"run_type"`
	EventSequence           uint64                    `bson:"event_sequence"`
	Symbol                  string                    `bson:"symbol,omitempty"`
	EventType               CapitalReservoirEventType `bson:"event_type"`
	Cause                   CapitalReservoirCause     `bson:"cause,omitempty"`
	StageEmitValueID        *bson.ObjectID            `bson:"stage_emit_value_id"`
	TriggerEventTime        *time.Time                `bson:"trigger_event_time"`
	ExecutionEventTime      *time.Time                `bson:"execution_event_time"`
	ProcessedAt             time.Time                 `bson:"processed_at"`
	Quantity                *uint64                   `bson:"quantity,omitempty"`
	ExecutionPrice          *float64                  `bson:"execution_price,omitempty"`
	SignedFlowAmount        *float64                  `bson:"signed_flow_amount,omitempty"`
	InitialReservoir        *float64                  `bson:"initial_reservoir,omitempty"`
	SymbolCount             *int32                    `bson:"symbol_count,omitempty"`
	InitialSymbolCapital    *float64                  `bson:"initial_symbol_capital,omitempty"`
	ReservoirBefore         float64                   `bson:"reservoir_before"`
	ReservoirAfter          float64                   `bson:"reservoir_after"`
	ActiveQuantityAfter     *uint64                   `bson:"active_quantity_after,omitempty"`
	ReservoirCash           *float64                  `bson:"reservoir_cash,omitempty"`
	TotalDeployedAfter      *float64                  `bson:"total_deployed_after,omitempty"`
	TotalMarkedCapitalAfter *float64                  `bson:"total_marked_capital_after,omitempty"`
	RealizedPnL             *float64                  `bson:"realized_pnl,omitempty"`
	UnrealizedPnL           *float64                  `bson:"unrealized_pnl,omitempty"`
	TotalPnL                *float64                  `bson:"total_pnl,omitempty"`
	ActivePositions         *int32                    `bson:"active_positions,omitempty"`
}

type CapitalReservoirWriter interface {
	Write(context.Context, CapitalReservoirEvent) error
	Close(context.Context) error
}

type reservoirPosition struct {
	quantity   uint64
	entryPrice float64
	markPrice  float64
}

type CapitalReservoir struct {
	mu                   sync.Mutex
	pipelineRunID        string
	collectionRunID      string
	runType              RunType
	symbolCount          int32
	initialSymbolCapital float64
	initialReservoir     float64
	cash                 float64
	realizedPnL          float64
	sequence             uint64
	started              bool
	ended                bool
	positions            map[string]reservoirPosition
	writer               CapitalReservoirWriter
}

func NewCapitalReservoir(pipelineRunID, collectionRunID string, runType RunType, symbolCount int32, initialSymbolCapital float64, writer CapitalReservoirWriter) (*CapitalReservoir, error) {
	if pipelineRunID == "" || collectionRunID == "" {
		return nil, fmt.Errorf("pipeline and collection run IDs are required")
	}
	if err := ValidateRunType(runType); err != nil {
		return nil, err
	}
	if symbolCount <= 0 || !positiveFinite(initialSymbolCapital) {
		return nil, fmt.Errorf("positive symbol count and initial symbol capital are required")
	}
	initialReservoir := float64(symbolCount) * initialSymbolCapital
	return &CapitalReservoir{
		pipelineRunID: pipelineRunID, collectionRunID: collectionRunID, runType: runType,
		symbolCount: symbolCount, initialSymbolCapital: initialSymbolCapital,
		initialReservoir: initialReservoir, cash: initialReservoir,
		positions: make(map[string]reservoirPosition), writer: writer,
	}, nil
}

func (reservoir *CapitalReservoir) Start(ctx context.Context, processedAt time.Time) (CapitalReservoirEvent, error) {
	reservoir.mu.Lock()
	defer reservoir.mu.Unlock()
	if reservoir.started {
		return CapitalReservoirEvent{}, fmt.Errorf("capital reservoir already started")
	}
	event := reservoir.baseEvent(CapitalReservoirEventRunStart, processedAt)
	event.InitialReservoir = &reservoir.initialReservoir
	event.SymbolCount = &reservoir.symbolCount
	event.InitialSymbolCapital = &reservoir.initialSymbolCapital
	if err := reservoir.persist(ctx, event); err != nil {
		return CapitalReservoirEvent{}, err
	}
	reservoir.started = true
	reservoir.sequence = event.EventSequence
	return event, nil
}

func (reservoir *CapitalReservoir) RecordExecution(ctx context.Context, instruction *dsejehv1.GovernedExecutionInstruction, executionEvent *dsejehv1.ExecutionEvent, cause CapitalReservoirCause, stageEmitValueID bson.ObjectID, triggerEventTime, processedAt time.Time) (CapitalReservoirEvent, error) {
	reservoir.mu.Lock()
	defer reservoir.mu.Unlock()
	if !reservoir.started || reservoir.ended {
		return CapitalReservoirEvent{}, fmt.Errorf("capital reservoir is not active")
	}
	if stageEmitValueID.IsZero() {
		return CapitalReservoirEvent{}, fmt.Errorf("stage_emit_value_id is required")
	}
	if instruction == nil || executionEvent == nil || executionEvent.GetStatus() != dsejehv1.ExecutionEventStatus_EXECUTION_EVENT_STATUS_FILLED {
		return CapitalReservoirEvent{}, fmt.Errorf("confirmed execution event is required")
	}
	if executionEvent.GetGovernedExecutionInstructionId() != instruction.GetInstructionId() {
		return CapitalReservoirEvent{}, fmt.Errorf("execution event does not match instruction")
	}
	quantity, err := decimalQuantity(executionEvent.GetFilledQuantity())
	if err != nil {
		return CapitalReservoirEvent{}, err
	}
	price, err := decimalFloat(executionEvent.GetFillPrice(), "fill price")
	if err != nil || !positiveFinite(price) {
		return CapitalReservoirEvent{}, fmt.Errorf("fill price must be finite and positive")
	}
	flow := float64(quantity) * price
	eventType := CapitalReservoirEventInflow
	switch instruction.GetRequestedAction() {
	case dsejehv1.GovernedExecutionAction_GOVERNED_EXECUTION_ACTION_ALLOCATE:
		if cause != CapitalReservoirCauseHopOn {
			return CapitalReservoirEvent{}, fmt.Errorf("confirmed allocation requires HOP_ON cause")
		}
		if flow > reservoir.cash {
			return CapitalReservoirEvent{}, fmt.Errorf("confirmed allocation exceeds reservoir cash")
		}
		eventType = CapitalReservoirEventOutflow
		flow = -flow
	case dsejehv1.GovernedExecutionAction_GOVERNED_EXECUTION_ACTION_LIQUIDATE:
		if cause != CapitalReservoirCauseHopOff && cause != CapitalReservoirCauseSafetyLiquidation {
			return CapitalReservoirEvent{}, fmt.Errorf("confirmed liquidation requires HOP_OFF or SAFETY_LIQUIDATION cause")
		}
	default:
		return CapitalReservoirEvent{}, fmt.Errorf("unsupported confirmed action %s", instruction.GetRequestedAction())
	}
	executionTime := time.UnixMilli(executionEvent.GetExecutionUnixMs()).UTC()
	event := reservoir.baseEvent(eventType, processedAt)
	event.Symbol = executionEvent.GetEntityId()
	event.Cause = cause
	event.StageEmitValueID = &stageEmitValueID
	event.TriggerEventTime = timePointer(triggerEventTime)
	event.ExecutionEventTime = &executionTime
	event.Quantity = &quantity
	event.ExecutionPrice = &price
	event.SignedFlowAmount = &flow
	event.ReservoirAfter = reservoir.cash + flow
	event.ActiveQuantityAfter = &quantity
	if eventType == CapitalReservoirEventInflow {
		zero := uint64(0)
		event.ActiveQuantityAfter = &zero
	}
	if err := reservoir.persist(ctx, event); err != nil {
		return CapitalReservoirEvent{}, err
	}
	reservoir.cash = event.ReservoirAfter
	reservoir.sequence = event.EventSequence
	if eventType == CapitalReservoirEventOutflow {
		reservoir.positions[event.Symbol] = reservoirPosition{quantity: quantity, entryPrice: price, markPrice: price}
	} else {
		position := reservoir.positions[event.Symbol]
		reservoir.realizedPnL += float64(quantity) * (price - position.entryPrice)
		delete(reservoir.positions, event.Symbol)
	}
	return event, nil
}

func (reservoir *CapitalReservoir) Mark(symbol string, price float64) {
	reservoir.mu.Lock()
	defer reservoir.mu.Unlock()
	position, ok := reservoir.positions[symbol]
	if !ok || !positiveFinite(price) {
		return
	}
	position.markPrice = price
	reservoir.positions[symbol] = position
}

func (reservoir *CapitalReservoir) End(ctx context.Context, processedAt time.Time) (CapitalReservoirEvent, error) {
	reservoir.mu.Lock()
	defer reservoir.mu.Unlock()
	if !reservoir.started || reservoir.ended {
		return CapitalReservoirEvent{}, fmt.Errorf("capital reservoir is not active")
	}
	var deployed, unrealized float64
	for _, position := range reservoir.positions {
		deployed += float64(position.quantity) * position.markPrice
		unrealized += float64(position.quantity) * (position.markPrice - position.entryPrice)
	}
	event := reservoir.baseEvent(CapitalReservoirEventRunEnd, processedAt)
	totalMarked := reservoir.cash + deployed
	totalPnL := totalMarked - reservoir.initialReservoir
	activePositions := int32(len(reservoir.positions))
	event.ReservoirCash = &reservoir.cash
	event.TotalDeployedAfter = &deployed
	event.TotalMarkedCapitalAfter = &totalMarked
	event.RealizedPnL = &reservoir.realizedPnL
	event.UnrealizedPnL = &unrealized
	event.TotalPnL = &totalPnL
	event.ActivePositions = &activePositions
	if err := reservoir.persist(ctx, event); err != nil {
		return CapitalReservoirEvent{}, err
	}
	reservoir.sequence = event.EventSequence
	reservoir.ended = true
	return event, nil
}

func (reservoir *CapitalReservoir) AvailableCash() float64 {
	reservoir.mu.Lock()
	defer reservoir.mu.Unlock()
	return reservoir.cash
}

func (reservoir *CapitalReservoir) AllocationCapacity(symbolCeiling float64) float64 {
	reservoir.mu.Lock()
	defer reservoir.mu.Unlock()
	return minCapital(symbolCeiling, reservoir.cash)
}

func (event CapitalReservoirEvent) Proto() *dsejehv1.CapitalReservoirEvent {
	result := &dsejehv1.CapitalReservoirEvent{
		EventId:       evidence.ID("capital-reservoir-event", event.PipelineRunID, fmt.Sprint(event.EventSequence)),
		PipelineRunId: event.PipelineRunID, CollectionRunId: event.CollectionRunID, RunType: string(event.RunType),
		EventSequence: event.EventSequence, Symbol: event.Symbol, EventType: protoReservoirEventType(event.EventType),
		Cause: protoReservoirCause(event.Cause), ProcessedUnixMs: event.ProcessedAt.UnixMilli(),
		ReservoirBefore: event.ReservoirBefore, ReservoirAfter: event.ReservoirAfter,
	}
	if event.StageEmitValueID != nil {
		result.StageEmitValueId = event.StageEmitValueID.Hex()
	}
	if event.TriggerEventTime != nil {
		result.TriggerEventUnixMs = event.TriggerEventTime.UnixMilli()
	}
	if event.ExecutionEventTime != nil {
		result.ExecutionEventUnixMs = event.ExecutionEventTime.UnixMilli()
	}
	if event.Quantity != nil {
		result.Quantity = *event.Quantity
	}
	if event.ExecutionPrice != nil {
		result.ExecutionPrice = *event.ExecutionPrice
	}
	if event.SignedFlowAmount != nil {
		result.SignedFlowAmount = *event.SignedFlowAmount
	}
	if event.InitialReservoir != nil {
		result.InitialReservoir = *event.InitialReservoir
	}
	if event.SymbolCount != nil {
		result.SymbolCount = uint32(*event.SymbolCount)
	}
	if event.InitialSymbolCapital != nil {
		result.InitialSymbolCapital = *event.InitialSymbolCapital
	}
	if event.ActiveQuantityAfter != nil {
		result.ActiveQuantityAfter = *event.ActiveQuantityAfter
	}
	if event.ReservoirCash != nil {
		result.ReservoirCash = *event.ReservoirCash
	}
	if event.TotalDeployedAfter != nil {
		result.TotalDeployedAfter = *event.TotalDeployedAfter
	}
	if event.TotalMarkedCapitalAfter != nil {
		result.TotalMarkedCapitalAfter = *event.TotalMarkedCapitalAfter
	}
	if event.RealizedPnL != nil {
		result.RealizedPnl = *event.RealizedPnL
	}
	if event.UnrealizedPnL != nil {
		result.UnrealizedPnl = *event.UnrealizedPnL
	}
	if event.TotalPnL != nil {
		result.TotalPnl = *event.TotalPnL
	}
	if event.ActivePositions != nil {
		result.ActivePositions = uint32(*event.ActivePositions)
	}
	return result
}

func (reservoir *CapitalReservoir) baseEvent(eventType CapitalReservoirEventType, processedAt time.Time) CapitalReservoirEvent {
	return CapitalReservoirEvent{
		PipelineRunID: reservoir.pipelineRunID, CollectionRunID: reservoir.collectionRunID, RunType: reservoir.runType,
		EventSequence: reservoir.sequence + 1, EventType: eventType, ProcessedAt: processedAt.UTC(),
		ReservoirBefore: reservoir.cash, ReservoirAfter: reservoir.cash,
	}
}

func (reservoir *CapitalReservoir) persist(ctx context.Context, event CapitalReservoirEvent) error {
	if err := event.Validate(); err != nil {
		return err
	}
	if reservoir.writer == nil {
		return nil
	}
	if err := reservoir.writer.Write(ctx, event); err != nil {
		return fmt.Errorf("persist capital reservoir event: %w", err)
	}
	return nil
}

func (event CapitalReservoirEvent) Validate() error {
	if event.PipelineRunID == "" || event.CollectionRunID == "" || event.EventSequence == 0 || event.ProcessedAt.IsZero() {
		return fmt.Errorf("capital reservoir common envelope is incomplete")
	}
	if err := ValidateRunType(event.RunType); err != nil {
		return fmt.Errorf("capital reservoir run_type: %w", err)
	}
	if !isFinite(event.ReservoirBefore) || !isFinite(event.ReservoirAfter) {
		return fmt.Errorf("capital reservoir balances must be finite")
	}
	switch event.EventType {
	case CapitalReservoirEventRunStart:
		if event.InitialReservoir == nil || event.SymbolCount == nil || event.InitialSymbolCapital == nil || event.StageEmitValueID != nil || event.ReservoirBefore != event.ReservoirAfter {
			return fmt.Errorf("RUN_START fields are invalid")
		}
	case CapitalReservoirEventOutflow, CapitalReservoirEventInflow:
		if event.Symbol == "" || event.Cause == "" || event.StageEmitValueID == nil || event.StageEmitValueID.IsZero() || event.TriggerEventTime == nil || event.ExecutionEventTime == nil || event.Quantity == nil || event.ExecutionPrice == nil || event.SignedFlowAmount == nil || event.ActiveQuantityAfter == nil {
			return fmt.Errorf("execution-derived capital reservoir fields are incomplete")
		}
		if event.EventType == CapitalReservoirEventOutflow && *event.SignedFlowAmount >= 0 {
			return fmt.Errorf("OUTFLOW signed_flow_amount must be negative")
		}
		if event.EventType == CapitalReservoirEventInflow && *event.SignedFlowAmount <= 0 {
			return fmt.Errorf("INFLOW signed_flow_amount must be positive")
		}
	case CapitalReservoirEventRunEnd:
		if event.ReservoirCash == nil || event.TotalDeployedAfter == nil || event.TotalMarkedCapitalAfter == nil || event.RealizedPnL == nil || event.UnrealizedPnL == nil || event.TotalPnL == nil || event.ActivePositions == nil || event.StageEmitValueID != nil || event.ReservoirBefore != event.ReservoirAfter {
			return fmt.Errorf("RUN_END fields are invalid")
		}
	default:
		return fmt.Errorf("unsupported capital reservoir event type %q", event.EventType)
	}
	return nil
}

func timePointer(value time.Time) *time.Time {
	utc := value.UTC()
	return &utc
}

func minCapital(left, right float64) float64 {
	return math.Min(left, right)
}

func isFinite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func protoReservoirEventType(eventType CapitalReservoirEventType) dsejehv1.CapitalReservoirEventType {
	switch eventType {
	case CapitalReservoirEventRunStart:
		return dsejehv1.CapitalReservoirEventType_CAPITAL_RESERVOIR_EVENT_TYPE_RUN_START
	case CapitalReservoirEventOutflow:
		return dsejehv1.CapitalReservoirEventType_CAPITAL_RESERVOIR_EVENT_TYPE_OUTFLOW
	case CapitalReservoirEventInflow:
		return dsejehv1.CapitalReservoirEventType_CAPITAL_RESERVOIR_EVENT_TYPE_INFLOW
	case CapitalReservoirEventRunEnd:
		return dsejehv1.CapitalReservoirEventType_CAPITAL_RESERVOIR_EVENT_TYPE_RUN_END
	default:
		return dsejehv1.CapitalReservoirEventType_CAPITAL_RESERVOIR_EVENT_TYPE_UNSPECIFIED
	}
}

func protoReservoirCause(cause CapitalReservoirCause) dsejehv1.CapitalReservoirCause {
	switch cause {
	case CapitalReservoirCauseHopOn:
		return dsejehv1.CapitalReservoirCause_CAPITAL_RESERVOIR_CAUSE_HOP_ON
	case CapitalReservoirCauseHopOff:
		return dsejehv1.CapitalReservoirCause_CAPITAL_RESERVOIR_CAUSE_HOP_OFF
	case CapitalReservoirCauseSafetyLiquidation:
		return dsejehv1.CapitalReservoirCause_CAPITAL_RESERVOIR_CAUSE_SAFETY_LIQUIDATION
	default:
		return dsejehv1.CapitalReservoirCause_CAPITAL_RESERVOIR_CAUSE_UNSPECIFIED
	}
}
