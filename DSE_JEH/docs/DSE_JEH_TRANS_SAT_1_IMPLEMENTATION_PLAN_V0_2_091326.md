# DSE_JEH_TransSat_1 Implementation Plan V0.2

## 1. Document Control

| Document control | Value |
| --- | --- |
| Filename | `DSE_JEH_TRANS_SAT_1_IMPLEMENTATION_PLAN_V0_2_091326.md` |
| Date | 2026-09-13 |
| Version | V0.2 |
| Status | APPROVED |
| Architectural authority | `DSE_JEH_TRANS_SAT_1_SYSTEM_DESIGN_V0_1_091226.md` |
| Historical implementation-plan baseline | `DSE_JEH_TRANS_SAT_1_IMPLEMENTATION_PLAN_V0_1_091226.md` |
| Current implementation status | DEP-01 through DEP-04 and Rule #1 analytical-path behavior are implemented and validated; bounded `CalculatePhaseTransition`, deterministic ordered boundary derivation, and `EvaluateDynamicExecution` are complete and focused-tested; the explicit Production Eligibility Controller, state-action algorithms, execution path, and both runtime objectives remain incomplete |
| Scope | Complete implementation planning for `DSE_JEH_TransSat_1` and its two ordered paper-runtime objectives |

This V0.2 document is the current approved implementation-plan authority subordinate to the approved System Design V0.1. It does not overwrite, rename, or delete V0.1, which remains preserved as historical implementation-planning documentation.

V0.2 is required because the implementation target has materially expanded and been clarified since V0.1. The complete target now includes persistent application operation, comprehensive single-proto governance, explicit Production Eligibility event-flow control, an outcome-based `expr` architecture, component-scoped rule contexts, typed outcomes and rule evidence, the authoritative four-state Dynamic Execution Engine, every-bar accountability, a governed execution-adapter boundary, Runtime Objective 1 using stored bars and local paper execution, and Runtime Objective 2 using Alpaca Paper.

Normative terms `MUST`, `MUST NOT`, `SHOULD`, and `MAY` express implementation requirements. A planned declaration or package name is not approved financial semantics merely because it appears in this plan.

---

## 2. Executive Summary

`DSE_JEH_TransSat_1` will be implemented as one independently executable, persistent, startable, stoppable, observable, proto-governed Transformation Satellite. It is not a replay experiment, a collection of proving utilities, or two executor-specific applications. ONLINE and OFFLINE sources converge at the same `BarEvent` boundary and use the same application logic through the governed execution-instruction boundary.

Implementation now defines three ordered phases:

1. **PHASE 1 - BAR INPUT AND JEH ANALYTICAL PATH**: COMPLETE. DEP-01 through DEP-04 are integrated with the authoritative proto, persistent lifecycle, health, activity, and every-bar evidence model without rewriting validated JEH mathematics.
2. **PHASE 2 - FOUR-STATE DYNAMIC EXECUTION ENGINE**: PARTIALLY COMPLETE. The four-state contracts, bounded V1 `DynamicExecutionService.CalculatePhaseTransition`, deterministic ordered boundary facts, and forward governed `DynamicExecutionService.EvaluateDynamicExecution` policy are complete and focused-tested. State-action algorithms, `GovernedExecutionInstruction` generation behavior, execution adapters, `ExecutionEvent`, reconciliation, and complete E2E wiring remain unimplemented or gated.
3. **PHASE 3 - POST-E2E OBSERVABILITY AND DATA**: NOT STARTED. Rich stage-by-stage persistence, causal analytical datasets, and P&L/strategy-performance analysis follow complete governed mock-execution E2E validation.

The completed application has two ordered runtime objectives, not two implementation phases:

- **Runtime Objective 1 - Stored-Bar Local Paper** runs the complete application from the approved OFFLINE stored-bar adapter through `LocalPaperExecutor` / `MockExecutor` and `ExecutionEvent`.
- **Runtime Objective 2 - Alpaca Paper** uses the same complete application and replaces only the executor with `AlpacaPaperExecutor`. It begins only after Objective 1 is implemented and validated.

The principal semantic comparison boundary is the governed execution instruction selected by the four-state engine. Equivalent input, initial state, configuration, rule set, and versions MUST produce equivalent upstream meanings through that boundary. Broker-specific acknowledgement, rejection, fill, partial fill, price, timing, cancellation, latency, and status differences begin after that boundary and belong to `ExecutionEvent` and reconciliation.

No live or funded-capital execution is authorized.

---

## 3. Authority and Relationship to System Design

Authority order for implementation work is:

1. The current approved `DSE_JEH_TRANS_SAT_1_SYSTEM_DESIGN_V0_1_091226.md` is the architectural authority.
2. This V0.2 plan is the current approved implementation-plan authority subordinate to that System Design.
3. Existing code, tests, and validation reports for implementation truth.
4. Supplied Rule Engine Pipeline Patterns material only for the approved Chain of Responsibility, Specification/rule, typed context, compile-once `expr`, local paper, and Alpaca Paper adapter patterns.

The System Design controls financial and operational semantics. Reference examples do not override Rule #1, crossover semantics, strategy regions, decisions, intents, or execution meanings. In particular, region membership alone does not prove `HOP_ON` or `HOP_OFF`, and `SequenceNo >= 64` is not a substitute for per-symbol contiguous causal history.

### 3.1 Bounded Consistency Check

The two required runtime objectives share one four-state engine and one governed execution boundary with substitutable local/mock and future approved Alpaca Paper executors.

The current System Design replaces the former post-phase pipeline decomposition with **PHASE 2 - FOUR-STATE DYNAMIC EXECUTION ENGINE**. Runtime Objectives 1 and 2 remain operating/proving objectives of one application and MUST NOT be renamed as implementation phases.

Unresolved mathematics and financial policy do not prevent this plan. They are retained as `BLOCKED PENDING DESIGN DECISION` at the component that requires them.

---

## 4. Current Implementation Baseline

### 4.1 Implemented and Validated

| Responsibility | Current implementation truth | V0.2 treatment |
| --- | --- | --- |
| DEP-01 Bar Event Admission | IMPLEMENTED AND VALIDATED | Preserve behavior; align contracts and complete every-bar evidence |
| DEP-02 Per-Entity Ordered Analytical State | IMPLEMENTED AND VALIDATED | Preserve entity isolation and causal ownership |
| DEP-03 JEH / Ehlers Phase Update | IMPLEMENTED AND VALIDATED | Do not rewrite validated mathematics without approved cause |
| DEP-04 Circular Phase State | IMPLEMENTED AND VALIDATED | Preserve explicit initialization/value presence and normalized phase |
| Rule #1 behavior | IMPLEMENTED AND VALIDATED in the analytical path | Preserve contiguous bars 1-63 initialization, first possible observable/eligible bar 64, current eligible `Bar[n]` ownership, and retained $\phi[n-1]$ as analytical context |
| Explicit proto-governed Production Eligibility Controller | NOT IMPLEMENTED as the complete target | Implement and wire `ProductionEligibilityService.EvaluateProductionEligibility` without conflating it with the existing Rule #1 analytical behavior |
| Rule Registry | IMPLEMENTED for the active Rule #1 set | Validate and retrieve the active typed rule set; no hot reload is authorized |
| ONLINE and OFFLINE adapters | IMPLEMENTED for the Phase 1 path | Both route through the same host-owned `BarEvent` pipeline; strong ONLINE continuity claims remain blocked |
| Authoritative proto and generated bindings | IMPLEMENTED | One proto defines 10 active services and 13 active RPCs; seven superseded services and one superseded executor RPC remain generated as deprecated compatibility surfaces |
| `DynamicExecutionService.CalculatePhaseTransition` | COMPLETE for the bounded V1 slice | Approved mathematics implemented; authoritative proto updated as required; generated Go regenerated by Buf; focused service tests pass |
| Deterministic ordered boundary facts | COMPLETE for the bounded service slice | Signed shortest directed arc; start excluded; end included; fixed 0/90/180/270 boundaries; direction and encounter order preserved; reverse and later-bar recross facts retained without suppression |
| `DynamicExecutionService.EvaluateDynamicExecution` | COMPLETE for the bounded evaluation slice | Forward governed policy and typed four-state/action outcomes implemented and focused-tested; reverse facts produce no forward firing; no `GovernedExecutionInstruction` is generated |
| Persistent runtime and observation services | IMPLEMENTED AND VALIDATED | Compiled host, lifecycle, shared status, heartbeat, evidence stream, zero-input persistence, and graceful stop |
| Built executable and start/stop scripts | IMPLEMENTED AND VALIDATED | Compiled-executable workflow with PID-scoped governed stop; no `go run` acceptance path |

