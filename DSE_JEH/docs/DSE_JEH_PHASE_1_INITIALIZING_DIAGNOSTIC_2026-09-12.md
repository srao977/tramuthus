# DSE_JEH Phase 1 INITIALIZING Diagnostic

| Item | Value |
| --- | --- |
| Date | 2026-09-12 |
| Scope | Focused OFFLINE Phase 1 diagnostic; no production behavior change |
| Evidence | `exports/initializing_diagnostic.jsonl` |
| Dataset | `bar_sequence_db.bar_sequence` |

## Executive Finding

The current built runtime does **not** emit 3,712 `INITIALIZING` records. It reads and admits 4,519 bars and emits **3,247 INITIALIZING** plus **1,272 OBSERVABLE** records. Invalid, rejected, duplicate, conflict, gap, and out-of-order counts are all zero.

The 465-record difference from 3,712 cannot be attributed to the current code and current dataset because the reproduced run and complete evidence classification account for every admitted bar:

```text
3,247 INITIALIZING + 1,272 OBSERVABLE = 4,519 admitted
3,712 - 3,247 = 465
```

The current 3,247 count is caused by repeated, intentional warm-up per analytical state key. The key is `source_provenance.entity_sequence_scope`, which OFFLINE constructs as:

```text
collection_run_id + "|" + normalized_symbol
```

There are 30 symbols but 63 such scopes across three collection runs. Each scope owns a fresh `jeh.Solver`, and each solver's first 63 admitted updates return no phase. Therefore:

```text
actual INITIALIZING = SUM over 63 scopes of min(admitted_in_scope, 63) = 3,247
```

The requested one-warm-up-per-symbol calculation is:

```text
expected INITIALIZING = SUM over 30 symbols of min(total_admitted_for_symbol, 63) = 1,890
excess under that comparison = 3,247 - 1,890 = 1,357
```

## Reproduced Runtime Counts

The normal executable was rebuilt with `go build` and run with `DSE_JEH_MODE=OFFLINE`; `go run` was not used.

| Classification | Count |
| --- | ---: |
| Total source bars | 4,519 |
| Admitted | 4,519 |
| INITIALIZING | 3,247 |
| OBSERVABLE | 1,272 |
| INVALID | 0 |
| Rejected | 0 |
| Duplicates | 0 |
| Conflicts | 0 |
| Gaps | 0 |
| Out of order | 0 |

## Authoritative INITIALIZING Decision

The only production decision that assigns `INITIALIZING` to a new `PhaseEvidence` is `analytical.Coordinator.Process` in `internal/analytical/coordinator.go`:

```go
result, err := state.solver.Update(event.GetHigh(), event.GetLow())
status := dsejehv1.PhaseStatus_PHASE_STATUS_INITIALIZING
var phaseDegrees *float64
if err != nil {
    status = dsejehv1.PhaseStatus_PHASE_STATUS_INVALID
} else if result.PhaseAngle != nil {
    status = dsejehv1.PhaseStatus_PHASE_STATUS_OBSERVABLE
    phaseDegrees = result.PhaseAngle
}
```

Exact pseudocode:

```text
result = scope_solver.Update(high, low)
status = INITIALIZING
if result has error:
    status = INVALID
else if result.PhaseAngle is present:
    status = OBSERVABLE
# otherwise status remains INITIALIZING
```

`jeh.Solver.Update` controls phase presence:

```text
index = solver.count
solver.count++
...
if index < 63:
    return Result with PhaseAngle absent
return Result with PhaseAngle present
```

Thus an admitted record is `INITIALIZING` exactly when its scope-owned solver update succeeds and its pre-increment zero-based `index` is less than 63. Updates 1 through 63 have indices 0 through 62; update 64 has index 63 and is the first possible observable result.

Other `INITIALIZING` occurrences are not independent production decisions:

| Location | Role |
| --- | --- |
| `api/proto/dse_jeh/v1/DSE_JEH_TransSat_1.proto` | Declares the enum value and presence contract. |
| `gen/dse_jeh/v1/DSE_JEH_TransSat_1.pb.go` | Generated enum mapping only. |
| `internal/runtime/runtime.go`, `Engine.process` | Counts an already assigned status. |
| `internal/analytical/coordinator_test.go` | Asserts expected behavior. |
| `internal/jeh/solver_test.go` | Asserts phase absence before lookback and presence afterward. |

