# DSE_JEH Dynamic Pipeline Execution Run Report

## Scope

Runtime Objective 1 was executed twice through the existing OFFLINE pipeline, V1 state actions, `GovernedExecutionInstruction`, deterministic local `MockExecutor`, authoritative confirmed `ExecutionEvent`, and MongoDB stage/Emit persistence. No broker or Alpaca endpoint was used.

The explicitly supplied `collection_run_id = 20260911T161623Z-1` was verified and frozen independently at startup for both runs. Each run consumed 3,120 usable records across 30 distinct symbols. No latest-run lookup or substitution was performed.

Before execution, the intentionally recreated `h_h_stage_emit_values` collection was verified once as empty and authoritative. Its validator requires top-level `run_type` with enum values `RUN_A` and `RUN_B`; its six indexes comprise MongoDB `_id_` plus the five governed trace indexes. No validator or index was replaced at runtime.

## Final Runs

| Run | run_type | pipeline_run_id | StartingCapital | AllocationPct | Risk R | Result |
| --- | --- | --- | ---: | ---: | ---: | --- |
| A | `RUN_A` | `DPE-GOVERNED-RUN-A-20260914-001` | $100,000.00 per symbol | 1.0 | 0.0 | PASS; 3,120 bars completed |
| B | `RUN_B` | `DPE-GOVERNED-RUN-B-20260914-001` | $100,000.00 per symbol | 1.0 | 1.0 | PASS; 3,120 bars completed |

`run_type` is run metadata and a persistence classification. It is not an architectural state or stage, and it does not replace `pipeline_run_id`, `collection_run_id`, or the independently persisted `capital_allocation.risk_r` value.

## Persisted Evidence

| Check | Run A | Run B |
| --- | ---: | ---: |
| Trace documents | 6,367 | 6,584 |
| Distinct post-classification symbols | 28 | 28 |
| Trace sequence range | 65-120 | 65-120 |
| Sequence 64 trace documents | 0 | 0 |
| Missing required causal-envelope fields | 0 | 0 |
| Distinct persisted `run_type` values | only `RUN_A` | only `RUN_B` |
| Mismatched `run_type` records | 0 | 0 |
| Distinct persisted `capital_allocation.risk_r` values | only `0` | only `1` |
| Mismatched persisted Risk R records | 0 | 0 |
| Distinct persisted `collection_run_id` values | only `20260911T161623Z-1` | only `20260911T161623Z-1` |
| `GovernedExecutionInstruction` records | 49 | 40 |
| Authoritative `ExecutionEvent` records | 49 | 40 |

`DIA` and `VXX` had initial classification but no later eligible bar from which to calculate an actual crossing. Their absence from post-classification traces is therefore expected and confirms that initial quadrant membership did not synthesize an action.

## Endpoint Comparison

| Measure | Run A, R=0.0 | Run B, R=1.0 |
| --- | ---: | ---: |
| Total final marked capital across 28 evaluated symbols | $2,800,541.400 | $2,799,821.535 |
| Total cumulative realized P&L | $526.875 | $777.700 |
| Active positions at source completion | 3 | 12 |
| Confirmed HOP_ON allocations | 26 | 26 |
| Confirmed normal HOP_OFF liquidations | 0 | 14 |
| Confirmed safety liquidations | 23 | 0 |
| Non-transactional HOLD records | 0 | 244 |

At `R=0.0`, the stop equals the high-water price, so active positions reached the safety condition before ordinary HOP_OFF in this source. At `R=1.0`, the stop is zero for positive prices, so no safety liquidation occurred; positions were held until an actual governed 90-degree HOP_OFF or source completion. These are endpoint observations, not efficacy or optimization claims.

The read-only script `scripts/report_dynamic_pipeline_runs_2026-09-14.js` produces the complete per-symbol comparison, including final marked capital, realized P&L, return percentage, active quantity, cash, action counts, and Run B minus Run A capital difference. It also repeats all envelope and source-identity checks.

## Validation

- `go test ./...`: PASS
- Governed `scripts/Build-DSEJEHTransSat1.ps1`: PASS for Buf lint/generation, `go test ./...`, and built executable creation.
- Both final runs used the governed `scripts/Start-DSEJEHTransSat1.ps1` launcher and `bin/dse-jeh-transsat-1.exe`.
- Initial classification regression across all four quadrants: PASS; state only, zero action counters, zero execution mutation, and zero trace records.
- Repeated identical `MockExecutor.Fill` calls: PASS; protobuf events are equal.
- MongoDB trace verification: PASS; the clean collection contains exactly the two authoritative pipeline IDs and 12,951 records. All records retain the fixed source ID and expected `run_type`, required envelope fields are present, independent Risk R values match their endpoints, and instruction/event counts match.

Runtime Objective 1 is implemented and validated for this bounded fixed-source local-paper path. Runtime Objective 2, Alpaca Paper, remains not started and requires separate authorization.
