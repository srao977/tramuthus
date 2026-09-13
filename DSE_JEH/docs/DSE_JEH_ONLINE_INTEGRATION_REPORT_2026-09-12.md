# DSE_JEH Online Integration Report

| Item | Value |
| --- | --- |
| Date | 2026-09-12 |
| Upstream contract | `finfeedsat.v1.IngestionService.StreamBars` |
| Live execution | EXTERNALLY UNAVAILABLE - MARKET CLOSED |
| Contract-level integration | PASS |

## Purpose

Validate that the Phase 1 ONLINE adapter uses the actual generated Fin gRPC contract, maps streamed bars to the common DSE_JEH BarEvent, and feeds the same admission and analytical interfaces used by OFFLINE mode.

## Implemented Behavior

- Connects to the configured gRPC address using the generated Fin client.
- Sends `StreamBarsRequest` with symbols, maximum bars, and finalized-only preference.
- Receives generated `finfeedsat.v1.Bar` messages.
- Maps Fin types only inside the ONLINE adapter.
- Assigns adapter-local per-symbol sequence because the upstream message has no accepted sequence.
- Reconnects up to five attempts with bounded exponential delay from 200 ms to 3 seconds.
- Preserves source, snapshot identity, instrument, OHLCV, interval, source timestamp, receipt time, final, and backfill fields where available.

## Validation Command

```powershell
go test ./internal/input/online ./internal/input ./internal/admission ./internal/analytical
```

The generated-contract test starts an in-process implementation of the actual Fin `IngestionService`, streams generated `Bar` messages through the DSE_JEH consumer, and verifies mapping and delivery. The test passed as part of `go test ./...`.

## Live Result

Live market data was unavailable on 2026-09-12 because the market was closed. Per task direction, runtime validation used `DSE_JEH_MODE=OFFLINE` and the actual `bar_sequence_db.bar_sequence` collection. No false live-success claim is made.

## Upstream Limitations

The current stream does not provide an accepted sequence, detectable queue-loss marker, resumable cursor, or guaranteed finalized-only delivery semantics. Adapter-local sequencing cannot prove upstream continuity. These limitations are surfaced here and do not cause source-mode branching in the analytical core.

## Failures And Next Actions

No compile or contract-test failures occurred. During the next market-open validation window:

1. Start the existing Fin service with valid upstream credentials.
2. Run the built DSE_JEH executable with `DSE_JEH_MODE=ONLINE` and configured symbols.
3. Confirm connection, received bars, admission evidence, initialization progression, observable evidence after sufficient history, and bounded reconnect diagnostics.