## Counter And State Ownership

| Property | Actual implementation |
| --- | --- |
| Counter | `jeh.Solver.count` |
| Type | `int` |
| Owner | One `jeh.Solver` inside one `analytical.entityState` |
| Initialized | Zero value when `jeh.NewSolver()` returns a fresh solver |
| Incremented | Once at the start of every successful `Solver.Update` after Median Price validation |
| Reset method | None |
| Retrieval key | `event.Provenance.EntitySequenceScope` |
| OFFLINE key construction | `analytical.Scope(collection_run_id, symbol)` |
| Actual counted quantity | Admitted solver updates per collection-run-and-symbol scope |
| Not counted | Global bars, raw `generator_sequence_no`, bars per symbol across all runs, partition, or provider |

`Coordinator.Process` returns before state lookup for non-admitted input. Therefore rejected, duplicate, conflict, gap, and out-of-order candidates do not increment `Solver.count`.

## State Creation And Reset Inventory

A fresh JEH state is created only in `analytical.Coordinator.Process` when `coordinator.entities[scope]` is absent. The state contains `solver: jeh.NewSolver()`. There is no map deletion, solver replacement, or reset method in production code.

| Trigger | Fresh solver? | Count restarts? | Status returns to INITIALIZING? |
| --- | --- | --- | --- |
| First admitted bar for a new `collection_run_id|symbol` scope | Yes | Yes, from zero | Yes |
| Collection run changes | Yes, because scope changes | Yes | Yes |
| Symbol changes | Yes, because scope changes | Yes | Yes |
| Generator sequence restarts inside a new collection run | Accompanies the new scope | Yes | Yes |
| Source/provider or partition changes alone | No; not part of the key | No | No |
| Duplicate, conflict, gap, or out-of-order input | No reset; input is not sent to solver | No | No new PhaseEvidence |
| ONLINE reconnect | No explicit reset while the same runtime/consumer/coordinator remains alive | No | No |
| OFFLINE runtime restart | Yes; a new coordinator map is constructed | Yes | Yes on the new execution |

The current run creates exactly **63 JEH solver instances**. All 30 symbols have more than one scope and therefore initialize more than once. AAPL, MSFT, and NVDA have three scopes; every other symbol has two.

# INITIALIZING Bars by Symbol

`First Seq`, `Last Seq`, and `First Observable Seq` are aggregate numeric extrema. Since sequences restart by collection run, the run-scoped ranges later in this report are authoritative.

## Alphabetical

| Symbol | Admitted | INITIALIZING | OBSERVABLE | Expected INITIALIZING | Excess | First Seq | Last Seq | First Observable Seq |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| AAPL | 173 | 116 | 57 | 63 | 53 | 1 | 120 | 64 |
| AMD | 160 | 110 | 50 | 63 | 47 | 1 | 113 | 64 |
| AMZN | 167 | 111 | 56 | 63 | 48 | 1 | 119 | 64 |
| AVGO | 166 | 111 | 55 | 63 | 48 | 1 | 118 | 64 |
| BAC | 164 | 110 | 54 | 63 | 47 | 1 | 117 | 64 |
| COST | 112 | 95 | 17 | 63 | 32 | 1 | 80 | 64 |
| DIA | 71 | 71 | 0 | 63 | 8 | 1 | 38 | N/A |
| EEM | 144 | 108 | 36 | 63 | 45 | 1 | 99 | 64 |
| GLD | 134 | 110 | 24 | 63 | 47 | 1 | 87 | 64 |
| GOOGL | 163 | 111 | 52 | 63 | 48 | 1 | 115 | 64 |
| HD | 133 | 108 | 25 | 63 | 45 | 1 | 88 | 64 |
| IWM | 169 | 113 | 56 | 63 | 50 | 1 | 119 | 64 |
| JNJ | 130 | 107 | 23 | 63 | 44 | 1 | 86 | 64 |
| JPM | 151 | 109 | 42 | 63 | 46 | 1 | 105 | 64 |
| MA | 132 | 106 | 26 | 63 | 43 | 1 | 89 | 64 |
| META | 168 | 111 | 57 | 63 | 48 | 1 | 120 | 64 |
| MSFT | 164 | 113 | 51 | 63 | 50 | 1 | 114 | 64 |
| NFLX | 169 | 112 | 57 | 63 | 49 | 1 | 120 | 64 |
| NVDA | 171 | 114 | 57 | 63 | 51 | 1 | 120 | 64 |
| QQQ | 164 | 111 | 53 | 63 | 48 | 1 | 116 | 64 |
| SPY | 168 | 111 | 57 | 63 | 48 | 1 | 120 | 64 |
| TLT | 154 | 114 | 40 | 63 | 51 | 1 | 103 | 64 |
| TSLA | 168 | 111 | 57 | 63 | 48 | 1 | 120 | 64 |
| UNH | 164 | 111 | 53 | 63 | 48 | 1 | 116 | 64 |
| V | 149 | 111 | 38 | 63 | 48 | 1 | 101 | 64 |
| VXX | 95 | 95 | 0 | 63 | 32 | 1 | 51 | N/A |
| WMT | 164 | 111 | 53 | 63 | 48 | 1 | 116 | 64 |
| XLF | 167 | 112 | 55 | 63 | 49 | 1 | 118 | 64 |
| XLK | 134 | 106 | 28 | 63 | 43 | 1 | 91 | 64 |
| XOM | 151 | 108 | 43 | 63 | 45 | 1 | 106 | 64 |

