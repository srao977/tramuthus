package jeh

import (
	"fmt"
	"math"
)

const (
	SolverName      = "EHLERS_DOMINANT_CYCLE_PHASE"
	SolverVersion   = "V0.1"
	Algorithm       = "TA_LIB_HT_DCPHASE"
	InputSeriesType = "MEDIAN_PRICE"
	Lookback        = 63
)

type Result struct {
	PhaseAngle *float64
	InPhase    float64
	Quadrature float64
}

type Solver struct {
	count int
	seed  []float64

	periodWMASum           float64
	periodWMASub           float64
	trailingWMAValue       float64
	trailingWMAQueue       [3]float64
	detrenderOdd           [3]float64
	detrenderEven          [3]float64
	q1Odd                  [3]float64
	q1Even                 [3]float64
	jIOdd                  [3]float64
	jIEven                 [3]float64
	jQOdd                  [3]float64
	jQEven                 [3]float64
	prevDetrenderOdd       float64
	prevDetrenderEven      float64
	prevDetrenderInputOdd  float64
	prevDetrenderInputEven float64
	prevQ1Odd              float64
	prevQ1Even             float64
	prevQ1InputOdd         float64
	prevQ1InputEven        float64
	prevJIOdd              float64
	prevJIEven             float64
	prevJIInputOdd         float64
	prevJIInputEven        float64
	prevJQOdd              float64
	prevJQEven             float64
	prevJQInputOdd         float64
	prevJQInputEven        float64
	period                 float64
	smoothPeriod           float64
	previousQ2             float64
	previousI2             float64
	re                     float64
	im                     float64
	i1OddPrev2             float64
	i1OddPrev3             float64
	i1EvenPrev2            float64
	i1EvenPrev3            float64
	smoothPrice            [50]float64
	hilbertIndex           int
	smoothPriceIndex       int
}

func NewSolver() *Solver {
	return &Solver{seed: make([]float64, 0, 37)}
}

func MedianPrice(high, low float64) (float64, error) {
	if !finite(high) || !finite(low) {
		return 0, fmt.Errorf("high and low must be finite")
	}
	return (high + low) / 2, nil
}

func NormalizeDegrees(angle float64) float64 {
	angle = math.Mod(angle, 360)
	if angle < 0 {
		angle += 360
	}
	if angle == 360 {
		return 0
	}
	return angle
}

