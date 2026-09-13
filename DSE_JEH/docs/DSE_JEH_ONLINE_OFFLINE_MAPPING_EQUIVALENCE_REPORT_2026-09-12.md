# DSE_JEH Online Offline Mapping Equivalence Report

| Item | Value |
| --- | --- |
| Date | 2026-09-12 |
| Purpose | Prove source adapters converge on one BarEvent semantic surface |
| Result | PASS |

## Scope

The comparison covers fields shared by equivalent ONLINE and OFFLINE source observations:

- canonical entity and symbol;
- interval and interval start/end in Unix milliseconds;
- open, high, low, close;
- optional volume and event count.

Mode-specific source observation IDs, collection-run identity, partition identity, upstream snapshot identity, receipt time, and replay policy are intentionally allowed to differ as provenance.

## Validation

`internal/input/mapping_test.go` constructs equivalent Fin and Mongo source observations, maps each through its real adapter mapper, and invokes `input.EquivalentBarValues`. The test passed under:

```powershell
go test ./internal/input ./internal/input/online
```

It also passed in the complete suite:

```powershell
go test ./...
```

The ONLINE adapter imports Fin generated types only in `internal/input/online`. The OFFLINE adapter imports Mongo document types only in `internal/input/offline`. Admission, the entity coordinator, JEH solver, circular phase representation, and PhaseEvidence generation receive only `dsejeh.v1.BarEvent` and contain no source-mode branch.

## Deviations And Limitations

ONLINE entity sequence is adapter-local because the upstream stream has no accepted sequence. OFFLINE sequence is authoritative within collection run and symbol. Mapping equivalence therefore asserts common analytical bar semantics, not identical transport identity or equivalent delivery guarantees.

## Next Action

Keep this test as a contract regression test. Add a captured live Fin observation only after market-open validation; it must exercise the same mapper and must not introduce Fin types into the analytical core.