## Sorted by Excess INITIALIZING Descending

| Symbol | Admitted | INITIALIZING | OBSERVABLE | Expected INITIALIZING | Excess | First Seq | Last Seq | First Observable Seq |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| AAPL | 173 | 116 | 57 | 63 | 53 | 1 | 120 | 64 |
| NVDA | 171 | 114 | 57 | 63 | 51 | 1 | 120 | 64 |
| TLT | 154 | 114 | 40 | 63 | 51 | 1 | 103 | 64 |
| IWM | 169 | 113 | 56 | 63 | 50 | 1 | 119 | 64 |
| MSFT | 164 | 113 | 51 | 63 | 50 | 1 | 114 | 64 |
| NFLX | 169 | 112 | 57 | 63 | 49 | 1 | 120 | 64 |
| XLF | 167 | 112 | 55 | 63 | 49 | 1 | 118 | 64 |
| AMZN | 167 | 111 | 56 | 63 | 48 | 1 | 119 | 64 |
| AVGO | 166 | 111 | 55 | 63 | 48 | 1 | 118 | 64 |
| GOOGL | 163 | 111 | 52 | 63 | 48 | 1 | 115 | 64 |
| META | 168 | 111 | 57 | 63 | 48 | 1 | 120 | 64 |
| QQQ | 164 | 111 | 53 | 63 | 48 | 1 | 116 | 64 |
| SPY | 168 | 111 | 57 | 63 | 48 | 1 | 120 | 64 |
| TSLA | 168 | 111 | 57 | 63 | 48 | 1 | 120 | 64 |
| UNH | 164 | 111 | 53 | 63 | 48 | 1 | 116 | 64 |
| V | 149 | 111 | 38 | 63 | 48 | 1 | 101 | 64 |
| WMT | 164 | 111 | 53 | 63 | 48 | 1 | 116 | 64 |
| AMD | 160 | 110 | 50 | 63 | 47 | 1 | 113 | 64 |
| BAC | 164 | 110 | 54 | 63 | 47 | 1 | 117 | 64 |
| GLD | 134 | 110 | 24 | 63 | 47 | 1 | 87 | 64 |
| JPM | 151 | 109 | 42 | 63 | 46 | 1 | 105 | 64 |
| EEM | 144 | 108 | 36 | 63 | 45 | 1 | 99 | 64 |
| HD | 133 | 108 | 25 | 63 | 45 | 1 | 88 | 64 |
| XOM | 151 | 108 | 43 | 63 | 45 | 1 | 106 | 64 |
| JNJ | 130 | 107 | 23 | 63 | 44 | 1 | 86 | 64 |
| MA | 132 | 106 | 26 | 63 | 43 | 1 | 89 | 64 |
| XLK | 134 | 106 | 28 | 63 | 43 | 1 | 91 | 64 |
| COST | 112 | 95 | 17 | 63 | 32 | 1 | 80 | 64 |
| VXX | 95 | 95 | 0 | 63 | 32 | 1 | 51 | N/A |
| DIA | 71 | 71 | 0 | 63 | 8 | 1 | 38 | N/A |

