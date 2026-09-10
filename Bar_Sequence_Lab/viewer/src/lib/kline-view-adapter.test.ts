import assert from "node:assert/strict";
import test from "node:test";
import type { TrajectoryObservation } from "./types.ts";
import {
  chartTimestampToSequence,
  formatBarIndexTick,
  klineToObservation,
  observationAtChartTimestamp,
  observationToKLine,
  sequenceToChartTimestamp,
} from "./kline-view-adapter.ts";

function sample(seq: number): TrajectoryObservation {
  return {
    generatorSequenceNo: seq,
    symbol: "AAPL",
    partitionId: "A",
    collectionRunId: "20260910T191246Z-1",
    y: 230.5,
    open: 229,
    high: 231,
    low: 228,
    close: 230.5,
    volume: 1_000_000,
    interval: "1Min",
    sourceEventTime: "2026-09-10T19:12:00.000Z",
    sourceTimestampText: "2026-09-10T19:12:00Z",
    receivedTime: "2026-09-10T19:12:01.000Z",
    persistedTime: "2026-09-10T19:12:02.000Z",
    sourceId: "alpaca",
    alpacaMessageType: "b",
    payloadHash: "abc",
    sourceTimeRegression: false,
    duplicateArrival: false,
    eventCount: 12,
  };
}

test("chart timestamp is generator_sequence_no, not source_event_time", () => {
  const mapped = observationToKLine(sample(17), "price");
  assert.equal(sequenceToChartTimestamp(17), 17);
  assert.equal(mapped.timestamp, 17);
  assert.equal(chartTimestampToSequence(mapped.timestamp), 17);
  assert.notEqual(mapped.timestamp, Date.parse(mapped.sourceEventTime));
  assert.equal(formatBarIndexTick(mapped.timestamp), "17");
});

test("price adapter plots raw close and keeps provenance + OHLC", () => {
  const mapped = observationToKLine(sample(17), "price");
  assert.equal(mapped.close, 230.5);
  assert.equal(mapped.rawClose, 230.5);
  assert.equal(mapped.open, 229);
  assert.equal(mapped.rawVolume, 1_000_000);
  assert.equal(mapped.sourceEventTime, "2026-09-10T19:12:00.000Z");
  assert.equal(mapped.collectionRunId, "20260910T191246Z-1");
});

test("volume adapter plots raw volume independently of price", () => {
  const mapped = observationToKLine(sample(17), "volume");
  assert.equal(mapped.close, 1_000_000);
  assert.equal(mapped.rawClose, 230.5);
  assert.equal(mapped.rawVolume, 1_000_000);
});

test("click mapping recovers stored observation from timestamp even if extra keys are stripped", () => {
  const stored = [sample(16), sample(17), sample(18)];
  const stripped = { timestamp: 17, open: 229, high: 231, low: 228, close: 230.5, volume: 1_000_000 };
  const mapped = klineToObservation(stripped, null, stored);
  assert.equal(mapped?.generatorSequenceNo, 17);
  assert.equal(mapped?.payloadHash, "abc");
  assert.equal(observationAtChartTimestamp(stored, 17)?.symbol, "AAPL");
});
