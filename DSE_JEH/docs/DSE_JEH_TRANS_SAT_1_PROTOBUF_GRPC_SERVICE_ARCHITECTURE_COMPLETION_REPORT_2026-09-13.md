# DSE_JEH_TransSat_1 Protobuf / gRPC Service Architecture Completion Report

| Field | Value |
| --- | --- |
| Date | 2026-09-13 |
| Scope | Bounded protobuf service-model review and generated Go contracts only |
| Authoritative design | `DSE_JEH_TRANS_SAT_1_SYSTEM_DESIGN_V0_1_091226.md`, approved four-state replacement |
| Authoritative proto | `api/proto/dse_jeh/v1/DSE_JEH_TransSat_1.proto` |
| Result | COMPLETE |

## 1. Executive Summary

The authoritative proto required changes. The earlier 16-service / 18-RPC model was completed before the approved four-state Dynamic Execution replacement and encoded the former DEP-05-through-DEP-11 chain as seven active services. That service chain is now historical rather than current architecture.

The current active contract has **10 services and 13 RPCs**. Post-eligibility behavior is governed cohesively by `DynamicExecutionService`, with separate RPCs for deterministic `PhaseTransitionState` calculation and governed four-state evaluation. `ExecutorService` now accepts `GovernedExecutionInstruction` directly and continues to publish separate `ExecutionEvent` outcomes.

For wire and generated-API compatibility, the seven former post-phase services and their existing messages remain present but are deprecated. `ExecutorService.SubmitExecutionIntent` is also retained and deprecated. Including these compatibility surfaces, the physical proto generates **17 services and 21 RPCs**.

No hand-written Go business logic, executor behavior, broker integration, deployment topology, listener, port, or runtime wiring changed.

## 2. Authority and Review Method

The review used this order of authority:

1. Current approved System Design, including the 2026-09-13 four-state replacement.
2. Current approved Implementation Plan V0.2.
3. The earlier protobuf/gRPC completion report as historical evidence only.
4. The authoritative proto and generated Go files as implementation facts.

The review classified every former post-phase service, derived the minimum current service model from governed responsibilities, checked compatibility before generation, regenerated only through Buf, and mechanically compared proto service/RPC declarations with generated gRPC surfaces.

## 3. Historical Model

The earlier report truthfully recorded its then-current result: 16 services and 18 RPCs, including this active chain:

```text
PhaseMotionService
  -> BoundaryCrossoverService
  -> StrategyRegionService
  -> UniverseStateService
  -> CandidateRankingService
  -> StrategyDecisionService
  -> ExecutionIntentService
  -> ExecutorService.SubmitExecutionIntent
```

That chain implemented the former DEP-05-through-DEP-11 decomposition. The approved four-state System Design later superseded it. It is not the current E2E service model, expr-bearing service matrix, or implementation sequence.

The even earlier pre-service baseline recorded by the original report was 3 services and 4 RPCs. That fact remains historical and is not the comparison baseline for this replacement review.

## 4. Review Matrix and Disposition

