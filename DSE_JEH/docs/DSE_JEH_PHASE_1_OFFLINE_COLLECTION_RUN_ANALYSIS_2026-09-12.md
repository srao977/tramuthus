# DSE_JEH Phase 1 OFFLINE Collection Run Analysis

| Item | Value |
| --- | --- |
| Date | 2026-09-12 |
| Purpose | Describe the current OFFLINE source dataset in collection-run-first order and reconcile Phase 1 initialization |
| Dataset | `bar_sequence_db.bar_sequence` |
| Classification evidence | `DSE_JEH/exports/initializing_diagnostic.jsonl` |
| Grouping hierarchy | `collection_run_id -> symbol -> generator_sequence_no -> received_time` |
| Scope key | `collection_run_id + "|" + normalized symbol` |

All timestamps are UTC. Mongo source identity and timing are carried into PhaseEvidence provenance. The report uses `received_time` as acquisition chronology and also reports `source_event_time` for comparison. No production code, JEH mathematics, warm-up rule, state ownership, or Phase 2 behavior was changed.

## 1. Executive Summary

| Metric | Value |
| --- | ---: |
| Total source observations | 4,519 |
| Total admitted observations | 4,519 |
| Distinct collection runs | 3 |
| Distinct symbols | 30 |
| Collection-run/symbol scopes | 63 |
| INITIALIZING | 3,247 |
| OBSERVABLE | 1,272 |
| INVALID | 0 |
| Rejected | 0 |

The current 3,247 INITIALIZING records are fully reconciled by the collection-run/symbol structure:

```text
3,247 + 1,272 = 4,519

SUM over collection_run_id + symbol scopes:
    MIN(admitted bars in scope, 63)
= 3,247
```

All 63 scopes have zero initialization difference. There is no unexplained initialization behavior in the current run.

## 2. Collection Run Inventory

Runs are sorted by earliest `received_time`.

| Collection run | Symbols | Source bars | Admitted | INITIALIZING | OBSERVABLE | INVALID | Earliest received | Latest received | Earliest source event | Latest source event | Received duration |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | --- | --- | --- | --- | --- |
| `20260910T190625Z-1` | 3 | 6 | 6 | 6 | 0 | 0 | 2026-09-10 19:07:00.557 | 2026-09-10 19:08:00.601 | 2026-09-10 19:06:00.000 | 2026-09-10 19:07:00.000 | 00:01:00.044 |
| `20260910T191246Z-1` | 30 | 1,393 | 1,393 | 1,393 | 0 | 0 | 2026-09-10 19:13:00.558 | 2026-09-10 20:05:00.476 | 2026-09-10 19:12:00.000 | 2026-09-10 20:04:00.000 | 00:51:59.918 |
| `20260911T161623Z-1` | 30 | 3,120 | 3,120 | 1,848 | 1,272 | 0 | 2026-09-11 16:17:01.027 | 2026-09-11 18:16:00.948 | 2026-09-11 16:16:00.000 | 2026-09-11 18:15:00.000 | 01:58:59.921 |

## 3. Collection Run `20260910T190625Z-1`

| Property | Value |
| --- | --- |
| Earliest/latest received | 2026-09-10 19:07:00.557 / 19:08:00.601 |
| Earliest/latest source event | 2026-09-10 19:06:00.000 / 19:07:00.000 |
| Symbols / bars | 3 / 6 |
| INITIALIZING / OBSERVABLE | 6 / 0 |

The table is alphabetical. `Exp I = min(admitted, 63)` and `Diff = actual I - Exp I`. `I recv` is first through final INITIALIZING receipt time. There is no observable range in this run.