Existing Phase 1 reports record passing `buf lint`, `buf generate`, `go test ./...`, and `go build` validation. Those reports establish the baseline; they do not prove the four-state Dynamic Execution Engine or either complete runtime objective.

### 4.2 Not Yet Implemented as the Complete Target

- state-associated ALLOCATE ranking, HOLD & TRAIL, and LIQUIDATE/reallocation algorithms;
- `GovernedExecutionInstruction` generation behavior and runtime integration of the bounded Dynamic Execution services;
- complete `LocalPaperExecutor` / `MockExecutor` E2E path;
- `AlpacaPaperExecutor`;
- complete `ExecutionEvent` and reconciliation path; and
- Runtime Objective 1 and Runtime Objective 2 completion evidence.

The runtime-integrated vertical slice ends deliberately after existing Rule #1 eligibility behavior. The bounded, currently unwired `CalculatePhaseTransition` and `EvaluateDynamicExecution` service implementations are complete and focused-tested, including deterministic ordered boundary facts and forward governed four-state outcomes. They do not implement state-action algorithms, instruction generation, executor behavior, or the complete E2E path. The existing phase-motion-unavailable runtime outcome is a deprecated compatibility result from the superseded decomposition.

---

## 5. Complete Application Runtime Objective

The implementation target is the complete common path:

```text
Application Startup / Lifecycle / Health
  -> ONLINE Fin_FeedSat_1 subscriber OR selected OFFLINE collection_run_id
  -> common BarEvent boundary and reception accounting
  -> DEP-01 Bar Event Admission
  -> DEP-02 Per-Entity Ordered Analytical State
  -> DEP-03 JEH / Ehlers Phase Update
  -> DEP-04 Circular Phase State
  -> ProductionEligibilityService.EvaluateProductionEligibility
       -> BLOCKED: typed evidence and terminal bar outcome
       -> ELIGIBLE: continue
    -> DynamicExecutionService.CalculatePhaseTransition
    -> DynamicExecutionService.EvaluateDynamicExecution
    -> typed boundary-policy outcome and four-state result
      -> DISREGARD: no allocation action
      -> ALLOCATE: rank by Phase Velocity
      -> HOLD & TRAIL: trail stops dynamically
      -> LIQUIDATE: reallocate freed capital
    -> GovernedExecutionInstruction when required
  -> ExecutorService.SubmitGovernedExecutionInstruction
  -> ExecutorService.StreamExecutionEvents
  -> ExecutionEvent
  -> reconciliation, evidence, telemetry, runtime state
```

Every received bar MUST have an attributable terminal result. A rejected, initializing, blocked, not-applicable, no-action, or error result is completed processing, not a silent disappearance.

The application remains running until governed stop, cancellation, or fatal application failure under approved lifecycle semantics. No incoming bars is a valid `RUNNING` condition. Liveness, source connectivity, and activity are separate signals; activity counters may remain at zero while the application remains healthy and subscribed.

---

## 6. Implementation Phases and E2E Sequence

### 6.1 PHASE 1 - BAR INPUT AND JEH ANALYTICAL PATH

Phase 1 owns source acquisition, common `BarEvent` mapping, DEP-01 admission, DEP-02 state ownership, DEP-03 JEH mathematics, DEP-04 circular phase state, and `PhaseEvidence`. The existing implementation is retained as the validated baseline.

Remaining Phase 1 integration work is limited to:

- adapting existing generated and Go contracts to the complete reviewed proto;
- preserving deterministic identity and source lineage;
- connecting source state to complete lifecycle/health;
- ensuring reception and rejection are accounted for even when no `PhaseEvidence` is produced;
- making zero-input ONLINE operation persistent;
- defining governed OFFLINE completion versus application stop behavior; and
- preserving existing analytical equivalence and causal-continuity tests.

Status: `COMPLETE` for the approved Phase 1 scope. Strong ONLINE continuity/recovery guarantees remain a separate blocked operational decision and do not alter the validated common-boundary implementation.

### 6.2 PHASE 2 - FOUR-STATE DYNAMIC EXECUTION ENGINE

Phase 2 starts after Production Eligibility and implements the authoritative four-state engine, its supporting transition mathematics, governed boundary policies, state-associated actions, execution boundary, and reconciliation path. The former DEP-05-through-DEP-11 conceptual decomposition is superseded and is not an implementation sequence.

Phase 2 contract review derived the minimum proto/service changes required by the approved replacement design. `DynamicExecutionService` now owns separate `CalculatePhaseTransition` and `EvaluateDynamicExecution` RPCs; `ExecutorService.SubmitGovernedExecutionInstruction` owns the active execution boundary. The seven former post-phase services remain generated only as deprecated compatibility surfaces. No unresolved action algorithm or execution policy was filled in for contract convenience.

Current status: `PARTIALLY COMPLETE`. The four-state proto contracts, bounded V1 `CalculatePhaseTransition`, deterministic ordered boundary derivation, and forward governed `EvaluateDynamicExecution` behavior are implemented and focused-tested. Existing Rule #1 behavior remains implemented in the analytical path, while the complete explicit proto-governed Production Eligibility Controller is not implemented. Runtime wiring, state-action algorithms, instruction generation, and both executor adapters do not exist.

### 6.3 PHASE 3 - POST-E2E OBSERVABILITY AND DATA

The immediate implementation objective is the complete governed path through `GovernedExecutionInstruction`, `MockExecutor`, `ExecutionEvent`, and E2E validation. Rich persistence is not a prerequisite for completing that path; existing runtime evidence and logging may support implementation debugging and validation.

Only after complete E2E validation should Phase 3 design and implement persistent stage-by-stage data for JEH phase, `PhaseTransitionState`, boundary facts, rule outcomes, strategy state, state-associated actions, governed execution instructions, execution events, and subsequent P&L/strategy-performance analysis. That later dataset must support distinguishing implementation correctness from economic strategy quality. This plan does not design or authorize that persistence subsystem now.

### 6.4 Phases Versus Runtime Objectives

The three implementation phases describe construction responsibility. Runtime Objectives 1 and 2 exercise the completed application with different executors. Objective 1 is a mandatory complete-application gate before Objective 2; neither objective is a renamed implementation phase.

---

## 7. Proto-First Contract Strategy

### 7.1 Single Authoritative Proto

The sole DSE_JEH proto remains:

`DSE_JEH/api/proto/dse_jeh/v1/DSE_JEH_TransSat_1.proto`

The previous contract workstream expanded this file in place to 16 services and 18 RPCs under the superseded post-phase design. The bounded four-state contract review replaced that active model additively: the current proto has 10 active services and 13 active RPCs, while seven former services and `ExecutorService.SubmitExecutionIntent` remain deprecated compatibility declarations. Including compatibility surfaces, generated code contains 17 services and 21 RPCs.

The implementation order is:

```text
replacement System Design approved
  -> replacement Implementation Plan approved
  -> minimum four-state proto/service changes reviewed
  -> authoritative proto updated additively
  -> buf lint / compatibility checks passed
  -> generated Go contracts regenerated
  -> Go implementation
  -> tests and build
  -> built executable
  -> governed startup/stop
  -> E2E runtime validation
```

### 7.2 Required Complete Coverage

