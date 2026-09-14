# Capital Reservoir Viewer User Guide

Date: 2026-09-14

This guide starts the read-only viewer before launching one governed DSE_JEH Dynamic Pipeline Execution Run.

## One-Time Prerequisites

1. MongoDB must be running at `mongodb://127.0.0.1:27017` unless server-only environment overrides are configured.
2. The required collections and indexes must already exist.
3. Before using `RUN_C` or another generic run label, manually run this script in MongoDB Compass Shell or `mongosh` and confirm `FINAL RESULT: PASS`:

   `C:\Users\chino\tramuthus\DSE_JEH\scripts\update_run_type_validation_2026-09-14.js1`

   The application does not execute this script. It preserves documents and indexes while updating `run_type` validation for `h_h_stage_emit_values` and `capital_reservoir_events`.
4. Build DSE_JEH after software changes. A new run label alone does not require another build:

```powershell
cd C:\Users\chino\tramuthus\DSE_JEH
.\scripts\Build-DSEJEHTransSat1.ps1
```

## Start The Viewer

Open **Terminal 1** and leave it running:

```powershell
cd C:\Users\chino\tramuthus\Capital_Reservoir_Viewer
npm run dev -- -H 127.0.0.1 -p 3010
```

Wait for Next.js to print:

```text
Ready
Local: http://127.0.0.1:3010
```

## Open The Browser

Open this address:

```text
http://127.0.0.1:3010
```

Port `3010` is the browser address. Port `50052` is the DSE_JEH gRPC endpoint used internally by the viewer server; do not open `50052` in the browser.

The viewer opens in **REPLAY** mode. Select **LIVE** before starting the pipeline. If LIVE reports that the runtime is unavailable, leave the page open and use the reconnect button after DSE_JEH reports that its gRPC endpoint is listening.

## Start A Pipeline Run

Open **Terminal 2** in DSE_JEH. Supply a new unique `PipelineRunID` for every execution:

```powershell
cd C:\Users\chino\tramuthus\DSE_JEH

.\scripts\Start-DSEJEHTransSat1.ps1 `
    -Mode OFFLINE `
    -OfflineCollectionRunID '20260911T161623Z-1' `
    -PipelineRunID 'DPE-GOVERNED-RUN-C-20260914-001' `
    -RunType RUN_C `
    -StartingCapital 100000 `
    -AllocationPct 1.0 `
    -RiskR 1.0
```

Keep Terminal 2 open. The launcher prints the process ID and runtime endpoint. When it reports `127.0.0.1:50052` as listening, press the viewer's reconnect button if the LIVE connection previously showed a fault.

`RunType` and `RiskR` are independent. For example, a later run can use:

```powershell
-PipelineRunID 'DPE-GOVERNED-RUN-D-20260914-001' `
-RunType RUN_D `
-RiskR 0.20
```

Valid `RunType` values match:

```text
^[A-Z0-9][A-Z0-9_-]*$
```

Valid Risk R values satisfy $0 \leq RiskR \leq 1$. Multiple runs may use the same Risk R and source collection, but every execution must have a different `PipelineRunID`.

## Reconnect Live Mode

If LIVE was selected before DSE_JEH was available, wait until Terminal 2 reports that `127.0.0.1:50052` is listening, then press the reconnect button in the viewer. The browser remains at `http://127.0.0.1:3010`.

## Use Live Mode

While DSE_JEH publishes Capital Reservoir events, LIVE mode updates:

- common reservoir metrics and continuity chart;
- selected-symbol signed flow;
- all 30 pipe states at the crosshair event sequence; and
- the normalized event inspector.

Moving either chart crosshair selects the related event sequence. Selecting a pipe changes the signed-flow chart to that symbol.

## Use Replay Mode

After the run has persisted events:

1. Select **REPLAY**.
2. Select the required `pipeline_run_id` from **PIPELINE RUN**.
3. Use Play, Pause, Restart, Step, `1x`, `5x`, `10x`, or `MAX`.

Replay reads `bar_sequence_db.capital_reservoir_events` and does not write to MongoDB. Use replay to inspect the complete run if the live subscription connected after early events were published.

## Shutdown

### Stop DSE_JEH

Open **Terminal 3**:

```powershell
cd C:\Users\chino\tramuthus\DSE_JEH
.\scripts\Stop-DSEJEHTransSat1.ps1
```

This requests a graceful shutdown through the governed stop file.

### Stop The Viewer

Return to Terminal 1 and press `Ctrl+C`. Stopping the viewer does not stop DSE_JEH or MongoDB.

## Quick Sequence

1. Apply the Mongo validator migration once and confirm PASS.
2. Start the Next.js viewer on port `3010`.
3. Open `http://127.0.0.1:3010`.
4. Select LIVE.
5. Start DSE_JEH from a second terminal with explicit run values.
6. Reconnect LIVE after `50052` is listening if required.
7. Inspect the run live, then use REPLAY for the complete persisted record.
8. Stop DSE_JEH gracefully and stop the viewer with `Ctrl+C`.