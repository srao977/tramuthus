import assert from "node:assert/strict";
import test from "node:test";
import { deriveCapitalSummary, derivePipeStates, mergeReservoirEvents, normalizeReservoirEvent } from "./capital-reservoir";

const mongoStart = {
  pipeline_run_id: "pipeline-1", collection_run_id: "collection-1", run_type: "RUN_A",
  event_sequence: 1, event_type: "RUN_START", processed_at: new Date("2026-09-14T12:00:00Z"),
  reservoir_before: 3_000_000, reservoir_after: 3_000_000, initial_reservoir: 3_000_000,
  symbol_count: 30, initial_symbol_capital: 100_000, stage_emit_value_id: null,
};

function flow(sequence: number, eventType: "OUTFLOW" | "INFLOW", amount: number, overrides: Record<string, unknown> = {}) {
  return normalizeReservoirEvent({
    ...mongoStart,
    event_sequence: sequence,
    event_type: eventType,
    symbol: "AAPL",
    cause: eventType === "OUTFLOW" ? "HOP_ON" : "HOP_OFF",
    quantity: 10,
    execution_price: Math.abs(amount) / 10,
    signed_flow_amount: amount,
    active_quantity_after: eventType === "OUTFLOW" ? 10 : 0,
    reservoir_before: 3_000_000,
    reservoir_after: 3_000_000 + amount,
    ...overrides,
  });
}

test("normalizes equivalent Mongo and gRPC events", () => {
  const mongo = normalizeReservoirEvent(mongoStart);
  const grpc = normalizeReservoirEvent({
    pipelineRunId: "pipeline-1", collectionRunId: "collection-1", runType: "RUN_A",
    eventSequence: "1", eventType: "CAPITAL_RESERVOIR_EVENT_TYPE_RUN_START",
    processedUnixMs: "1789387200000", reservoirBefore: 3_000_000, reservoirAfter: 3_000_000,
    initialReservoir: 3_000_000, symbolCount: 30, initialSymbolCapital: 100_000,
  });
  assert.deepEqual({ ...grpc, eventId: null }, mongo);
});

test("deduplicates by pipeline and sequence and orders events", () => {
  const start = normalizeReservoirEvent(mongoStart);
  const end = normalizeReservoirEvent({ ...mongoStart, event_sequence: 3, event_type: "RUN_END" });
  assert.deepEqual(mergeReservoirEvents([end, start, start]).map((event) => event.eventSequence), [1, 3]);
});

test("presents an active BUY as deployed capital rather than a loss", () => {
  const [state] = derivePipeStates([flow(2, "OUTFLOW", -99_848.54)], ["AAPL"]);
  assert.equal(state.activeQuantity, 10);
  assert.equal(state.capitalDeployed, 99_848.54);
  assert.equal(state.capitalReturned, null);
  assert.equal(state.realizedPnl, null);
});

test("derives a profitable completed cycle from confirmed flow magnitudes", () => {
  const [state] = derivePipeStates([flow(2, "OUTFLOW", -99_800), flow(3, "INFLOW", 99_920)], ["AAPL"]);
  assert.equal(state.capitalDeployed, 99_800);
  assert.equal(state.capitalReturned, 99_920);
  assert.equal(state.realizedPnl, 120);
});

test("derives a losing completed cycle from confirmed flow magnitudes", () => {
  const [state] = derivePipeStates([flow(2, "OUTFLOW", -99_800), flow(3, "INFLOW", 99_770)], ["AAPL"]);
  assert.equal(state.capitalDeployed, 99_800);
  assert.equal(state.capitalReturned, 99_770);
  assert.equal(state.realizedPnl, -30);
});

test("does not leak a later SELL into an earlier replay prefix", () => {
  const events = [flow(2, "OUTFLOW", -99_800), flow(3, "INFLOW", 99_920)];
  const [state] = derivePipeStates(events.filter((event) => event.eventSequence <= 2), ["AAPL"]);
  assert.equal(state.capitalReturned, null);
  assert.equal(state.realizedPnl, null);
  assert.equal(state.currentCycle?.exitCause, null);
});

