import assert from "node:assert/strict";
import test from "node:test";
import { currentCounts, polarPoint } from "./presentation.ts";
import type { RunEvidence } from "./types.ts";

function near(actual: number, expected: number) {
  assert.ok(Math.abs(actual - expected) < 1e-9, `${actual} != ${expected}`);
}

test("polar orientation is top zero and clockwise", () => {
  const center = 100;
  const radius = 50;
  const zero = polarPoint(0, center, radius);
  const ninety = polarPoint(90, center, radius);
  const oneEighty = polarPoint(180, center, radius);
  const twoSeventy = polarPoint(270, center, radius);
  near(zero.x, 100); near(zero.y, 50);
  near(ninety.x, 150); near(ninety.y, 100);
  near(oneEighty.x, 100); near(oneEighty.y, 150);
  near(twoSeventy.x, 50); near(twoSeventy.y, 100);
});

test("359 and 1 degrees render adjacent at the circular boundary", () => {
  const before = polarPoint(359, 100, 50);
  const after = polarPoint(1, 100, 50);
  assert.ok(Math.hypot(before.x - after.x, before.y - after.y) < 2);
});

test("initializing inputs are counted but never treated as observable zero", () => {
  const run = {
    inputs: [{ symbol: "DIA", phase_observable: false, phase_angle_degrees: null }, { symbol: "AAPL", phase_observable: true, phase_angle_degrees: 0 }],
    states: { AAPL: { symbol: "AAPL", current_zone: "MOMENTUM_HOLD" } },
  } as unknown as RunEvidence;
  assert.deepEqual(currentCounts(run), {
    total: 2,
    observable: 1,
    initializing: 1,
    zones: { HOP_ON: 0, MOMENTUM_HOLD: 1, HOP_OFF: 0, DISREGARD: 0 },
  });
});
