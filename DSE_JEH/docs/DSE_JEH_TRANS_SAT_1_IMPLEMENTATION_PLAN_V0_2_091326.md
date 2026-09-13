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
| Current implementation status | PHASE 1 and the persistent runtime slice through Production Eligibility are implemented and validated; PHASE 2 and both complete runtime objectives remain incomplete |
| Scope | Complete implementation planning for `DSE_JEH_TransSat_1` and its two ordered paper-runtime objectives |

This V0.2 document is the current approved implementation-plan authority subordinate to the approved System Design V0.1. It does not overwrite, rename, or delete V0.1, which remains preserved as historical implementation-planning documentation.

V0.2 is required because the implementation target has materially expanded and been clarified since V0.1. The complete target now includes persistent application operation, comprehensive single-proto governance, explicit Production Eligibility event-flow control, an outcome-based `expr` architecture, component-scoped rule contexts, typed outcomes and rule evidence, the authoritative four-state Dynamic Execution Engine, every-bar accountability, a governed execution-adapter boundary, Runtime Objective 1 using stored bars and local paper execution, and Runtime Objective 2 using Alpaca Paper.

Normative terms `MUST`, `MUST NOT`, `SHOULD`, and `MAY` express implementation requirements. A planned declaration or package name is not approved financial semantics merely because it appears in this plan.

---

## 2. Executive Summary

`DSE_JEH_TransSat_1` will be implemented as one independently executable, persistent, startable, stoppable, observable, proto-governed Transformation Satellite. It is not a replay experiment, a collection of proving utilities, or two executor-specific applications. ONLINE and OFFLINE sources converge at the same `BarEvent` boundary and use the same application logic through the governed execution-instruction boundary.

Implementation preserves two established phases:

1. **PHASE 1 - BAR INPUT AND JEH ANALYTICAL PATH**: COMPLETE. DEP-01 through DEP-04 are integrated with the authoritative proto, persistent lifecycle, health, activity, and every-bar evidence model without rewriting validated JEH mathematics.
2. **PHASE 2 - FOUR-STATE DYNAMIC EXECUTION ENGINE**: PARTIALLY COMPLETE only at the shared Rule Registry and Production Eligibility infrastructure boundary. The four-state engine, its governed boundary policies and state actions, the design-derived contract update, execution adapters, `ExecutionEvent`, and reconciliation remain not implemented.

The completed application has two ordered runtime objectives, not two implementation phases:

- **Runtime Objective 1 - Stored-Bar Local Paper** runs the complete application from the approved OFFLINE stored-bar adapter through `LocalPaperExecutor` / `MockExecutor` and `ExecutionEvent`.
- **Runtime Objective 2 - Alpaca Paper** uses the same complete application and replaces only the executor with `AlpacaPaperExecutor`. It begins only after Objective 1 is implemented and validated.

The principal semantic comparison boundary is the governed execution instruction selected by the four-state engine. Equivalent input, initial state, configuration, rule set, and versions MUST produce equivalent upstream meanings through that boundary. Broker-specific acknowledgement, rejection, fill, partial fill, price, timing, cancellation, latency, and status differences begin after that boundary and belong to `ExecutionEvent` and reconciliation.

No live or funded-capital execution is authorized.

---

## 3. Authority and Relationship to System Design

Authority order for implementation work is:

1. Current modified `DSE_JEH_TRANS_SAT_1_SYSTEM_DESIGN_V0_1_091226.md`.
2. This V0.2 plan after human approval; until approval, current modified V0.1 remains the implementation-plan baseline.
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
| DEP-01 Bar Event Admission | IMPLEMENTED AND VALIDATED | Preserve behavior; reconcile contracts and complete every-bar evidence |
| DEP-02 Per-Entity Ordered Analytical State | IMPLEMENTED AND VALIDATED | Preserve entity isolation and causal ownership |
| DEP-03 JEH / Ehlers Phase Update | IMPLEMENTED AND VALIDATED | Do not rewrite validated mathematics without approved cause |
| DEP-04 Circular Phase State | IMPLEMENTED AND VALIDATED | Preserve explicit initialization/value presence and normalized phase |
| Rule #1 behavior | IMPLEMENTED AND VALIDATED as an explicit `ProductionEligibilityService` | Compile once with `expr`, emit typed eligibility and rule evidence, and preserve established 63-bar semantics |
| Rule Registry | IMPLEMENTED for the active Rule #1 set | Validate and retrieve the active typed rule set; no hot reload is authorized |
| ONLINE and OFFLINE adapters | IMPLEMENTED for the Phase 1 path | Both route through the same host-owned `BarEvent` pipeline; strong ONLINE continuity claims remain blocked |
| Authoritative proto and generated bindings | IMPLEMENTED | One proto defines 16 services and 18 RPCs; Go message and gRPC bindings are generated |
| Persistent runtime and observation services | IMPLEMENTED AND VALIDATED | Compiled host, lifecycle, shared status, heartbeat, evidence stream, zero-input persistence, and graceful stop |
| Built executable and start/stop scripts | IMPLEMENTED AND VALIDATED | Compiled-executable workflow with PID-scoped governed stop; no `go run` acceptance path |

