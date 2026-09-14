package execution

import (
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	dsejehv1 "tramuthus/dse-jeh-transsat-1/gen/dse_jeh/v1"
)

type recordingReservoirWriter struct {
	events []CapitalReservoirEvent
}

func (writer *recordingReservoirWriter) Write(_ context.Context, event CapitalReservoirEvent) error {
	writer.events = append(writer.events, event)
	return nil
}

func (*recordingReservoirWriter) Close(context.Context) error { return nil }

func TestCapitalReservoirConfirmedFlowAndContinuity(t *testing.T) {
	writer := &recordingReservoirWriter{}
	reservoir, err := NewCapitalReservoir("pipeline", "source", RunTypeA, 30, 100_000, writer)
	if err != nil {
		t.Fatal(err)
	}
	when := time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC)
	start, err := reservoir.Start(context.Background(), when)
	if err != nil {
		t.Fatal(err)
	}
	if start.InitialReservoir == nil || *start.InitialReservoir != 3_000_000 || start.ReservoirBefore != 3_000_000 || start.ReservoirAfter != 3_000_000 {
		t.Fatalf("RUN_START = %+v", start)
	}

	stageID := bson.NewObjectID()
	buy := reservoirInstruction("buy", dsejehv1.GovernedExecutionAction_GOVERNED_EXECUTION_ACTION_ALLOCATE, "9985")
	buyEvent := reservoirExecutionEvent("buy", "9985", "10", when)
	outflow, err := reservoir.RecordExecution(context.Background(), buy, buyEvent, CapitalReservoirCauseHopOn, stageID, when, when)
	if err != nil {
		t.Fatal(err)
	}
	if outflow.SignedFlowAmount == nil || *outflow.SignedFlowAmount != -99_850 || outflow.ReservoirAfter != 2_900_150 || outflow.Cause != CapitalReservoirCauseHopOn || outflow.StageEmitValueID == nil || *outflow.StageEmitValueID != stageID {
		t.Fatalf("OUTFLOW = %+v", outflow)
	}

	sell := reservoirInstruction("sell", dsejehv1.GovernedExecutionAction_GOVERNED_EXECUTION_ACTION_LIQUIDATE, "9985")
	sellEvent := reservoirExecutionEvent("sell", "9985", "10.13520280420631", when.Add(time.Minute))
	inflow, err := reservoir.RecordExecution(context.Background(), sell, sellEvent, CapitalReservoirCauseHopOff, bson.NewObjectID(), when.Add(time.Minute), when.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if inflow.SignedFlowAmount == nil || *inflow.SignedFlowAmount != 101_200 || inflow.ReservoirAfter != 3_001_350 || inflow.Cause != CapitalReservoirCauseHopOff {
		t.Fatalf("INFLOW = %+v", inflow)
	}
	if start.ReservoirAfter != outflow.ReservoirBefore || outflow.ReservoirAfter != inflow.ReservoirBefore {
		t.Fatalf("reservoir continuity failed: %+v", writer.events)
	}
	for index, event := range writer.events {
		if event.EventSequence != uint64(index+1) {
			t.Fatalf("event sequence at %d = %d", index, event.EventSequence)
		}
	}
}

func TestCapitalReservoirRunEndDoesNotLiquidateActivePosition(t *testing.T) {
	writer := &recordingReservoirWriter{}
	reservoir, err := NewCapitalReservoir("pipeline", "source", RunTypeB, 30, 100_000, writer)
	if err != nil {
		t.Fatal(err)
	}
	when := time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC)
	if _, err := reservoir.Start(context.Background(), when); err != nil {
		t.Fatal(err)
	}
	buy := reservoirInstruction("buy", dsejehv1.GovernedExecutionAction_GOVERNED_EXECUTION_ACTION_ALLOCATE, "100")
	if _, err := reservoir.RecordExecution(context.Background(), buy, reservoirExecutionEvent("buy", "100", "10", when), CapitalReservoirCauseHopOn, bson.NewObjectID(), when, when); err != nil {
		t.Fatal(err)
	}
	reservoir.Mark("AAPL", 12)
	end, err := reservoir.End(context.Background(), when.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if end.EventType != CapitalReservoirEventRunEnd || end.ActivePositions == nil || *end.ActivePositions != 1 || end.ReservoirCash == nil || *end.ReservoirCash != 2_999_000 || end.TotalDeployedAfter == nil || *end.TotalDeployedAfter != 1_200 || end.TotalMarkedCapitalAfter == nil || *end.TotalMarkedCapitalAfter != 3_000_200 || end.UnrealizedPnL == nil || *end.UnrealizedPnL != 200 || end.TotalPnL == nil || *end.TotalPnL != 200 {
		t.Fatalf("RUN_END = %+v", end)
	}
	if len(writer.events) != 3 || writer.events[2].Cause != "" || writer.events[2].StageEmitValueID != nil {
		t.Fatalf("RUN_END synthesized liquidation: %+v", writer.events)
	}
}

