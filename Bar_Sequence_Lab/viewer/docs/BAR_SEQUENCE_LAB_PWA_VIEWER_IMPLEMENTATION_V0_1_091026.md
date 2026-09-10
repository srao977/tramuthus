# Bar Sequence Lab — portable Next.js PWA viewer V0.1

Date: 2026-09-10  
Module: `Bar_Sequence_Lab/viewer`  
Verdict: **FUNCTIONALLY REAL V0.1** (read-only Mongo waves; independent of Fin / DSE)

The ordered bar sequence **is** the wave.  
x-axis = `generator_sequence_no` (accepted-arrival per `collection_run_id` + symbol).  
Price and Volume are independent trajectories.  
Maths / Ehlers-Hilbert / Hop / DSE science were **not** implemented.

Fin_FeedSat_1, Fin_FeedSat_1_Viewer, DSE_TransSat_1_viewer, DSE_TransSat_1_worker, HACCAM, and `Bar_Sequence_Lab/generator` were **not** modified.  
Not committed. Not pushed.

## Reference inspection (read-only)

| Path | Result |
| --- | --- |
| `C:\Users\chino\Fin_FeedSat_1_Viewer` | Next **16.3.3**, React 19, Tailwind 4, App Router, `src/`. Charts: **klinecharts** (time-axis candlesticks). Theme: `next-themes` + custom CSS tokens. **No** web manifest, **no** service worker, **no** PWA. APIs are Node `force-dynamic` relays to Fin gRPC — observation-only. |
| `C:\Users\chino\tramuthus\DSE_TransSat_1_viewer` | Empty folder. No Next/PWA/chart stack to copy. |

This viewer is an **independent** Next.js module. It does not import Fin or DSE packages, proto, or runtime URLs. Visual language is a light lab identity (teal/gold), not a Fin dashboard clone. Chart is a sequence polyline, not klinecharts, because the x-axis must be bar index, not timestamp.

## Scaffold

`create-next-app@latest` in `Bar_Sequence_Lab/viewer` after moving the empty `docs/` hold folder aside.

| Item | Value |
| --- | --- |
| Next | 16.3.4 |
| React | 19.2.8 |
| TypeScript | yes |
| Tailwind | 4 |
| App Router | yes (`src/app`) |
| ESLint | yes |
| package name | `bar-sequence-lab-viewer` |

Added: `mongodb` (server-only driver), `server-only`.

Did **not** add `next-pwa` (obsolete relative to this Next version). PWA is Next `metadata` + `public/manifest.webmanifest` + `public/sw.js` maintained for Next 16.

## PWA identity

| Field | Value |
| --- | --- |
| name | Bar Sequence Lab |
| short_name | BarSeq Lab |
| display | standalone |
| theme_color | `#1d4e4a` (root `viewport.themeColor` + manifest) |
| background_color | `#f4f6f5` |
| icons | locally generated `public/icons/icon-192.png`, `icon-512.png`, `icon-maskable-512.png` (`npm run icons`) |
| SW | `public/sw.js` registered from `PwaRegister` |

Service worker caches static shell (`/`, manifest, icons). Observation `/api/*` is **network-first** and is **not** written into Cache Storage. Scientific/raw series are never SW-cached.

Installability on HTTP LAN is best-effort. Chromium generally requires HTTPS (or localhost) for install prompts. No device-specific PWA forks.

## Architecture

Portable types in `src/lib/types.ts`:

- `CollectionRunSummary`
- `GroupDescriptor`
- `SymbolDescriptor`
- `TrajectoryDescriptor` (`price` \| `volume`)
- `TrajectoryObservation`
- `TrajectorySeries`
- `SequenceSummary`

`BarSequenceDataProvider` (`src/lib/provider.ts`) is the portable contract.  
`labMongoProvider` is the Lab Mongo adapter (`src/lib/lab-mongo-provider.ts`), imported only from App Router route handlers (`import "server-only"`).

Browser talks only to same-origin `/api/*`. Browser does **not** open Mongo. Env vars are **not** `NEXT_PUBLIC_*`.

### Env (server-only)

```
BAR_SEQ_LAB_MONGO_URI=mongodb://127.0.0.1:27017
BAR_SEQ_LAB_MONGO_DB=bar_sequence_db
BAR_SEQ_LAB_MONGO_COLLECTION=bar_sequence
```

See `.env.example`. Defaults match the proved Lab collection. Viewer never writes.

### APIs

