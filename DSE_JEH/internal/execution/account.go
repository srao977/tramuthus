package execution

import (
	"fmt"
	"strconv"

	dsejehv1 "tramuthus/dse-jeh-transsat-1/gen/dse_jeh/v1"
)

type Account struct {
	StartingCapital    float64
	AvailableCapital   float64
	CashRemaining      float64
	ActiveQuantity     uint64
	EntryPrice         float64
	PeakPrice          float64
	LastPrice          float64
	CumulativeRealized float64
}

func NewAccount(startingCapital float64) (*Account, error) {
	if !positiveFinite(startingCapital) {
		return nil, fmt.Errorf("starting capital must be finite and positive")
	}
	return &Account{StartingCapital: startingCapital, AvailableCapital: startingCapital, CashRemaining: startingCapital}, nil
}

func (account *Account) ApplyConfirmed(instruction *dsejehv1.GovernedExecutionInstruction, event *dsejehv1.ExecutionEvent) error {
	if instruction == nil || event == nil {
		return fmt.Errorf("instruction and execution event are required")
	}
	if event.GetStatus() != dsejehv1.ExecutionEventStatus_EXECUTION_EVENT_STATUS_FILLED {
		return fmt.Errorf("execution event must be FILLED")
	}
	if event.GetGovernedExecutionInstructionId() != instruction.GetInstructionId() {
		return fmt.Errorf("execution event does not match instruction")
	}
	quantity, err := decimalQuantity(event.GetFilledQuantity())
	if err != nil {
		return err
	}
	price, err := decimalFloat(event.GetFillPrice(), "fill price")
	if err != nil || !positiveFinite(price) {
		return fmt.Errorf("fill price must be finite and positive")
	}

	switch instruction.GetRequestedAction() {
	case dsejehv1.GovernedExecutionAction_GOVERNED_EXECUTION_ACTION_ALLOCATE:
		if account.ActiveQuantity != 0 {
			return fmt.Errorf("cannot allocate while a position is active")
		}
		cost := float64(quantity) * price
		if cost > account.StartingCapital {
			return fmt.Errorf("confirmed allocation cost exceeds governed symbol ceiling")
		}
		account.AvailableCapital = account.StartingCapital
		account.ActiveQuantity = quantity
		account.EntryPrice = price
		account.PeakPrice = price
		account.CashRemaining = account.AvailableCapital - cost
	case dsejehv1.GovernedExecutionAction_GOVERNED_EXECUTION_ACTION_LIQUIDATE:
		if quantity == 0 || quantity != account.ActiveQuantity {
			return fmt.Errorf("confirmed liquidation must fill the complete active quantity")
		}
		exit, exitErr := ConfirmedExit(account.StartingCapital, account.CashRemaining, account.EntryPrice, price, quantity)
		if exitErr != nil {
			return exitErr
		}
		account.AvailableCapital = exit.EndingCapital
		account.CashRemaining = exit.EndingCapital
		account.CumulativeRealized += exit.RealizedPnL
		account.ActiveQuantity = 0
		account.EntryPrice = 0
		account.PeakPrice = 0
	default:
		return fmt.Errorf("unsupported confirmed action %s", instruction.GetRequestedAction())
	}
	account.LastPrice = price
	return nil
}

func (account *Account) CurrentCapital(price float64) (float64, error) {
	if account.ActiveQuantity == 0 {
		return account.AvailableCapital, nil
	}
	return MarkToMarket(account.CashRemaining, account.ActiveQuantity, price)
}

func decimalQuantity(value *dsejehv1.DecimalValue) (uint64, error) {
	if value == nil || value.GetValue() == "" {
		return 0, fmt.Errorf("filled quantity is required")
	}
	quantity, err := strconv.ParseUint(value.GetValue(), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse filled quantity: %w", err)
	}
	return quantity, nil
}

func decimalFloat(value *dsejehv1.DecimalValue, name string) (float64, error) {
	if value == nil || value.GetValue() == "" {
		return 0, fmt.Errorf("%s is required", name)
	}
	parsed, err := strconv.ParseFloat(value.GetValue(), 64)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", name, err)
	}
	return parsed, nil
}