| Symbol | Part. | Adm. | I | O | Exp I | Diff | Seq | Received first..last | Source first..last | I range | I recv first..final | First O seq/time |
| --- | --- | ---: | ---: | ---: | ---: | ---: | --- | --- | --- | --- | --- | --- |
| AAPL | A | 2 | 2 | 0 | 2 | 0 | 1-2 | 19:07:00.557..19:08:00.601 | 19:06:00..19:07:00 | 1-2 | 19:07:00.557..19:08:00.601 | N/A |
| MSFT | B | 2 | 2 | 0 | 2 | 0 | 1-2 | 19:07:00.560..19:08:00.601 | 19:06:00..19:07:00 | 1-2 | 19:07:00.560..19:08:00.601 | N/A |
| NVDA | C | 2 | 2 | 0 | 2 | 0 | 1-2 | 19:07:00.567..19:08:00.601 | 19:06:00..19:07:00 | 1-2 | 19:07:00.567..19:08:00.601 | N/A |
| **Total** |  | **6** | **6** | **0** | **6** | **0** |  |  |  |  |  |  |

Every symbol starts at sequence 1. Each sequence is contiguous and unique; no gap, duplicate, out-of-order value, or in-run restart exists. Actual and expected INITIALIZING both equal 6.

## 4. Collection Run `20260910T191246Z-1`

| Property | Value |
| --- | --- |
| Earliest/latest received | 2026-09-10 19:13:00.558 / 20:05:00.476 |
| Earliest/latest source event | 2026-09-10 19:12:00.000 / 20:04:00.000 |
| Symbols / bars | 30 / 1,393 |
| INITIALIZING / OBSERVABLE | 1,393 / 0 |

All dates in this table are 2026-09-10. Every scope is shorter than 64 observations, so every admitted bar is INITIALIZING.