Existing Phase 1 reports record passing `buf lint`, `buf generate`, `go test ./...`, and `go build` validation. Those reports establish the baseline; they do not prove the four-state Dynamic Execution Engine or either complete runtime objective.

### 4.2 Not Yet Implemented as the Complete Target

- the four-state Dynamic Execution Engine and its state-associated actions;
- complete `LocalPaperExecutor` / `MockExecutor` E2E path;
- `AlpacaPaperExecutor`;
- complete `ExecutionEvent` and reconciliation path; and
- Runtime Objective 1 and Runtime Objective 2 completion evidence.

The implemented vertical slice ends deliberately after Production Eligibility. Its existing phase-motion-unavailable outcome is a legacy contract result from the superseded decomposition, not implementation of the replacement engine. No four-state transition, state action, governed execution instruction, or execution semantics are inferred.

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
  -> Production Eligibility Controller
       -> BLOCKED: typed evidence and terminal bar outcome
       -> ELIGIBLE: continue
    -> deterministic PhaseTransitionState calculation
    -> component-scoped boundary policy rule
    -> four-state Dynamic Execution Engine
      -> DISREGARD: no allocation action
      -> ALLOCATE: rank by Phase Velocity
      -> HOLD & TRAIL: trail stops dynamically
      -> LIQUIDATE: reallocate freed capital
    -> governed execution instruction when required
  -> selected Executor
  -> ExecutionEvent
  -> reconciliation, evidence, telemetry, runtime state
```

Every received bar MUST have an attributable terminal result. A rejected, initializing, blocked, not-applicable, no-action, or error result is completed processing, not a silent disappearance.

The application remains running until governed stop, cancellation, or fatal application failure under approved lifecycle semantics. No incoming bars is a valid `RUNNING` condition. Liveness, source connectivity, and activity are separate signals; activity counters may remain at zero while the application remains healthy and subscribed.

---

## 6. Two Implementation Phases

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

Phase 2 begins by deriving the minimum proto/service changes required by the approved replacement design. Existing post-phase services remain generated but have no continuing design authority merely because they exist. This plan does not decide which survive, combine, disappear, or are replaced. No unresolved action algorithm or execution policy may be filled in for implementation convenience.

Current status: `PARTIALLY COMPLETE`. The Rule Registry, compile-once Rule #1 program, and Production Eligibility Controller exist. The post-phase contract requires review; the four-state engine, its actions, and both executor adapters do not exist.

### 6.3 Phases Versus Runtime Objectives

The two implementation phases describe construction responsibility. Runtime Objectives 1 and 2 exercise the completed application with different executors. Objective 1 is a mandatory complete-application gate before Objective 2; neither objective is a renamed implementation phase.

---

## 7. Proto-First Contract Strategy

### 7.1 Single Authoritative Proto

The sole DSE_JEH proto remains:

`DSE_JEH/api/proto/dse_jeh/v1/DSE_JEH_TransSat_1.proto`

The previous contract workstream expanded this file in place to 16 services and 18 RPCs under the superseded post-phase design. The authoritative proto and generated bindings remain unchanged in this documentation task. After approval of the replacement documents, a bounded contract-design task MUST derive the minimum service/message changes required by the four-state engine while preserving compatibility where required.

The implementation order is:

```text
replacement System Design approved
  -> replacement Implementation Plan approved
  -> derive minimum four-state proto/service changes
  -> update human-reviewed authoritative proto
  -> buf lint / compatibility checks
  -> generated Go contracts
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