## Totals And Reconciliation

| Metric | Total |
| --- | ---: |
| Number of symbols | 30 |
| Total admitted bars | 4,519 |
| Total INITIALIZING bars | 3,247 |
| Total expected INITIALIZING by requested symbol formula | 1,890 |
| Total excess INITIALIZING by requested symbol formula | 1,357 |
| Total OBSERVABLE bars | 1,272 |

Reconciliations:

```text
SUM(symbol.initializing_bars) = 3,247 = runtime INITIALIZING
SUM(min(symbol.total_admitted_bars, 63)) = 1,890
SUM(symbol.excess_initializing_bars) = 1,357 = 3,247 - 1,890
3,247 + 1,272 = 4,519 admitted
```

## INITIALIZING Ranges for Every Symbol

Each line names the collection run, then the actual entity-scoped sequence ranges and counts.

### AAPL
- `20260910T190625Z-1`: INITIALIZING 1-2 (2); no observable phase.
- `20260910T191246Z-1`: INITIALIZING 1-51 (51); no observable phase.
- `20260911T161623Z-1`: INITIALIZING 1-63 (63); OBSERVABLE 64-120 (57).

### AMD
- `20260910T191246Z-1`: INITIALIZING 1-47 (47); no observable phase.
- `20260911T161623Z-1`: INITIALIZING 1-63 (63); OBSERVABLE 64-113 (50).

### AMZN
- `20260910T191246Z-1`: INITIALIZING 1-48 (48); no observable phase.
- `20260911T161623Z-1`: INITIALIZING 1-63 (63); OBSERVABLE 64-119 (56).

### AVGO
- `20260910T191246Z-1`: INITIALIZING 1-48 (48); no observable phase.
- `20260911T161623Z-1`: INITIALIZING 1-63 (63); OBSERVABLE 64-118 (55).

### BAC
- `20260910T191246Z-1`: INITIALIZING 1-47 (47); no observable phase.
- `20260911T161623Z-1`: INITIALIZING 1-63 (63); OBSERVABLE 64-117 (54).

### COST
- `20260910T191246Z-1`: INITIALIZING 1-32 (32); no observable phase.
- `20260911T161623Z-1`: INITIALIZING 1-63 (63); OBSERVABLE 64-80 (17).

### DIA
- `20260910T191246Z-1`: INITIALIZING 1-38 (38); no observable phase.
- `20260911T161623Z-1`: INITIALIZING 1-33 (33); no observable phase.

### EEM
- `20260910T191246Z-1`: INITIALIZING 1-45 (45); no observable phase.
- `20260911T161623Z-1`: INITIALIZING 1-63 (63); OBSERVABLE 64-99 (36).

### GLD
- `20260910T191246Z-1`: INITIALIZING 1-47 (47); no observable phase.
- `20260911T161623Z-1`: INITIALIZING 1-63 (63); OBSERVABLE 64-87 (24).

### GOOGL
- `20260910T191246Z-1`: INITIALIZING 1-48 (48); no observable phase.
- `20260911T161623Z-1`: INITIALIZING 1-63 (63); OBSERVABLE 64-115 (52).

### HD
- `20260910T191246Z-1`: INITIALIZING 1-45 (45); no observable phase.
- `20260911T161623Z-1`: INITIALIZING 1-63 (63); OBSERVABLE 64-88 (25).

### IWM
- `20260910T191246Z-1`: INITIALIZING 1-50 (50); no observable phase.
- `20260911T161623Z-1`: INITIALIZING 1-63 (63); OBSERVABLE 64-119 (56).

### JNJ
- `20260910T191246Z-1`: INITIALIZING 1-44 (44); no observable phase.
- `20260911T161623Z-1`: INITIALIZING 1-63 (63); OBSERVABLE 64-86 (23).