| Symbol | Part. | Adm. | I | O | Exp I | Diff | Seq | Received first..last | Source first..last | I range | I recv first..final | First O seq/time |
| --- | --- | ---: | ---: | ---: | ---: | ---: | --- | --- | --- | --- | --- | --- |
| AAPL | A | 51 | 51 | 0 | 51 | 0 | 1-51 | 19:13:00.562..20:04:00.470 | 19:12..20:03 | 1-51 | 19:13:00.562..20:04:00.470 | N/A |
| AMD | A | 47 | 47 | 0 | 47 | 0 | 1-47 | 19:13:00.569..20:00:00.784 | 19:12..19:59 | 1-47 | 19:13:00.569..20:00:00.784 | N/A |
| AMZN | A | 48 | 48 | 0 | 48 | 0 | 1-48 | 19:13:00.558..20:00:00.829 | 19:12..19:59 | 1-48 | 19:13:00.558..20:00:00.829 | N/A |
| AVGO | A | 48 | 48 | 0 | 48 | 0 | 1-48 | 19:13:00.558..20:00:00.762 | 19:12..19:59 | 1-48 | 19:13:00.558..20:00:00.762 | N/A |
| BAC | C | 47 | 47 | 0 | 47 | 0 | 1-47 | 19:13:00.565..20:00:00.776 | 19:12..19:59 | 1-47 | 19:13:00.565..20:00:00.776 | N/A |
| COST | C | 32 | 32 | 0 | 32 | 0 | 1-32 | 19:13:00.574..20:00:00.776 | 19:12..19:59 | 1-32 | 19:13:00.574..20:00:00.776 | N/A |
| DIA | B | 38 | 38 | 0 | 38 | 0 | 1-38 | 19:13:00.562..20:00:00.829 | 19:12..19:59 | 1-38 | 19:13:00.562..20:00:00.829 | N/A |
| EEM | B | 45 | 45 | 0 | 45 | 0 | 1-45 | 19:13:00.565..20:00:00.829 | 19:12..19:59 | 1-45 | 19:13:00.565..20:00:00.829 | N/A |
| GLD | B | 47 | 47 | 0 | 47 | 0 | 1-47 | 19:13:00.562..20:00:00.795 | 19:12..19:59 | 1-47 | 19:13:00.562..20:00:00.795 | N/A |
| GOOGL | A | 48 | 48 | 0 | 48 | 0 | 1-48 | 19:13:00.558..20:00:00.829 | 19:12..19:59 | 1-48 | 19:13:00.558..20:00:00.829 | N/A |
| HD | C | 45 | 45 | 0 | 45 | 0 | 1-45 | 19:13:00.575..20:00:00.835 | 19:12..19:59 | 1-45 | 19:13:00.575..20:00:00.835 | N/A |
| IWM | B | 50 | 50 | 0 | 50 | 0 | 1-50 | 19:13:00.565..20:03:00.325 | 19:12..20:02 | 1-50 | 19:13:00.565..20:03:00.325 | N/A |
| JNJ | C | 44 | 44 | 0 | 44 | 0 | 1-44 | 19:13:00.562..20:00:00.765 | 19:12..19:59 | 1-44 | 19:13:00.562..20:00:00.765 | N/A |
| JPM | C | 46 | 46 | 0 | 46 | 0 | 1-46 | 19:13:00.558..20:00:00.772 | 19:12..19:59 | 1-46 | 19:13:00.558..20:00:00.772 | N/A |
| MA | C | 43 | 43 | 0 | 43 | 0 | 1-43 | 19:13:00.569..20:00:00.781 | 19:12..19:59 | 1-43 | 19:13:00.569..20:00:00.781 | N/A |
| META | A | 48 | 48 | 0 | 48 | 0 | 1-48 | 19:13:00.562..20:00:00.768 | 19:12..19:59 | 1-48 | 19:13:00.562..20:00:00.768 | N/A |
| MSFT | A | 48 | 48 | 0 | 48 | 0 | 1-48 | 19:13:00.562..20:00:00.835 | 19:12..19:59 | 1-48 | 19:13:00.562..20:00:00.835 | N/A |
| NFLX | A | 49 | 49 | 0 | 49 | 0 | 1-49 | 19:13:00.574..20:00:30.393 | 19:12..19:59 | 1-49 | 19:13:00.574..20:00:30.393 | N/A |
| NVDA | A | 49 | 49 | 0 | 49 | 0 | 1-49 | 19:13:00.558..20:00:00.835 | 19:12..19:59 | 1-49 | 19:13:00.558..20:00:00.835 | N/A |
| QQQ | B | 48 | 48 | 0 | 48 | 0 | 1-48 | 19:13:00.575..20:00:00.829 | 19:12..19:59 | 1-48 | 19:13:00.575..20:00:00.829 | N/A |
| SPY | B | 48 | 48 | 0 | 48 | 0 | 1-48 | 19:13:00.558..20:00:00.789 | 19:12..19:59 | 1-48 | 19:13:00.558..20:00:00.789 | N/A |
| TLT | B | 51 | 51 | 0 | 51 | 0 | 1-51 | 19:13:00.562..20:05:00.476 | 19:12..20:04 | 1-51 | 19:13:00.562..20:05:00.476 | N/A |
| TSLA | A | 48 | 48 | 0 | 48 | 0 | 1-48 | 19:13:00.569..20:00:00.784 | 19:12..19:59 | 1-48 | 19:13:00.569..20:00:00.784 | N/A |
| UNH | C | 48 | 48 | 0 | 48 | 0 | 1-48 | 19:13:00.558..20:00:00.796 | 19:12..19:59 | 1-48 | 19:13:00.558..20:00:00.796 | N/A |
| V | C | 48 | 48 | 0 | 48 | 0 | 1-48 | 19:13:00.562..20:00:00.829 | 19:12..19:59 | 1-48 | 19:13:00.562..20:00:00.829 | N/A |
| VXX | B | 44 | 44 | 0 | 44 | 0 | 1-44 | 19:14:00.572..20:00:00.765 | 19:13..19:59 | 1-44 | 19:14:00.572..20:00:00.765 | N/A |
| WMT | C | 48 | 48 | 0 | 48 | 0 | 1-48 | 19:13:00.575..20:00:00.829 | 19:12..19:59 | 1-48 | 19:13:00.575..20:00:00.829 | N/A |
| XLF | B | 49 | 49 | 0 | 49 | 0 | 1-49 | 19:13:00.562..20:00:00.795 | 19:12..19:59 | 1-49 | 19:13:00.562..20:00:00.795 | N/A |
| XLK | B | 43 | 43 | 0 | 43 | 0 | 1-43 | 19:13:00.569..20:00:00.796 | 19:12..19:59 | 1-43 | 19:13:00.569..20:00:00.796 | N/A |
| XOM | C | 45 | 45 | 0 | 45 | 0 | 1-45 | 19:13:00.575..20:00:00.829 | 19:12..19:59 | 1-45 | 19:13:00.575..20:00:00.829 | N/A |
| **Total** |  | **1,393** | **1,393** | **0** | **1,393** | **0** |  |  |  |  |  |  |

