export type ReservoirEventType = "RUN_START" | "OUTFLOW" | "INFLOW" | "RUN_END";
export type ReservoirCause = "HOP_ON" | "HOP_OFF" | "SAFETY_LIQUIDATION";

export type CapitalReservoirEvent = {
  eventId: string | null;
  pipelineRunId: string;
  collectionRunId: string;
  runType: string;
  eventSequence: number;
  symbol: string | null;
  eventType: ReservoirEventType;
  cause: ReservoirCause | null;
  stageEmitValueId: string | null;
  triggerEventTime: string | null;
  executionEventTime: string | null;
  processedAt: string;
  quantity: number | null;
  executionPrice: number | null;
  signedFlowAmount: number | null;
  initialReservoir: number | null;
  symbolCount: number | null;
  initialSymbolCapital: number | null;
  reservoirBefore: number;
  reservoirAfter: number;
  activeQuantityAfter: number | null;
  reservoirCash: number | null;
  totalDeployedAfter: number | null;
  totalMarkedCapitalAfter: number | null;
  realizedPnl: number | null;
  unrealizedPnl: number | null;
  totalPnl: number | null;
  activePositions: number | null;
};

export type PipeState = {
  symbol: string;
  activeQuantity: number;
  capitalDeployed: number | null;
  capitalReturned: number | null;
  realizedPnl: number | null;
  lastExecutionPrice: number | null;
  cumulativeReservoirFlow: number;
  lastCause: ReservoirCause | null;
  lastEventSequence: number;
  currentCycle: PipeCycle | null;
  completedCycles: PipeCycle[];
};

export type PipeCycle = {
  entryCause: "HOP_ON";
  quantity: number;
  capitalDeployed: number;
  entryPrice: number | null;
  entryEventSequence: number;
  triggerTime: string | null;
  entryExecutionTime: string | null;
  capitalReturned: number | null;
  realizedPnl: number | null;
  exitCause: "HOP_OFF" | "SAFETY_LIQUIDATION" | null;
  exitEventSequence: number | null;
  exitExecutionTime: string | null;
};

export type CapitalSummary = {
  initialCapital: number | null;
  reservoirCash: number | null;
  deployedMarkedCapital: number | null;
  totalMarkedCapital: number | null;
  realizedPnl: number | null;
  unrealizedPnl: number | null;
  totalPnl: number | null;
  activeSymbols: number;
  participatingSymbols: number;
  utilizationPct: number | null;
  availablePct: number | null;
};

export const GOVERNED_SYMBOLS = [
  "AAPL", "AMD", "AMZN", "AVGO", "BAC", "COST", "DIA", "EEM", "GLD", "GOOGL",
  "HD", "IWM", "JNJ", "JPM", "MA", "META", "MSFT", "NFLX", "NVDA", "QQQ",
  "SPY", "TLT", "TSLA", "UNH", "V", "VXX", "WMT", "XLF", "XLK", "XOM",
] as const;

const eventTypes = new Set<ReservoirEventType>(["RUN_START", "OUTFLOW", "INFLOW", "RUN_END"]);
const causes = new Set<ReservoirCause>(["HOP_ON", "HOP_OFF", "SAFETY_LIQUIDATION"]);

function field(record: Record<string, unknown>, camel: string, snake: string): unknown {
  return record[camel] ?? record[snake];
}

function text(value: unknown): string | null {
  if (value === null || value === undefined || value === "") return null;
  if (typeof value === "object" && value && "toHexString" in value) {
    return String((value as { toHexString(): string }).toHexString());
  }
  return String(value);
}

function numberValue(value: unknown): number | null {
  if (value === null || value === undefined || value === "") return null;
  const candidate = typeof value === "object" && value && "toString" in value ? value.toString() : value;
  const parsed = Number(candidate);
  return Number.isFinite(parsed) ? parsed : null;
}

function iso(value: unknown): string | null {
  if (value === null || value === undefined || value === "" || value === 0 || value === "0") return null;
  const date = value instanceof Date ? value : new Date(typeof value === "string" && !/^\d+$/.test(value) ? value : Number(value));
  return Number.isNaN(date.valueOf()) ? null : date.toISOString();
}

