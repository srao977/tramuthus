# Bar Sequence Lab — multi-wave klinecharts replay

Date: 2026-09-10  
Module: `Bar_Sequence_Lab/viewer` only  
Verdict: **FUNCTIONALLY REAL** raw four-wave bar-index replay over Mongo `bar_sequence_db.bar_sequence`

THE BAR SEQUENCE IS THE EXPERIMENTAL SOURCE OF TRUTH.  
x-axis = `generator_sequence_no` (per `collection_run_id` + symbol).  
Price and Volume are independent trajectories.  
Maths / Ehlers-Hilbert / Hop / strategy overlays were **not** implemented.

Fin_FeedSat_1, Fin_FeedSat_1_Viewer, DSE_TransSat_1_viewer, DSE_TransSat_1_worker, HACCAM, and `Bar_Sequence_Lab/generator` were **not** modified.  
Not committed. Not pushed.

## Chart decision

Custom SVG is no longer the primary chart.

klinecharts **10.0.3** is the rendering mechanism because:

- future derived overlays (Ehlers / Hop) can be layers, not a second chart stack
- touch / crosshair / tooltip already exist
- DSE / later satellites can reuse the same canvas contract

klinecharts does **not** redefine the analytical axis. The library requires a time-like `timestamp` field. The Lab adapter maps:

```
timestamp := generator_sequence_no
```

It does **not** use `Date.parse(source_event_time)` as the chart coordinate. `formatter.formatDate` prints the bar index. Replay is BAR-INDEX, not market-time sync.

## Fin UX reuse (shell only)

Read-only visual language from `C:\Users\chino\Fin_FeedSat_1_Viewer`:

- topbar `#132126` + gold `#e8b931`
- left rail `.view-button` / `.is-active` / `.is-disabled`
- workspace, health-banner, stream-pill, ops-grid
- `next-themes` `attribute="class"` default dark

Not copied:

- Fin science, gRPC, EventSource live stream
- Fin market-time kline axis
- Fin product names
- Ehlers / Hop as working views (rail buttons exist, disabled)

PWA identity remains **Bar Sequence Lab**.

## Architecture

| Layer | Role |
| --- | --- |
| Mongo `bar_sequence` | source of truth; viewer is read-only |
| `LabMongoProvider` | server-only adapter (`import "server-only"`) |
| App Router `/api/runs*` | `runtime=nodejs`, `force-dynamic`, `Cache-Control: no-store` |
| Browser | same-origin fetch only; never opens Mongo |
| `KlineChartsViewAdapter` | sequence → `KLineData.timestamp`; provenance on extra keys |
| Four `WaveLaneChart` instances | independent y-scales |
| Client replay | prefix `generator_sequence_no <= cursor`; no Mongo per frame |

Env (server-only, never `NEXT_PUBLIC_*`):

```
BAR_SEQ_LAB_MONGO_URI=mongodb://127.0.0.1:27017
BAR_SEQ_LAB_MONGO_DB=bar_sequence_db
BAR_SEQ_LAB_MONGO_COLLECTION=bar_sequence
```

## Four presentation slots

Slots are **not** Alpha/Beta/Gamma/Delta science. They are four independently scaled canvases.

Default symbols when present: AAPL, MSFT, SPY, JPM.  
Replay cursor uses the **longest** wave `maxSequence`. Shorter waves stay complete; they are not stretched, interpolated, or time-aligned.

Speeds: 0.5 / 1 / 2 / 4. `BASE_MS_PER_BAR = 140` at 1x (presentation only).

Volume trajectory plots raw volume as the area `close` while keeping `rawClose` / `rawVolume` / OHLC / provenance on extra KLine keys.

## Reference Mongo (`20260910T191246Z-1`)

Live mongosh against `127.0.0.1:27017` / `bar_sequence_db.bar_sequence`:

| Query | Count |
| --- | ---: |
| run total | 1393 |
| AAPL | 51 (seq 1..51) |
| MSFT | 48 (seq 1..48) |
| SPY | 48 (seq 1..48) |
| JPM | 46 (seq 1..46) |

Price and volume share the same observation documents; they are display trajectories, not separate Mongo collections. Counts match the generator live collection report.

## Validation

| Check | Result |
| --- | --- |
| `npm test` | 12 pass (numbers, replay, adapter) |
| `npx tsc --noEmit` | pass |
| `npm run lint` | pass after epoch-safe replay + client-only chart |
| `npm run build` | pass after `dynamic(..., { ssr: false })` for WaveLaneChart (klinecharts touches `window`) |
| Mongo URI in `.next/static` | **not present** |
| Mongo URI in `.next/server` | present (expected; Node routes only) |

## Files (viewer only)

Added / primary:

- `src/lib/kline-view-adapter.ts` + test
- `src/lib/replay.ts` + test
- `src/components/WaveLaneChart.tsx`
- `src/components/ReplayBar.tsx`
- `src/components/ThemeProvider.tsx`
- `src/components/LabApp.tsx` (Fin shell + four lanes)
- `src/app/globals.css` (Fin shell + wave-grid / replay)

Unchanged contract: `src/lib/mongo.ts`, `lab-mongo-provider.ts`, `/api/*`.  
Legacy `WaveChart.tsx` (SVG) remains on disk and is **not** imported.

## Explicitly not implemented

- Ehlers / I/Q / phase
- Hop-On / Hop-Off / strategy metrics
- market-time synchronization across symbols
- writes to Mongo
- generator / Fin / DSE / HACCAM edits