Every symbol starts at sequence 1. Every scope is contiguous and unique; no gap, duplicate, out-of-order sequence, or in-run restart exists. Actual and expected INITIALIZING both equal 1,393.

## 5. Collection Run `20260911T161623Z-1`

| Property | Value |
| --- | --- |
| Earliest/latest received | 2026-09-11 16:17:01.027 / 18:16:00.948 |
| Earliest/latest source event | 2026-09-11 16:16:00.000 / 18:15:00.000 |
| Symbols / bars | 30 / 3,120 |
| INITIALIZING / OBSERVABLE | 1,848 / 1,272 |

All dates in this table are 2026-09-11. `O range` is the complete observable sequence range.

| Symbol | Part. | Adm. | I | O | Exp I | Diff | Seq | Received first..last | Source first..last | I range / I recv first..final | O range / first O received |
| --- | --- | ---: | ---: | ---: | ---: | ---: | --- | --- | --- | --- | --- |
| AAPL | A | 120 | 63 | 57 | 63 | 0 | 1-120 | 16:17:01.032..18:16:00.941 | 16:16..18:15 | 1-63 / 16:17:01.032..17:19:00.985 | 64-120 / 17:20:00.987 |
| AMD | A | 113 | 63 | 50 | 63 | 0 | 1-113 | 16:17:01.027..18:16:00.936 | 16:16..18:15 | 1-63 / 16:17:01.027..17:25:00.973 | 64-113 / 17:26:00.988 |
| AMZN | A | 119 | 63 | 56 | 63 | 0 | 1-119 | 16:17:01.041..18:16:00.941 | 16:16..18:15 | 1-63 / 16:17:01.041..17:20:00.987 | 64-119 / 17:21:00.981 |
| AVGO | A | 118 | 63 | 55 | 63 | 0 | 1-118 | 16:17:01.035..18:16:00.941 | 16:16..18:15 | 1-63 / 16:17:01.035..17:21:00.972 | 64-118 / 17:22:00.985 |
| BAC | C | 117 | 63 | 54 | 63 | 0 | 1-117 | 16:17:01.035..18:16:00.935 | 16:16..18:15 | 1-63 / 16:17:01.035..17:20:00.978 | 64-117 / 17:21:00.975 |
| COST | C | 80 | 63 | 17 | 63 | 0 | 1-80 | 16:20:01.023..18:16:00.944 | 16:19..18:15 | 1-63 / 16:20:01.023..17:41:00.962 | 64-80 / 17:42:00.958 |
| DIA | B | 33 | 33 | 0 | 33 | 0 | 1-33 | 16:17:01.027..18:14:00.939 | 16:16..18:13 | 1-33 / 16:17:01.027..18:14:00.939 | None |
| EEM | B | 99 | 63 | 36 | 63 | 0 | 1-99 | 16:17:01.027..18:16:00.941 | 16:16..18:15 | 1-63 / 16:17:01.027..17:35:00.995 | 64-99 / 17:36:00.787 |
| GLD | B | 87 | 63 | 24 | 63 | 0 | 1-87 | 16:17:01.037..18:16:00.941 | 16:16..18:15 | 1-63 / 16:17:01.037..17:44:00.797 | 64-87 / 17:45:00.984 |
| GOOGL | A | 115 | 63 | 52 | 63 | 0 | 1-115 | 16:17:01.037..18:16:00.945 | 16:16..18:15 | 1-63 / 16:17:01.037..17:23:00.970 | 64-115 / 17:24:00.985 |
| HD | C | 88 | 63 | 25 | 63 | 0 | 1-88 | 16:18:00.915..18:16:00.944 | 16:17..18:15 | 1-63 / 16:18:00.915..17:43:00.977 | 64-88 / 17:44:00.782 |
| IWM | B | 119 | 63 | 56 | 63 | 0 | 1-119 | 16:17:01.032..18:16:00.941 | 16:16..18:15 | 1-63 / 16:17:01.032..17:20:00.987 | 64-119 / 17:21:00.981 |
| JNJ | C | 86 | 63 | 23 | 63 | 0 | 1-86 | 16:19:00.908..18:15:00.937 | 16:18..18:14 | 1-63 / 16:19:00.908..17:43:00.968 | 64-86 / 17:45:00.980 |
| JPM | C | 105 | 63 | 42 | 63 | 0 | 1-105 | 16:17:01.032..18:16:00.941 | 16:16..18:15 | 1-63 / 16:17:01.032..17:29:00.986 | 64-105 / 17:30:00.972 |
| MA | C | 89 | 63 | 26 | 63 | 0 | 1-89 | 16:17:01.027..18:16:00.945 | 16:16..18:15 | 1-63 / 16:17:01.027..17:37:00.970 | 64-89 / 17:38:00.950 |
| META | A | 120 | 63 | 57 | 63 | 0 | 1-120 | 16:17:01.027..18:16:00.944 | 16:16..18:15 | 1-63 / 16:17:01.027..17:19:00.985 | 64-120 / 17:20:00.987 |
| MSFT | A | 114 | 63 | 51 | 63 | 0 | 1-114 | 16:17:01.027..18:16:00.935 | 16:16..18:15 | 1-63 / 16:17:01.027..17:20:00.978 | 64-114 / 17:21:00.972 |
| NFLX | A | 120 | 63 | 57 | 63 | 0 | 1-120 | 16:17:01.034..18:16:00.935 | 16:16..18:15 | 1-63 / 16:17:01.034..17:19:00.987 | 64-120 / 17:20:00.978 |
| NVDA | A | 120 | 63 | 57 | 63 | 0 | 1-120 | 16:17:01.027..18:16:00.935 | 16:16..18:15 | 1-63 / 16:17:01.027..17:19:00.991 | 64-120 / 17:20:00.978 |
| QQQ | B | 116 | 63 | 53 | 63 | 0 | 1-116 | 16:17:01.037..18:16:00.936 | 16:16..18:15 | 1-63 / 16:17:01.037..17:21:00.972 | 64-116 / 17:22:00.973 |
| SPY | B | 120 | 63 | 57 | 63 | 0 | 1-120 | 16:17:01.037..18:16:00.936 | 16:16..18:15 | 1-63 / 16:17:01.037..17:19:00.985 | 64-120 / 17:20:00.987 |
| TLT | B | 103 | 63 | 40 | 63 | 0 | 1-103 | 16:17:01.027..18:16:00.944 | 16:16..18:15 | 1-63 / 16:17:01.027..17:29:00.979 | 64-103 / 17:30:00.977 |
| TSLA | A | 120 | 63 | 57 | 63 | 0 | 1-120 | 16:17:01.034..18:16:00.948 | 16:16..18:15 | 1-63 / 16:17:01.034..17:19:00.987 | 64-120 / 17:20:00.987 |
| UNH | C | 116 | 63 | 53 | 63 | 0 | 1-116 | 16:17:01.027..18:16:00.935 | 16:16..18:15 | 1-63 / 16:17:01.027..17:19:00.980 | 64-116 / 17:20:00.978 |
| V | C | 101 | 63 | 38 | 63 | 0 | 1-101 | 16:18:00.915..18:16:00.944 | 16:17..18:15 | 1-63 / 16:18:00.915..17:34:00.995 | 64-101 / 17:35:00.995 |
| VXX | B | 51 | 51 | 0 | 51 | 0 | 1-51 | 16:22:01.038..18:16:00.945 | 16:21..18:15 | 1-51 / 16:22:01.038..18:16:00.945 | None |
| WMT | C | 116 | 63 | 53 | 63 | 0 | 1-116 | 16:17:01.027..18:16:00.936 | 16:16..18:15 | 1-63 / 16:17:01.027..17:20:00.987 | 64-116 / 17:21:00.975 |
| XLF | B | 118 | 63 | 55 | 63 | 0 | 1-118 | 16:18:00.852..18:16:00.944 | 16:17..18:15 | 1-63 / 16:18:00.852..17:21:00.975 | 64-118 / 17:22:00.980 |
| XLK | B | 91 | 63 | 28 | 63 | 0 | 1-91 | 16:18:00.915..18:16:00.941 | 16:17..18:15 | 1-63 / 16:18:00.915..17:32:00.986 | 64-91 / 17:33:00.973 |
| XOM | C | 106 | 63 | 43 | 63 | 0 | 1-106 | 16:18:00.915..18:16:00.945 | 16:17..18:15 | 1-63 / 16:18:00.915..17:27:00.982 | 64-106 / 17:28:00.973 |
| **Total** |  | **3,120** | **1,848** | **1,272** | **1,848** | **0** |  |  |  |  |  |

