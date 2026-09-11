# Bar Sequence Lab - 2h or 120-bar collection

Date: 2026-09-11  
Run: `20260911T161623Z-1`  
Verdict: **PASS**  
Stop reason: `TWO_HOUR_LIMIT`

## Executive summary

A new 30-symbol Alpaca IEX raw collection ran for the configured two-hour limit. The early stop condition was not met because only 6 of 30 symbols reached 120 accepted bars. The generator shut down cleanly and reconciled 3,120 accepted observations to 3,120 JSONL records and 3,120 MongoDB documents. No raw observations from the prior reference run were changed. Phase calculation was not run.

## Command and runtime

```powershell
$env:BAR_SEQ_LAB_GROUP_A = "AAPL,AMZN,MSFT,META,TSLA,NVDA,GOOGL,NFLX,AMD,AVGO"
$env:BAR_SEQ_LAB_GROUP_B = "SPY,QQQ,DIA,IWM,EEM,VXX,GLD,TLT,XLF,XLK"
$env:BAR_SEQ_LAB_GROUP_C = "JPM,XOM,UNH,JNJ,V,MA,COST,HD,BAC,WMT"
$env:BAR_SEQ_LAB_MONGO_ENABLED = "true"
$env:BAR_SEQ_LAB_MONGO_URI = "mongodb://127.0.0.1:27017"
$env:BAR_SEQ_LAB_MONGO_DB = "bar_sequence_db"
$env:BAR_SEQ_LAB_MONGO_COLLECTION = "bar_sequence"
$env:BAR_SEQ_LAB_METRICS_INTERVAL = "1m"
.\scripts\Start-BarSequenceLab.ps1 -Feed iex -Duration 2h -TargetBarsPerSymbol 120 -MaxBars 0 A B C
```

| Item | Result |
| --- | --- |
| Collection run ID | `20260911T161623Z-1` |
| Process/timer start | 2026-09-11 09:16:23 -07:00 / 16:16:23Z |
| Alpaca subscription confirmed | 2026-09-11 09:16:24 -07:00 / 16:16:24Z |
| Stop | 2026-09-11 11:16:23 -07:00 / 18:16:23Z |
| Configured elapsed duration | 2h from engine dispatch start |
| Healthy subscribed duration | approximately 1h 59m 59s |
| Stop policy | elapsed 2h OR all 30 symbols each reach 120 accepted bars |
| Stop reason | `TWO_HOUR_LIMIT` |
| Source/feed | Alpaca IEX / `ALPACA_IEX` |
| Mongo destination | `bar_sequence_db.bar_sequence` |
| JSONL destination | `data/20260911T161623Z-1/partition_{A,B,C}.jsonl` |

The duration timer begins at engine dispatch startup, not at confirmed subscription health. Subscription confirmation followed one second later.

## Symbol universe

All 30 requested subscriptions were confirmed together.

- A: AAPL, AMZN, MSFT, META, TSLA, NVDA, GOOGL, NFLX, AMD, AVGO
- B: SPY, QQQ, DIA, IWM, EEM, VXX, GLD, TLT, XLF, XLK
- C: JPM, XOM, UNH, JNJ, V, MA, COST, HD, BAC, WMT

A/B/C are ingestion partitions only. No analytical or Greek segmentation was applied.

## Reconciliation

| Partition | Accepted | JSONL | Mongo | Critical | Pending JSONL | Pending Mongo | Mongo failed |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| A | 1,179 | 1,179 | 1,179 | 0 | 0 | 0 | 0 |
| B | 937 | 937 | 937 | 0 | 0 | 0 | 0 |
| C | 1,004 | 1,004 | 1,004 | 0 | 0 | 0 | 0 |
| **Total** | **3,120** | **3,120** | **3,120** | **0** | **0** | **0** | **0** |

`dropped_valid_count=0`, unknown symbols=0, duplicate arrivals=0, source-time regressions=0, failed JSONL batches=0, and Mongo failures=0. Alpaca was healthy at shutdown with no last error and zero heartbeat misses.

Observed buffer high-water marks at shutdown were A1/A2/A3=6, B1/B2/B3=7, and C1=5, C2/C3=4, all against capacity 256. Every buffer depth was zero at shutdown. The empty persist-stage `oldest` metric retained an elapsed age; depth and pending counters are the authoritative drain indicators.