func TestCapitalReservoirSafetyCauseRemainsDistinct(t *testing.T) {
	reservoir, err := NewCapitalReservoir("pipeline", "source", RunTypeA, 30, 100_000, nil)
	if err != nil {
		t.Fatal(err)
	}
	when := time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC)
	_, _ = reservoir.Start(context.Background(), when)
	buy := reservoirInstruction("buy", dsejehv1.GovernedExecutionAction_GOVERNED_EXECUTION_ACTION_ALLOCATE, "1")
	_, _ = reservoir.RecordExecution(context.Background(), buy, reservoirExecutionEvent("buy", "1", "10", when), CapitalReservoirCauseHopOn, bson.NewObjectID(), when, when)
	sell := reservoirInstruction("sell", dsejehv1.GovernedExecutionAction_GOVERNED_EXECUTION_ACTION_LIQUIDATE, "1")
	event, err := reservoir.RecordExecution(context.Background(), sell, reservoirExecutionEvent("sell", "1", "10", when), CapitalReservoirCauseSafetyLiquidation, bson.NewObjectID(), when, when)
	if err != nil {
		t.Fatal(err)
	}
	if event.Cause != CapitalReservoirCauseSafetyLiquidation {
		t.Fatalf("cause = %s", event.Cause)
	}
}

func TestCapitalReservoirBSONReferenceAndBoundaryNulls(t *testing.T) {
	when := time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC)
	reservoir, err := NewCapitalReservoir("pipeline", "source", RunTypeA, 30, 100_000, nil)
	if err != nil {
		t.Fatal(err)
	}
	start, err := reservoir.Start(context.Background(), when)
	if err != nil {
		t.Fatal(err)
	}
	startBytes, err := bson.Marshal(start)
	if err != nil {
		t.Fatal(err)
	}
	var startDocument bson.M
	if err := bson.Unmarshal(startBytes, &startDocument); err != nil {
		t.Fatal(err)
	}
	if value, exists := startDocument["stage_emit_value_id"]; !exists || value != nil {
		t.Fatalf("RUN_START stage_emit_value_id = %#v, exists=%v; want explicit null", value, exists)
	}

	stageID := bson.NewObjectID()
	buy := reservoirInstruction("buy", dsejehv1.GovernedExecutionAction_GOVERNED_EXECUTION_ACTION_ALLOCATE, "1")
	outflow, err := reservoir.RecordExecution(context.Background(), buy, reservoirExecutionEvent("buy", "1", "10", when), CapitalReservoirCauseHopOn, stageID, when, when)
	if err != nil {
		t.Fatal(err)
	}
	outflowBytes, err := bson.Marshal(outflow)
	if err != nil {
		t.Fatal(err)
	}
	var outflowDocument bson.M
	if err := bson.Unmarshal(outflowBytes, &outflowDocument); err != nil {
		t.Fatal(err)
	}
	if value, ok := outflowDocument["stage_emit_value_id"].(bson.ObjectID); !ok || value != stageID {
		t.Fatalf("OUTFLOW stage_emit_value_id = %#v, want ObjectId %s", outflowDocument["stage_emit_value_id"], stageID.Hex())
	}
}

func TestCapitalReservoirAllowsRepeatedGenericRunMetadata(t *testing.T) {
	for _, pipelineRunID := range []string{"DPE-GOVERNED-RUN-C-20260914-001", "DPE-GOVERNED-RUN-C-20260914-002"} {
		reservoir, err := NewCapitalReservoir(pipelineRunID, "20260911T161623Z-1", "RUN_C", 30, 100_000, nil)
		if err != nil {
			t.Fatalf("NewCapitalReservoir(%q): %v", pipelineRunID, err)
		}
		event, err := reservoir.Start(context.Background(), time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC))
		if err != nil {
			t.Fatalf("Start(%q): %v", pipelineRunID, err)
		}
		if event.PipelineRunID != pipelineRunID || event.CollectionRunID != "20260911T161623Z-1" || event.RunType != "RUN_C" {
			t.Fatalf("independent run metadata = %+v", event)
		}
	}
}

func reservoirInstruction(id string, action dsejehv1.GovernedExecutionAction, quantity string) *dsejehv1.GovernedExecutionInstruction {
	return &dsejehv1.GovernedExecutionInstruction{
		InstructionId: id, EntityId: "AAPL", RequestedAction: action,
		RequestedQuantity: &dsejehv1.DecimalValue{Value: quantity, Unit: "SHARES"},
	}
}

func reservoirExecutionEvent(instructionID, quantity, price string, when time.Time) *dsejehv1.ExecutionEvent {
	return &dsejehv1.ExecutionEvent{
		EntityId: "AAPL", Status: dsejehv1.ExecutionEventStatus_EXECUTION_EVENT_STATUS_FILLED,
		GovernedExecutionInstructionId: instructionID,
		FilledQuantity:                 &dsejehv1.DecimalValue{Value: quantity, Unit: "SHARES"},
		FillPrice:                      &dsejehv1.DecimalValue{Value: price, Unit: "USD"},
		ExecutionUnixMs:                when.UnixMilli(),
	}
}
