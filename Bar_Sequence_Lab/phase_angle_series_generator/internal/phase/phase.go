// Package phase calculates price-only John Ehlers dominant-cycle phase.
// Inputs are ordered median prices; outputs preserve one result per input bar.
// Bars before the recognized lookback are returned as INITIALIZING with no angle.
package phase

import (
	"fmt"
	"math"
)

const (
	SolverName      = "EHLERS_DOMINANT_CYCLE_PHASE"
	SolverVersion   = "V0.1"
	InputSeriesType = "MEDIAN_PRICE"
	Lookback        = 63

	Initializing = "INITIALIZING"
	Observable   = "OBSERVABLE"
)

// Result is phase evidence for the input at the same slice index. PhaseAngle
// is nil during initialization; InPhase and Quadrature are solver diagnostics.
type Result struct {
	PhaseAngle *float64
	Observable bool
	State      string
	InPhase    float64
	Quadrature float64
}

// MedianPrice returns the Lab V0.1 price trajectory value for one raw bar.
func MedianPrice(high, low float64) (float64, error) {
	if math.IsNaN(high) || math.IsInf(high, 0) || math.IsNaN(low) || math.IsInf(low, 0) {
		return 0, fmt.Errorf("high and low must be finite")
	}
	return (high + low) / 2, nil
}

// SmoothFour applies Ehlers' 4-bar weighted price smoother in newest-first order.
func SmoothFour(current, previous1, previous2, previous3 float64) float64 {
	return (4*current + 3*previous1 + 2*previous2 + previous3) / 10
}

// NormalizeDegrees maps any finite angle deterministically into [0, 360).
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

