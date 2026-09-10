# Bar Sequence Lab Generator — A/B/C Live Proving Report

**Date:** 2026-09-10  
**Component:** Bar_Sequence_Lab / generator  
**Environment:** Live Alpaca IEX  
**MongoDB:** DEFERRED  
**Persistence:** LOCAL JSONL  
**Git:** NOT COMMITTED. NOT PUSHED.

---

## Executive Summary

Live Alpaca IEX proving confirmed the implemented A/B/C startup selection and partition architecture with real 1-minute bars.

- Group A independently: **PASS** (prior run `20260910T175222Z-1`; implementation unchanged).
- Group B independently: **PASS** (`20260910T180559Z-1`).
- Group C independently: **PASS** (`20260910T180813Z-1`).
- A+B in one Go process: **PASS** (`20260910T181056Z-1`).
- A+B+C in one Go process: **PASS** (`20260910T181804Z-1`).

All performed live runs had `dropped_valid_count=0`, pending=0 at shutdown, no unknown symbols, no cross-partition contamination, and per-symbol `generator_sequence_no` starting at 1 independently.

One implementation defect was found and fixed before the B/C/multi-group live runs: PowerShell unnamed groups bound to `-Feed`/`-DataDir` instead of remaining arguments, so documented `.\scripts\Start-BarSequenceLab.ps1 A` / `-Feed iex A B` printed usage and exited 1. Minimal positional-binding fix applied. Selector revalidated. Live proving then used the official script.

Duplicates and source-time regressions: **NOT OBSERVED DURING LIVE PROVING**. Deterministic unit tests already cover those flags.

MongoDB was not created. Viewer was not modified. Fin_FeedSat_1 and DSE were not modified.

---

## Configured groups (before live proving)

Leftover session env from the earlier AAPL-only run had `BAR_SEQ_LAB_GROUP_B` and `BAR_SEQ_LAB_GROUP_C` empty. Those leftovers were cleared. For this architectural proof a small controlled subset was set (IEX defaults; no 30-symbol load test):

GROUP A:  
AAPL

GROUP B:  
MSFT

GROUP C:  
NVDA

Feed: `iex` (`wss://stream.data.alpaca.markets/v2/iex`). Credentials present (not printed).

---

## Exact PowerShell syntax confirmed

Intended / documented human-facing interface (unchanged):

```powershell
.\scripts\Start-BarSequenceLab.ps1 A
.\scripts\Start-BarSequenceLab.ps1 B
.\scripts\Start-BarSequenceLab.ps1 C
.\scripts\Start-BarSequenceLab.ps1 A B
.\scripts\Start-BarSequenceLab.ps1 A C
.\scripts\Start-BarSequenceLab.ps1 B C
.\scripts\Start-BarSequenceLab.ps1 A B C
.\scripts\Start-BarSequenceLab.ps1 -Feed iex A B
```

**Defect (pre-fix):** unnamed remaining groups were bound as positional `-Feed` / `-DataDir` / `-MaxBars`. `.\scripts\Start-BarSequenceLab.ps1 B -Feed iex ...` and `-Feed iex A B` therefore treated groups as missing and printed usage (exit 1).

**Fix:** `PositionalBinding = $false` plus `Groups` `Position = 0, ValueFromRemainingArguments`. Temporary `BAR_SEQ_LAB_PARSE_ONLY=1` exits after printing `PARSE_OK` so selector tests do not start a live process.

Post-fix parse-only:

| Command | Result |
|---|---|
| (no groups) | usage, exit 1 |
| `A` `B` `C` `A B` `A C` `B C` `A B C` | PARSE_OK, exit 0 |
| `-Feed iex A B` | groups=A,B feed=iex |
| `B -Feed iex -MaxBars 2 -Duration 2m` | groups=B |
| `-Feed iex -MaxBars 2 -Duration 2m B` | groups=B |
| `D` | invalid group, exit 1 |

---

## Test matrix

| Test | A active | B active | C active | Result |
|---|---|---|---|---|
| A | yes | no | no | PASS (prior live evidence) |
| B | no | yes | no | PASS |
| C | no | no | yes | PASS |
| A+B | yes | yes | no | PASS |
| A+B+C | yes | yes | yes | PASS |

---

## TEST 1 — GROUP A (prior live evidence)

Relied on previous successful live IEX run. Generator implementation of A/B/C pipelines was not reimplemented for this task. The only subsequent code change was the PowerShell positional-binding fix, which does not alter Go ingest/persist behavior.