| Method | Path | Notes |
| --- | --- | --- |
| GET | `/api/runs` | Distinct `collection_run_id`, newest lexicographic first (not ObjectId) |
| GET | `/api/runs/{run}/groups` | Distinct `partition_id` from Mongo |
| GET | `/api/runs/{run}/symbols?group=` | Optional A/B/C filter; omit or ALL = all symbols |
| GET | `/api/runs/{run}/trajectory?symbol=&type=price\|volume` | Sorted `generator_sequence_no` ASC |

All handlers: `runtime = "nodejs"`, `dynamic = "force-dynamic"`, `Cache-Control: no-store`.

y for `price` = `close`. y for `volume` = `volume`. Not combined. Not OHLC-derived science.

## UI

- `RunSelector` — `<select>` of Mongo runs; default newest `collection_run_id` lexicographically (`20260910T191246Z-1` before `20260910T190625Z-1`)
- `GroupSelector` — `[ALL][A][B][C]` from Mongo partitions
- `SymbolSelector` — symbols in the selected group
- `TrajectorySelector` — Price \| Volume
- `WaveChart` — SVG polyline; x = `generatorSequenceNo`; click/tap selects a point (not hover-only)
- `ObservationDetails` — raw fields for the selected observation
- `SequenceSummary` — count, min/max seq, contiguous 1..N, partition

Responsive: chart dominant on desktop; stacked at 768px; compact at 390px. Light lab theme. No Fin dark cockpit.

## Validation against `20260910T191246Z-1`

Production server: `npm run start -- -H 0.0.0.0 -p 3000`  
Mongo: `bar_sequence_db.bar_sequence` @ `127.0.0.1:27017` (read-only)

| Check | Result |
| --- | --- |
| GET `/` | 200; title Bar Sequence Lab; no `mongodb://`; no `BAR_SEQ_LAB_MONGO` in HTML |
| manifest | name Bar Sequence Lab, short_name BarSeq Lab, display standalone |
| icons / sw.js | 200; SW bypasses `/api/` |
| GET `/api/runs` | `20260910T191246Z-1` first (1393 obs, 30 symbols, A,B,C); prior prove run 6 obs |
| groups | A 10/484, B 10/463, C 10/446 |
| symbols ALL | 30 |
| symbols A | AAPL,AMD,AMZN,AVGO,GOOGL,META,MSFT,NFLX,NVDA,TSLA |
| symbols C | BAC,COST,HD,JNJ,JPM,MA,UNH,V,WMT,XOM |
| AAPL price | n=51, seq 1..51, contiguous, partition A |
| AAPL volume | n=51, independent y (volume), same x |
| MSFT price | n=48, 1..48, A |
| SPY price | n=48, 1..48, B |
| COST volume | n=32, 1..32, C |
| client bundle | no `mongodb://` / `BAR_SEQ_LAB_MONGO` / `27017` in `.next/static/chunks` |
| `npm run build` | pass (Next 16.3.4) |
| `npm test` | 3 pass (newest-run lexicographic; 1..N; Long coerce) |
| `npm run lint` | pass after effect/unused-var fixes |
| `tsc --noEmit` | pass |

Matches generator recon report: AAPL/TLT 51 … COST 32; totals 484+463+446=1393.

## URLs

| Surface | URL |
| --- | --- |
| Desktop / localhost | http://127.0.0.1:3000 |
| Next printed local | http://localhost:3000 |
| LAN bind | `0.0.0.0:3000` |
| Host Wi-Fi IPv4 | http://192.168.0.21:3000 |
| Host VirtualBox IPv4 | http://192.168.56.1:3000 |

Install-to-homescreen typically needs HTTPS or localhost. HTTP LAN viewing works; install prompt may not. No per-device PWA logic.

## Out of scope (held)

- Maths, Ehlers, Hilbert I/Q, phase, Hop-On/Hop-Off
- Total Return / Active Position / Engine Signal / Portfolio Wave Vector
- Alpha/Beta/Gamma/Delta naming
- Fake sinusoids / invented Wave Creation transform
- Combining price×volume
- SSE / WebSocket / change streams
- Cloudflare
- CSV
- Generator changes
- Writes to Mongo
- Fin_FeedSat_1 / DSE / HACCAM edits

JSONL audit path remains on the generator; viewer reads Mongo only.

## How to run

From `Bar_Sequence_Lab/viewer`:

```
npm install
npm run build
npm run start -- -H 0.0.0.0 -p 3000
```

Dev: `npm run dev -- -H 0.0.0.0 -p 3000`

Mongo must be reachable with the server-only env vars above.

## Git

`git diff --check` / `git status --short` after this work. **Not committed. Not pushed.**