test("preserves SAFETY_LIQUIDATION as the completed cycle exit cause", () => {
  const safetyExit = flow(3, "INFLOW", 99_770, { cause: "SAFETY_LIQUIDATION" });
  const [state] = derivePipeStates([flow(2, "OUTFLOW", -99_800), safetyExit], ["AAPL"]);
  assert.equal(state.capitalReturned, 99_770);
  assert.equal(state.realizedPnl, -30);
  assert.equal(state.completedCycles[0].exitCause, "SAFETY_LIQUIDATION");
});

test("keeps a completed cycle while presenting a later BUY as the current cycle", () => {
  const events = [flow(2, "OUTFLOW", -99_800), flow(3, "INFLOW", 99_920), flow(4, "OUTFLOW", -88_000)];
  const [state] = derivePipeStates(events, ["AAPL"]);
  assert.equal(state.activeQuantity, 10);
  assert.equal(state.capitalDeployed, 88_000);
  assert.equal(state.capitalReturned, null);
  assert.equal(state.realizedPnl, null);
  assert.equal(state.completedCycles.length, 1);
  assert.equal(state.completedCycles[0].realizedPnl, 120);
});

test("preserves authoritative signed reservoir flow while deriving UX values", () => {
  const buy = flow(2, "OUTFLOW", -99_800);
  const sell = flow(3, "INFLOW", 99_920);
  const [state] = derivePipeStates([buy, sell], ["AAPL"]);
  assert.equal(buy.signedFlowAmount, -99_800);
  assert.equal(sell.signedFlowAmount, 99_920);
  assert.equal(state.cumulativeReservoirFlow, 120);
});

test("maintains authoritative reservoir continuity", () => {
  const events = [
    normalizeReservoirEvent(mongoStart),
    normalizeReservoirEvent({ ...mongoStart, event_sequence: 2, event_type: "OUTFLOW", reservoir_before: 3_000_000, reservoir_after: 2_999_000 }),
    normalizeReservoirEvent({ ...mongoStart, event_sequence: 3, event_type: "INFLOW", reservoir_before: 2_999_000, reservoir_after: 3_000_050 }),
  ];
  for (let index = 1; index < events.length; index += 1) assert.equal(events[index - 1].reservoirAfter, events[index].reservoirBefore);
});

test("normalizes a generic run classification", () => {
  const event = normalizeReservoirEvent({ ...mongoStart, run_type: "RESERVOIR_VALIDATION_001" });
  assert.equal(event.runType, "RESERVOIR_VALIDATION_001");
});

test("derives reconciled endpoint capital metrics from authoritative RUN_END", () => {
  const events = [
    normalizeReservoirEvent(mongoStart),
    flow(2, "OUTFLOW", -99_800),
    normalizeReservoirEvent({
      ...mongoStart,
      event_sequence: 3,
      event_type: "RUN_END",
      reservoir_before: 2_900_200,
      reservoir_after: 2_900_200,
      reservoir_cash: 2_900_200,
      total_deployed_after: 100_000,
      total_marked_capital_after: 3_000_200,
      realized_pnl: 0,
      unrealized_pnl: 200,
      total_pnl: 200,
      active_positions: 1,
    }),
  ];
  const summary = deriveCapitalSummary(events, ["AAPL", "AMD"]);
  assert.equal(summary.reservoirCash, 2_900_200);
  assert.equal(summary.deployedMarkedCapital, 100_000);
  assert.equal(summary.totalMarkedCapital, 3_000_200);
  assert.equal(summary.realizedPnl, 0);
  assert.equal(summary.unrealizedPnl, 200);
  assert.equal(summary.totalPnl, 200);
  assert.equal(summary.activeSymbols, 1);
  assert.equal(summary.participatingSymbols, 1);
  assert.equal(summary.utilizationPct, 10 / 3);
});

test("does not expose future RUN_END marked capital or P&L in an earlier replay prefix", () => {
  const summary = deriveCapitalSummary([normalizeReservoirEvent(mongoStart), flow(2, "OUTFLOW", -99_800)], ["AAPL"]);
  assert.equal(summary.reservoirCash, 2_900_200);
  assert.equal(summary.deployedMarkedCapital, null);
  assert.equal(summary.totalMarkedCapital, null);
  assert.equal(summary.realizedPnl, null);
  assert.equal(summary.unrealizedPnl, null);
  assert.equal(summary.totalPnl, null);
  assert.equal(summary.participatingSymbols, 1);
});