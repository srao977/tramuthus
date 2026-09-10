# Bar Sequence Lab Generator — Phase 1.5 JSONL Implementation Report

**Date:** 2026-09-10  
**Status:** LOCAL IMPLEMENTATION COMPLETE (not committed, not pushed)  
**Persistence proving mode:** JSONL  
**MongoDB:** NOT created, NOT implemented, NOT connected

---

1. **Standalone module**  
   [Bar_Sequence_Lab/generator](Bar_Sequence_Lab/generator) `module bar_sequence_lab/generator`. No Fin import/`replace`. Cmd: [cmd/generator/main.go](Bar_Sequence_Lab/generator/cmd/generator/main.go). Packages: `internal/{config,alpaca,buffer,sequence,persistence,pipeline,runtime,types}`.

2. **PowerShell startup**  
   [scripts/Start-BarSequenceLab.ps1](Bar_Sequence_Lab/generator/scripts/Start-BarSequenceLab.ps1). Analog of Fin script (read-only). Remaining args = groups. No args → usage + exit 1. No default A+B+C.

3. **Group selection**  
   Combinations A | B | C | A B | A C | B C | A B C. Invalid name fails. Empty selected group fails. Duplicate symbol across selected groups fails. One process. Subscribe = union of selected groups only.

4. **Lab-owned group env**  
   `BAR_SEQ_LAB_GROUP_A/B/C` comma-separated (Fin split style). Fin has no A/B/C lists. Test-feed defaults: A=`FAKEPACA`, B/C empty. IEX defaults: AAPL / MSFT / NVDA (operator-readable, not frozen scientific assignment; LAB-GEN-OE-02 still open).

5. **JSONL persistence**  
   `data\<collection_run_id>\partition_{A\|B\|C}.jsonl` for active partitions only. One JSON object per line. No fake Mongo fields. Independent writer per active partition.

6. **Approved V0.1**  
   Architectural authority unchanged. Mongo remains the plan’s initial store. This increment documents JSONL as **proving deviation**. V0.1 was not rewritten.

7. **Alpaca**  
   Adapted WS auth/subscribe/decode/heartbeat/reconnect from Fin. Subscribe selected union only. T=b and T=u preserved (OE-10 open). No gap fill. No silent drop of valid bars.

8. **Sequence**  
   `generator_sequence_no` = monotonic accepted-arrival per symbol per `collection_run_id`. `source_event_time` preserved. Regression flagged (`source_time_regression`) and counted; no reorder/renumber.

9. **Duplicates**  
   Detected, persisted with `duplicate_arrival`, counted. Evidence kept. OE-08 not closed.

10. **Buffers**  
    Per active group: ingress → sequence → ready → persist batch → JSONL writer. Bounded; full = critical error, not overwrite/drop.

11. **Inactive partitions**  
    No pipeline, no JSONL file, not in dispatcher map.

12. **Observability**  
    Startup banner (no secrets), periodic metrics, shutdown flush report, `dropped_valid target=0`.

13. **Ctrl+C**  
    Stop ingest, close ingress, drain A1→A3 independently, flush JSONL, print paths/counts.

14. **Tests**  
    Config groups/duplicates/empty/secrets; decode b/u + malformed; sequence arrival/regression/duplicate; buffer order + full; JSONL lines; partition pipeline; engine selected-only / independent persist / new run id / concurrent writers.

15. **Live Alpaca**  
    Credentials were present (values not printed). Short proving run `groups=A feed=test max_bars=1 duration=25s`: auth/subscribe succeeded (`FAKEPACA`, heartbeat healthy, no secrets logged). No bar arrived in the 25s window, so first JSONL observation is still pending. Connection proof is not a bar-persist proof.

16. **Not done / not authorized**  
    Viewer, maths, Ehlers, wave, phase, Hop, DSE, Fin modifications, HACCAM, Mongo.

17. **OE still open**  
    OE-02 assignment algorithm, OE-04 schema names, OE-05 numeric flush defaults (placeholders 50 / 200ms), OE-08 duplicate disposition, OE-09 later analytical handling, OE-10 forming vs final persist policy, OE-11 retry policy detail, OE-12 exact concurrency mechanism (progress independence closed).

18. **Fin unmodified**  
    Both tramuthus `Fin_FeedSat_1/` and standalone `C:\Users\chino\Fin_FeedSat_1` treated read-only.

19. **DSE unmodified**  
    Process model and uncommitted architectural design left untouched.

20. **Viewer unmodified**  
    `Bar_Sequence_Lab/viewer` not modified.

21. **Git**  
    NOT COMMITTED. NOT PUSHED.

22. **Contradiction check**  
    JSONL vs Mongo is an authorized proving override in the implementation task, not a V0.1 rewrite. No STOP-level contradiction treated as requiring plan edit.

23. **Physical names frozen by this increment**  
    `cmd/generator`, `internal/config`, `internal/alpaca`, `internal/buffer`, `internal/sequence`, `internal/persistence`, `internal/pipeline`, `internal/runtime`, `scripts/Start-BarSequenceLab.ps1`.

24. **JSONL fields written**  
    collection_run_id, partition_id, symbol, generator_sequence_no, source_event_time, source_timestamp_text, received_time, persisted_time, interval, OHLC, volume, event_count, source_id, alpaca_message_type, payload_hash, source_time_regression, duplicate_arrival.

25. **Unknown symbols**  
    Rejected with diagnostic; not a valid-bar drop.

26. **Restart**  
    New process ⇒ new `collection_run_id` (UTC timestamp). No silent continuity.

27. **Symbol cap**  
    Operational MaxSymbolsBasicPlan=30 from Fin evidence; not an architectural law.

28. **Dependencies**  
    Go 1.25.3; `github.com/gorilla/websocket` (ingestion only).

29. **How to run**  
    See [README.md](Bar_Sequence_Lab/generator/README.md).

30. **Closeout**  
    NOT COMMITTED. NOT PUSHED. LOCAL JSONL PERSISTENCE IMPLEMENTED FOR LAB PROVING. MONGODB NOT CREATED OR IMPLEMENTED. A/B/C GROUPS SELECTABLE FROM POWERSHELL STARTUP COMMAND. BAR_SEQUENCE_LAB/VIEWER NOT MODIFIED. FIN_FEEDSAT_1 NOT MODIFIED. DSE NOT MODIFIED.