Exact declaration names and the minimum service decomposition require later proto review. Financially or operationally meaningful Go enums, states, statuses, events, outcomes, and execution meanings MUST NOT exist as parallel ungoverned vocabulary.

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
| Four-state Dynamic Execution Engine | exactly four states, four boundary policies, and state-associated actions | `PhaseTransitionState`, typed state/policy/action outcomes, and separate audit records | Existing post-phase services require later design-derived review; no service split is selected here | deterministic transition facts plus governed state/action changes |
| Governed execution boundary | convey required external action without claiming execution | execution-instruction identity/status and execution outcome | Existing execution-related services require later review; executor remains a separate adapter boundary | deterministic instruction identity/idempotency and adapter routing |
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
3. Reconcile DEP-01 reception/admission so every candidate, including rejection, has terminal evidence.
4. Reconcile DEP-02 state ownership with lifecycle, drain, recovery, and eligibility-controller inputs.
5. Reconcile DEP-03/04 output with distinct mathematical status and production eligibility contracts.
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
\omega[n] = \operatorname{circular\_delta}(\phi[n-1], \phi[n])
$$

Its unit is degrees per bar with $\Delta Bar=1$. Phase Velocity supports the `ALLOCATE` ranking action; it is not a strategy state or separate Dynamic Execution stage. The exact 180-degree signed-direction tie remains unresolved.

### 11.2 History-Free Transition Model

Each state transition is defined only by its input state and resulting output state. Historical transition values may be retained for telemetry, diagnostics, auditability, or analysis, but MUST NOT alter an already-established state. Do not implement historical trajectory inference, multi-transition lookback, smoothing, acceleration as a state determinant, or downstream reconstruction of earlier state.

### 11.3 Governed Boundary Policies

Deterministic Go/domain code computes factual transition and boundary inputs. Component-scoped `expr` rules evaluate these authoritative policies and produce typed outcomes:

- cross 270 degrees -> `HOP_ON` -> `ALLOCATE`;
- cross 0 degrees -> `HOLD_AND_TRAIL`;
- cross 90 degrees -> `HOP_OFF` -> `LIQUIDATE`; and
- cross 180 degrees -> `DISREGARD`.

The policy meanings are resolved. The deterministic algorithm for wrap, reverse movement, exact-boundary samples, large or multiple-boundary transitions, and jitter/recrossing remains to be specified. `expr` MUST NOT implement that mathematics.

### 11.4 Four Persistent States and Actions

| Interval | Persistent strategy state | Authoritative action | Remaining implementation detail |
| --- | --- | --- | --- |
| $180 \le \phi < 270$ | `DISREGARD` | No allocation action | None beyond typed no-action behavior |
| $270 \le \phi < 360$ | `ALLOCATE` | Rank by Phase Velocity | Formula, timing, ties, staleness, candidate eligibility |
| $0 \le \phi < 90$ | `HOLD_AND_TRAIL` | Trail stops dynamically | Exact trailing algorithm |
| $90 \le \phi < 180$ | `LIQUIDATE` | Reallocate freed capital | Capital, capacity, sizing, sequencing, reconciliation |

`HOP_ON` and `HOP_OFF` are rule-policy firing events, not states. No additional state or HOP terminology is authorized.

### 11.5 Contract and Execution Integration

After this replacement plan is approved, derive the minimum proto/service changes from the four-state model. Review `PhaseMotionService`, `BoundaryCrossoverService`, `StrategyRegionService`, `UniverseStateService`, `CandidateRankingService`, `StrategyDecisionService`, and `ExecutionIntentService` without presuming that any must survive, combine, disappear, or be replaced.

The engine must still produce the governed execution instruction required to cross into the external executor. Do not force the former `StrategyDecision`/`ExecutionIntent` two-stage model into the design. Mock/local paper remains the first target; future Alpaca PAPER remains later. No live/funded execution is authorized.

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
- `execution_instructions`; and
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
  -> four-state Dynamic Execution Engine
  -> state-associated action
  -> governed execution instruction, when required
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

Current status: `PARTIALLY COMPLETE`. The stored-bar path is validated through Production Eligibility, including collection run `20260911T161623Z-1` with 3,120 received/admitted bars, but the four-state engine, `LocalPaperExecutor`, `ExecutionEvent`, reconciliation, and full repeated E2E comparison remain absent. Runtime Objective 1 is not accepted.

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
- `PhaseTransitionState` circular-delta vectors after direction-policy approval;
- boundary-policy vectors for 270/0/90/180 transitions after deterministic crossing semantics are approved;
- history-free state-transition and boundary-event tests;
- ALLOCATE ranking, HOLD & TRAIL, LIQUIDATE reallocation, and execution-instruction idempotency tests after action-policy approval;
- lifecycle, zero-input, stop/drain, failure, and activity-accounting tests; and
- race tests for entity and universe ownership.