The expanded proto must describe all governed meanings required for:

- runtime lifecycle, mode, health, degradation, and diagnostics;
- source/subscription state and inbound bar reception;
- bar admission and admission evidence;
- per-entity analytical identity/state and JEH phase evidence;
- Production Eligibility inputs, statuses, typed outcomes, and evidence;
- rule identity/version, rule-set identity, evaluation status, typed outcomes, and `RuleEvaluationEvidence`;
- operational `PhaseTransitionState` and separate audit/trace records;
- the four strategy states and four boundary-policy outcomes;
- ALLOCATE ranking, HOLD & TRAIL, LIQUIDATE reallocation, and DISREGARD no-action results;
- the governed execution instruction required at the external boundary;
- execution outcome, `ExecutionEvent`, and reconciliation;
- runtime activity and every-bar terminal outcome; and
- governed outbound evidence/publication behavior.

The minimum declaration and service decomposition is now reviewed: `PhaseTransitionState`, `BoundaryPolicyOutcome`, `DynamicExecutionState`, `StateActionOutcome`, `DynamicExecutionResult`, and `GovernedExecutionInstruction` are current authoritative types under `DynamicExecutionService` and `ExecutorService`. Financially or operationally meaningful Go enums, states, statuses, events, outcomes, and execution meanings MUST NOT exist as parallel ungoverned vocabulary.

### 7.3 Proto Services Versus Network Services

Proto governance describes the complete machine vocabulary and operational responsibilities. Every major governed process MUST have an authoritative protobuf service with at least one meaningful typed RPC contract. This requirement applies equally to external, approved internal gRPC, and internal-only services.

A protobuf service definition is not a deployment declaration and does not require a separate executable, process, container, port, listener, or remote hop. Network exposure is a separate implementation and deployment decision. An internal-only Production Eligibility Controller or ranking service remains defined and generated from proto but need not be registered on an externally reachable `grpc.Server`. Conversely, declaring a protobuf service does not authorize deploying or exposing it independently.

### 7.4 Major Contract and Implementation Boundaries

| Component | Governed responsibility | Required proto vocabulary/evidence | Proto service contract / network exposure | Go responsibility |
| --- | --- | --- | --- | --- |
| Lifecycle/runtime | start, run, drain, stop, fail, health, zero-input liveness | lifecycle/health/status/config/activity messages and enums | `RuntimeOperationsService`; exposure separately approved | process ownership, signals, cancellation, drain |
| Input adapters / reception | source state and mapping to the common `BarEvent` boundary | source/subscription status, provenance, reception evidence | `BarReceptionService`; internal-only. ONLINE separately consumes upstream Fin gRPC | transport/database access and mapping only |
| DEP-01 | deterministic admission before mutation | `BarEvent`, admission status/findings/evidence | `BarAdmissionService`; internal-only | validation, sequence disposition, routing |
| DEP-02 | ordered analytical state | analytical identity/state evidence | `AnalyticalStateService`; internal-only | serialized entity ownership and recurrence state |
| DEP-03 through DEP-04 | deterministic JEH update and circular phase state | `PhaseEvidence`, phase status, solver identity | `JehPhaseService`; internal-only | approved JEH algorithms and phase normalization |
| Production Eligibility | Rule #1 event-flow gate | scoped input, eligibility status/outcome/evidence, rule evidence | `ProductionEligibilityService`; internal-only | context, evaluation, typed routing |
| Governed Rule Registry | attributable rule configuration and validation | rule identity/version, definitions, typed mappings, evaluation evidence | `RuleRegistryService`; internal-only | registry, compile/cache, validate, retrieve active set |
| Four-state Dynamic Execution Engine | exactly four states, four boundary policies, and state-associated actions | `PhaseTransitionState`, `BoundaryPolicyOutcome`, `DynamicExecutionState`, `StateActionOutcome`, `DynamicExecutionResult` | `DynamicExecutionService.CalculatePhaseTransition`, `EvaluateDynamicExecution`; internal-only | deterministic transition facts plus governed state/action changes |
| Governed execution boundary | convey required external action without claiming execution | `GovernedExecutionInstruction`, `ExecutionEvent` | `ExecutorService.SubmitGovernedExecutionInstruction`, `StreamExecutionEvents`; approved internal gRPC when registered | deterministic instruction identity/idempotency and adapter routing |
| Executors | attempt requested action and report result | execution outcome, `ExecutionEvent`, reconciliation evidence | `ExecutorService`; approved internal gRPC when registered, otherwise internal-only | local simulation or separately approved Alpaca Paper transport only |
| Evidence/telemetry | publish typed lifecycle, activity, processing, rule, and execution evidence | `RuntimeEvidenceEnvelope` and typed evidence | `RuntimeEvidenceService`; external exposure separately approved | publication ordering, delivery, and recovery policy |

---

## 8. Governed Rule Engine / expr Architecture

### 8.1 Normative Evaluation Flow

Every `expr`-governed component uses this flow:

```text
Typed Input
  -> Component Context Builder
  -> Component-Scoped Dynamic Variables
  -> Governed Rule Selection (rule_id + rule_version)
  -> Precompiled expr Program
  -> expr Evaluation
  -> Raw Evaluation Result
  -> Rule-Specific Outcome Mapper
  -> Typed Proto Outcome
  -> Governed Go State Transition / Routing
  -> RuleEvaluationEvidence
  -> Next Governed Component OR Attributable Terminal Outcome
```

`github.com/antonmedv/expr` is the mandatory evaluator for approved dynamic rule conditions. It is not the architecture and has no independent financial authority.

- `expr` = condition evaluator.
- proto = authoritative vocabulary and typed outcomes.
- Go = orchestration, state ownership, and approved deterministic algorithms.
- typed rule outcome = governed routing or state-transition result.

A raw `true` or `false` has no global business meaning. Every result maps through its owning rule to an approved typed proto outcome before routing or mutation.

### 8.2 Rule Registry and Compiled Programs

One common Governed Rule Engine contains:

- Rule Registry;
- `expr` compiler;
- compiled-program cache;
- evaluator;
- rule-specific typed-outcome mapper; and
- `RuleEvaluationEvidence` producer.

Each rule is attributable by rule ID, rule version, owning component/DEP, name, purpose, expression, authorized variables, expected result type, typed success outcome, typed failure/block/no-action outcome, routing behavior, permitted state mutation, evidence requirements, and rule-set/configuration identity.

Expressions compile once at startup or once at a separately approved governed load/reload operation. Programs are cached by rule identity/version or equivalent immutable key. Compilation failure is a visible attributable configuration/runtime failure; invalid rules are never silently skipped. This plan does not authorize hot reload.

Anonymous expression strings scattered through component packages are prohibited.

Implementation status: Rule #1 is compiled once when `ProductionEligibility` is constructed. `RuleRegistryService` validates definitions and returns the active immutable set. A generalized reloadable compiled-program cache is neither needed for the single active rule nor authorized; additional rule programs remain gated by their owning approved semantics.

### 8.3 Component-Scoped Contexts

There is no unrestricted global `PipelineContext`. Conceptual context boundaries are:

| Context | Authorized scope | Explicitly excluded authority |
| --- | --- | --- |
| Production Eligibility | phase status/presence, contiguous valid-bar count, sequence integrity, current validity, prior analytical validity | broker state, fills, cash, quantity |
| Boundary policy | current production eligibility, current strategy state, current `PhaseTransitionState`, deterministic boundary facts | historical transition inference, circular mathematics, broker state, or execution |
| ALLOCATE action | current `ALLOCATE` state, eligible candidates, current Phase Velocity values, approved ranking policy | extra strategy states or direct execution |
| HOLD & TRAIL action | current `HOLD_AND_TRAIL` state and approved trailing-policy inputs | invented trailing mathematics or broker fills |
| LIQUIDATE action | current `LIQUIDATE` state and approved capital/reallocation inputs | invented capital policy or direct broker mutation |
| Execution instruction eligibility | typed current state/action result and pending/reconciliation facts | broker fill claims or adapter-specific responses |