### JPM
- `20260910T191246Z-1`: INITIALIZING 1-46 (46); no observable phase.
- `20260911T161623Z-1`: INITIALIZING 1-63 (63); OBSERVABLE 64-105 (42).

### MA
- `20260910T191246Z-1`: INITIALIZING 1-43 (43); no observable phase.
- `20260911T161623Z-1`: INITIALIZING 1-63 (63); OBSERVABLE 64-89 (26).

### META
- `20260910T191246Z-1`: INITIALIZING 1-48 (48); no observable phase.
- `20260911T161623Z-1`: INITIALIZING 1-63 (63); OBSERVABLE 64-120 (57).

### MSFT
- `20260910T190625Z-1`: INITIALIZING 1-2 (2); no observable phase.
- `20260910T191246Z-1`: INITIALIZING 1-48 (48); no observable phase.
- `20260911T161623Z-1`: INITIALIZING 1-63 (63); OBSERVABLE 64-114 (51).

### NFLX
- `20260910T191246Z-1`: INITIALIZING 1-49 (49); no observable phase.
- `20260911T161623Z-1`: INITIALIZING 1-63 (63); OBSERVABLE 64-120 (57).

### NVDA
- `20260910T190625Z-1`: INITIALIZING 1-2 (2); no observable phase.
- `20260910T191246Z-1`: INITIALIZING 1-49 (49); no observable phase.
- `20260911T161623Z-1`: INITIALIZING 1-63 (63); OBSERVABLE 64-120 (57).

### QQQ
- `20260910T191246Z-1`: INITIALIZING 1-48 (48); no observable phase.
- `20260911T161623Z-1`: INITIALIZING 1-63 (63); OBSERVABLE 64-116 (53).

### SPY
- `20260910T191246Z-1`: INITIALIZING 1-48 (48); no observable phase.
- `20260911T161623Z-1`: INITIALIZING 1-63 (63); OBSERVABLE 64-120 (57).

### TLT
- `20260910T191246Z-1`: INITIALIZING 1-51 (51); no observable phase.
- `20260911T161623Z-1`: INITIALIZING 1-63 (63); OBSERVABLE 64-103 (40).

### TSLA
- `20260910T191246Z-1`: INITIALIZING 1-48 (48); no observable phase.
- `20260911T161623Z-1`: INITIALIZING 1-63 (63); OBSERVABLE 64-120 (57).

### UNH
- `20260910T191246Z-1`: INITIALIZING 1-48 (48); no observable phase.
- `20260911T161623Z-1`: INITIALIZING 1-63 (63); OBSERVABLE 64-116 (53).

### V
- `20260910T191246Z-1`: INITIALIZING 1-48 (48); no observable phase.
- `20260911T161623Z-1`: INITIALIZING 1-63 (63); OBSERVABLE 64-101 (38).

### VXX
- `20260910T191246Z-1`: INITIALIZING 1-44 (44); no observable phase.
- `20260911T161623Z-1`: INITIALIZING 1-51 (51); no observable phase.

### WMT
- `20260910T191246Z-1`: INITIALIZING 1-48 (48); no observable phase.
- `20260911T161623Z-1`: INITIALIZING 1-63 (63); OBSERVABLE 64-116 (53).

### XLF
- `20260910T191246Z-1`: INITIALIZING 1-49 (49); no observable phase.
- `20260911T161623Z-1`: INITIALIZING 1-63 (63); OBSERVABLE 64-118 (55).

### XLK
- `20260910T191246Z-1`: INITIALIZING 1-43 (43); no observable phase.
- `20260911T161623Z-1`: INITIALIZING 1-63 (63); OBSERVABLE 64-91 (28).

### XOM
- `20260910T191246Z-1`: INITIALIZING 1-45 (45); no observable phase.
- `20260911T161623Z-1`: INITIALIZING 1-63 (63); OBSERVABLE 64-106 (43).

Range reconciliation:

```text
SUM(INITIALIZING range counts) = 3,247
SUM(OBSERVABLE range counts) = 1,272
number of scopes = 63
scopes shorter than 63 bars = 35
scopes exactly 63 bars = 0
scopes longer than 63 bars = 28
```

## Representative Traces

### AMZN: ordinary two-scope behavior

