import assert from "node:assert/strict";
import test from "node:test";
import { derivePipeStates, mergeReservoirEvents, normalizeReservoirEvent } from "./capital-reservoir";

const mongoStart = {
  pipeline_run_id: "pipeline-1", collection_run_id: "collection-1", run_type: "RUN_A",
  event_sequence: 1, event_type: "RUN_START", processed_at: new Date("2026-09-14T12:00:00Z"),
  reservoir_before: 3_000_000, reservoir_after: 3_000_000, initial_reservoir: 3_000_000,
  symbol_count: 30, initial_symbol_capital: 100_000, stage_emit_value_id: null,
};

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

test("derives pipe state from confirmed flows without fabricating RUN_END liquidation", () => {
  const start = normalizeReservoirEvent(mongoStart);
  const outflow = normalizeReservoirEvent({
    ...mongoStart, event_sequence: 2, event_type: "OUTFLOW", symbol: "AAPL", cause: "HOP_ON",
    quantity: 10, execution_price: 100, signed_flow_amount: -1000, active_quantity_after: 10,
    reservoir_before: 3_000_000, reservoir_after: 2_999_000,
  });
  const end = normalizeReservoirEvent({
    ...mongoStart, event_sequence: 3, event_type: "RUN_END", reservoir_before: 2_999_000,
    reservoir_after: 2_999_000, active_positions: 1, total_deployed_after: 1_050,
  });
  const state = derivePipeStates([start, outflow, end], ["AAPL", "MSFT"]);
  assert.equal(state.length, 2);
  assert.deepEqual(state[0], {
    symbol: "AAPL", activeQuantity: 10, deployedCapital: 1000, lastExecutionPrice: 100,
    netFlow: -1000, lastCause: "HOP_ON", lastEventSequence: 2,
  });
  assert.equal(state[1].activeQuantity, 0);
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