| Former service | Former responsibility | Current disposition | Compatibility consequence | Reason |
| --- | --- | --- | --- | --- |
| `PhaseMotionService` | DEP-05 motion stage | **RENAMED / RESCOPED and REPLACED** by `DynamicExecutionService.CalculatePhaseTransition` | Service and old motion messages retained deprecated | The operational value is `PhaseTransitionState`, not evidence or a strategy stage; deterministic current-transition mathematics has no expr calculation and no predecessor-eligibility requirement. |
| `BoundaryCrossoverService` | DEP-06 crossover stage | **MERGED / REPLACED** by `DynamicExecutionService.EvaluateDynamicExecution` | Service and old crossover messages retained deprecated | Deterministic `BoundaryFact` inputs feed component-scoped policy; fixed typed outcomes drive the four-state result without a separate active stage. |
| `StrategyRegionService` | DEP-07 region stage | **REPLACED** by four-state membership in `DynamicExecutionResult` | Service and old region messages retained deprecated | `DynamicExecutionState` contains exactly four persistent states; HOP_ON and HOP_OFF are policy outcomes, not states. |
| `UniverseStateService` | DEP-08 universe stage | **MERGED** into `DynamicExecutionContext` | Service and old universe evidence retained deprecated | Asynchronous candidate, holding, freshness, and capacity inputs remain representable without forcing an independent active RPC stage. |
| `CandidateRankingService` | DEP-09 ranking stage | **MERGED** into the ALLOCATE `StateActionOutcome` | Service and old ranking evidence retained deprecated | Phase Velocity ranking is an action associated with `ALLOCATE`; the unresolved algorithm is not invented. |
| `StrategyDecisionService` | DEP-10 decision stage | **REMOVED FROM ACTIVE CONTRACT / REPLACED** by `DynamicExecutionResult` | Service and `StrategyDecision` retained deprecated | The four-state result already owns governed state and action meaning; an extra decision stage would duplicate that authority. |
| `ExecutionIntentService` | DEP-11 intent-generation stage | **REMOVED FROM ACTIVE CONTRACT / REPLACED** by `GovernedExecutionInstruction` | Service, `ExecutionIntent`, and `SubmitExecutionIntent` retained deprecated | The approved model needs one governed instruction boundary, not a forced decision-to-intent chain. |

No former service was retained as active merely because generated Go code existed.

## 5. Current Active Service Inventory

| Active service | RPCs | Governed responsibility |
| --- | --- | --- |
| `RuntimeOperationsService` | `GetRuntimeStatus` | Lifecycle, health, source, version, and activity observation |
| `RuntimeEvidenceService` | `SubscribeRuntimeEvidence` | Typed evidence publication |
| `BarReceptionService` | `ReceiveBar` | Common ONLINE/OFFLINE `BarEvent` reception |
| `BarAdmissionService` | `AdmitBar` | DEP-01 deterministic admission |
| `AnalyticalStateService` | `UpdateAnalyticalState` | DEP-02 ordered per-entity analytical state |
| `JehPhaseService` | `UpdateJehPhase` | DEP-03 JEH update and DEP-04 circular phase state |
| `ProductionEligibilityService` | `EvaluateProductionEligibility` | Rule #1 gate; current eligible Bar[n] owns downstream processing |
| `DynamicExecutionService` | `CalculatePhaseTransition`, `EvaluateDynamicExecution` | Deterministic current-transition facts, boundary policy, exactly four states, and associated action/result boundary |
| `RuleRegistryService` | `ValidateRuleSet`, `GetActiveRuleSet` | Governed rule validation and active-set observation |
| `ExecutorService` | `SubmitGovernedExecutionInstruction`, `StreamExecutionEvents` | Governed instruction receipt and separate executor outcomes |

**Current active total: 10 services / 13 RPCs.**

The service declarations govern vocabulary and responsibility. They do not require separate processes, network hops, listeners, or externally exposed endpoints.

## 6. Current Dynamic Execution Model

```text
ProductionEligibilityService.EvaluateProductionEligibility
  -> DynamicExecutionService.CalculatePhaseTransition
       -> PhaseTransitionState (deterministic mathematics)
  -> DynamicExecutionService.EvaluateDynamicExecution
       -> BoundaryPolicyRuleOutcome (governed expr policy)
       -> DynamicExecutionResult (four-state membership)
       -> StateActionOutcome
       -> GovernedExecutionInstruction when required
  -> ExecutorService.SubmitGovernedExecutionInstruction
  -> ExecutorService.StreamExecutionEvents
       -> ExecutionEvent
```

### 6.1 PhaseTransitionState

`PhaseTransitionState` is an operational current-transition value for eligible Bar[n]. It identifies retained prior and current `PhaseEvidence`, current production-eligibility evidence, signed circular displacement, magnitude, direction, Phase Velocity in degrees per bar, validity, deterministic algorithm identity, and causal entity identity.