function enumValue(value: unknown, prefix: string): string {
  const raw = String(value ?? "");
  return raw.startsWith(prefix) ? raw.slice(prefix.length) : raw;
}

export function normalizeReservoirEvent(input: unknown): CapitalReservoirEvent {
  if (!input || typeof input !== "object") throw new Error("Capital Reservoir event must be an object");
  const record = input as Record<string, unknown>;
  const eventType = enumValue(
    field(record, "eventType", "event_type"),
    "CAPITAL_RESERVOIR_EVENT_TYPE_",
  ) as ReservoirEventType;
  const rawCause = enumValue(field(record, "cause", "cause"), "CAPITAL_RESERVOIR_CAUSE_");
  const runType = text(field(record, "runType", "run_type"));
  const eventSequence = numberValue(field(record, "eventSequence", "event_sequence"));
  const pipelineRunId = text(field(record, "pipelineRunId", "pipeline_run_id"));
  const collectionRunId = text(field(record, "collectionRunId", "collection_run_id"));
  const processedAt = iso(field(record, "processedUnixMs", "processed_at"));
  if (!pipelineRunId || !collectionRunId || !processedAt || !eventTypes.has(eventType) || !eventSequence || eventSequence < 1) {
    throw new Error("Capital Reservoir event is missing required identity, sequence, type, or time");
  }
  if (!runType || !/^[A-Z0-9][A-Z0-9_-]*$/.test(runType)) throw new Error("Capital Reservoir event has invalid run type");

  return {
    eventId: text(field(record, "eventId", "event_id")),
    pipelineRunId,
    collectionRunId,
    runType,
    eventSequence,
    symbol: text(field(record, "symbol", "symbol")),
    eventType,
    cause: causes.has(rawCause as ReservoirCause) ? (rawCause as ReservoirCause) : null,
    stageEmitValueId: text(field(record, "stageEmitValueId", "stage_emit_value_id")),
    triggerEventTime: iso(field(record, "triggerEventUnixMs", "trigger_event_time")),
    executionEventTime: iso(field(record, "executionEventUnixMs", "execution_event_time")),
    processedAt,
    quantity: numberValue(field(record, "quantity", "quantity")),
    executionPrice: numberValue(field(record, "executionPrice", "execution_price")),
    signedFlowAmount: numberValue(field(record, "signedFlowAmount", "signed_flow_amount")),
    initialReservoir: numberValue(field(record, "initialReservoir", "initial_reservoir")),
    symbolCount: numberValue(field(record, "symbolCount", "symbol_count")),
    initialSymbolCapital: numberValue(field(record, "initialSymbolCapital", "initial_symbol_capital")),
    reservoirBefore: numberValue(field(record, "reservoirBefore", "reservoir_before")) ?? 0,
    reservoirAfter: numberValue(field(record, "reservoirAfter", "reservoir_after")) ?? 0,
    activeQuantityAfter: numberValue(field(record, "activeQuantityAfter", "active_quantity_after")),
    reservoirCash: numberValue(field(record, "reservoirCash", "reservoir_cash")),
    totalDeployedAfter: numberValue(field(record, "totalDeployedAfter", "total_deployed_after")),
    totalMarkedCapitalAfter: numberValue(field(record, "totalMarkedCapitalAfter", "total_marked_capital_after")),
    realizedPnl: numberValue(field(record, "realizedPnl", "realized_pnl")),
    unrealizedPnl: numberValue(field(record, "unrealizedPnl", "unrealized_pnl")),
    totalPnl: numberValue(field(record, "totalPnl", "total_pnl")),
    activePositions: numberValue(field(record, "activePositions", "active_positions")),
  };
}

export function mergeReservoirEvents(events: CapitalReservoirEvent[]): CapitalReservoirEvent[] {
  const keyed = new Map<string, CapitalReservoirEvent>();
  for (const event of events) keyed.set(`${event.pipelineRunId}:${event.eventSequence}`, event);
  return [...keyed.values()].sort((left, right) => left.eventSequence - right.eventSequence);
}

