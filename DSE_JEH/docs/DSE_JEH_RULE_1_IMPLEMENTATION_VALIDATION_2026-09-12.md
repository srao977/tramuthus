# DSE_JEH Rule #1 Implementation Validation

| Item | Value |
| --- | --- |
| Date | 2026-09-12 |
| Rule | 63-CONTIGUOUS-BAR ELIGIBILITY |
| Mode | OFFLINE |
| Selected collection run | `20260911T161623Z-1` |
| Embedded run timestamp | 2026-09-11 16:16:23 UTC |
| Status | PASS |

## Scope

This change makes Phase 1 production eligibility explicit and requires OFFLINE execution to select one collection run. It does not change JEH mathematics, the 63-observation lookback, the common analytical path, the protobuf contract, ONLINE input behavior, or any Phase 2 strategy behavior.

The selected OFFLINE path remains:

```text
selected Mongo collection run
    -> OFFLINE observation mapper
    -> common BarEvent
    -> common admission
    -> per-symbol analytical state
    -> JEH Solver.Update
    -> PhaseEvidence
```

## Rule #1

For each symbol, admitted causal observations advance an independent contiguous-valid-bar count and the existing JEH state:

```text
contiguous valid bars 1..63 -> INITIALIZING; phase unavailable
contiguous valid bar 64     -> OBSERVABLE; phase production eligible
contiguous valid bar 65+    -> OBSERVABLE; phase production eligible
```

Continuity is determined by the symbol-scoped causal sequence. Elapsed source or receipt time does not reset history. Duplicate, conflict, gap, out-of-order, invalid, and rejected candidates do not mutate analytical state. No missing observations are synthesized or interpolated.

## Code Changes

| Area | Change |
| --- | --- |
| `internal/config` | Added canonical `DSE_JEH_OFFLINE_COLLECTION_RUN_ID`; retained `DSE_JEH_COLLECTION_RUN_ID` as a compatible alias; rejected conflicting values; required a non-empty selector in OFFLINE mode; kept ONLINE isolated from OFFLINE-only configuration. |
| `internal/adapter/mongo` | Rejected empty `collection_run_id`; Mongo queries now always include an exact `collection_run_id` filter. |
| `internal/input/offline` | Rejected a missing selector before repository access and tested mixed-run source selection. |
| `internal/analytical` | Added explicit `ProductionEligibilityBar = Lookback + 1` and a per-scope contiguous admitted-bar count. Phase becomes OBSERVABLE only at count 64 or later with an available solver result. |
| `scripts/Start-DSEJEHTransSat1.ps1` | Added `-OfflineCollectionRunID` and deterministic OFFLINE preflight validation. |
| Existing system design | Added Rule #1 in place; no new design version was created. |

The authoritative proto already distinguishes INITIALIZING from OBSERVABLE and represents phase as optional. It was not changed. Zero degrees remains valid when `phase_degrees` is present.

## Configuration

Canonical environment configuration:

```powershell
$env:DSE_JEH_MODE = 'OFFLINE'
$env:DSE_JEH_OFFLINE_COLLECTION_RUN_ID = '20260911T161623Z-1'
```

Normal startup script invocation:

```powershell
.\scripts\Start-DSEJEHTransSat1.ps1 `
    -Mode OFFLINE `
    -OfflineCollectionRunID '20260911T161623Z-1'
```

OFFLINE startup fails if neither the canonical variable nor its existing compatibility alias supplies a run ID. If both are supplied with different values, startup fails. An empty selector also fails at the producer and Mongo query boundaries; there is no silent all-run fallback.

## Deterministic Tests

| Requirement | Test/result |
| --- | --- |
| Exact boundary | Updates 1-63 are INITIALIZING with absent phase; update 64 is OBSERVABLE with present phase. PASS. |
| Non-linear time | Synthetic sequence 1-64 uses increasingly large source/receipt gaps and transitions at 64 without reset. PASS. |
| Symbol isolation | AAPL reaching 64 does not advance MSFT, whose first bar remains INITIALIZING. PASS. |
| Sparse symbol | A 51-bar VXX sequence remains entirely INITIALIZING with absent phase. PASS. |
| Collection filtering | Mixed records from all three runs emit only the configured run. PASS. |
| Missing selector | OFFLINE config and producer reject an empty collection run. PASS. |
| Admission integrity | Existing duplicate, conflict, gap, and sequencing tests pass unchanged. PASS. |
| ONLINE isolation | Conflicting OFFLINE-only variables do not affect ONLINE configuration. PASS. |

## Build And Execution

No `go run` command was used.