### 8.4 Evidence and Outcomes

`RuleEvaluationEvidence` must identify rule ID/version, owning component, causal input identity, evaluation sequence, evaluation status, typed outcome, reason/diagnostic, and rule-set/configuration identity. Exact enum names are reviewed in proto; statuses conceptually support pass, block, not applicable, no action, and error.

A blocked or no-action result is a valid completed outcome. It remains visible in evidence and bar accountability.

### 8.5 Deterministic Algorithm Boundary

`expr` MUST NOT replace JEH/Ehlers mathematics, Hilbert recurrence, phase normalization, signed circular delta, Phase Velocity, factual boundary detection, approved ranking calculation, quantity sizing, or deterministic identity generation. Those remain deterministic Go/domain algorithms. Rules govern eligibility, the four authoritative boundary policies, action-policy conditions, routing, and typed outcome selection around them.

No expression may calculate a quantity and submit a buy/sell, call Alpaca, mutate authoritative holdings/cash/broker state, bypass the four-state action and governed execution-instruction boundary, or create `ExecutionEvent`.

---

## 9. Production Eligibility Controller

Production Eligibility is an explicit runtime responsibility between DEP-04 and the four-state Dynamic Execution Engine:

```text
DEP-03 JEH Phase Update
  -> DEP-04 Circular Phase State
  -> Production Eligibility Controller
       -> BLOCKED: typed outcome + RuleEvaluationEvidence + bar completion
      -> ELIGIBLE: typed outcome + RuleEvaluationEvidence -> four-state engine
```

### 9.1 Rule #1

For each symbol independently:

- contiguous means the valid symbol-specific causal sequence, not wall-clock minute adjacency;
- `generator_sequence_no` is the primary OFFLINE continuity signal;
- bars 1 through 63 execute JEH mathematics and update analytical state;
- those bars remain `INITIALIZING` for production and cannot route to Dynamic Execution;
- bar 64 is the first possible production-eligible `OBSERVABLE` Phase Angle;
- bar 65 and later remain eligible while causal continuity remains valid;
- duplicate, conflicting, gapped, out-of-order, or invalid candidates do not mutate JEH state; and
- missing bars are never synthesized or interpolated.

A genuine causal continuity break requires production eligibility to be re-established under the approved continuity rule. Mathematical phase calculability is not production runtime eligibility.

### 9.2 Current Eligible Bar Ownership

Production Eligibility applies to the current `Bar[n]`. Once `Bar[n]` is production eligible, that bar owns its complete downstream Dynamic Execution processing. Prior analytical state is input context for `Bar[n]`; the predecessor bar does not need to have entered Dynamic Execution and does not retroactively become production eligible when its retained state is consumed.

Bar 64 is therefore the first possible bar that may enter the four-state engine. Transition mathematics may consume retained predecessor analytical phase $\phi[n-1]$ from bar 63 and current phase $\phi[n]$ from eligible bar 64. Bar 65 and later follow the same current-bar ownership model.

---

## 10. PHASE 1 Existing Implementation and Remaining Integration Work

The implementation SHALL preserve the existing Phase 1 behavior while adapting it to complete contracts.

1. Freeze current solver vectors, initialization behavior, phase normalization, identity, and mode-equivalence evidence as regression baselines.
2. Review complete proto changes against existing generated types and identify adapter changes without changing JEH recurrence.
3. Align DEP-01 reception/admission so every candidate, including rejection, has terminal evidence.
4. Align DEP-02 state ownership with lifecycle, drain, recovery, and eligibility-controller inputs.
5. Align DEP-03/04 output with distinct mathematical status and production eligibility contracts.
6. Preserve the common ONLINE/OFFLINE `BarEvent` boundary.
7. Preserve OFFLINE selection by one explicit `DSE_JEH_OFFLINE_COLLECTION_RUN_ID`; no selector never means all runs.
8. Re-run Phase 1 unit, race, deterministic replay, ONLINE mapping, and phase-equivalence validation after contract integration.

The validated JEH solver is not rewritten merely to adopt new orchestration or evidence contracts.

Items 1 through 7 are complete. Ordinary unit/integration tests, deterministic OFFLINE replay, ONLINE/OFFLINE mapping, phase comparison, vet, build, and proto checks pass. The Go race test remains environment-blocked because CGO/GCC is unavailable on the validation host.

---

## 11. PHASE 2 Four-State Dynamic Execution Implementation

### 11.1 Operational Phase Transition Value

For current production-eligible `Bar[n]`, deterministic Go/domain mathematics consumes retained predecessor analytical phase $\phi[n-1]$, current phase $\phi[n]$, and sequence integrity. The predecessor phase does not need to belong to a production-eligible bar. The operational result is `PhaseTransitionState`, not experimental evidence.

$$
\phi_{normalized} = ((\phi \bmod 360) + 360) \bmod 360
$$

$$
\Delta\phi[n] = \operatorname{circular\_delta}(\phi[n-1], \phi[n]) \in (-180°,180°]
$$

$$
\omega[n] = \Delta\phi[n]
$$

Its unit is degrees per bar with $\Delta Bar=1$. Every exact 180-degree tie is canonically $+180°$; magnitude is $|\Delta\phi|$; direction follows the sign; zero displacement is `STATIONARY`; and non-finite phase input is invalid. Phase Velocity supports the `ALLOCATE` ranking action; it is not a strategy state or separate Dynamic Execution stage.

Status: `COMPLETE` for the bounded V1 `DynamicExecutionService.CalculatePhaseTransition` slice. The authoritative proto was updated as required, generated Go was regenerated by Buf, and focused service tests pass. V1 excludes hysteresis, Schmitt-trigger deadbands, degree buffers, N-bar confirmation, anti-jitter history, smoothing, acceleration, higher-order motion, historical direction lookup, trajectory reconstruction, additional eligibility bars, and a separate Edge Case Handler stage.

### 11.2 History-Free Transition Model

Each state transition is defined only by its input state and resulting output state. Historical transition values may be retained for telemetry, diagnostics, auditability, or analysis, but MUST NOT alter an already-established state. Do not implement historical trajectory inference, multi-transition lookback, smoothing, acceleration as a state determinant, or downstream reconstruction of earlier state.

### 11.3 Governed Boundary Policies

Deterministic Go/domain code computes factual transition and boundary inputs. Component-scoped `expr` rules evaluate these authoritative policies and produce typed outcomes:

- cross 270 degrees -> `HOP_ON` -> `ALLOCATE`;
- cross 0 degrees -> `HOLD_AND_TRAIL`;
- cross 90 degrees -> `HOP_OFF` -> `LIQUIDATE`; and
- cross 180 degrees -> `DISREGARD`.

The deterministic boundary algorithm is resolved and implemented for the bounded service slice. It traverses only the signed shortest directed arc in `PhaseTransitionState`, excludes the start, includes the end, and emits every encountered boundary from 0, 90, 180, and 270 degrees exactly once in strict directed encounter order. Direction is preserved; wrap through 0 degrees is treated identically; `STATIONARY` emits no facts; reverse crossings remain facts but do not fire the forward policy; and a later-bar recross is a new fact. No hysteresis, deadband, degree buffer, smoothing, N-bar confirmation, acceleration, historical direction inference, or cross-bar suppression is authorized. `expr` evaluates governed meaning and MUST NOT implement the circular mathematics.

### 11.4 Four Persistent States and Actions

| Interval | Persistent strategy state | Authoritative action | Remaining implementation detail |
| --- | --- | --- | --- |
| $180 \le \phi < 270$ | `DISREGARD` | No allocation action | None beyond typed no-action behavior |
| $270 \le \phi < 360$ | `ALLOCATE` | Rank by Phase Velocity | Formula, timing, ties, staleness, candidate eligibility |
| $0 \le \phi < 90$ | `HOLD_AND_TRAIL` | Trail stops dynamically | Exact trailing algorithm |
| $90 \le \phi < 180$ | `LIQUIDATE` | Reallocate freed capital | Capital, capacity, sizing, sequencing, reconciliation |