Every symbol starts at sequence 1. Every scope is contiguous and unique; no gap, duplicate, out-of-order sequence, or in-run restart exists. Actual and expected INITIALIZING both equal 1,848. DIA and VXX do not reach sequence 64 and therefore have no observable observations.

## 6. Sequence And Time Integrity

The §3-5 tables establish the sequence and receipt bounds for all 63 scopes. Machine checks over all evidence found:

- starts above 1: none;
- sequence gaps: none;
- duplicate sequence numbers: none;
- out-of-order sequences: none;
- sequence restarts inside a collection run/symbol scope: none;
- sequence restarts between collection runs: yes, expected and visible; each new scope starts at 1.

The hierarchy is therefore exactly `collection_run_id -> symbol -> generator_sequence_no`, with `received_time` documenting receipt chronology for each sequence.

## 7. INITIALIZING by Collection Run and Symbol

The complete tables in §3, §4, and §5 are the principal collection-run/symbol artifact. They contain every required row and include admitted, actual and expected INITIALIZING, initialization difference, observable count, sequence bounds, receipt bounds, source-event bounds, and first observable sequence/time.

All 63 `Diff` values are zero. No exception requires highlighting beyond this explicit result.

## 8. Initialization Ranges

The `I range / I recv` and `O range / first O received` columns in §3-5 are the complete range report for every collection-run/symbol scope.