func (solver *Solver) Update(high, low float64) (Result, error) {
	price, err := MedianPrice(high, low)
	if err != nil {
		return Result{}, err
	}
	index := solver.count
	solver.count++
	if index < 37 {
		solver.seed = append(solver.seed, price)
		if index == 36 {
			solver.initializeWMA()
		}
		return Result{}, nil
	}

	adjustedPeriod := 0.075*solver.period + 0.54
	solver.periodWMASub += price - solver.trailingWMAValue
	solver.periodWMASum += price * 4
	solver.trailingWMAValue = solver.trailingWMAQueue[0]
	solver.trailingWMAQueue[0] = solver.trailingWMAQueue[1]
	solver.trailingWMAQueue[1] = solver.trailingWMAQueue[2]
	solver.trailingWMAQueue[2] = price
	smoothed := solver.periodWMASum * 0.1
	solver.periodWMASum -= solver.periodWMASub
	solver.smoothPrice[solver.smoothPriceIndex] = smoothed

	var detrender, q1, jI, jQ, i2, q2 float64
	if index%2 == 0 {
		detrender, solver.prevDetrenderEven, solver.prevDetrenderInputEven = hilbert(smoothed, adjustedPeriod, solver.hilbertIndex, &solver.detrenderEven, solver.prevDetrenderEven, solver.prevDetrenderInputEven)
		q1, solver.prevQ1Even, solver.prevQ1InputEven = hilbert(detrender, adjustedPeriod, solver.hilbertIndex, &solver.q1Even, solver.prevQ1Even, solver.prevQ1InputEven)
		jI, solver.prevJIEven, solver.prevJIInputEven = hilbert(solver.i1EvenPrev3, adjustedPeriod, solver.hilbertIndex, &solver.jIEven, solver.prevJIEven, solver.prevJIInputEven)
		jQ, solver.prevJQEven, solver.prevJQInputEven = hilbert(q1, adjustedPeriod, solver.hilbertIndex, &solver.jQEven, solver.prevJQEven, solver.prevJQInputEven)
		solver.hilbertIndex = (solver.hilbertIndex + 1) % 3
		q2 = 0.2*(q1+jI) + 0.8*solver.previousQ2
		i2 = 0.2*(solver.i1EvenPrev3-jQ) + 0.8*solver.previousI2
		solver.i1OddPrev3, solver.i1OddPrev2 = solver.i1OddPrev2, detrender
	} else {
		detrender, solver.prevDetrenderOdd, solver.prevDetrenderInputOdd = hilbert(smoothed, adjustedPeriod, solver.hilbertIndex, &solver.detrenderOdd, solver.prevDetrenderOdd, solver.prevDetrenderInputOdd)
		q1, solver.prevQ1Odd, solver.prevQ1InputOdd = hilbert(detrender, adjustedPeriod, solver.hilbertIndex, &solver.q1Odd, solver.prevQ1Odd, solver.prevQ1InputOdd)
		jI, solver.prevJIOdd, solver.prevJIInputOdd = hilbert(solver.i1OddPrev3, adjustedPeriod, solver.hilbertIndex, &solver.jIOdd, solver.prevJIOdd, solver.prevJIInputOdd)
		jQ, solver.prevJQOdd, solver.prevJQInputOdd = hilbert(q1, adjustedPeriod, solver.hilbertIndex, &solver.jQOdd, solver.prevJQOdd, solver.prevJQInputOdd)
		q2 = 0.2*(q1+jI) + 0.8*solver.previousQ2
		i2 = 0.2*(solver.i1OddPrev3-jQ) + 0.8*solver.previousI2
		solver.i1EvenPrev3, solver.i1EvenPrev2 = solver.i1EvenPrev2, detrender
	}

	solver.re = 0.2*(i2*solver.previousI2+q2*solver.previousQ2) + 0.8*solver.re
	solver.im = 0.2*(i2*solver.previousQ2-q2*solver.previousI2) + 0.8*solver.im
	solver.previousQ2, solver.previousI2 = q2, i2
	previousPeriod := solver.period
	if solver.im != 0 && solver.re != 0 {
		solver.period = 360 / (math.Atan(solver.im/solver.re) * 180 / math.Pi)
	}
	solver.period = math.Min(solver.period, 1.5*previousPeriod)
	solver.period = math.Max(solver.period, 0.67*previousPeriod)
	solver.period = math.Max(6, math.Min(50, solver.period))
	solver.period = 0.2*solver.period + 0.8*previousPeriod
	solver.smoothPeriod = 0.33*solver.period + 0.67*solver.smoothPeriod

	cycleLength := int(solver.smoothPeriod + 0.5)
	realPart, imagPart := 0.0, 0.0
	priceIndex := solver.smoothPriceIndex
	for cycleIndex := 0; cycleIndex < cycleLength; cycleIndex++ {
		angle := float64(cycleIndex) * 2 * math.Pi / float64(cycleLength)
		realPart += math.Sin(angle) * solver.smoothPrice[priceIndex]
		imagPart += math.Cos(angle) * solver.smoothPrice[priceIndex]
		priceIndex--
		if priceIndex < 0 {
			priceIndex = len(solver.smoothPrice) - 1
		}
	}
	phaseAngle := 0.0
	if math.Abs(imagPart) > 0 {
		phaseAngle = math.Atan(realPart/imagPart) * 180 / math.Pi
	} else if realPart < 0 {
		phaseAngle -= 90
	} else if realPart > 0 {
		phaseAngle += 90
	}
	phaseAngle += 90 + 360/solver.smoothPeriod
	if imagPart < 0 {
		phaseAngle += 180
	}
	if phaseAngle > 315 {
		phaseAngle -= 360
	}
	solver.smoothPriceIndex = (solver.smoothPriceIndex + 1) % len(solver.smoothPrice)
	if index < Lookback {
		return Result{InPhase: i2, Quadrature: q2}, nil
	}
	normalized := NormalizeDegrees(phaseAngle)
	return Result{PhaseAngle: &normalized, InPhase: i2, Quadrature: q2}, nil
}

func (solver *Solver) initializeWMA() {
	solver.periodWMASub = solver.seed[0] + solver.seed[1] + solver.seed[2]
	solver.periodWMASum = solver.seed[0] + 2*solver.seed[1] + 3*solver.seed[2]
	trailingValue := 0.0
	trailingIndex := 0
	for index := 3; index < len(solver.seed); index++ {
		value := solver.seed[index]
		solver.periodWMASub += value - trailingValue
		solver.periodWMASum += value * 4
		trailingValue = solver.seed[trailingIndex]
		trailingIndex++
		_ = solver.periodWMASum * 0.1
		solver.periodWMASum -= solver.periodWMASub
	}
	solver.trailingWMAValue = solver.seed[33]
	copy(solver.trailingWMAQueue[:], solver.seed[34:37])
	solver.seed = nil
}

func hilbert(input, adjustedPeriod float64, index int, buffer *[3]float64, previous, previousInput float64) (float64, float64, float64) {
	weighted := 0.0962 * input
	output := -buffer[index] + weighted - previous + 0.5769*previousInput
	buffer[index] = weighted
	return output * adjustedPeriod, 0.5769 * previousInput, input
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
