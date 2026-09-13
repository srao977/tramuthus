package input_test

import (
	"testing"
	"time"

	finfeedsatv1 "fin_feedsat_1/gen/fin_feedsat/v1"

	"tramuthus/dse-jeh-transsat-1/internal/adapter/mongo"
	inputcommon "tramuthus/dse-jeh-transsat-1/internal/input"
	"tramuthus/dse-jeh-transsat-1/internal/input/offline"
	"tramuthus/dse-jeh-transsat-1/internal/input/online"
)

func TestOnlineOfflineMappingEquivalence(t *testing.T) {
	start := time.UnixMilli(1_700_000_000_000).UTC()
	offlineEvent := offline.MapObservation(mongo.Observation{
		CollectionRunID: "run-1", Symbol: "AAPL", GeneratorSequenceNo: 1,
		SourceEventTime: start, SourceTimestampText: start.Format(time.RFC3339Nano),
		ReceivedTime: start.Add(time.Second), Interval: "1Min",
		Open: 100, High: 102, Low: 99, Close: 101, Volume: 42, EventCount: 7,
		SourceID: "provider", PayloadHash: "payload",
	})
	onlineEvent := online.MapBar(&finfeedsatv1.Bar{
		Symbol: "AAPL", Interval: "1Min", IntervalStartUnixMs: start.UnixMilli(),
		IntervalEndUnixMs: start.Add(time.Minute).UnixMilli(), Open: 100, High: 102,
		Low: 99, Close: 101, Volume: 42, EventCount: 7, Source: "provider",
		SourceTimestamp: start.Format(time.RFC3339Nano), ReceiptUnixMs: start.Add(time.Second).UnixMilli(),
		MarketSnapshotId: "payload", IsFinal: true,
	}, 1)
	if err := inputcommon.EquivalentBarValues(offlineEvent, onlineEvent); err != nil {
		t.Fatal(err)
	}
}