- Runs `20260910T190625Z-1` and `20260910T191246Z-1` end before any scope reaches update 64; every range is INITIALIZING only.
- In `20260911T161623Z-1`, 28 scopes reach update 64 and transition exactly from INITIALIZING 1-63 to OBSERVABLE 64-N.
- DIA and VXX in that run contain 33 and 51 bars respectively and remain INITIALIZING throughout.

## 9. Secondary Cross-Run Symbol View

This view is intentionally secondary; each semicolon-separated contribution is `run: initializing + observable`.

| Symbol | Runs | Admitted | INITIALIZING | OBSERVABLE | Run contributions |
| --- | ---: | ---: | ---: | ---: | --- |
| AAPL | 3 | 173 | 116 | 57 | `190625: 2I+0O; 191246: 51I+0O; 161623: 63I+57O` |
| AMD | 2 | 160 | 110 | 50 | `191246: 47I+0O; 161623: 63I+50O` |
| AMZN | 2 | 167 | 111 | 56 | `191246: 48I+0O; 161623: 63I+56O` |
| AVGO | 2 | 166 | 111 | 55 | `191246: 48I+0O; 161623: 63I+55O` |
| BAC | 2 | 164 | 110 | 54 | `191246: 47I+0O; 161623: 63I+54O` |
| COST | 2 | 112 | 95 | 17 | `191246: 32I+0O; 161623: 63I+17O` |
| DIA | 2 | 71 | 71 | 0 | `191246: 38I+0O; 161623: 33I+0O` |
| EEM | 2 | 144 | 108 | 36 | `191246: 45I+0O; 161623: 63I+36O` |
| GLD | 2 | 134 | 110 | 24 | `191246: 47I+0O; 161623: 63I+24O` |
| GOOGL | 2 | 163 | 111 | 52 | `191246: 48I+0O; 161623: 63I+52O` |
| HD | 2 | 133 | 108 | 25 | `191246: 45I+0O; 161623: 63I+25O` |
| IWM | 2 | 169 | 113 | 56 | `191246: 50I+0O; 161623: 63I+56O` |
| JNJ | 2 | 130 | 107 | 23 | `191246: 44I+0O; 161623: 63I+23O` |
| JPM | 2 | 151 | 109 | 42 | `191246: 46I+0O; 161623: 63I+42O` |
| MA | 2 | 132 | 106 | 26 | `191246: 43I+0O; 161623: 63I+26O` |
| META | 2 | 168 | 111 | 57 | `191246: 48I+0O; 161623: 63I+57O` |
| MSFT | 3 | 164 | 113 | 51 | `190625: 2I+0O; 191246: 48I+0O; 161623: 63I+51O` |
| NFLX | 2 | 169 | 112 | 57 | `191246: 49I+0O; 161623: 63I+57O` |
| NVDA | 3 | 171 | 114 | 57 | `190625: 2I+0O; 191246: 49I+0O; 161623: 63I+57O` |
| QQQ | 2 | 164 | 111 | 53 | `191246: 48I+0O; 161623: 63I+53O` |
| SPY | 2 | 168 | 111 | 57 | `191246: 48I+0O; 161623: 63I+57O` |
| TLT | 2 | 154 | 114 | 40 | `191246: 51I+0O; 161623: 63I+40O` |
| TSLA | 2 | 168 | 111 | 57 | `191246: 48I+0O; 161623: 63I+57O` |
| UNH | 2 | 164 | 111 | 53 | `191246: 48I+0O; 161623: 63I+53O` |
| V | 2 | 149 | 111 | 38 | `191246: 48I+0O; 161623: 63I+38O` |
| VXX | 2 | 95 | 95 | 0 | `191246: 44I+0O; 161623: 51I+0O` |
| WMT | 2 | 164 | 111 | 53 | `191246: 48I+0O; 161623: 63I+53O` |
| XLF | 2 | 167 | 112 | 55 | `191246: 49I+0O; 161623: 63I+55O` |
| XLK | 2 | 134 | 106 | 28 | `191246: 43I+0O; 161623: 63I+28O` |
| XOM | 2 | 151 | 108 | 43 | `191246: 45I+0O; 161623: 63I+43O` |
| **Total** | **63 scopes** | **4,519** | **3,247** | **1,272** |  |