- **Run ID:** `20260910T175222Z-1`
- **Command:** `go run ./cmd/generator` with `BAR_SEQ_LAB_FEED=iex`, `BAR_SEQ_LAB_SELECTED_GROUPS=A`, `BAR_SEQ_LAB_GROUP_A=AAPL`, `MAX_BARS=3`, `DURATION=2m`
- **Selected groups:** A
- **Active partitions:** A
- **Inactive:** B, C (no pipelines, no JSONL)
- **Subscribe:** AAPL (`alpaca subscribed ... bars=[AAPL]`)
- **Duration:** ~2m (duration cap; 2 of 3 requested bars)
- **accepted / persisted / pending / dropped_valid:** 2 / 2 / 0 / 0
- **duplicates / regressions / unknown:** 0 / 0 / 0
- **JSONL:** `data\20260910T175222Z-1\partition_A.jsonl` (2 records)
- **Files not created:** partition_B.jsonl, partition_C.jsonl
- **Result:** PASS

AAPL sequence: 1, 2 (`T=b`, source times 17:52:00Z and 17:53:00Z).

---

## TEST 2 — GROUP B

- **Run ID:** `20260910T180559Z-1`
- **Command:** `.\scripts\Start-BarSequenceLab.ps1 B -Feed iex -MaxBars 2 -Duration 2m`
- **Selected groups:** B
- **Active partitions:** B (`B1/B2/B3` only)
- **Inactive:** A, C
- **Subscribe:** MSFT
- **Duration:** 2m (duration reached; 1 of 2 requested bars)
- **accepted / persisted / pending / dropped_valid:** 1 / 1 / 0 / 0
- **duplicates / regressions / unknown:** 0 / 0 / 0
- **JSONL:** `data\20260910T180559Z-1\partition_B.jsonl` (1 record)
- **Files not created:** partition_A.jsonl, partition_C.jsonl
- **Result:** PASS
- **Sequence proof:** PARTIAL for MSFT (only one observation in the 2-minute window). seq=1.

MSFT 18:06:00Z O/H/L/C 491.805 / 491.855 / 491.645 / 491.645 vol 280 `T=b`.

---

## TEST 3 — GROUP C

- **Run ID:** `20260910T180813Z-1`
- **Command:** `.\scripts\Start-BarSequenceLab.ps1 C -Feed iex -MaxBars 2 -Duration 2m`
- **Selected groups:** C
- **Active partitions:** C (`C1/C2/C3` only)
- **Inactive:** A, B
- **Subscribe:** NVDA
- **Duration:** stopped at `max_bars=2`
- **accepted / persisted / pending / dropped_valid:** 2 / 2 / 0 / 0
- **duplicates / regressions / unknown:** 0 / 0 / 0
- **JSONL:** `data\20260910T180813Z-1\partition_C.jsonl` (2 records)
- **Files not created:** partition_A.jsonl, partition_B.jsonl
- **Result:** PASS

NVDA sequence: 1, 2 (18:08:00Z, 18:09:00Z).

---

## TEST 4 — A+B (one process)

```text
                    Alpaca IEX
                        |
                 symbols(A ∪ B) = AAPL,MSFT
                        |
                     ONE process
                    /           \
                   A             B
             A1→A2→A3       B1→B2→B3
              Writer A       Writer B
        partition_A.jsonl  partition_B.jsonl
```

Partition C remained inactive.

- **Run ID:** `20260910T181056Z-1`
- **Command:** `.\scripts\Start-BarSequenceLab.ps1 A B -Feed iex -MaxBars 4 -Duration 3m`
- **Selected groups:** A,B
- **Subscribe:** AAPL,MSFT (`bars=[AAPL MSFT]`)
- **Stopped:** `max_bars=4`
- **A accepted/persisted/pending:** 2 / 2 / 0
- **B accepted/persisted/pending:** 2 / 2 / 0
- **dropped_valid / unknown / dup / regression:** 0 / 0 / 0 / 0
- **JSONL files:** partition_A.jsonl (2), partition_B.jsonl (2)
- **partition_C.jsonl:** NOT created
- **Independent progress:** both writers persisted the 18:10 bar in the same metrics interval, then both persisted the 18:11 bar at shutdown (`last_latency` 503.4µs each). Separate A1–A3 and B1–B3 metrics throughout.
- **Result:** PASS

---

## TEST 5 — A+B+C (one process)

- **Run ID:** `20260910T181804Z-1`
- **Command:** `.\scripts\Start-BarSequenceLab.ps1 A B C -Feed iex -MaxBars 6 -Duration 3m`
- **Selected groups:** A,B,C
- **Subscribe:** AAPL,MSFT,NVDA (`bars=[AAPL MSFT NVDA]`)
- **Stopped:** `max_bars=6`
- **A/B/C each:** accepted=2 persisted=2 pending=0 critical=0
- **dropped_valid / unknown / dup / regression:** 0 / 0 / 0 / 0
- **JSONL:** partition_A.jsonl (2), partition_B.jsonl (2), partition_C.jsonl (2)
- **Independent progress:** first bar for all three partitions persisted in the same 11:19:04 metrics snapshot (separate last_latency 14.3765ms / 13.8735ms / 13.3219ms). Second bars all persisted at shutdown. Separate A1–A3, B1–B3, C1–C3 metrics.
- **Result:** PASS

---

## Per-symbol sequence evidence

There is no cross-symbol sequence authority and no cross-partition total sequence.