| Validation | Result |
| --- | --- |
| `go test ./...` | PASS: 13 packages; 8 tested, 5 without test files |
| `go build ./...` | PASS |
| `go build -o bin/dse-jeh-transsat-1.exe ./cmd/dse-jeh-transsat-1` | PASS |
| Normal executable through startup script | PASS, exit code 0 |
| Executable size | 23,309,824 bytes |

## Runtime Classification

| Metric | Count |
| --- | ---: |
| Source/read | 3,120 |
| Admitted | 3,120 |
| INITIALIZING | 1,848 |
| OBSERVABLE | 1,272 |
| INVALID | 0 |
| Rejected | 0 |
| Duplicates | 0 |
| Conflicts | 0 |
| Gaps | 0 |
| Out-of-order | 0 |

The output contains exactly 3,120 PhaseEvidence records for 3,120 admitted source observations. No synthetic records were created. The output digest is `9be155eff0627ad056b60c8c93e721b2811366e5ce2ff5d2243b65e5a98c9c6c`.

## Per-Symbol Validation

`Contiguous` is the verified length of the exact sequence `1..N`. A dash means the symbol never reaches an observable result.

| Symbol | Admitted | Contiguous | INITIALIZING | OBSERVABLE | First seq | Last seq | First observable seq | First received UTC | Last received UTC |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- | --- |
| AAPL | 120 | 120 | 63 | 57 | 1 | 120 | 64 | 2026-09-11 16:17:01.032 | 2026-09-11 18:16:00.941 |
| AMD | 113 | 113 | 63 | 50 | 1 | 113 | 64 | 2026-09-11 16:17:01.027 | 2026-09-11 18:16:00.936 |
| AMZN | 119 | 119 | 63 | 56 | 1 | 119 | 64 | 2026-09-11 16:17:01.041 | 2026-09-11 18:16:00.941 |
| AVGO | 118 | 118 | 63 | 55 | 1 | 118 | 64 | 2026-09-11 16:17:01.035 | 2026-09-11 18:16:00.941 |
| BAC | 117 | 117 | 63 | 54 | 1 | 117 | 64 | 2026-09-11 16:17:01.035 | 2026-09-11 18:16:00.935 |
| COST | 80 | 80 | 63 | 17 | 1 | 80 | 64 | 2026-09-11 16:20:01.023 | 2026-09-11 18:16:00.944 |
| DIA | 33 | 33 | 33 | 0 | 1 | 33 | - | 2026-09-11 16:17:01.027 | 2026-09-11 18:14:00.939 |
| EEM | 99 | 99 | 63 | 36 | 1 | 99 | 64 | 2026-09-11 16:17:01.027 | 2026-09-11 18:16:00.941 |
| GLD | 87 | 87 | 63 | 24 | 1 | 87 | 64 | 2026-09-11 16:17:01.037 | 2026-09-11 18:16:00.941 |
| GOOGL | 115 | 115 | 63 | 52 | 1 | 115 | 64 | 2026-09-11 16:17:01.037 | 2026-09-11 18:16:00.945 |
| HD | 88 | 88 | 63 | 25 | 1 | 88 | 64 | 2026-09-11 16:18:00.915 | 2026-09-11 18:16:00.944 |
| IWM | 119 | 119 | 63 | 56 | 1 | 119 | 64 | 2026-09-11 16:17:01.032 | 2026-09-11 18:16:00.941 |
| JNJ | 86 | 86 | 63 | 23 | 1 | 86 | 64 | 2026-09-11 16:19:00.908 | 2026-09-11 18:15:00.937 |
| JPM | 105 | 105 | 63 | 42 | 1 | 105 | 64 | 2026-09-11 16:17:01.032 | 2026-09-11 18:16:00.941 |
| MA | 89 | 89 | 63 | 26 | 1 | 89 | 64 | 2026-09-11 16:17:01.027 | 2026-09-11 18:16:00.945 |
| META | 120 | 120 | 63 | 57 | 1 | 120 | 64 | 2026-09-11 16:17:01.027 | 2026-09-11 18:16:00.944 |
| MSFT | 114 | 114 | 63 | 51 | 1 | 114 | 64 | 2026-09-11 16:17:01.027 | 2026-09-11 18:16:00.935 |
| NFLX | 120 | 120 | 63 | 57 | 1 | 120 | 64 | 2026-09-11 16:17:01.034 | 2026-09-11 18:16:00.935 |
| NVDA | 120 | 120 | 63 | 57 | 1 | 120 | 64 | 2026-09-11 16:17:01.027 | 2026-09-11 18:16:00.935 |
| QQQ | 116 | 116 | 63 | 53 | 1 | 116 | 64 | 2026-09-11 16:17:01.037 | 2026-09-11 18:16:00.936 |
| SPY | 120 | 120 | 63 | 57 | 1 | 120 | 64 | 2026-09-11 16:17:01.037 | 2026-09-11 18:16:00.936 |
| TLT | 103 | 103 | 63 | 40 | 1 | 103 | 64 | 2026-09-11 16:17:01.027 | 2026-09-11 18:16:00.944 |
| TSLA | 120 | 120 | 63 | 57 | 1 | 120 | 64 | 2026-09-11 16:17:01.034 | 2026-09-11 18:16:00.948 |
| UNH | 116 | 116 | 63 | 53 | 1 | 116 | 64 | 2026-09-11 16:17:01.027 | 2026-09-11 18:16:00.935 |
| V | 101 | 101 | 63 | 38 | 1 | 101 | 64 | 2026-09-11 16:18:00.915 | 2026-09-11 18:16:00.944 |
| VXX | 51 | 51 | 51 | 0 | 1 | 51 | - | 2026-09-11 16:22:01.038 | 2026-09-11 18:16:00.945 |
| WMT | 116 | 116 | 63 | 53 | 1 | 116 | 64 | 2026-09-11 16:17:01.027 | 2026-09-11 18:16:00.936 |
| XLF | 118 | 118 | 63 | 55 | 1 | 118 | 64 | 2026-09-11 16:18:00.852 | 2026-09-11 18:16:00.944 |
| XLK | 91 | 91 | 63 | 28 | 1 | 91 | 64 | 2026-09-11 16:18:00.915 | 2026-09-11 18:16:00.941 |
| XOM | 106 | 106 | 63 | 43 | 1 | 106 | 64 | 2026-09-11 16:18:00.915 | 2026-09-11 18:16:00.945 |
| **Total** | **3,120** | **3,120** | **1,848** | **1,272** |  |  |  |  |  |