## 10. Received-Time Chronology

| Run | Received range | Duration | Relationship to preceding run |
| --- | --- | --- | --- |
| `20260910T190625Z-1` | 19:07:00.557..19:08:00.601 | 00:01:00.044 | First run |
| `20260910T191246Z-1` | 19:13:00.558..20:05:00.476 | 00:51:59.918 | Separate; gap 00:04:59.957 |
| `20260911T161623Z-1` | 16:17:01.027..18:16:00.948 next day | 01:58:59.921 | Separate; gap 20:12:00.551 |

The received-time ranges do not overlap and are not contiguous. They are three clearly separate capture intervals, with a roughly five-minute gap between the first two and a 20-hour-12-minute gap before the third.

Each run's source-event range begins approximately one minute before its received range and ends approximately one minute before it. The source spans are 1 minute, 52 minutes, and 1 hour 59 minutes respectively; received spans differ only by sub-second receipt timing. Nothing in these timestamps establishes operator intent, but they do establish separate acquisition sessions.

## 11. Reconciliation of INITIALIZING = 3,247

```text
20260910T190625Z-1    INITIALIZING =     6
20260910T191246Z-1    INITIALIZING = 1,393
20260911T161623Z-1    INITIALIZING = 1,848
                                      -----
TOTAL                               = 3,247
```

