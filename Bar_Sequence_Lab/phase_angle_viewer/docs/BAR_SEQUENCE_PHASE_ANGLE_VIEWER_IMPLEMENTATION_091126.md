# Bar Sequence Phase Angle Viewer Implementation — 2026-09-11

## Purpose

`phase_angle_viewer` is an independent, presentation-only Next.js PWA for inspecting persisted Bar Sequence Lab phase-angle records. It does not replay bars, calculate phase, write MongoDB, or attach trading and strategy semantics.

## Inputs and experiment identity

The viewer reads these MongoDB collections from `bar_sequence_db`:

- `bar_sequence_phase_angle_series`: derived phase records.
- `bar_sequence`: raw lineage retrieved on demand for a selected point.

Every series query is constrained by `collection_run_id`, `series_size`, `solver_name`, `solver_version`, `input_series_type`, `partition_id`, and `symbol`. The initial preferred experiment is run `20260911T161623Z-1`, series size `120`, solver `EHLERS_DOMINANT_CYCLE_PHASE` V0.1, input `MEDIAN_PRICE`.

## Architecture

The browser calls read-only Next.js API routes. Node-runtime server providers own MongoDB access and configuration; credentials are not exposed as public environment variables.

- `/api/phase/experiments`: discover persisted experiment identities.
- `/api/phase/symbols`: list symbol counts for one complete experiment identity.
- `/api/phase/series`: load one partition/symbol series.
- `/api/raw-observation`: retrieve one raw observation by run, symbol, and generator sequence.

## Presentation semantics

- Three panels provide one selected symbol for each A/B/C partition.
- The complete selected series loads directly; there is no player, animation, or replay cursor.
- `generator_sequence_no` is the analytical x-coordinate and is never presented as physical time.
- Only records with `phase_observable=true` and a numeric `phase_angle_degrees` enter chart data.
- Initialization nulls are omitted without renumbering and never converted to zero.
- Crosshair tooltips carry generator sequence, persisted angle, validity state, and series size.
- Selecting a point retrieves its raw OHLCV and provenance on demand.
- Symbols with zero observable records display an explicit initialization-only state.

## Validation scope

Infrastructure, MongoDB persistence, filtering, coordinate preservation, null handling, API behavior, and visual inspection are validated as viewer concerns. The viewer displays the frozen Lab V0.1 solver output and does not claim independent scientific validation of the John Ehlers implementation. That remains a separate analytical activity.

## Operational limits

- MongoDB must be reachable from the Next.js server.
- The viewer is read-only but does not implement authentication or authorization.
- Chart coordinates are intentionally synthetic rendering coordinates equal to generator sequence numbers.
- Raw lineage depends on the source observation remaining available in `bar_sequence`.