The request deliberately carries no prior production-eligibility evidence. Bar 64 may be the first engine entry using retained phi[63] as analytical context. No multi-transition history, smoothing, acceleration, or downstream reconstruction field was added.

### 6.2 Boundary Policies

`BoundaryFact` represents deterministic crossing input without specifying the unresolved crossing algorithm. `BoundaryPolicyRuleOutcome` and `BoundaryPolicyOutcome` represent the fixed governed meanings:

| Boundary | Typed outcome | Resulting persistent state |
| --- | --- | --- |
| 270 degrees | `HOP_ON_ALLOCATE` | `ALLOCATE` |
| 0 degrees | `ENTER_HOLD_AND_TRAIL` | `HOLD_AND_TRAIL` |
| 90 degrees | `HOP_OFF_LIQUIDATE` | `LIQUIDATE` |
| 180 degrees | `ENTER_DISREGARD` | `DISREGARD` |
| No firing | `NO_TRANSITION` | Current state persists |

HOP_ON and HOP_OFF exist only in policy outcomes. They are not persistent states. `RuleOutcomeMapping` and `RuleEvaluationEvidence` now admit the boundary-policy typed outcome; no generic `ExprService` exists.

### 6.3 Four States and Actions

`DynamicExecutionState` has four persistent values: `DISREGARD`, `ALLOCATE`, `HOLD_AND_TRAIL`, and `LIQUIDATE`. Its `UNSPECIFIED` zero value is protobuf absence safety, not a fifth state.

`StateActionOutcome` represents the approved associated action boundary:

| State | `StateActionType` |
| --- | --- |
| `DISREGARD` | `NO_ALLOCATION_ACTION` |
| `ALLOCATE` | `RANK_BY_PHASE_VELOCITY` |
| `HOLD_AND_TRAIL` | `TRAIL_STOPS_DYNAMICALLY` |
| `LIQUIDATE` | `REALLOCATE_FREED_CAPITAL` |

`DynamicExecutionContext`, `AllocationCandidate`, and `PhaseVelocityRankingEntry` preserve required universe and ranking inputs/results inside the cohesive engine boundary. They do not define ranking timing, formula, ties, staleness, trailing, sizing, capital, reallocation, or reconciliation algorithms.

### 6.4 Governed Execution Boundary

`GovernedExecutionInstruction` is the active external request contract. It carries deterministic identity, entity, requested external action, causal four-state result/action, configuration and policy lineage, idempotency/correlation identity, optional validity, and optional quantity when a future approved sizing policy supplies one.

`ExecutorService.SubmitGovernedExecutionInstruction` acknowledges receipt only. `ExecutionEvent` remains the sole execution-outcome contract and now has `governed_execution_instruction_id`; the legacy `execution_intent_id` remains deprecated for compatibility. Strategy state/action, execution instruction, and execution outcome remain distinct.

## 7. Telemetry and Evidence

`RuntimeActivitySnapshot` now adds current counters for phase-transition calculations, boundary-policy evaluations, entries and persistence for each of the four states, state-action outcomes, and governed execution instructions. The old motion, crossover, strategy-decision, and execution-intent counters remain at their original field numbers but are deprecated.

`RuntimeEvidenceEnvelope` and `BarProcessingOutcomeEvidence` add current four-state evidence references without changing existing field numbers. Old post-phase evidence variants remain deprecated compatibility fields.

## 8. Compatibility

The replacement is additive and passed Buf breaking analysis.

- No existing field number was changed or reused.
- No existing enum number was changed or reused.
- No existing message, service, or RPC was removed.
- Seven superseded services remain generated with `deprecated = true`.
- Historical messages, evidence variants, counters, and outcomes remain at their existing numbers and are deprecated where they conflict with current terminology.
- `ExecutorService.SubmitExecutionIntent` remains generated and deprecated while the active service adds `SubmitGovernedExecutionInstruction`.
- New implementations must use the current contracts; compatibility declarations do not preserve the retired architecture as active design.

