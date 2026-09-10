# Bar Sequence Lab generator — Mongo raw persistence live proving

Date: 2026-09-10  
Run: `20260910T190625Z-1`  
Verdict: **PASS** (3-symbol A+B+C IEX dual-sink; stop here)

JSONL audit remains always-on. Mongo is an independent raw sink into existing `bar_sequence_db.bar_sequence`. Maths, waves, Ehlers, Hop, viewer, Fin_FeedSat_1, and DSE were not implemented or modified.

## What was implemented

- Opt-in Mongo via `BAR_SEQ_LAB_MONGO_ENABLED=true` plus `BAR_SEQ_LAB_MONGO_URI` and `BAR_SEQ_LAB_MONGO_DB`. Collection default `bar_sequence`.
- One shared Mongo client/pool/database/collection. Per-partition `MongoWriter` handles share that store.
- Dual write in each partition persist loop: JSONL first, then Mongo. Slow Mongo on one partition does not serialize other partitions (three persist goroutines).
- JSONL write failures retry (no silent drop). Mongo write failures retry except duplicate-key (code 11000), which is CRITICAL and stops that persist loop. Existing indexes were not modified.
- Sequence index is **not unique**. Duplicate arrivals still receive the next `generator_sequence_no`; uniqueness would destroy evidence.
- Startup/metrics/shutdown report JSONL and Mongo separately.

Driver: `go.mongodb.org/mongo-driver/v2 v2.9.1`.

## Live run

Command (credentials not printed):

```
.\scripts\Start-BarSequenceLab.ps1 A B C -Feed iex -MaxBars 6 -Duration 3m
```

with `BAR_SEQ_LAB_MONGO_ENABLED=true`, URI/DB set, groups AAPL/MSFT/NVDA.

| Item | Value |
| --- | --- |
| collection_run_id | `20260910T190625Z-1` |
| feed | iex (`ALPACA_IEX`) |
| subscribe | AAPL, MSFT, NVDA |
| persistence | jsonl+mongo host=`127.0.0.1:27017` db=`bar_sequence_db` collection=`bar_sequence` |
| start | 2026-09-10 12:06:25 local / 19:06:25Z |
| stop | max_bars=6 at 12:08:00 local |
| alpaca | healthy; heartbeat_misses=0 |
| dropped_valid | 0 |
| unknown_symbol | 0 |
| source_time_regressions | 0 |
| duplicate_arrivals | 0 |
| jsonl_unflushed_total | 0 |
| mongo_unflushed_total | 0 |
| mongo_failed_total | 0 |

### Per partition (accepted = JSONL persisted = Mongo persisted)

| Partition | Symbol | accepted | jsonl | mongo | critical | pending |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| A | AAPL | 2 | 2 | 2 | 0 | 0 |
| B | MSFT | 2 | 2 | 2 | 0 | 0 |
| C | NVDA | 2 | 2 | 2 | 0 | 0 |

JSONL paths:

- `data/20260910T190625Z-1/partition_A.jsonl`
- `data/20260910T190625Z-1/partition_B.jsonl`
- `data/20260910T190625Z-1/partition_C.jsonl`

## Mongo query proof

Existing Compass collection used. No drop/recreate. No `bar_sequence_A/B/C`. No `collection_runs` metadata invented.

After the run:

- `countDocuments({collection_run_id: "20260910T190625Z-1"})` = **6**
- `countDocuments({})` = **6** (collection was empty before this run)
- distinct symbols = AAPL, MSFT, NVDA
- distinct partition_id = A, B, C
- each symbol ordered by `generator_sequence_no`: 1 then 2, no gaps

Indexes after the run (unchanged; `unique` not set):

- `_id_`
- `ix_run_symbol_sequence` `{collection_run_id, symbol, generator_sequence_no}` not unique
- `ix_run_partition` `{collection_run_id, partition_id}` not unique

## Representative field match (AAPL seq 1)

JSONL and Mongo agree on identity and payload. BSON `Long`/`ISODate` are type encoding only.

| Field | JSONL | Mongo |
| --- | --- | --- |
| collection_run_id | 20260910T190625Z-1 | 20260910T190625Z-1 |
| partition_id | A | A |
| symbol | AAPL | AAPL |
| generator_sequence_no | 1 | 1 |
| source_event_time / source_timestamp_text | 2026-09-10T19:06:00Z | 2026-09-10T19:06:00Z |
| interval | 1Min | 1Min |
| open/high/low/close | 325.065 / 325.275 / 324.99 / 325.22 | same |
| volume | 3107 | 3107 |
| event_count | 58 | 58 |
| source_id | ALPACA_IEX | ALPACA_IEX |
| alpaca_message_type | b | b |
| payload_hash | 9306699ab6b3c5128214e0704a565100361a580b6e7bfb3d18f1cf3d8a79c5cc | same |
| source_time_regression | false | false |
| duplicate_arrival | false | false |

Same identity match holds for AAPL seq 2, MSFT 1–2, NVDA 1–2 (`payload_hash` equal per record).

Received/persisted times match at millisecond resolution; JSONL stores extra sub-ms digits that BSON Date truncates.

## Scope held

- Stopped after 3-symbol success. Did not expand to 30 symbols.
- Did not commit. Did not push.
- Did not modify Fin_FeedSat_1, DSE worker/viewer, or approved generator V0.1 design (no contradiction found).
- JSONL retained. Mongo opt-in. No silent drop.
