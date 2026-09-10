import assert from "node:assert/strict";
import test from "node:test";
import {
  clampCursor,
  maxSequenceOf,
  pickDefaultWaveSymbols,
  reconcileSlotsWithAvailable,
  replaceWaveSymbol,
  visiblePrefix,
  waveComplete,
} from "./replay.ts";

test("defaults prefer AAPL MSFT SPY JPM when all exist", () => {
  const available = ["AMD", "AAPL", "JPM", "MSFT", "NVDA", "SPY"];
  assert.deepEqual(pickDefaultWaveSymbols(available), ["AAPL", "MSFT", "SPY", "JPM"]);
});

test("defaults fill from available when preferred symbols are missing", () => {
  const available = ["COST", "HD", "WMT"];
  assert.deepEqual(pickDefaultWaveSymbols(available), ["COST", "HD", "WMT", ""]);
});

test("visiblePrefix is bar-index, does not fabricate", () => {
  const obs = [
    { generatorSequenceNo: 1, y: 10 },
    { generatorSequenceNo: 2, y: 11 },
    { generatorSequenceNo: 3, y: 12 },
  ];
  assert.equal(visiblePrefix(obs, 0).length, 0);
  assert.deepEqual(
    visiblePrefix(obs, 2).map((row) => row.generatorSequenceNo),
    [1, 2],
  );
  assert.equal(visiblePrefix(obs, 9).length, 3);
});

test("shorter wave stays complete while longer waves continue", () => {
  assert.equal(waveComplete(32, 32), true);
  assert.equal(waveComplete(32, 51), true);
  assert.equal(waveComplete(51, 32), false);
  assert.equal(maxSequenceOf([51, 48, 48, 32]), 51);
});

test("cursor clamp and slot replace stay in four presentation slots", () => {
  assert.equal(clampCursor(99, 51), 51);
  assert.equal(clampCursor(-2, 51), 0);
  assert.deepEqual(replaceWaveSymbol(["AAPL", "MSFT", "SPY", "JPM"], 2, "QQQ"), [
    "AAPL",
    "MSFT",
    "QQQ",
    "JPM",
  ]);
});

test("group-filtered availability replaces missing symbols without inventing", () => {
  assert.deepEqual(
    reconcileSlotsWithAvailable(["AAPL", "MSFT", "SPY", "JPM"], ["AAPL", "AMD", "MSFT", "NVDA"]),
    ["AAPL", "MSFT", "AMD", "NVDA"],
  );
});
