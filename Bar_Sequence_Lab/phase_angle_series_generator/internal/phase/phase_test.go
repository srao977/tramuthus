// Package phase tests the Lab's scientific input, initialization, Hilbert,
// determinism, and stored-angle conventions without external runtime services.
package phase

import (
	"math"
	"testing"
)

func TestMedianPrice(t *testing.T) {
	got, err := MedianPrice(112.5, 107.5)
	if err != nil || got != 110 {
		t.Fatalf("MedianPrice() = %v, %v; want 110, nil", got, err)
	}
}

func TestFourBarSmoothing(t *testing.T) {
	if got := SmoothFour(4, 3, 2, 1); got != 3 {
		t.Fatalf("SmoothFour() = %v; want 3", got)
	}
}

func TestNormalizeDegrees(t *testing.T) {
	cases := map[float64]float64{-45: 315, 0: 0, 360: 0, 725: 5}
	for input, want := range cases {
		if got := NormalizeDegrees(input); got != want {
			t.Errorf("NormalizeDegrees(%v) = %v; want %v", input, got, want)
		}
	}
}

func TestInitializingHasNullPhase(t *testing.T) {
	results, err := DominantCyclePhase([]float64{1, 2, 3, 4, 5, 6})
	if err != nil {
		t.Fatal(err)
	}
	for i, result := range results {
		if result.PhaseAngle != nil || result.Observable || result.State != Initializing {
			t.Fatalf("result[%d] = %+v; want initializing with nil phase", i, result)
		}
	}
}

func TestObservableAfterTALibLookbackAndDeterministic(t *testing.T) {
	prices := make([]float64, 128)
	for i := range prices {
		prices[i] = 100 + 3*math.Sin(2*math.Pi*float64(i)/20) + 0.02*float64(i)
	}
	first, err := DominantCyclePhase(prices)
	if err != nil {
		t.Fatal(err)
	}
	second, err := DominantCyclePhase(prices)
	if err != nil {
		t.Fatal(err)
	}
	if first[Lookback-1].Observable || first[Lookback-1].PhaseAngle != nil {
		t.Fatal("bar before lookback must be initializing")
	}
	for i := Lookback; i < len(first); i++ {
		if !first[i].Observable || first[i].PhaseAngle == nil {
			t.Fatalf("result[%d] should be observable", i)
		}
		if *first[i].PhaseAngle < 0 || *first[i].PhaseAngle >= 360 {
			t.Fatalf("result[%d] angle %v outside [0,360)", i, *first[i].PhaseAngle)
		}
		if *first[i].PhaseAngle != *second[i].PhaseAngle {
			t.Fatalf("result[%d] is not deterministic", i)
		}
	}
	if first[Lookback].InPhase == first[Lookback].Quadrature {
		t.Fatal("recognized Hilbert process must produce distinct I/Q evidence")
	}
}

func TestTALib071ReferenceVector(t *testing.T) {
	prices := make([]float64, 128)
	for i := range prices {
		prices[i] = 100 + 3*math.Sin(2*math.Pi*float64(i)/20) + 0.02*float64(i)
	}
	results, err := DominantCyclePhase(prices)
	if err != nil {
		t.Fatal(err)
	}
	expected := map[int]float64{
		63:  19.32903537839626,
		64:  40.90279790358568,
		65:  52.49272351484598,
		100: NormalizeDegrees(-0.3054191102075947),
		125: 92.7725808838555,
		126: 110.74745522082993,
		127: 128.47981591897434,
	}
	const tolerance = 1e-9
	for index, want := range expected {
		if results[index].PhaseAngle == nil {
			t.Fatalf("result[%d] unexpectedly initializing", index)
		}
		if difference := math.Abs(*results[index].PhaseAngle - want); difference > tolerance {
			t.Errorf("result[%d] = %.15f; TA-Lib 0.7.1 = %.15f; error %.3g exceeds %.3g", index, *results[index].PhaseAngle, want, difference, tolerance)
		}
	}
}
