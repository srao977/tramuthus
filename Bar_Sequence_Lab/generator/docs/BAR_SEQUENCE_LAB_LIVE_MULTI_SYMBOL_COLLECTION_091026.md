# Bar Sequence Lab — live multi-symbol collection

Date: 2026-09-10  
collection_run_id: `20260910T191246Z-1`  
Verdict: **PASS**

LIVE MULTI-SYMBOL BAR-SEQUENCE COLLECTION COMPLETED.  
RAW BAR-SEQUENCE WAVES PRESERVED IN MONGODB.  
JSONL AUDIT PATH PRESERVED.  
ACCEPTED / JSONL / MONGODB RECONCILIATION PASSED.  
DROPPED_VALID = 0.

Maths / Ehlers-Hilbert / Hop-On/Hop-Off were not implemented. Viewer, Fin_FeedSat_1, DSE, and HACCAM were not modified.

## Run identity

| Item | Value |
| --- | --- |
| start | 2026-09-10 12:12:46 local (US Mountain) / 19:12:46Z |
| end | 2026-09-10 13:07:46 local / 20:07:46Z (`duration reached`) |
| duration | 55m configured; ~54m 60s wall |
| feed | iex (`ALPACA_IEX`, `wss://stream.data.alpaca.markets/v2/iex`) |
| MaxBars | 0 (unlimited) |
| configured symbols | 30 (Alpaca subscription accepted all 30) |
| Mongo | `bar_sequence_db.bar_sequence` host `127.0.0.1:27017` |
| JSONL directory | `data/20260910T191246Z-1/` |

### A (10)

AAPL, AMZN, MSFT, META, TSLA, NVDA, GOOGL, NFLX, AMD, AVGO

### B (10)

SPY, QQQ, DIA, IWM, EEM, VXX, GLD, TLT, XLF, XLK

### C (10)

JPM, XOM, UNH, JNJ, V, MA, COST, HD, BAC, WMT

No duplicate ownership. No empty group. No 30-symbol list existed in-repo; universe is the proved project instruments plus liquid U.S. names to fill the plan cap.

JSONL paths:

- `data/20260910T191246Z-1/partition_A.jsonl`
- `data/20260910T191246Z-1/partition_B.jsonl`
- `data/20260910T191246Z-1/partition_C.jsonl`

## Shutdown totals

Generator SHUTDOWN:

| Partition | accepted | JSONL persisted | Mongo persisted | pending JSONL | pending Mongo | failed Mongo | critical |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| A | 484 | 484 | 484 | 0 | 0 | 0 | 0 |
| B | 463 | 463 | 463 | 0 | 0 | 0 | 0 |
| C | 446 | 446 | 446 | 0 | 0 | 0 | 0 |
| **total** | **1393** | **1393** | **1393** | **0** | **0** | **0** | **0** |

| Integrity | Value |
| --- | --- |
| dropped_valid | 0 |
| unknown_symbol | 0 |
| duplicate_arrivals | 0 |
| source_time_regressions | 0 |
| mongo_failed_total | 0 |
| jsonl_unflushed_total | 0 |
| mongo_unflushed_total | 0 |
| Alpaca | healthy through shutdown; heartbeat_misses=0; last_error empty |
| reconnects | none observed |

Buffer high-water (capacity 256): A1=4, A2=4, A3=4, B1=3, B2=3, B3=3, C1=4, C2=3, C3=3. Depth 0 at shutdown.

## Per-symbol reconciliation

Generator `latest` sequence at shutdown equals JSONL count and Mongo count for every symbol. Each symbol starts at `generator_sequence_no=1` and is contiguous `1..N` independently (not a global sequence). Unequal N across symbols is IEX activity, not a defect. Zero-bar symbols: **none**.

| Symbol | Partition | accepted / JSONL / Mongo | seq |
| --- | --- | ---: | --- |
| AAPL | A | 51 | 1..51 |
| AMD | A | 47 | 1..47 |
| AMZN | A | 48 | 1..48 |
| AVGO | A | 48 | 1..48 |
| META | A | 48 | 1..48 |
| MSFT | A | 48 | 1..48 |
| NFLX | A | 49 | 1..49 |
| NVDA | A | 49 | 1..49 |
| TSLA | A | 48 | 1..48 |
| GOOGL | A | 48 | 1..48 |
| DIA | B | 38 | 1..38 |
| EEM | B | 45 | 1..45 |
| GLD | B | 47 | 1..47 |
| IWM | B | 50 | 1..50 |
| QQQ | B | 48 | 1..48 |
| SPY | B | 48 | 1..48 |
| TLT | B | 51 | 1..51 |
| VXX | B | 44 | 1..44 |
| XLF | B | 49 | 1..49 |
| XLK | B | 43 | 1..43 |
| BAC | C | 47 | 1..47 |
| COST | C | 32 | 1..32 |
| HD | C | 45 | 1..45 |
| JNJ | C | 44 | 1..44 |
| JPM | C | 46 | 1..46 |
| MA | C | 43 | 1..43 |
| UNH | C | 48 | 1..48 |
| V | C | 48 | 1..48 |
| WMT | C | 48 | 1..48 |
| XOM | C | 45 | 1..45 |

JSONL line counts: A=484, B=463, C=446.  
Mongo `countDocuments({collection_run_id:"20260910T191246Z-1"})` = 1393.  
Collection total 1399 = this run 1393 + prior proving run `20260910T190625Z-1` (6 docs). Prior run was not merged, deleted, or renumbered.

## Sequence integrity

Per symbol: `minseq=1`, `maxseq=N`, `count=N`. JSONL sequences consecutive. Mongo sequences consecutive. No invented observations. No global merge by `_id`, `persisted_time`, `received_time`, or `source_event_time`.

## Defects

None demonstrated. `dropped_valid=0`. Persistence lag during the run was in-flight only; shutdown exact.

## Deferred (not done during collection)

- Formal 30-symbol universe freeze in config (today used env lists)
- CSV export
- Maths / Ehlers-Hilbert / Hop
- Viewer
- `collection_runs` metadata collection
- Persistence fan-out redesign (JSONL-then-Mongo in each partition persist loop remained as proved)

Not committed. Not pushed.
