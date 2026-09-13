# DSE_JEH Phase Equivalence Report

| Item | Value |
| --- | --- |
| Date | 2026-09-12 |
| Purpose | Compare bar-by-bar runtime phase with the independent phase-angle series reference |
| Tolerance | Absolute `1e-9` |

## Inputs

- Reference: `Bar_Sequence_Lab/phase_angle_series_generator/exports/bar_sequence_phase_angle_series_2026-09-11.csv`.
- Runtime: `DSE_JEH/exports/phase_evidence_run1.jsonl` and `phase_evidence_run2.jsonl`.
- Join identity: collection run, normalized entity/symbol, and entity-scoped generator sequence.

## Method

The runtime processed source bars through admission and the incremental JEH solver. The comparator matched initializing and observable evidence against the independent batch reference. Observable values were compared using absolute difference.

## Commands

```powershell
$env:DSE_JEH_MODE='OFFLINE'
$env:DSE_JEH_REFERENCE_CSV='..\Bar_Sequence_Lab\phase_angle_series_generator\exports\bar_sequence_phase_angle_series_2026-09-11.csv'
$env:DSE_JEH_COMPARISON_REPORT='exports\phase_equivalence_run1.json'
.\bin\dse-jeh-transsat-1.exe
```

The command was repeated for run 2 with separate output paths.

## Results

| Metric | Run 1 | Run 2 |
| --- | ---: | ---: |
| Reference rows | 3,480 | 3,480 |
| Matched rows | 3,480 | 3,480 |
| Missing runtime rows | 0 | 0 |
| Status mismatches | 0 | 0 |
| Observable values compared | 1,272 | 1,272 |
| Numeric mismatches | 0 | 0 |
| Maximum absolute difference | 0 | 0 |

The reference contains 3,480 rows from the applicable collection runs. The runtime processed all 4,519 Mongo observations; the additional runtime rows are outside the reference export and are not treated as mismatches.

## Conclusion

The runtime reproduces the reference phase exactly for every comparable row. Initialization status also matches. The observed maximum numeric difference is `0`, which passes the engineering tolerance of `1e-9`.

## Limitations And Next Actions

The `1e-9` value is the Phase 1 engineering validation tolerance, not a claim of universal scientific tolerance. Retain the raw reports and rerun comparison after any solver, compiler, dependency, or reference-data change.
