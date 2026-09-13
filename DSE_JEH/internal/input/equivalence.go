package input

import (
	"fmt"

	dsejehv1 "tramuthus/dse-jeh-transsat-1/gen/dse_jeh/v1"
)

func EquivalentBarValues(left, right *dsejehv1.BarEvent) error {
	if left.GetEntityId() != right.GetEntityId() || left.GetSymbol() != right.GetSymbol() {
		return fmt.Errorf("entity mismatch")
	}
	if left.GetInterval() != right.GetInterval() || left.GetIntervalStartUnixMs() != right.GetIntervalStartUnixMs() || left.GetIntervalEndUnixMs() != right.GetIntervalEndUnixMs() {
		return fmt.Errorf("interval mismatch")
	}
	if left.GetOpen() != right.GetOpen() || left.GetHigh() != right.GetHigh() || left.GetLow() != right.GetLow() || left.GetClose() != right.GetClose() {
		return fmt.Errorf("OHLC mismatch")
	}
	if left.GetVolume() != right.GetVolume() || left.GetEventCount() != right.GetEventCount() {
		return fmt.Errorf("volume or event-count mismatch")
	}
	return nil
}