All 30 symbols start at sequence 1 and contain every integer sequence through their final value exactly once. Twenty-eight symbols reach bar 64. DIA and VXX are the two valid sparse histories that remain INITIALIZING.

## Reference Equivalence

| Metric | Result |
| --- | ---: |
| Tolerance | `1e-9` |
| Reference rows | 3,480 |
| Matched reference rows | 3,300 |
| Missing runtime rows | 180 |
| Status mismatches | 0 |
| Observable values compared | 1,272 |
| Numeric mismatches | 0 |
| Maximum phase difference | 0 |

The existing reference CSV contains 180 rows from earlier run `20260910T191246Z-1` and 3,300 rows labeled with the selected run. The selected-run reference portion contains 3,120 unique `(collection_run_id, symbol, generator_sequence_no)` identities and 180 duplicate identity rows. The existing comparator evaluates every reference row against its identity-indexed runtime map, which explains 3,300 matched reference rows for 3,120 runtime rows. The 180 missing rows are from the unselected earlier run. All overlapping status and numerical comparisons pass exactly.

The reference generator remains validation-only and is not a runtime dependency.

## Deviations

- No implementation deviation from Rule #1 was found.
- The independent reference CSV contains 180 duplicate selected-run identities; this affects comparator row accounting but produces no status or numeric mismatch.
- Live ONLINE execution was not repeated. ONLINE production code was not changed, and its existing tests pass.
- No proto generation was needed because the authoritative contract already represents the required states.

## Final Console Summary

```text
DSE_JEH RULE #1 IMPLEMENTATION VALIDATION

Rule:
63-CONTIGUOUS-BAR ELIGIBILITY

OFFLINE collection_run_id:
20260911T161623Z-1

Source bars: 3120
Admitted: 3120
INITIALIZING: 1848
OBSERVABLE: 1272
INVALID: 0
Rejected: 0
Duplicates: 0
Conflicts: 0
Gaps: 0
Out-of-order: 0

Symbols: 30
Symbols reaching bar 64: 28
Symbols remaining INITIALIZING: 2

DIA:
    admitted: 33
    INITIALIZING: 33
    OBSERVABLE: 0

VXX:
    admitted: 51
    INITIALIZING: 51
    OBSERVABLE: 0

Bar 63 production eligible:
NO

Bar 64 production eligible:
YES

Wall-clock gap causes reset:
NO

Sequence continuity governs eligibility:
YES

Reference status mismatches: 0
Reference numeric mismatches: 0
Maximum phase difference: 0

go test ./...: PASS
go build ./...: PASS
Executable run: PASS

ONLINE behavior changed:
NO

Phase 2 implemented:
NO
```