`HOP_ON` and `HOP_OFF` are rule-policy firing events, not states. No additional state or HOP terminology is authorized.

The bounded evaluation implementation produces the resulting four-state value and typed state-action outcome. It marks unresolved ALLOCATE, HOLD & TRAIL, and LIQUIDATE action algorithms as blocked rather than implementing them; `DISREGARD` remains typed no-action behavior. Runtime state persistence and action execution remain future integration work.

### 11.5 Contract and Execution Integration

The bounded contract review replaced the active seven-service post-phase chain with cohesive `DynamicExecutionService.CalculatePhaseTransition` and `EvaluateDynamicExecution` contracts. Universe/candidate inputs and Phase Velocity ranking results are represented inside the engine context/action boundary. The former seven services remain deprecated only for compatibility.

`DynamicExecutionService.EvaluateDynamicExecution` is implemented and focused-tested for the bounded evaluation slice. It consumes current four-state strategy state and `PhaseTransitionState`, derives ordered deterministic boundary facts, evaluates each fact in encounter order through the forward governed policy, and produces typed boundary-policy, resulting-state, and state-action outcomes. It does not generate `GovernedExecutionInstruction`; ranking, trailing, reallocation, instruction, and execution behavior remain gated and unimplemented.

The engine produces `GovernedExecutionInstruction` when external execution is required. `ExecutorService.SubmitGovernedExecutionInstruction` accepts it without a forced `StrategyDecision`/`ExecutionIntent` chain, and `ExecutionEvent` remains separate. Mock/local paper remains the first target; future Alpaca PAPER remains later. No live/funded execution is authorized.

---

## 12. Persistent Runtime / Lifecycle / Telemetry

### 12.1 Lifecycle

The normal workflow is:

```text
buf lint
  -> buf generate
  -> go test ./...
  -> go build
  -> built executable
  -> governed PowerShell start
  -> governed PowerShell stop
```

Runtime and acceptance validation MUST NOT use `go run`. The application starts only after required configuration and rule compilation validate. It remains running until governed stop, cancellation, or fatal application failure. Shutdown stops intake, drains or explicitly cancels accepted work under approved policy, flushes evidence, reconciles pending execution state where possible, and publishes final lifecycle status.

The target includes `Start-DSEJEHTransSat1.ps1` and `Stop-DSEJEHTransSat1.ps1` or an explicitly approved equivalent stop mechanism. This plan does not modify scripts.

Implementation status: the target scripts now build/start the compiled executable and request PID-scoped graceful stop through a stop file. Runtime shutdown transitions through `STOPPING` and `DRAINING`, stops gRPC, flushes phase evidence, and reaches `STOPPED`.

### 12.2 Health and Zero Input

Health distinguishes process liveness, selected mode, source connectivity/subscription, input activity, per-entity readiness, rule-registry readiness, universe readiness, evidence/publication health, executor health, reconciliation, and active versions.

Zero bars is valid while `RUNNING`. It produces a flat activity signal and does not itself mean stopped, failed, completed, or disconnected. OFFLINE end-of-selection semantics must distinguish source completion from application stop.

Validation status: a compiled OFFLINE run with an intentionally absent collection-run ID produced zero bars, reported source `COMPLETED`, remained `RUNNING` and `HEALTHY` for 15 one-second heartbeats, and then stopped through the governed script.

### 12.3 Activity and Every-Bar Accountability

The proto workstream must define exact safe names and semantics corresponding to:

- `bars_received`;
- `bars_admitted`;
- `bars_rejected`;
- `bars_initializing`;
- `bars_phase_eligible`;
- `phase_transition_calculations`;
- `boundary_policy_evaluations`;
- `hop_on_events`;
- `hop_off_events`;
- `state_entries` by four-state value;
- `state_action_outcomes` by action;
- `governed_execution_instructions`; and
- `execution_events`.

Counters are not interchangeable. Evidence joins must reconstruct each bar's path and terminal outcome, including blocked/no-action cases.

Implementation status: one concurrency-safe runtime state owns the currently implemented counters. The terminal heartbeat and `RuntimeOperationsService.GetRuntimeStatus` read snapshots from that same state. The implemented pipeline publishes typed reception, admission, analytical, phase, rule, eligibility, and terminal bar evidence. Four-state and downstream counters do not yet exist and MUST NOT be simulated.

---

## 13. Execution Adapter Architecture

```text
Four-state action outcome
  -> governed execution instruction, when required
  -> Executor interface
       -> LocalPaperExecutor / MockExecutor
       -> AlpacaPaperExecutor
  -> ExecutionEvent
  -> reconciliation and evidence
```

The two objectives are executor selections within one application:

```text
                   DSE_JEH_TransSat_1
                            |
                         BarEvent
                            |
                    DEP-01 ... DEP-04
                            |
                      Production Eligibility
                      |
                  Four-State Dynamic Execution Engine
                      |
                  governed execution instruction
                       /          \
                      /            \
                     v              v
          LocalPaperExecutor   AlpacaPaperExecutor
            Objective 1          Objective 2
                     \            /
                      \          /
                       v        v
                     ExecutionEvent
```

They are not two applications, two rule engines, or two strategy implementations.

Executors do not calculate phase, derive `PhaseTransitionState`, evaluate boundary policies, select the four-state strategy state, rank candidates, trail stops, or decide reallocation. They consume the governed execution instruction selected by the engine, attempt or simulate the requested action, and emit `ExecutionEvent`.

`ExecutionEvent` is semantically separate from phase evidence, `PhaseTransitionState`, boundary-policy outcomes, strategy state, and state-action outcomes. Broker or executor failure never rewrites the meaning of the upstream instruction. External calls occur outside entity and shared-state locks.

The rule engine and analytical engine MUST NOT depend on Alpaca. Live/funded endpoints and orders are outside scope and require a separate design review and explicit human authorization.

---

## 14. Runtime Objective 1 - Stored-Bar Local Paper

### 14.1 Required Path

```text
Approved stored Bar Sequence
  -> OFFLINE input adapter
  -> common BarEvent boundary
  -> complete DEP-01 through DEP-04
  -> Production Eligibility Controller
  -> DynamicExecutionService.CalculatePhaseTransition
  -> DynamicExecutionService.EvaluateDynamicExecution
  -> four-state result
  -> state-associated action
  -> GovernedExecutionInstruction, when required
  -> LocalPaperExecutor / MockExecutor
  -> ExecutionEvent
  -> local reconciliation, runtime evidence, telemetry, and resulting state
```

This is the actual application, not a replay trading application, alternate strategy, alternate rule engine, experiment, or `DSE_JEH_provers` harness. The stored source is only an input adapter. The local executor is only an execution adapter.

### 14.2 Local Executor Responsibility

`LocalPaperExecutor` / `MockExecutor` receives a governed execution instruction, applies approved deterministic simulated acceptance/rejection/fill behavior, and emits an `ExecutionEvent`. It MUST NOT independently decide `HOP_ON`, `HOP_OFF`, `ALLOCATE`, `LIQUIDATE`, `HOLD_AND_TRAIL`, `DISREGARD`, ranking, trailing, or reallocation.

### 14.3 Reproducibility and Acceptance

For identical stored input, initial state, and governed configuration, Objective 1 must reproduce the chain from `BarEvent` through admission, JEH state, `PhaseEvidence`, eligibility, `PhaseTransitionState`, boundary-policy outcomes, four-state membership, state-associated actions, governed execution instructions when required, and simulated `ExecutionEvent`.

Objective 1 acceptance requires deterministic E2E evidence, no silent bar loss, stable causal identities, expected counter relationships, executor reconciliation, clean governed start/stop, and repeated-run comparison. Objective 2 cannot begin until human review accepts Objective 1.