Physical generated inventory, including compatibility surfaces: **17 services / 21 RPCs**.

## 9. Generated Artifacts

Both generated files changed and were regenerated only by `buf generate`:

| Generated file | Result | SHA-256 after generation | Lines |
| --- | --- | --- | --- |
| `gen/dse_jeh/v1/DSE_JEH_TransSat_1.pb.go` | CHANGED | `04b72697cd3d39a6023e79f4e82c69db6b21967043dfaecde704a9ff8319833f` | 13,624 |
| `gen/dse_jeh/v1/DSE_JEH_TransSat_1_grpc.pb.go` | CHANGED | `4469b6ba822334b5880790158308006d0a8273a32fff0f22d7e10aec1b001aea` | 2,082 |

Pre-review hashes were respectively `c86a74d5442d9f9963de39dba549a93e180ced611460aea43dfe847b3b6314d0` and `7d46225d1f3a9ceaf285a1c625d8d7f6044a92b0747cc3ce0b331c90aa2f623b`.

Mechanical inspection found all **17/17** service descriptors/registration surfaces and all **21/21** RPC method constants/handlers. The intended current `DynamicExecutionService` and executor instruction RPC are present. No generated service is missing.

## 10. Validation

Commands were run from `DSE_JEH` with Buf 1.59.0:

| Command | Result | Diagnostics |
| --- | --- | --- |
| `buf lint` | PASS | None |
| `buf breaking --against '../.git#subdir=DSE_JEH'` | PASS | None; no breaking-check bypass or weakening |
| `buf generate` | PASS | Both Go outputs regenerated |
| Generated contract inventory | PASS | 17/17 services and 21/21 RPCs generated; active inventory is 10/13 |
| VS Code generated-file diagnostics | PASS | No proto, pb.go, or grpc.pb.go errors |

`go test ./...` was not required: generation produced no compilation diagnostics, and the bounded task requires it only for regeneration-caused failures or a normal contract procedure that mandates it.

## 11. Remaining Human Decisions

The contract intentionally does not resolve:

- exact signed direction at 180 degrees;
- wrap, reverse, exact-boundary, multiple-crossing, or jitter handling;
- ALLOCATE ranking timing/formula/ties/staleness;
- HOLD & TRAIL algorithm;
- LIQUIDATE capital, sizing, sequencing, and reallocation mechanics;
- executor behavior, retries, recovery, and reconciliation; or
- broker behavior.

These are implementation/design blockers at their owning boundaries, not reasons to retain the retired seven-stage service architecture. The four boundary meanings and bar-64 ownership are resolved and are not listed as open questions.

## 12. Files Changed and Scope Confirmation

Changed by this review:

- `api/proto/dse_jeh/v1/DSE_JEH_TransSat_1.proto`
- `gen/dse_jeh/v1/DSE_JEH_TransSat_1.pb.go` through Buf only
- `gen/dse_jeh/v1/DSE_JEH_TransSat_1_grpc.pb.go` through Buf only
- this report, updated in place
- the Implementation Plan and narrow stale factual proto references in the System Design

No hand-written Dynamic Execution Go implementation, executor behavior, broker integration, Alpaca call/configuration, deployment change, runtime replay, MongoDB access, staging, commit, or push occurred.

## 13. Change History

| Date | Change |
| --- | --- |
| 2026-09-13, earlier | Completed the historical DEP-05-through-DEP-11 service architecture with 16 services / 18 RPCs. |
| 2026-09-13, four-state review | Re-reviewed the service model against the later approved four-state System Design; added the active 10-service / 13-RPC model, retained seven old services and one old executor RPC as deprecated compatibility surfaces, regenerated both Go files, and passed lint and breaking validation. |

## 14. Final Verdict

**COMPLETE:** The active protobuf/gRPC model follows the approved four-state Dynamic Execution Engine rather than the retired DEP-05-through-DEP-11 chain. Compatibility is preserved additively, generated files match the authoritative proto, and no unresolved algorithm or hand-written runtime behavior was invented.