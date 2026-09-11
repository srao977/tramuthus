import assert from "node:assert/strict";
import test from "node:test";
import { fittedBarSpace, klineToPhasePoint, phasePointAtCoordinate, phasePointsToKLines } from "./phase-adapter.ts";
import type { PhasePoint } from "./types.ts";

const point: PhasePoint = {
  collectionRunId: "run",
  partitionId: "A",
  symbol: "GOOGL",
  generatorSequenceNo: 64,
  seriesSize: 120,
  solverName: "EHLERS_DOMINANT_CYCLE_PHASE",
  solverVersion: "V0.1",
  inputSeriesType: "MEDIAN_PRICE",
  phaseAngleDegrees: 327.675,
  phaseObservable: true,
  validityState: "OBSERVABLE",
};

test("preserves generator_sequence_no as the chart coordinate", () => {
  const [mapped] = phasePointsToKLines([point]);
  assert.equal(mapped.timestamp, 64);
  assert.equal(mapped.close, point.phaseAngleDegrees);
  assert.equal(phasePointAtCoordinate([point], mapped.timestamp), point);
});

test("does not invent records before the first observable phase point", () => {
  const mapped = phasePointsToKLines([{ ...point, generatorSequenceNo: 81 }]);
  assert.deepEqual(mapped.map((row) => row.timestamp), [81]);
  assert.equal(phasePointAtCoordinate([point], 1), null);
});

test("maps a clicked kline back to the exact persisted phase point", () => {
  const points = [point, { ...point, generatorSequenceNo: 82 }];
  const clicked = phasePointsToKLines(points)[1];

  assert.equal(klineToPhasePoint(clicked, points), points[1]);
});

test("fits short observable series without changing their sequence coordinates", () => {
  assert.equal(fittedBarSpace(1024, 52), (1024 - 76) / 52);
  assert.equal(fittedBarSpace(1024, 10), 24);
  assert.equal(fittedBarSpace(320, 120), 4);
});