// DominantCyclePhase implements TA-Lib HT_DCPHASE with unstable period zero.
// It uses Ehlers' 4-bar WMA, Hilbert coefficients 0.0962/0.5769, homodyne
// discriminator, period bounds 6..50, and TA-Lib's 63-bar lookback. The
// returned slice always matches prices in length and never fabricates angles.
func DominantCyclePhase(prices []float64) ([]Result, error) {
	results := make([]Result, len(prices))
	for i := range results {
		results[i].State = Initializing
	}
	for i, price := range prices {
		if math.IsNaN(price) || math.IsInf(price, 0) {
			return nil, fmt.Errorf("price[%d] must be finite", i)
		}
	}
	if len(prices) <= Lookback {
		return results, nil
	}

	const a, b = 0.0962, 0.5769
	var periodWMASum, periodWMASub float64
	today := 0
	periodWMASub = prices[today]
	periodWMASum = prices[today]
	today++
	periodWMASub += prices[today]
	periodWMASum += prices[today] * 2
	today++
	periodWMASub += prices[today]
	periodWMASum += prices[today] * 3
	today++
	trailingWMAIdx := 0
	trailingWMAValue := 0.0
	for i := 0; i < 34; i++ {
		value := prices[today]
		today++
		periodWMASub += value - trailingWMAValue
		periodWMASum += value * 4
		trailingWMAValue = prices[trailingWMAIdx]
		trailingWMAIdx++
		_ = periodWMASum * 0.1
		periodWMASum -= periodWMASub
	}

	var detrenderOdd, detrenderEven [3]float64
	var q1Odd, q1Even [3]float64
	var jIOdd, jIEven [3]float64
	var jQOdd, jQEven [3]float64
	var prevDetrenderOdd, prevDetrenderEven float64
	var prevDetrenderInputOdd, prevDetrenderInputEven float64
	var prevQ1Odd, prevQ1Even, prevQ1InputOdd, prevQ1InputEven float64
	var prevJIOdd, prevJIEven, prevJIInputOdd, prevJIInputEven float64
	var prevJQOdd, prevJQEven, prevJQInputOdd, prevJQInputEven float64
	var period, smoothPeriod, previousQ2, previousI2, re, im float64
	var i1OddPrev2, i1OddPrev3, i1EvenPrev2, i1EvenPrev3 float64
	var smoothPrice [50]float64
	hilbertIndex, smoothPriceIndex := 0, 0

	for ; today < len(prices); today++ {
		adjustedPeriod := 0.075*period + 0.54
		value := prices[today]
		periodWMASub += value - trailingWMAValue
		periodWMASum += value * 4
		trailingWMAValue = prices[trailingWMAIdx]
		trailingWMAIdx++
		smoothed := periodWMASum * 0.1
		periodWMASum -= periodWMASub
		smoothPrice[smoothPriceIndex] = smoothed

		var detrender, q1, jI, jQ, i2, q2 float64
		if today%2 == 0 {
			detrender, prevDetrenderEven, prevDetrenderInputEven = hilbert(smoothed, adjustedPeriod, hilbertIndex, &detrenderEven, prevDetrenderEven, prevDetrenderInputEven)
			q1, prevQ1Even, prevQ1InputEven = hilbert(detrender, adjustedPeriod, hilbertIndex, &q1Even, prevQ1Even, prevQ1InputEven)
			jI, prevJIEven, prevJIInputEven = hilbert(i1EvenPrev3, adjustedPeriod, hilbertIndex, &jIEven, prevJIEven, prevJIInputEven)
			jQ, prevJQEven, prevJQInputEven = hilbert(q1, adjustedPeriod, hilbertIndex, &jQEven, prevJQEven, prevJQInputEven)
			hilbertIndex = (hilbertIndex + 1) % 3
			q2 = 0.2*(q1+jI) + 0.8*previousQ2
			i2 = 0.2*(i1EvenPrev3-jQ) + 0.8*previousI2
			i1OddPrev3, i1OddPrev2 = i1OddPrev2, detrender
		} else {
			detrender, prevDetrenderOdd, prevDetrenderInputOdd = hilbert(smoothed, adjustedPeriod, hilbertIndex, &detrenderOdd, prevDetrenderOdd, prevDetrenderInputOdd)
			q1, prevQ1Odd, prevQ1InputOdd = hilbert(detrender, adjustedPeriod, hilbertIndex, &q1Odd, prevQ1Odd, prevQ1InputOdd)
			jI, prevJIOdd, prevJIInputOdd = hilbert(i1OddPrev3, adjustedPeriod, hilbertIndex, &jIOdd, prevJIOdd, prevJIInputOdd)
			jQ, prevJQOdd, prevJQInputOdd = hilbert(q1, adjustedPeriod, hilbertIndex, &jQOdd, prevJQOdd, prevJQInputOdd)
			q2 = 0.2*(q1+jI) + 0.8*previousQ2
			i2 = 0.2*(i1OddPrev3-jQ) + 0.8*previousI2
			i1EvenPrev3, i1EvenPrev2 = i1EvenPrev2, detrender
		}

		re = 0.2*(i2*previousI2+q2*previousQ2) + 0.8*re
		im = 0.2*(i2*previousQ2-q2*previousI2) + 0.8*im
		previousQ2, previousI2 = q2, i2
		previousPeriod := period
		if im != 0 && re != 0 {
			period = 360 / (math.Atan(im/re) * 180 / math.Pi)
		}
		period = math.Min(period, 1.5*previousPeriod)
		period = math.Max(period, 0.67*previousPeriod)
		period = math.Max(6, math.Min(50, period))
		period = 0.2*period + 0.8*previousPeriod
		smoothPeriod = 0.33*period + 0.67*smoothPeriod

		cycleLength := int(smoothPeriod + 0.5)
		realPart, imagPart := 0.0, 0.0
		index := smoothPriceIndex
		for i := 0; i < cycleLength; i++ {
			angle := float64(i) * 2 * math.Pi / float64(cycleLength)
			realPart += math.Sin(angle) * smoothPrice[index]
			imagPart += math.Cos(angle) * smoothPrice[index]
			index--
			if index < 0 {
				index = len(smoothPrice) - 1
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
		phaseAngle += 90 + 360/smoothPeriod
		if imagPart < 0 {
			phaseAngle += 180
		}
		if phaseAngle > 315 {
			phaseAngle -= 360
		}
		if today >= Lookback {
			normalized := NormalizeDegrees(phaseAngle)
			results[today] = Result{PhaseAngle: &normalized, Observable: true, State: Observable, InPhase: i2, Quadrature: q2}
		}
		smoothPriceIndex = (smoothPriceIndex + 1) % len(smoothPrice)
	}
	return results, nil
}

func hilbert(input, adjustedPeriod float64, index int, buffer *[3]float64, previous, previousInput float64) (output, nextPrevious, nextInput float64) {
	weighted := 0.0962 * input
	output = -buffer[index] + weighted - previous + 0.5769*previousInput
	buffer[index] = weighted
	return output * adjustedPeriod, 0.5769 * previousInput, input
}
