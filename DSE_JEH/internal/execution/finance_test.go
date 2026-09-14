package execution

import (
	"math"
	"testing"
)

func TestSizeAllocationUsesWholeSharesAndPreservesResidualCash(t *testing.T) {
	allocation, err := SizeAllocation(100_000, 1, 123.45)
	if err != nil {
		t.Fatal(err)
	}
	if allocation.Quantity != 810 {
		t.Fatalf("quantity = %d, want 810", allocation.Quantity)
	}
	assertClose(t, allocation.CashRemaining, 5.5)
}

func TestTrailingPricesRiskEndpoints(t *testing.T) {
	tests := []struct {
		name     string
		risk     float64
		wantStop float64
	}{
		{name: "no retreat", risk: 0, wantStop: 110},
		{name: "non binding", risk: 1, wantStop: 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			peak, stop, err := TrailingPrices(100, 110, test.risk)
			if err != nil {
				t.Fatal(err)
			}
			assertClose(t, peak, 110)
			assertClose(t, stop, test.wantStop)
		})
	}
}

func TestMarkToMarketAndConfirmedExit(t *testing.T) {
	currentCapital, err := MarkToMarket(5.5, 810, 130)
	if err != nil {
		t.Fatal(err)
	}
	assertClose(t, currentCapital, 105_305.5)

	exit, err := ConfirmedExit(100_000, 5.5, 123.45, 130, 810)
	if err != nil {
		t.Fatal(err)
	}
	assertClose(t, exit.EndingCapital, 105_305.5)
	assertClose(t, exit.RealizedPnL, 5_305.5)
	assertClose(t, exit.ReturnPct, 5.3055)
}

func TestRunParametersRejectOutOfRangeRisk(t *testing.T) {
	parameters := RunParameters{StartingCapital: 100_000, AllocationPct: 1, Risk: 1.01}
	if err := parameters.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want out-of-range risk error")
	}
}

func assertClose(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("got %.12f, want %.12f", got, want)
	}
}
