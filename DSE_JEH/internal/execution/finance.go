package execution

import (
	"fmt"
	"math"
)

type RunParameters struct {
	StartingCapital float64
	AllocationPct   float64
	Risk            float64
}

type Allocation struct {
	Quantity      uint64
	CashRemaining float64
}

type ExitValues struct {
	EndingCapital float64
	RealizedPnL   float64
	ReturnPct     float64
}

func (parameters RunParameters) Validate() error {
	if !positiveFinite(parameters.StartingCapital) {
		return fmt.Errorf("starting capital must be finite and positive")
	}
	if !finiteInUnitInterval(parameters.AllocationPct) {
		return fmt.Errorf("allocation percentage must be finite and in [0,1]")
	}
	if !finiteInUnitInterval(parameters.Risk) {
		return fmt.Errorf("risk must be finite and in [0,1]")
	}
	return nil
}

func SizeAllocation(availableCapital, allocationPct, entryPrice float64) (Allocation, error) {
	if !nonNegativeFinite(availableCapital) {
		return Allocation{}, fmt.Errorf("available capital must be finite and non-negative")
	}
	if !finiteInUnitInterval(allocationPct) {
		return Allocation{}, fmt.Errorf("allocation percentage must be finite and in [0,1]")
	}
	if !positiveFinite(entryPrice) {
		return Allocation{}, fmt.Errorf("entry price must be finite and positive")
	}
	quantityValue := math.Floor((availableCapital * allocationPct) / entryPrice)
	if quantityValue > math.MaxUint64 {
		return Allocation{}, fmt.Errorf("allocation quantity exceeds uint64")
	}
	quantity := uint64(quantityValue)
	return Allocation{
		Quantity:      quantity,
		CashRemaining: availableCapital - float64(quantity)*entryPrice,
	}, nil
}

func TrailingPrices(previousPeak, price, risk float64) (peak float64, stop float64, err error) {
	if !positiveFinite(price) {
		return 0, 0, fmt.Errorf("price must be finite and positive")
	}
	if previousPeak != 0 && !positiveFinite(previousPeak) {
		return 0, 0, fmt.Errorf("previous peak must be zero or finite and positive")
	}
	if !finiteInUnitInterval(risk) {
		return 0, 0, fmt.Errorf("risk must be finite and in [0,1]")
	}
	peak = math.Max(previousPeak, price)
	return peak, peak * (1 - risk), nil
}

func MarkToMarket(cashRemaining float64, activeQuantity uint64, price float64) (float64, error) {
	if !nonNegativeFinite(cashRemaining) {
		return 0, fmt.Errorf("cash remaining must be finite and non-negative")
	}
	if !positiveFinite(price) {
		return 0, fmt.Errorf("price must be finite and positive")
	}
	return cashRemaining + float64(activeQuantity)*price, nil
}

func ConfirmedExit(startingCapital, cashRemaining, entryPrice, exitPrice float64, activeQuantity uint64) (ExitValues, error) {
	if !positiveFinite(startingCapital) {
		return ExitValues{}, fmt.Errorf("starting capital must be finite and positive")
	}
	if !nonNegativeFinite(cashRemaining) {
		return ExitValues{}, fmt.Errorf("cash remaining must be finite and non-negative")
	}
	if !positiveFinite(entryPrice) || !positiveFinite(exitPrice) {
		return ExitValues{}, fmt.Errorf("entry and exit prices must be finite and positive")
	}
	endingCapital := cashRemaining + float64(activeQuantity)*exitPrice
	return ExitValues{
		EndingCapital: endingCapital,
		RealizedPnL:   (exitPrice - entryPrice) * float64(activeQuantity),
		ReturnPct:     ((endingCapital - startingCapital) / startingCapital) * 100,
	}, nil
}

func finiteInUnitInterval(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= 0 && value <= 1
}

func positiveFinite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value > 0
}

func nonNegativeFinite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= 0
}