export function deriveCapitalSummary(events: CapitalReservoirEvent[], symbols: string[] = []): CapitalSummary {
  const ordered = mergeReservoirEvents(events);
  const start = ordered.find((event) => event.eventType === "RUN_START");
  const end = [...ordered].reverse().find((event) => event.eventType === "RUN_END");
  const initialCapital = start?.initialReservoir ?? null;
  const reservoirCash = ordered.at(-1)?.reservoirAfter ?? null;
  const pipeStates = derivePipeStates(ordered, symbols);
  const participatingSymbols = new Set(
    ordered.filter((event) => event.symbol && (event.eventType === "OUTFLOW" || event.eventType === "INFLOW")).map((event) => event.symbol),
  ).size;
  return {
    initialCapital,
    reservoirCash,
    deployedMarkedCapital: end?.totalDeployedAfter ?? null,
    totalMarkedCapital: end?.totalMarkedCapitalAfter ?? null,
    realizedPnl: end?.realizedPnl ?? null,
    unrealizedPnl: end?.unrealizedPnl ?? null,
    totalPnl: end?.totalPnl ?? null,
    activeSymbols: pipeStates.filter((pipe) => pipe.activeQuantity > 0).length,
    participatingSymbols,
    utilizationPct: initialCapital && end?.totalDeployedAfter !== null && end?.totalDeployedAfter !== undefined
      ? (end.totalDeployedAfter / initialCapital) * 100
      : null,
    availablePct: initialCapital && reservoirCash !== null ? (reservoirCash / initialCapital) * 100 : null,
  };
}

export function derivePipeStates(events: CapitalReservoirEvent[], symbols: string[] = []): PipeState[] {
  const states = new Map<string, PipeState>();
  for (const symbol of symbols) {
    states.set(symbol, emptyPipeState(symbol));
  }
  for (const event of mergeReservoirEvents(events)) {
    if (!event.symbol || (event.eventType !== "OUTFLOW" && event.eventType !== "INFLOW")) continue;
    const current = states.get(event.symbol) ?? emptyPipeState(event.symbol);
    current.activeQuantity = event.activeQuantityAfter ?? current.activeQuantity;
    current.lastExecutionPrice = event.executionPrice;
    current.cumulativeReservoirFlow += event.signedFlowAmount ?? 0;
    current.lastCause = event.cause;
    current.lastEventSequence = event.eventSequence;
    if (event.eventType === "OUTFLOW") {
      current.currentCycle = {
        entryCause: "HOP_ON",
        quantity: event.quantity ?? current.activeQuantity,
        capitalDeployed: Math.abs(event.signedFlowAmount ?? 0),
        entryPrice: event.executionPrice,
        entryEventSequence: event.eventSequence,
        triggerTime: event.triggerEventTime,
        entryExecutionTime: event.executionEventTime,
        capitalReturned: null,
        realizedPnl: null,
        exitCause: null,
        exitEventSequence: null,
        exitExecutionTime: null,
      };
    } else if (current.currentCycle) {
      const capitalReturned = Math.abs(event.signedFlowAmount ?? 0);
      const completedCycle: PipeCycle = {
        ...current.currentCycle,
        capitalReturned,
        realizedPnl: capitalReturned - current.currentCycle.capitalDeployed,
        exitCause: event.cause === "HOP_OFF" || event.cause === "SAFETY_LIQUIDATION" ? event.cause : null,
        exitEventSequence: event.eventSequence,
        exitExecutionTime: event.executionEventTime,
      };
      current.completedCycles.push(completedCycle);
      current.currentCycle = null;
    }
    const latestCycle = current.currentCycle ?? current.completedCycles.at(-1) ?? null;
    current.capitalDeployed = latestCycle?.capitalDeployed ?? null;
    current.capitalReturned = latestCycle?.capitalReturned ?? null;
    current.realizedPnl = latestCycle?.realizedPnl ?? null;
    states.set(event.symbol, current);
  }
  return [...states.values()].sort((left, right) => left.symbol.localeCompare(right.symbol));
}

function emptyPipeState(symbol: string): PipeState {
  return {
    symbol,
    activeQuantity: 0,
    capitalDeployed: null,
    capitalReturned: null,
    realizedPnl: null,
    lastExecutionPrice: null,
    cumulativeReservoirFlow: 0,
    lastCause: null,
    lastEventSequence: 0,
    currentCycle: null,
    completedCycles: [],
  };
}