Current status: `PARTIALLY COMPLETE`. The stored-bar runtime path is validated through existing Rule #1 eligibility behavior, including collection run `20260911T161623Z-1` with 3,120 received/admitted bars. The bounded `CalculatePhaseTransition` and `EvaluateDynamicExecution` services are complete and focused-tested but not integrated into this E2E runtime path. State-action algorithms, instruction generation, `LocalPaperExecutor`, `ExecutionEvent`, reconciliation, and full repeated E2E comparison remain absent. Runtime Objective 1 is not accepted.

---

## 15. Runtime Objective 2 - Alpaca Paper

### 15.1 Required Path

```text
Approved BarEvent input
  -> same complete DSE_JEH_TransSat_1
  -> same DEP-01 through DEP-04 and Production Eligibility
  -> same four-state Dynamic Execution Engine
  -> same governed execution-instruction boundary
  -> AlpacaPaperExecutor
  -> Alpaca PAPER API
  -> broker response/state
  -> ExecutionEvent
  -> reconciliation and runtime evidence
```

Alpaca begins at the Executor boundary. There is no Alpaca-specific Dynamic Execution Engine, rule registry, boundary policy, strategy-state logic, ranking logic, or execution-instruction semantics.

### 15.2 Prerequisites

Objective 2 requires:

- accepted Objective 1 evidence;
- frozen validated upstream behavior through the governed execution-instruction boundary;
- separately approved Alpaca Paper adapter contract and configuration policy;
- paper-endpoint enforcement and protection against live endpoints;
- approved order mapping, idempotency, timeout, retry, cancellation, partial-fill, and reconciliation behavior;
- secret handling that never places credentials in proto evidence or logs; and
- explicit human authorization to perform Alpaca Paper validation.

This plan does not configure credentials or call Alpaca.

Current status: `NOT STARTED / BLOCKED UNTIL OBJECTIVE 1 APPROVAL`.

### 15.3 Execution-Instruction Comparison Boundary

For equivalent input, initial state, configuration, rule identities/versions, and deterministic policies, upstream evidence and governed execution-instruction semantics remain common across Objectives 1 and 2. Expected differences after the boundary include broker order identity, acknowledgement, rejection, fill, partial fill, fill price/time, cancellation, latency, and broker-side status. Those differences belong only to `ExecutionEvent` and reconciliation.

---

## 16. E2E Validation and Comparison Strategy

### 16.1 Contract and Component Validation

- proto lint, generation, compatibility, presence, units, identity, lifecycle, and outcome tests;
- Rule Registry schema, authorized-variable, compile failure, cache identity, result-type, outcome-mapping, and evidence tests;
- Rule #1 bars 1 through 63, bar 64, continuity break, and re-establishment tests;
- completed focused `PhaseTransitionState` vectors for normalization, wrap/reverse displacement, zero, exact +180-degree tie handling, magnitude, direction, velocity, causal references, and non-finite input;
- completed focused boundary-policy vectors for no crossing, forward/reverse crossings, forward/reverse wrap, start exclusion, end inclusion, `STATIONARY`, later-bar recross, ordered multiple-boundary traversal, exact +180-degree behavior, and ordered typed policy evaluations;
- completed focused history-free state-transition and boundary-event tests for the bounded service slice;
- ALLOCATE ranking, HOLD & TRAIL, LIQUIDATE reallocation, and execution-instruction idempotency tests after action-policy approval;
- lifecycle, zero-input, stop/drain, failure, and activity-accounting tests; and
- race tests for entity and universe ownership.

Current result: authoritative proto lint/generation/breaking checks and focused service tests pass for the bounded `CalculatePhaseTransition`, deterministic boundary, and `EvaluateDynamicExecution` slices; earlier Go tests, vet, and build pass for their recorded implemented scopes. The race test is environment-blocked by unavailable CGO/GCC. Tests for blocked state-action and execution behavior cannot exist until the controlling semantics are approved.

### 16.2 Runtime Objective 1 Validation

Run the built application against the explicitly selected approved stored collection run and `LocalPaperExecutor`. Compare repeated runs by causal evidence identity and semantic digest. Validate all terminal paths: admission rejection, initialization block, eligibility block, transition unavailable, unchanged state, each boundary-policy event, each state action, no action, governed execution instruction, simulated execution, and reconciliation.

Current result: runtime-slice validation through motion-unavailable is complete; complete Objective 1 validation is `BLOCKED`.

### 16.3 Runtime Objective 2 Validation

After Objective 1 approval, run the same built application with `AlpacaPaperExecutor` against the approved paper endpoint. Validate adapter mapping, paper-only enforcement, broker outcomes, retries/idempotency, partial fills, cancellation, failures, and reconciliation without changing upstream strategy meaning.

Current result: `NOT STARTED`.

### 16.4 Cross-Objective Comparison

Join equivalent runs at the governed execution-instruction boundary using causal input, initial state, configuration, solver/rule/action-policy versions, and instruction identity. Upstream mismatches are application defects or configuration differences. Downstream differences are valid only when attributable to executor environment and represented in `ExecutionEvent`/reconciliation evidence.

Current result: `NOT STARTED`; neither objective has reached the comparison boundary.

---

## 17. Design-to-Implementation Traceability Matrix

The required chain is `System Design -> governed rule/outcome -> proto -> generated Go -> Go component -> test -> evidence -> runtime objective`.

| Plan Section / Sequence Item | Governed Responsibility | Proto Service / RPC | Go Implementation | Runtime Wiring | Test / Validation Evidence | Current Status | Blocker / Remaining Work |
| --- | --- | --- | --- | --- | --- | --- | --- |
| §12 / 8 | Lifecycle, health, activity | `RuntimeOperationsService.GetRuntimeStatus` | `internal/runtime/state.go`, `internal/runtime/terminal.go`, `internal/app/application.go` | External observation RPC; shared state; governed start/stop | Host integration test; compiled zero-input and 3,120-bar runs | IMPLEMENTED | Recovery/checkpoint policy remains open |
| §12 / 8 | Runtime evidence publication | `RuntimeEvidenceService.SubscribeRuntimeEvidence` | `internal/evidence/bus.go`, `internal/services/services.go` | External server stream; bounded non-blocking subscriber queues | Host integration stream test | IMPLEMENTED | Durable publication/recovery policy is not approved |
| §7, §10 / 4-7 | Common input and reception | `BarReceptionService.ReceiveBar` | `internal/input/online`, `internal/input/offline`, `internal/services/services.go` | Internal direct call from one host pipeline | Mapping, ONLINE contract, OFFLINE selection, 3,120-bar run | IMPLEMENTED | Strong ONLINE continuity claims remain blocked |
| §10 / 7 | DEP-01 admission | `BarAdmissionService.AdmitBar` | `internal/admission`, `internal/services/services.go` | Internal direct call | Admission vectors and runtime counters | IMPLEMENTED | None for current scope |
| §10 / 7 | DEP-02 analytical ownership | `AnalyticalStateService.UpdateAnalyticalState` | `internal/analytical/coordinator.go`, `internal/services/services.go` | Internal direct call; per-entity coordinator | Isolation, causal sequence, sparse-symbol tests | IMPLEMENTED | Persistence/recovery policy remains open |
| §10 / 7 | DEP-03/04 JEH and phase | `JehPhaseService.UpdateJehPhase` | `internal/jeh`, `internal/analytical/coordinator.go`, `internal/services/services.go` | Internal direct call; durable phase JSONL | Solver vectors, phase comparison, 3,120-record digest | IMPLEMENTED | None for current Phase 1 semantics |
| §8 / 9 | Governed Rule Registry | `RuleRegistryService.ValidateRuleSet`, `GetActiveRuleSet` | `internal/services/services.go` | Internal-only concrete service; Rule #1 supplied at startup | Rule compilation exercised by service construction/tests | IMPLEMENTED FOR RULE #1 | Additional rules await owning semantics |
| §9 / 10 | Rule #1 analytical-path behavior | `ProductionEligibilityService.EvaluateProductionEligibility` is the target explicit contract | Existing analytical status/state handling | Existing runtime path after phase | Bar 63/64 test; 1,848 initializing and 1,272 eligible replay outcomes | RULE #1 BEHAVIOR IMPLEMENTED | Complete explicit proto-governed controller remains not implemented |
| §11.1-11.2 / 6 | `PhaseTransitionState` and history-free transition ownership | `DynamicExecutionService.CalculatePhaseTransition` | `internal/services/phase_transition.go` | Not yet wired into the runtime path | Focused service vectors pass; proto lint/breaking/generation previously passed | COMPLETE FOR BOUNDED V1 SLICE | Runtime integration follows the engine path; V1 transition mathematics is closed |
| §11.3-11.4 / 4-7 | Ordered boundary facts, forward boundary policy, and bounded four-state evaluation | `DynamicExecutionService.EvaluateDynamicExecution` | `internal/services/dynamic_execution.go` | Not yet wired into the runtime path | Focused boundary and ordered typed-policy tests pass | COMPLETE FOR BOUNDED EVALUATION SLICE | State-action algorithms and runtime integration remain |
| §11.5, §13-14 / 10-11 | Governed instruction, local execution, reconciliation, and Objective 1 | `ExecutorService.SubmitGovernedExecutionInstruction`, `StreamExecutionEvents` | No replacement implementation | Not wired | Generated instruction/event separation verified | CONTRACT COMPLETE / IMPLEMENTATION NOT STARTED | Executor/reconciliation semantics |
| §15 / 12 | Alpaca Paper and Objective 2 | Same governed execution boundary | No adapter | Not wired | No paper integration test | NOT STARTED | Objective 1 approval and paper-only adapter policy |

