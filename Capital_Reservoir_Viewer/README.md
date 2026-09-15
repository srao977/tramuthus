# Capital Reservoir Viewer

Independent, read-only Next.js viewer for DSE_JEH Capital Reservoir evidence. Implemented 2026-09-14.

The app has two source adapters and one normalized browser model:

- **Replay** reads ordered events from `bar_sequence_db.capital_reservoir_events` through server-only MongoDB routes.
- **Live** subscribes to `RuntimeEvidenceService.SubscribeRuntimeEvidence` on the server and bridges typed `CapitalReservoirEvent` publications to the browser over SSE.
- **KlineCharts** renders common-reservoir continuity and selected-symbol signed flow. Moving either crosshair selects one authoritative event sequence and resolves the related reservoir event plus all 30 pipe states from that prefix.

The viewer does not calculate reservoir mathematics, submit trades, write MongoDB, start DSE_JEH, or synthesize liquidation at `RUN_END`.

## Reservoir Flow Semantics

Presentation clarified 2026-09-15:

- Signed flow is always shown from the common reservoir perspective. A confirmed BUY is negative reservoir flow and a confirmed SELL or liquidation is positive reservoir flow.
- A negative BUY flow is capital deployment into a position, not a loss. The pipe table displays its positive magnitude as **Capital Deployed**.
- A positive SELL flow is capital returned to the reservoir, not automatically profit. The pipe table displays its positive magnitude as **Capital Returned**.
- **Realized P&L** is presented separately for a completed position cycle as capital returned minus capital deployed in the current zero-cost MockExecutor evidence.
- Every pipe value at replay event sequence N is derived only from authoritative events at or before N. Raw `signedFlowAmount` remains available in the event inspector.

## Capital Summary Semantics

UX refined 2026-09-15 around the single common-reservoir identity:

`Initial Capital = Reservoir Cash + Deployed Marked Capital = Total Marked Capital`

- Current reservoir cash comes from the latest event at or before the selected sequence.
- Deployed marked capital, total marked capital, realized P&L, unrealized P&L, total P&L, and utilization are displayed only when the selected prefix includes the authoritative `RUN_END` fields. The viewer does not estimate them earlier.
- Reservoir utilization is `RUN_END totalDeployedAfter / RUN_START initialReservoir`. Reservoir availability is current `reservoirAfter / RUN_START initialReservoir`.
- **Capital Currently Deployed** counts prefix-derived symbols with active quantity. **Symbols Participated** counts distinct symbols with confirmed OUTFLOW or INFLOW events in the selected prefix; it is not the endpoint active-position count.
- The compact pipe topology represents 30 bidirectional connections to one reservoir. It does not represent 30 isolated cash accounts.
- `CapitalReservoirEvent` does not contain the four-state Dynamic Execution state. The viewer therefore exposes the authoritative capital state (`IDLE` or `DEPLOYED`) without guessing or renaming DSE state.

No Dynamic Risk Interceptor fields, scores, gains, or behavior are implemented.

## Run

```powershell
npm install
npm test
npm run lint
npx tsc --noEmit
npm run build
npm run dev -- -H 127.0.0.1 -p 3010
```

Open `http://127.0.0.1:3010`.

See [Capital Reservoir Viewer User Guide](docs/CAPITAL_RESERVOIR_VIEWER_USER_GUIDE_2026-09-14.md) for the complete viewer, browser, pipeline-run, replay, and shutdown sequence.

Copy environment values from `.env.example` only when overriding the local defaults. All connection settings are server-only and must never use a `NEXT_PUBLIC_` prefix.

## Routes

- `GET /api/runs` lists persisted pipeline runs using aggregation only.
- `GET /api/runs/[run]/events` returns that run ordered by `event_sequence` using `find` only.
- `GET /api/live` opens the server-side gRPC subscription and emits normalized SSE events.

MongoDB credentials and the gRPC client remain outside the browser bundle.