## Per-symbol results

Counts independently matched between MongoDB and JSONL. Every symbol had `generator_sequence_no` exactly contiguous from 1 through the reported last value, with the number of distinct sequence values equal to the record count.

| Partition | Symbol | Accepted / JSONL / Mongo | First sequence | Last sequence |
| --- | --- | ---: | ---: | ---: |
| A | AAPL | 120 | 1 | 120 |
| A | AMD | 113 | 1 | 113 |
| A | AMZN | 119 | 1 | 119 |
| A | AVGO | 118 | 1 | 118 |
| A | GOOGL | 115 | 1 | 115 |
| A | META | 120 | 1 | 120 |
| A | MSFT | 114 | 1 | 114 |
| A | NFLX | 120 | 1 | 120 |
| A | NVDA | 120 | 1 | 120 |
| A | TSLA | 120 | 1 | 120 |
| B | DIA | 33 | 1 | 33 |
| B | EEM | 99 | 1 | 99 |
| B | GLD | 87 | 1 | 87 |
| B | IWM | 119 | 1 | 119 |
| B | QQQ | 116 | 1 | 116 |
| B | SPY | 120 | 1 | 120 |
| B | TLT | 103 | 1 | 103 |
| B | VXX | 51 | 1 | 51 |
| B | XLF | 118 | 1 | 118 |
| B | XLK | 91 | 1 | 91 |
| C | BAC | 117 | 1 | 117 |
| C | COST | 80 | 1 | 80 |
| C | HD | 88 | 1 | 88 |
| C | JNJ | 86 | 1 | 86 |
| C | JPM | 105 | 1 | 105 |
| C | MA | 89 | 1 | 89 |
| C | UNH | 116 | 1 | 116 |
| C | V | 101 | 1 | 101 |
| C | WMT | 116 | 1 | 116 |
| C | XOM | 106 | 1 | 106 |

## Readiness summary

| Measure | Result |
| --- | ---: |
| Minimum bars per symbol | 33 (DIA) |
| Maximum bars per symbol | 120 |
| Average bars per symbol | 104.0 |
| Symbols >=63 | 28 |
| Symbols >=72 | 28 |
| Symbols >=90 | 23 |
| Symbols >=120 | 6 |

DIA (33) and VXX (51) did not reach the phase solver's 63-bar lookback. This is preserved source behavior; no bars were fabricated, repeated, interpolated, forward-filled, or renumbered. These thresholds indicate phase-analysis readiness only, not strategy eligibility.

## Sequence and persistence integrity

- All 30 configured symbols appeared; no unexpected symbol appeared.
- Every per-symbol sequence began at 1 and was contiguous through its own final value.
- No global cross-symbol sequence was introduced.
- Mongo count, JSONL count, and accepted count reconcile per symbol and partition.
- Shutdown pending/unflushed count was zero for both sinks.
- Duplicate and source-time-regression flags were zero across the run.

## Reference-run protection

Before and after this collection, reference run `20260910T191246Z-1` remained:

| Scope | Count |
| --- | ---: |
| Total | 1,393 |
| A | 484 |
| B | 463 |
| C | 446 |

Its 30 per-symbol counts also matched the established reference baseline. No existing raw record was appended, modified, deleted, renumbered, merged, or normalized.

## Operational enhancement

The pre-existing `MaxBars` setting counts aggregate accepted observations and could not represent the requested per-symbol condition. An optional `TargetBarsPerSymbol` runner parameter and `BAR_SEQ_LAB_TARGET_BARS_PER_SYMBOL` configuration value were added. A zero value preserves prior behavior. Runtime tests verify that a fast symbol reaching the target cannot stop collection while another configured symbol remains below it.

Periodic diagnostics now report minimum/maximum per-symbol counts and counts at 63, 72, 90, and 120 bars. Stop output reports an explicit reason. Tests, `go vet ./...`, and `go build ./cmd/generator` passed before launch.

## Limitations

- IEX activity differs by symbol, so a fixed wall-clock period does not guarantee 63 or 120 observations for every symbol.
- The duration timer starts at engine dispatch startup; it is not deferred until the subscription acknowledgment.
- The run ended by duration with only 6 symbols at 120, as required by the all-symbol early-stop definition.
- Phase Angle Series Generator was not run. No derived phase records were created or modified by this task.