Generated stubs prove contract availability, not implemented business behavior. Service definition remains distinct from network exposure.

---

## 18. Dependency-Ordered Implementation Sequence

1. Preserve the validated DEP-01-through-DEP-04 implementation and Rule #1 analytical-path behavior, including current eligible `Bar[n]` ownership and retained $\phi[n-1]$ context; keep the complete explicit Production Eligibility Controller status distinct.
2. Retain the completed four-state proto contract review: 10 active services / 13 active RPCs and 17 services / 21 RPCs including deprecated compatibility surfaces. Make future proto changes only for a concrete implementation blocker or approved new contract requirement.
3. Retain the completed bounded V1 `DynamicExecutionService.CalculatePhaseTransition` implementation and focused validation without reopening its mathematics.
4. Retain the completed bounded `DynamicExecutionService.EvaluateDynamicExecution` implementation and focused validation.
5. Retain the completed deterministic start-excluded/end-included boundary-fact handling, including ordered multiple boundaries, reverse facts without forward firing, and later-bar recrosses without suppression.
6. Retain the completed forward governed boundary-policy evaluation and typed outcomes for the fixed 270, 0, 90, and 180-degree meanings.
7. Retain the completed bounded four-state result evaluation; integrate runtime state persistence only with the later approved engine wiring.
8. Integrate state-associated actions only after their gates: DISREGARD no action; ALLOCATE after ranking formula/timing/tie/candidate/freshness approval; HOLD & TRAIL after trailing-algorithm approval; LIQUIDATE after capital/sizing/sequencing/reallocation approval.
9. Generate `GovernedExecutionInstruction` where required after instruction behavior is approved.
10. Implement the approved `MockExecutor` behavior and emit `ExecutionEvent`, preserving the executor as an external action boundary.
11. Execute and accept the complete governed E2E path through mock execution and `ExecutionEvent` before beginning rich persistence work.
12. After E2E acceptance, design and implement stage-by-stage persistence, causal execution lineage, and P&L/strategy-performance datasets for implementation-correctness and strategy-efficacy analysis.
13. Only after explicit approval, implement and validate the paper-only Alpaca adapter and Runtime Objective 2.

The sequence MUST NOT jump from isolated DEP tests directly to Alpaca Paper. Objective 1 is the required complete-application proving gate.

### 18.1 Current Sequence Status

| Item | Current status | Evidence or blocker |
| --- | --- | --- |
| 1 | COMPLETE | DEP-01 through DEP-04 and Rule #1 analytical behavior are implemented and validated; the complete explicit controller remains separate and unimplemented |
| 2 | COMPLETE | Active post-phase model is 10 services / 13 RPCs; generated compatibility surface is 17 services / 21 RPCs; old post-phase services are deprecated only |
| 3 | COMPLETE | Approved V1 transition mathematics is implemented and focused-tested; Buf validation and generation previously passed |
| 4 | COMPLETE | `EvaluateDynamicExecution` is implemented and focused-tested for the bounded evaluation slice |
| 5 | COMPLETE | Directed start-excluded/end-included facts, exact landings/departures, ordered multiple boundaries, reverse facts, wrap, `STATIONARY`, and later-bar recross behavior are implemented |
| 6 | COMPLETE | Fixed forward policy meanings are evaluated through compile-once `expr` and mapped to typed outcomes in encounter order |
| 7 | COMPLETE FOR BOUNDED EVALUATION | Exactly four resulting states and history-free persistence outcomes are produced; runtime state-store wiring remains future work |
| 8 | GATED BY ACTION DESIGNS | Ranking, trailing, capital, sizing, sequencing, candidate eligibility, freshness, and reallocation require approval |
| 9 | CONTRACT COMPLETE / BEHAVIOR GATED | `GovernedExecutionInstruction` exists; generation behavior awaits approved state actions and execution policy |
| 10 | NOT STARTED / GATED | Mock executor and event behavior require approved execution semantics |
| 11 | PARTIAL | Existing runtime validates through Rule #1 eligibility; both bounded Dynamic Execution services are focused-tested but not E2E-wired; complete Objective 1 is not accepted |
| 12 | POST-E2E | Rich persistence and P&L datasets deliberately follow complete mock-execution E2E validation |
| 13 | NOT STARTED | Requires accepted Objective 1 and explicit Alpaca Paper authorization |

---

## 19. Blocking Decisions and Open Questions

The following remain open and do not prevent approval of this plan. Work stops only at the implementation boundary that needs the decision.

| Boundary | Status | Decision or remaining requirement |
| --- | --- | --- |
| Four-state entry ownership | RESOLVED | Current eligible `Bar[n]` owns processing; retained $\phi[n-1]$ is predecessor context; bar 64 is the first possible engine entry |
| `PhaseTransitionState` | RESOLVED AND IMPLEMENTED FOR V1 | Normalize to $[0°,360°)$; shortest signed displacement in $(-180°,180°]$; exact tie $+180°$; magnitude $|\Delta\phi|$; sign-derived direction; zero `STATIONARY`; $\omega=\Delta\phi$ degrees/bar; non-finite input invalid; history-free $\phi[n-1]\rightarrow\phi[n]$ |
| Boundary facts and policies | RESOLVED AND IMPLEMENTED FOR BOUNDED SLICE | Directed shortest arc; start excluded/end included; ordered 0/90/180/270 facts; wrap and direction retained; reverse facts do not fire forward policy; later-bar recrosses are not suppressed; forward meanings produce typed outcomes |
| Four-state evaluation and persistence | BOUNDED EVALUATION IMPLEMENTED / RUNTIME WIRING OPEN | Exactly four resulting states and history-free persistence outcomes are focused-tested; integration with runtime state ownership remains incomplete |
| ALLOCATE action | BLOCKED PENDING DETAIL | Candidate snapshot, freshness, Phase Velocity ranking formula/timing/ties, capacity, and sizing |
| HOLD & TRAIL action | BLOCKED PENDING DETAIL | Exact dynamic trailing-stop algorithm and execution-update semantics |
| LIQUIDATE action | BLOCKED PENDING DETAIL | Capital representation, reallocation, capacity, sizing, sequencing, and reconciliation |
| Governed execution instruction | MINIMUM CONTRACT RESOLVED; IMPLEMENTATION POLICY BLOCKED | `GovernedExecutionInstruction` defines identity, causality, action, idempotency, correlation, optional validity/quantity, and status; sizing, expiry/cancellation behavior, and adapter policy remain unresolved |
| Executors/reconciliation | BLOCKED PENDING DESIGN DECISION | Execution outcome state machine, retries, duplicate events, recovery and reconciliation |
| Lifecycle/recovery | PARTIALLY RESOLVED | OFFLINE completion is distinct from application stop and accepted work drains before stop; checkpoint, restart, pending-work recovery, and future execution reconciliation remain open |
| ONLINE continuity | BLOCKED FOR STRONG CONTINUITY CLAIMS | Upstream loss detection, finalized-bar guarantee, resume cursor, backpressure |
| Alpaca Paper | BLOCKED UNTIL OBJECTIVE 1 APPROVAL | Paper-only endpoint/configuration, order mapping, idempotency, fill/cancel/reconciliation policy |
| Future adaptive policy geometry | FUTURE / NOT IMPLEMENTED | Any controlled, versioned response coefficients must preserve deterministic phase-transition and boundary facts; no coefficient names, formulas, learning logic, or persistence are approved |