**A-only `20260910T175222Z-1`**

- AAPL: 1, 2

**B-only `20260910T180559Z-1`**

- MSFT: 1 (PARTIAL; only one live bar in window)

**C-only `20260910T180813Z-1`**

- NVDA: 1, 2

**A+B `20260910T181056Z-1`**

- AAPL: 1, 2 (partition A only)
- MSFT: 1, 2 (partition B only)

**A+B+C `20260910T181804Z-1`**

- AAPL: 1, 2 (partition A)
- MSFT: 1, 2 (partition B)
- NVDA: 1, 2 (partition C)

This is the required independent pattern (each symbol 1, 2, …), not AAPL=1 / MSFT=2 / NVDA=3.

---

## Partition isolation evidence

For every run directory, only the selected partitions created JSONL files. Record `partition_id` matched the file. Symbols matched assignment.

| Run | Files present | Symbols in A | Symbols in B | Symbols in C |
|---|---|---|---|---|
| 20260910T175222Z-1 | A only | AAPL | — | — |
| 20260910T180559Z-1 | B only | — | MSFT | — |
| 20260910T180813Z-1 | C only | — | — | NVDA |
| 20260910T181056Z-1 | A, B | AAPL only | MSFT only | file absent |
| 20260910T181804Z-1 | A, B, C | AAPL only | MSFT only | NVDA only |

No AAPL in B/C files. No MSFT in A/C files. No NVDA in A/B files. All records: valid JSON, one object per line, `collection_run_id` matches directory, `interval=1Min`, OHLCV present, `alpaca_message_type=b`, `source_event_time` / `received_time` present.

---

## Independent A/B/C path proof (live)

- Inactive groups: no pipeline start, no JSONL file.
- Active groups: dedicated A1/A2/A3, B1/B2/B3, C1/C2/C3 buffer metrics.
- Separate writer paths and record counts.
- Multi-group runs showed simultaneous persist progress (same minute close written by each active writer without waiting on a shared persist loop).
- No destructive writer fault injection in this task.

---

## Reconciliation

Under all completed live runs:

`accepted_for_partition = locally_persisted_for_partition + pending + explicitly_failed`

At graceful completion: pending=0, failed_batches=0, accepted = locally persisted, `dropped_valid_count=0`.

JSONL line counts matched shutdown persisted counts.

---

## Duplicates / source-time regressions

**NOT OBSERVED DURING LIVE PROVING.**

Existing deterministic tests:

- [internal/sequence/authority_test.go](../internal/sequence/authority_test.go) `TestAcceptedArrivalNotSourceSorted`, `TestDuplicatePreserved`, `TestIndependentSymbols`

Live test does not claim validation of conditions that did not occur.

---

## Validation after change

- `gofmt` on `./cmd` `./internal`
- `go test ./...` — PASS
- `go vet ./...` — PASS (`VET_OK`)
- `go build ./...` — PASS (`BUILD_OK`)

---

## Defects and fixes

1. **PowerShell group remaining-arguments binding**  
   Violated authorized startup interface (groups required; documented `A` / `-Feed iex A B`).  
   Smallest fix in [scripts/Start-BarSequenceLab.ps1](../scripts/Start-BarSequenceLab.ps1): disable default positional binding; Groups remain remaining-arguments at position 0. Parse-only env for selector tests.  
   Approved V0.1 design not changed.

No Go pipeline defects observed in live proving.

---

## Remaining proving work

- Larger symbol-count / plan-cap stress (not this task).
- Natural duplicate / source-time regression live observation (not manufactured).
- Forming vs final (`T=u` vs `T=b`) live policy (OE-10); this session persisted `T=b` only.
- Mongo persistence (deferred).
- Viewer (not this task).
- Restart continuity across process restart (LAB-GEN-OE-13).
- A C and B C live combinations not separately live-tested; A+C isolation is covered by unit test `TestEnginePersistsSelectedPartitionsOnly`; live A+B and A+B+C cover multi-group isolation.

---

## Git

**Before work:**

```text
?? Bar_Sequence_Lab/
?? DSE_TransSat_1_worker/docs/DSE_TRANS_SAT_1_SYSTEM_ARCHITECTURAL_DESIGN_V0_1_091026.md
```

**After work:** `git diff --check` clean. Status still untracked `Bar_Sequence_Lab/` plus pre-existing uncommitted DSE architectural design. Unrelated files not cleaned or reverted.

NOT COMMITTED.  
NOT PUSHED.

---

## Explicit statement

NOT COMMITTED.  
NOT PUSHED.  
A/B/C LIVE PROVING COMPLETED TO THE EXTENT REPORTED.  
LOCAL JSONL REMAINS THE LAB PERSISTENCE ADAPTER.  
MONGODB REMAINS DEFERRED.  
BAR_SEQUENCE_LAB/VIEWER NOT MODIFIED.  
FIN_FEEDSAT_1 NOT MODIFIED.  
DSE NOT MODIFIED.
