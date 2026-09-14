package execution

import (
	"testing"

	"google.golang.org/protobuf/proto"

	dsejehv1 "tramuthus/dse-jeh-transsat-1/gen/dse_jeh/v1"
)

func TestAccountChangesOnlyForMatchingConfirmedFill(t *testing.T) {
	account, err := NewAccount(100_000)
	if err != nil {
		t.Fatal(err)
	}
	buy := instruction("buy-1", dsejehv1.GovernedExecutionAction_GOVERNED_EXECUTION_ACTION_ALLOCATE, "810")
	pending := executionEvent("buy-1", dsejehv1.ExecutionEventStatus_EXECUTION_EVENT_STATUS_PENDING, "810", "123.45")
	if err := account.ApplyConfirmed(buy, pending); err == nil {
		t.Fatal("pending event changed account without a confirmed fill")
	}
	if account.ActiveQuantity != 0 || account.AvailableCapital != 100_000 {
		t.Fatalf("account changed after pending event: %+v", account)
	}

	filled := executionEvent("buy-1", dsejehv1.ExecutionEventStatus_EXECUTION_EVENT_STATUS_FILLED, "810", "123.45")
	if err := account.ApplyConfirmed(buy, filled); err != nil {
		t.Fatal(err)
	}
	if account.ActiveQuantity != 810 {
		t.Fatalf("active quantity = %d, want 810", account.ActiveQuantity)
	}
	assertClose(t, account.CashRemaining, 5.5)

	sell := instruction("sell-1", dsejehv1.GovernedExecutionAction_GOVERNED_EXECUTION_ACTION_LIQUIDATE, "810")
	if err := account.ApplyConfirmed(sell, executionEvent("sell-1", dsejehv1.ExecutionEventStatus_EXECUTION_EVENT_STATUS_FILLED, "810", "130")); err != nil {
		t.Fatal(err)
	}
	if account.ActiveQuantity != 0 {
		t.Fatalf("active quantity = %d, want 0", account.ActiveQuantity)
	}
	assertClose(t, account.AvailableCapital, 105_305.5)
	assertClose(t, account.CumulativeRealized, 5_305.5)
}

func TestMockExecutorProducesAuthoritativeFilledEvent(t *testing.T) {
	instructionValue := instruction("buy-1", dsejehv1.GovernedExecutionAction_GOVERNED_EXECUTION_ACTION_ALLOCATE, "10")
	event, err := (MockExecutor{}).Fill(instructionValue, 125.25)
	if err != nil {
		t.Fatal(err)
	}
	if event.GetStatus() != dsejehv1.ExecutionEventStatus_EXECUTION_EVENT_STATUS_FILLED || event.GetGovernedExecutionInstructionId() != "buy-1" {
		t.Fatalf("execution event = %+v", event)
	}
	if event.GetFillPrice().GetValue() != "125.25" || event.GetFilledQuantity().GetValue() != "10" {
		t.Fatalf("fill values = price %q quantity %q", event.GetFillPrice().GetValue(), event.GetFilledQuantity().GetValue())
	}
	repeated, err := (MockExecutor{}).Fill(instructionValue, 125.25)
	if err != nil {
		t.Fatal(err)
	}
	if !proto.Equal(event, repeated) {
		t.Fatalf("repeated fill differs:\nfirst: %v\nsecond: %v", event, repeated)
	}
}

func instruction(id string, action dsejehv1.GovernedExecutionAction, quantity string) *dsejehv1.GovernedExecutionInstruction {
	return &dsejehv1.GovernedExecutionInstruction{
		InstructionId: id, RuntimeId: "runtime-test", EntityId: "AAPL", RequestedAction: action,
		Status:            dsejehv1.GovernedExecutionInstructionStatus_GOVERNED_EXECUTION_INSTRUCTION_STATUS_ELIGIBLE,
		RequestedQuantity: &dsejehv1.DecimalValue{Value: quantity, Unit: "SHARES"}, ConfigurationId: "config-test",
		IdempotencyKey: id, CorrelationId: "correlation-test", ProducedUnixMs: 1_789_321_600_000,
	}
}

func executionEvent(instructionID string, status dsejehv1.ExecutionEventStatus, quantity, price string) *dsejehv1.ExecutionEvent {
	return &dsejehv1.ExecutionEvent{
		Status: status, GovernedExecutionInstructionId: instructionID,
		FilledQuantity: &dsejehv1.DecimalValue{Value: quantity, Unit: "SHARES"},
		FillPrice:      &dsejehv1.DecimalValue{Value: price, Unit: "USD"},
	}
}