Current result: authoritative proto lint/generation/breaking checks, Go tests, vet, and build pass for the implemented slice. The race test is environment-blocked by unavailable CGO/GCC. Tests for blocked Phase 2 behavior cannot exist until the controlling semantics are approved.

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
| §9 / 10 | Production Eligibility / Rule #1 | `ProductionEligibilityService.EvaluateProductionEligibility` | `internal/services/services.go` | Internal direct call after phase | Bar 63/64 test; 1,848 initializing and 1,272 eligible replay outcomes | IMPLEMENTED | None for current-bar eligibility ownership |
| §11.1-11.2 / 6-7 | `PhaseTransitionState` and history-free transition ownership | Contract changes derived after replacement approval | Not implemented | Not wired; current terminal outcome reflects the superseded contract | No four-state tests | NOT STARTED | Circular direction/tie and transition contract review |
| §11.3-11.4 / 7-9 | Four boundary policies, four states, and state-associated actions | Contract changes derived after replacement approval | Not implemented | Not wired | No boundary/state/action tests | BLOCKED BY SPECIFIED DETAILS | Crossing algorithm, ranking, trailing, capital, sizing, and sequencing details |
| §11.5, §13-14 / 10-11 | Governed instruction, local execution, reconciliation, and Objective 1 | Existing execution services require design-derived review | No replacement implementation | Not wired | No complete E2E test | NOT STARTED | Contract review and executor/reconciliation semantics |
| §15 / 12 | Alpaca Paper and Objective 2 | Same governed execution boundary | No adapter | Not wired | No paper integration test | NOT STARTED | Objective 1 approval and paper-only adapter policy |

Generated stubs prove contract availability, not implemented business behavior. Service definition remains distinct from network exposure.

---

## 18. Dependency-Ordered Implementation Sequence

1. Approve the replacement System Design and this replacement Implementation Plan as the governing architecture.
2. Preserve the validated DEP-01-through-DEP-04, Production Eligibility, persistent-runtime, and compile-once Rule Registry implementation.
3. Derive the minimum authoritative proto/service changes from the four-state model; do not preserve the former service split by default.
4. Human-review the replacement state, boundary-policy, action, execution-instruction, event, status, identity, and audit contracts.
5. Run approved proto compatibility/lint checks and regenerate bindings only after that review.
6. Implement deterministic `PhaseTransitionState` calculation for the current eligible bar using retained predecessor analytical context.
7. Specify and implement deterministic boundary facts plus the four component-scoped `expr` boundary policies for 270, 0, 90, and 180 degrees.
8. Implement exactly four persistent strategy states with history-free transitions and separate audit/telemetry history.
9. Implement the four state-associated actions: DISREGARD no action, ALLOCATE ranking by Phase Velocity, HOLD & TRAIL dynamic trailing, and LIQUIDATE capital reallocation.
10. Implement the governed execution-instruction boundary, `LocalPaperExecutor` / `MockExecutor`, `ExecutionEvent`, and local reconciliation.
11. Execute and accept Runtime Objective 1 with deterministic E2E, lifecycle, telemetry, every-bar accountability, and repeated-run evidence; freeze the governed instruction semantics.
12. After explicit approval, implement and validate the paper-only Alpaca adapter, execute Runtime Objective 2, compare at the governed execution-instruction boundary, and produce completion evidence.

The sequence MUST NOT jump from isolated DEP tests directly to Alpaca Paper. Objective 1 is the required complete-application proving gate.

### 18.1 Current Sequence Status

| Item | Current status | Evidence or blocker |
| --- | --- | --- |
| 1 | COMPLETE | Replacement System Design is documented; this plan now reflects it |
| 2 | COMPLETE | Validated implementation remains through Production Eligibility; Rule #1 and persistent runtime are retained |
| 3 | NOT STARTED | Existing post-phase proto reflects the superseded design and requires bounded review |
| 4 | NOT STARTED | Depends on replacement contract proposal |
| 5 | NOT STARTED | No proto or generated files are changed by this documentation task |
| 6 | BLOCKED | Circular direction convention, including exact 180-degree handling, remains pending |
| 7 | BLOCKED | Deterministic wrap, reverse, exact-boundary, jitter, and multi-boundary crossing semantics remain pending |
| 8 | NOT STARTED | Four states and history-free principle are resolved; implementation follows items 3-7 |
| 9 | BLOCKED BY ACTION DETAILS | Ranking, trailing, capital, capacity, sizing, sequencing, freshness, and reconciliation policies remain to be specified |
| 10 | NOT STARTED | Replacement instruction contract and executor/reconciliation semantics require review |
| 11 | PARTIAL | Existing stored-bar run validates only through Production Eligibility; full Objective 1 is not accepted |
| 12 | NOT STARTED | Requires accepted Objective 1 and explicit Alpaca Paper authorization |

