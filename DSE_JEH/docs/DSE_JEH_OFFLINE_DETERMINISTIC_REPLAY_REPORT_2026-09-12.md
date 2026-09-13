# DSE_JEH Offline Deterministic Replay Report

| Item | Value |
| --- | --- |
| Date | 2026-09-12 |
| Source | `bar_sequence_db.bar_sequence` |
| Mode | `OFFLINE` |
| Replay policy | `COLLECTION_RUN_SYMBOL_ENTITY_SEQUENCE_PAYLOAD_V1` |

## Purpose

Validate the complete OFFLINE Phase 1 path using the actual local Mongo collection and prove repeatable PhaseEvidence output.

## Replay Policy

The source has authoritative order only within collection run and entity. The producer sorts deterministically by collection run, normalized symbol, entity-scoped `generator_sequence_no`, and payload hash. This is a versioned replay order, not a reconstructed global market clock. Mongo data is read only.

## Commands

```powershell
$env:DSE_JEH_MODE='OFFLINE'
$env:BAR_SEQ_LAB_MONGO_URI='mongodb://127.0.0.1:27017'
$env:BAR_SEQ_LAB_MONGO_DB='bar_sequence_db'
$env:BAR_SEQ_LAB_MONGO_COLLECTION='bar_sequence'
$env:DSE_JEH_COLLECTION_RUN_ID=''
$env:DSE_JEH_OUTPUT='exports\phase_evidence_final1.jsonl'
.\bin\dse-jeh-transsat-1.exe

$env:DSE_JEH_OUTPUT='exports\phase_evidence_final2.jsonl'
.\bin\dse-jeh-transsat-1.exe
```

## Results

Both executions exited with code `0`.

| Metric | Run 1 | Run 2 |
| --- | ---: | ---: |
| Source observations read | 4,519 | 4,519 |
| Admitted | 4,519 | 4,519 |
| Initializing evidence | 3,247 | 3,247 |
| Observable evidence | 1,272 | 1,272 |
| Invalid | 0 | 0 |
| Duplicate | 0 | 0 |
| Conflict | 0 | 0 |
| Gap | 0 | 0 |
| Out of order | 0 | 0 |
| Rejected | 0 | 0 |
| Evidence lines | 4,519 | 4,519 |

Run 1 SHA-256:

```text
f30759c4b164c84a57f108f2817bf6eeadad435f97c3fd54708be3f7982f504e
```

Run 2 SHA-256:

```text
f30759c4b164c84a57f108f2817bf6eeadad435f97c3fd54708be3f7982f504e
```

The files are byte-for-byte equal.

## Initialization

Every admitted bar emits PhaseEvidence. The first 63 observations for each sufficiently long entity trajectory are `INITIALIZING`; observation 64 is the first possible `OBSERVABLE` result. The aggregate counts reflect multiple collection runs and entity trajectories.

## Failures And Deviations

None occurred during either replay. The deterministic replay order is intentionally not claimed to be original cross-entity arrival order.

## Next Action

Retain both evidence files and their digest as the Phase 1 replay baseline. Recompute the baseline only under an explicitly reviewed contract, solver, source-data, or replay-policy change.