No unresolved item may be silently converted into a default implementation assumption.

---

## 20. Continuing Non-Goals

Neither this plan nor the completed runtime slice authorizes:

- creation of another DSE_JEH proto;
- implementation of state-action, instruction-generation, or remaining runtime behavior before their required policy details are approved;
- restoration of DEP-05 through DEP-11 as the active architectural decomposition;
- MongoDB mutation;
- Alpaca credential configuration or API calls;
- paper or live order submission;
- live/funded-capital design or execution;
- experimental `go run` programs or trading experiments;
- viewer changes;
- staging, commit, push, or deployment; or
- invention of unresolved mathematics or financial policy.

---

## 21. Change Log

| Version | Date | Change |
| --- | --- | --- |
| V0.2 | 2026-09-13 | Created as a new proposed implementation-plan authority because the complete target was materially expanded and clarified beyond V0.1: persistent application operation; comprehensive single-proto governance; explicit Production Eligibility control; outcome-based compile-once `expr` rules with scoped contexts, typed outcomes, and rule evidence; complete DEP-01 through DEP-11 integration; every-bar accountability; governed execution adapters; Runtime Objective 1 for Stored-Bar Local Paper; and Runtime Objective 2 for Alpaca Paper. Preserves validated DEP-01 through DEP-04 and unresolved mathematical/policy blockers. |
| V0.2 approval | 2026-09-13 | Human reviewer approved V0.2 as the current implementation-plan authority subordinate to the approved DSE_JEH_TransSat_1 System Design V0.1. V0.1 remains preserved as historical documentation. |
| V0.2 service-contract clarification | 2026-09-13 | Clarified that every major governed process requires a generated protobuf service/RPC contract, including internal-only processes, while service definition remains independent from registration, listener, process, port, and network exposure decisions. Updated the component responsibility matrix without resolving existing mathematical, policy, or execution blockers. |
| V0.2 implementation reconciliation | 2026-09-13 | Recorded the implemented persistent Go host, two externally exposed observation services, direct internal service composition through Production Eligibility, Rule #1 and Rule Registry support, governed scripts, zero-input persistence, graceful shutdown, and verified 3,120-bar OFFLINE run. Replaced proposed traceability with actual paths and classified sequence items 1-27. DEP-05 through DEP-11, Runtime Objective 1 completion, and Runtime Objective 2 remain blocked or not started. |
| V0.2 DEP-05/06 entry clarification | 2026-09-13 | Established current eligible `Bar[n]` ownership, removed the bar-64/bar-65 entry blocker, and fixed the DEP-05 relation as single-bar signed circular delta in degrees per bar while retaining only the exact 180-degree signed-direction tie policy as pending. |
| V0.2 four-state design replacement | 2026-09-13 | Superseded the active DEP-05-through-DEP-11 decomposition with one history-free four-state Dynamic Execution Engine: `DISREGARD`, `ALLOCATE`, `HOLD_AND_TRAIL`, and `LIQUIDATE`, with fixed governed boundary policies, state-associated actions, and one governed external execution boundary. Proto changes are deferred to a later design-derived review. |
| V0.2 four-state contract reconciliation | 2026-09-13 | Recorded the completed bounded proto review: 10 active services / 13 active RPCs; cohesive `DynamicExecutionService`; operational `PhaseTransitionState`; typed four-state, boundary-policy, and state-action contracts; direct `GovernedExecutionInstruction` executor boundary; seven former services and one former executor RPC retained deprecated for compatibility; Buf lint, breaking, and generation passed. |
| V0.2 four-state implementation alignment | 2026-09-13 | Aligned the Implementation Plan with the approved replacement Dynamic Execution architecture and current implementation state; marked `CalculatePhaseTransition` complete; established `EvaluateDynamicExecution` as the next bounded implementation target; preserved unresolved state-action algorithms as explicit design gates; sequenced governed mock execution and complete E2E validation before stage-by-stage persistent analytical data work. No code, proto, generated artifacts, tests, scripts, or runtime behavior changed. |
| V0.2 editorial consistency correction | 2026-09-13 | Corrected the Executive Summary to identify three implementation phases, corrected duplicate §6.3 numbering by renumbering "Phases Versus Runtime Objectives" to §6.4, and removed stale pre-approval wording from the authority hierarchy now that V0.2 is approved. No architecture, implementation status, mathematics, algorithms, contracts, code, tests, runtime behavior, or implementation sequence changed. |
| V0.2 editorial phase-count correction | 2026-09-13 | Corrected the remaining stale reference in §6.4 from "two implementation phases" to "three implementation phases" so the section is consistent with the approved three-phase implementation structure. No architecture, implementation status, mathematics, algorithms, contracts, code, tests, runtime behavior, authorization, or implementation sequence changed. |
| V0.2 Dynamic Execution implementation status alignment | 2026-09-13 | Updated the Implementation Plan to reflect the completed bounded `DynamicExecutionService.EvaluateDynamicExecution` implementation, deterministic ordered boundary-fact derivation, forward governed boundary-policy evaluation, reverse-crossing no-fire behavior, multiple-boundary directed encounter ordering, later-bar recross handling, and focused validation. Preserved unresolved state-action, execution, executor, E2E, persistence, P&L, and Alpaca gates. No code, proto, generated artifacts, tests, scripts, configuration, or runtime behavior changed in this documentation-only alignment. |

---

## 22. Authorization Statement

This Implementation Plan V0.2 is **APPROVED** as the current implementation-plan authority subordinate to the approved `DSE_JEH_TransSat_1` System Design V0.1.

V0.1 remains preserved as historical documentation. The approved System Design now governs the four-state replacement. This plan does not resolve any mathematical, action-policy, contract, lifecycle, or execution detail explicitly retained as open.

Approval of this document authorizes implementation planning to proceed through the dependency gates stated here; it does not by itself approve unresolved mathematics, live trading, production deployment, Alpaca Paper execution, or any broker credentials. The four-state contract gate is complete; each remaining algorithmic/implementation boundary and Runtime Objective 2 requires its stated review and authorization.

Implementation progress recorded here does not expand authorization. Current eligible `Bar[n]` ownership, bar 64 as the first possible four-state-engine entry, the V1 `PhaseTransitionState` mathematics, deterministic ordered boundary facts, bounded `EvaluateDynamicExecution`, exactly four states, history-free transitions, the four fixed forward boundary-policy meanings, and the minimum service model are resolved and implemented for their bounded service slices. State-action algorithms, `GovernedExecutionInstruction` generation, runtime integration, executor/reconciliation behavior, persistence, P&L analysis, adaptive coefficients, and broker interaction remain subject to their stated approvals. No paper/live order submission, staging, commit, push, or deployment is approved by this replacement.

Overall implementation verdict: **V0.2 PLAN REMAINS IN PROGRESS**.
