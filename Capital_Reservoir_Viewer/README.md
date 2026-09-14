# Capital Reservoir Viewer

Independent, read-only Next.js viewer for DSE_JEH Capital Reservoir evidence. Implemented 2026-09-14.

The app has two source adapters and one normalized browser model:

- **Replay** reads ordered events from `bar_sequence_db.capital_reservoir_events` through server-only MongoDB routes.
- **Live** subscribes to `RuntimeEvidenceService.SubscribeRuntimeEvidence` on the server and bridges typed `CapitalReservoirEvent` publications to the browser over SSE.
- **KlineCharts** renders common-reservoir continuity and selected-symbol signed flow. Moving either crosshair selects one authoritative event sequence and resolves the related reservoir event plus all 30 pipe states from that prefix.

The viewer does not calculate reservoir mathematics, submit trades, write MongoDB, start DSE_JEH, or synthesize liquidation at `RUN_END`.

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