Equivalently:

```text
SUM over collection_run_id:
    SUM over symbol:
        MIN(admitted_bars, 63)
= 6 + 1,393 + 1,848
= 3,247
```

This equals the actual Phase 1 classification exactly.

## 12. Why 1,890 Describes a Different Aggregation

`30 symbols x 63 = 1,890` assumes one continuous analytical history per symbol over the combined collection. That is not the state topology used by the current OFFLINE run. Its key is `collection_run_id|symbol`, so sequence restarts in a different collection run select a new solver state rather than extending an earlier run's symbol history.

The current collection has 63 run-symbol scopes: 3 scopes in the first run, 30 in the second, and 30 in the third. The report does not judge whether a future replay should select, merge, or independently process runs; it only distinguishes the two counting models.

## 13. Historical 3,712 Figure

```text
Historical/previously observed INITIALIZING: approximately 3,712
Current reproduced INITIALIZING:                         3,247
Difference:                                                465
```

The current 4,519-record source dataset and current PhaseEvidence classifications do not reproduce 3,712 INITIALIZING records. No direct evidence inspected in this analysis establishes the historical cause, so none is invented.

## 14. Factual Conclusions

1. Three collection runs are present.
2. They are `20260910T190625Z-1`, `20260910T191246Z-1`, and `20260911T161623Z-1`.
3. Their received ranges are 19:07:00.557-19:08:00.601, 19:13:00.558-20:05:00.476, and next-day 16:17:01.027-18:16:00.948 UTC.
4. They contain 3, 30, and 30 symbols.
5. They contain 6, 1,393, and 3,120 bars.
6. They contribute 6, 1,393, and 1,848 INITIALIZING bars.
7. Every symbol's run-specific contribution is enumerated in §3-5 and summarized across runs in §9.
8. `generator_sequence_no` restarts at 1 between collection runs.
9. No sequence restarts inside any collection-run/symbol scope.
10. Every scope obeys the existing 63-bar initialization rule: actual INITIALIZING equals `min(scope admitted, 63)`.
11. Scope-level arithmetic explains all 3,247 INITIALIZING records exactly.
12. No unexplained initialization behavior remains in the current run.

This report does not decide whether future OFFLINE execution should replay all runs, select one run, merge runs, or use another policy.

## 15. Console Summary

```text
DSE_JEH PHASE 1 OFFLINE COLLECTION RUN ANALYSIS

Source bars: 4519
Collection runs: 3
Symbols: 30
Collection-run/symbol scopes: 63

Run 1:
    collection_run_id: 20260910T190625Z-1
    received_time range: 2026-09-10T19:07:00.557Z to 2026-09-10T19:08:00.601Z
    symbols: 3
    bars: 6
    INITIALIZING: 6
    OBSERVABLE: 0

Run 2:
    collection_run_id: 20260910T191246Z-1
    received_time range: 2026-09-10T19:13:00.558Z to 2026-09-10T20:05:00.476Z
    symbols: 30
    bars: 1393
    INITIALIZING: 1393
    OBSERVABLE: 0

Run 3:
    collection_run_id: 20260911T161623Z-1
    received_time range: 2026-09-11T16:17:01.027Z to 2026-09-11T18:16:00.948Z
    symbols: 30
    bars: 3120
    INITIALIZING: 1848
    OBSERVABLE: 1272

Total INITIALIZING: 3247
Expected by collection-run/symbol scope: 3247
Difference: 0

Historical 3,712 reproduced: NO

Unexplained initialization behavior: NO

Production code changed: NO
```
