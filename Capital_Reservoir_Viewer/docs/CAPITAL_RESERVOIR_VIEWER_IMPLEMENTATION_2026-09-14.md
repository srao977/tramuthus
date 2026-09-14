# Capital Reservoir Viewer Implementation - 2026-09-14

## Scope

`Capital_Reservoir_Viewer` is an independent Next.js application. It observes the common DSE_JEH Capital Reservoir in live and replay modes without changing the governed runtime or persistence contract.

## Evidence handling

Mongo BSON documents and gRPC proto-loader objects are normalized to the same `CapitalReservoirEvent` model. Events are keyed by `pipeline_run_id + event_sequence`, deduplicated, and ordered by `event_sequence`. Pipe state at crosshair sequence $n$ is derived only from the authoritative prefix with sequence $\le n$.

`RUN_END` accounting is displayed as published. It does not close active pipes or manufacture an `INFLOW` event.

## Presentation

- KlineCharts common-reservoir continuity lane.
- KlineCharts selected-pipe signed-flow lane with a zero baseline encoded by each bar's open value.
- Shared crosshair sequence selection.
- Stable 30-pipe overview and detailed table.
- Event inspector showing the normalized authoritative event.
- Replay play, pause, restart, step, and `1x` / `5x` / `10x` / `MAX` rates.
- Explicit empty, loading, connection, and fault states.

## Security and operations

MongoDB and gRPC configuration is server-only. Replay routes issue no write operations. The browser connects only to same-origin HTTP/SSE routes. Development defaults target MongoDB at `127.0.0.1:27017` and DSE_JEH at `127.0.0.1:50052`; neither service is started by the viewer.