---

## 19. Blocking Decisions and Open Questions

The following remain open and do not prevent approval of this plan. Work stops only at the implementation boundary that needs the decision.

| Boundary | Status | Required decision |
| --- | --- | --- |
| Four-state entry ownership | RESOLVED | Current eligible `Bar[n]` owns processing; retained $\phi[n-1]$ is predecessor context; bar 64 is the first possible engine entry |
| `PhaseTransitionState` | PARTIALLY RESOLVED | $\omega[n]=\operatorname{circular\_delta}(\phi[n-1],\phi[n])$ in degrees per bar with $\Delta Bar=1$; approve the circular direction convention, including exact 180-degree handling |
| Boundary facts and policies | POLICY MEANINGS RESOLVED; ALGORITHM BLOCKED | Implement governed 270 -> `HOP_ON` -> `ALLOCATE`, 0 -> `HOLD_AND_TRAIL`, 90 -> `HOP_OFF` -> `LIQUIDATE`, and 180 -> `DISREGARD`; specify wrap, reverse, exact samples, jitter, and large/multiple-boundary transitions |
| Four-state persistence | RESOLVED | Exactly four states and history-free transition semantics; implementation and contract review remain |
| ALLOCATE action | BLOCKED PENDING DETAIL | Candidate snapshot, freshness, Phase Velocity ranking formula/timing/ties, capacity, and sizing |
| HOLD & TRAIL action | BLOCKED PENDING DETAIL | Exact dynamic trailing-stop algorithm and execution-update semantics |
| LIQUIDATE action | BLOCKED PENDING DETAIL | Capital representation, reallocation, capacity, sizing, sequencing, and reconciliation |
| Governed execution instruction | BLOCKED PENDING CONTRACT REVIEW | Minimum replacement contract, identity, idempotency, expiry/cancellation, sizing, and correlation |
| Executors/reconciliation | BLOCKED PENDING DESIGN DECISION | Execution outcome state machine, retries, duplicate events, recovery and reconciliation |
| Lifecycle/recovery | PARTIALLY RESOLVED | OFFLINE completion is distinct from application stop and accepted work drains before stop; checkpoint, restart, pending-work recovery, and future execution reconciliation remain open |
| ONLINE continuity | BLOCKED FOR STRONG CONTINUITY CLAIMS | Upstream loss detection, finalized-bar guarantee, resume cursor, backpressure |
| Alpaca Paper | BLOCKED UNTIL OBJECTIVE 1 APPROVAL | Paper-only endpoint/configuration, order mapping, idempotency, fill/cancel/reconciliation policy |

No unresolved item may be silently converted into a default implementation assumption.

---

## 20. Continuing Non-Goals

Neither this plan nor the completed runtime slice authorizes:

- creation of another DSE_JEH proto;
- implementation of the four-state engine before the required mathematical, policy, action, and contract details are approved;
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

---

## 22. Authorization Statement

This Implementation Plan V0.2 is **APPROVED** as the current implementation-plan authority subordinate to the approved `DSE_JEH_TransSat_1` System Design V0.1.

V0.1 remains preserved as historical documentation. The approved System Design now governs the four-state replacement. This plan does not resolve any mathematical, action-policy, contract, lifecycle, or execution detail explicitly retained as open.

Approval of this document authorizes implementation planning to proceed through the dependency gates stated here; it does not by itself approve unresolved mathematics, live trading, production deployment, Alpaca Paper execution, or any broker credentials. Each blocked boundary and Runtime Objective 2 requires its stated review and authorization.

Implementation progress recorded here does not expand authorization. Current eligible `Bar[n]` ownership, bar 64 as the first possible four-state-engine entry, the `PhaseTransitionState` relation, exactly four states, history-free transitions, and the four fixed boundary-policy meanings are resolved. Circular direction/tie handling, deterministic boundary algorithms, state-action details, replacement proto/service contracts, executor/reconciliation behavior, and broker interaction remain subject to their stated approvals. No paper/live order submission, staging, commit, push, or deployment is approved by this replacement.

Overall implementation verdict: **V0.2 PLAN REMAINS IN PROGRESS**.
