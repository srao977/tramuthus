package jeh

import (
	"math"
	"testing"
)

func TestIncrementalTALibReferenceVector(t *testing.T) {
	expected := map[int]float64{
		63:  19.32903537839626,
		64:  40.90279790358568,
		65:  52.49272351484598,
		100: NormalizeDegrees(-0.3054191102075947),
		125: 92.7725808838555,
		126: 110.74745522082993,
		127: 128.47981591897434,
	}
	solver := NewSolver()
	for index := range 128 {
		price := 100 + 3*math.Sin(2*math.Pi*float64(index)/20) + 0.02*float64(index)
		result, err := solver.Update(price, price)
		if err != nil {
			t.Fatal(err)
		}
		if index < Lookback && result.PhaseAngle != nil {
			t.Fatalf("result[%d] must be initializing", index)
		}
		if want, ok := expected[index]; ok {
			if result.PhaseAngle == nil {
				t.Fatalf("result[%d] unexpectedly initializing", index)
			}
			if difference := math.Abs(*result.PhaseAngle - want); difference > 1e-9 {
				t.Errorf("result[%d] = %.15f; want %.15f; difference %.3g", index, *result.PhaseAngle, want, difference)
			}
		}
	}
}

func TestNormalizeDegreesPreservesZero(t *testing.T) {
	for input, want := range map[float64]float64{-45: 315, 0: 0, 360: 0, 725: 5} {
		if got := NormalizeDegrees(input); got != want {
			t.Errorf("NormalizeDegrees(%v) = %v; want %v", input, got, want)
		}
	}
}