Run `20260910T191246Z-1` creates key `20260910T191246Z-1|AMZN` and a fresh solver. Sequence 1 has high 251.93 and low 251.83; admission accepts sequence 1; `solver.count` starts at 0, `index` is 0, and the result has no phase. Sequences 1-48 are therefore INITIALIZING. The run ends before update 64, so it never becomes observable.

Run `20260911T161623Z-1` creates a different key, `20260911T161623Z-1|AMZN`, and another fresh solver. At sequence 63, high is 256.4 and low is 256.37; pre-increment `index` is 62, so `index < 63` and status remains INITIALIZING. At sequence 64, high is 256.445 and low is 256.405; pre-increment `index` is 63, the solver returns a phase pointer, and the coordinator assigns OBSERVABLE. This scope remains observable through sequence 119.

### AAPL: largest symbol-level excess

AAPL contributes 116 INITIALIZING records versus the requested symbol-level expectation of 63, an excess of 53. The excess is exactly the two earlier run scopes: 2 + 51 = 53.

- Run `20260910T190625Z-1`, key `20260910T190625Z-1|AAPL`: a new solver processes sequences 1-2, both INITIALIZING. Sequence 1 has high 325.275 and low 324.99.
- Run `20260910T191246Z-1`, key `20260910T191246Z-1|AAPL`: another new solver processes sequences 1-51, all INITIALIZING. Sequence 51 has high and low 326.43.
- Run `20260911T161623Z-1`, key `20260911T161623Z-1|AAPL`: a third new solver processes sequence 63 with high 334.27 and low 334.08 at `index=62`, so it remains INITIALIZING. Sequence 64 has high 334.07 and low 333.93 at `index=63`, so it becomes OBSERVABLE and remains observable through sequence 120.

In every trace the path is:

```text
Mongo observation
-> OFFLINE MapObservation assigns entity_sequence_scope = run_id|symbol
-> admission accepts the next sequence in that scope
-> Coordinator retrieves or creates state[entity_sequence_scope]
-> that scope's Solver.count determines phase presence
-> Coordinator maps absent phase to INITIALIZING
-> PhaseEvidence carries the same scope and entity sequence
```

## Why 3,712 Is Not the Current Count

No combination of current PhaseEvidence classifications totals 3,712 INITIALIZING records. The complete current evidence contains 3,247, and every record is reconciled by the 63 scope-level ranges above. A 3,712 observation must have come from a different dataset snapshot, different scope construction, different code revision, or a prior counting method. The present evidence cannot select among those historical possibilities.

The evidence-based explanation for the current high INITIALIZING total is precise: the dataset contains multiple collection runs, the analytical state key includes `collection_run_id`, and all 30 symbols therefore receive a fresh 63-update warm-up in more than one scope. This is repeated initialization by run-scoped solver ownership, not one long initialization period and not a change to the 63-bar rule.

## Required Console Summary

```text
DSE_JEH PHASE 1 INITIALIZING DIAGNOSTIC

Total admitted bars: 4519
Actual INITIALIZING: 3247
Expected INITIALIZING: 1890
Difference: 1357
OBSERVABLE: 1272

INITIALIZING decision:
status remains INITIALIZING when Solver.Update succeeds and returns PhaseAngle == nil; this occurs when the solver's pre-increment index is < 63.

State/counter used:
jeh.Solver.count (int), counting admitted Solver.Update calls per solver instance; index := count, then count++.

Analytical state key:
event.provenance.entity_sequence_scope = collection_run_id + "|" + normalized symbol

Number of JEH state instances created:
63

Number of JEH state resets:
0 in-place resets; 63 fresh states are created in this runtime, and a runtime restart reconstructs all state.

Entities initialized more than once:
30: AAPL, AMD, AMZN, AVGO, BAC, COST, DIA, EEM, GLD, GOOGL, HD, IWM, JNJ, JPM, MA, META, MSFT, NFLX, NVDA, QQQ, SPY, TLT, TSLA, UNH, V, VXX, WMT, XLF, XLK, XOM

Primary reason for 3,712 INITIALIZING bars:
3,712 is not reproduced. The current 3,247 records result from 63 collection-run-and-symbol solver instances, each independently applying up to 63 INITIALIZING updates.

Production code changed:
NO
```
