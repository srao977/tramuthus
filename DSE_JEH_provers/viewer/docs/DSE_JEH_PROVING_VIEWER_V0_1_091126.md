# DSE_JEH Proving Viewer

**Date:** 2026-09-11  
**Version:** V0.1  
**Status:** EXPERIMENTAL / PROVING  
**Implementation:** AUTHORIZED FOR DSE_JEH PROVING VIEWER ONLY

## 1. Purpose

This viewer provides a practical, read-only inspection surface for evidence emitted by the Go DSE_JEH CSV proving engine. It exists to make current phase state, history, transitions, candidate events, decisions, and source lineage inspectable without changing the authoritative evidence.

## 2. Scope

Implementation is limited to `DSE_JEH/viewer`. The viewer does not modify Bar Sequence Lab, Fin_FeedSat_1, HACCAM, or either DSE_TransSat_1 system.

## 3. Authority Boundary

The Go DSE_JEH implementation is authoritative. TypeScript loads and presents persisted records. It does not calculate phase, classify zones, advance strategy state, rank candidates, or create decisions.

## 4. Current Data Flow

The proving flow is:

```text
phase-angle CSV
    -> Go DSE_JEH proving engine
    -> deterministic local evidence files
    -> Next.js proving viewer
```

## 5. Evidence Root

The server reads `DSE_JEH_EVIDENCE_ROOT` when set. Otherwise it resolves `../output` from the viewer process. A missing root is valid and produces a clear no-data state.

## 6. Run Discovery

Only directories matching `csv-prove-[a-f0-9]{16}` are admitted. `run_metadata.json` is loaded first, and its `run_id` must match the directory name.

## 7. Required Artifacts

Each run contains:

- `admitted_phase_input_evidence.jsonl`
- `strategy_state_evidence.jsonl`
- `decision_evidence.jsonl`
- `final_entity_states.json`
- `run_metadata.json`

## 8. Run Selection

The run selector writes `?run=<run_id>` to the URL. The server then reloads all five artifacts for that identity, preventing evidence from different runs from being combined in one view.

## 9. Polar Phase Diagram

The primary current-state visualization is SVG. Zero degrees is at the top, angles increase clockwise, 90 degrees is right, 180 degrees is bottom, and 270 degrees is left. Every observable entity uses the same radius, so radial position carries no rank or magnitude meaning.

## 10. Zone Presentation

Zone color and labels display the persisted `current_zone` value. Sector graphics are orientation aids only. The viewer does not reproduce the Go classifier or infer a zone from an angle.

## 11. Initialization Handling

An entity is shown as initializing when it occurs in admitted input but has no persisted final observable state. Initializing entities are listed outside the phase circle and are never plotted at zero degrees.

## 12. Entity Selection

An observable entity can be selected from the phase circle. Any admitted entity can be selected from the entity menu. Selection updates details, history, and event evidence without changing the loaded run.

## 13. Linear Phase History

Klinecharts renders selected-entity history with `generator_sequence_no` as the analytical x value and persisted `phase_angle_degrees` as the y value. Only observable records are supplied. Equal OHLC values are a rendering adapter for the chart library, not synthesized market data.

## 14. Circular Boundary

Phase remains normalized to `[0,360)`, where zero and 360 degrees are equivalent. The viewer does not smooth or unwrap history, so transitions such as 359 to 1 degrees remain visible in the linear chart.

## 15. Current State Detail

The entity panel presents current and previous phase, current and previous zone, sequence, validity, solver identity, input series, collection run, partition, analysis time, source identity, and references to persisted transition, candidate, and decision evidence.

## 16. Event and Decision History

The event table combines persisted strategy and decision records for display ordering. It does not infer transitions from repeated zone membership and does not promote candidate evidence into an execution order.

## 17. Ranking Limitation

`PROVING_ONLY_ORDER` is displayed verbatim. It is a stable proving presentation order and is not scientific ranking, portfolio priority, expected return, or execution preference.

## 18. Undefined Measures

Phase velocity is explicitly `NOT YET DEFINED`. The viewer does not calculate velocity, acceleration, profitability, positions, portfolio state, orders, fills, or P&L.

## 19. Local Operation

Install and validate from `DSE_JEH/viewer`:

```powershell
npm install
npm test
npm run lint
npx tsc --noEmit
npm run build
npm run dev
```

To use another deterministic evidence root:

```powershell
$env:DSE_JEH_EVIDENCE_ROOT = "C:\path\to\output"
npm run dev
```

## 20. Future Live Evolution

MongoDB is not part of this CSV proving viewer. The approved Foundation Design retains MongoDB for a later live, event-driven DSE_JEH runtime where durable audit, recovery, and replay are required. That future system should replace the repository adapter while preserving the viewer's observational boundary; it must not move scientific or strategy authority into TypeScript.