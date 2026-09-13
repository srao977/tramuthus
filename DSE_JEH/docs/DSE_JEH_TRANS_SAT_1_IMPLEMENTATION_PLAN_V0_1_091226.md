# DSE_JEH_TransSat_1 Implementation Plan V0.1

## 1. Document Control

| Document control | Value |
| --- | --- |
| Filename | `DSE_JEH_TRANS_SAT_1_IMPLEMENTATION_PLAN_V0_1_091226.md` |
| Date | 2026-09-13 |
| Version | V0.1 |
| Status | APPROVED |
| Implementation status | DEP-01 through DEP-04 implemented and validated; complete application architecture not implemented |
| Architectural authority | `DSE_JEH_TRANS_SAT_1_SYSTEM_DESIGN_V0_1_091226.md` |
| Scope | Implementation planning for `DSE_JEH_TransSat_1` |

This is an implementation plan subordinate to the System Design. It is not a replacement design, new architecture, code or protobuf implementation, validation report, authorization to build, or new System Design version.

---

## 2. Executive Summary

Implementation is sequenced in exactly two phases:

1. **PHASE 1 - BAR INPUT AND JEH ANALYTICAL PATH** has implemented ONLINE and OFFLINE bar-stream acquisition, common `BarEvent` admission, per-entity analytical state, the approved JEH/Ehlers solver, circular phase state, and validated `PhaseEvidence`. Its current lifecycle, health, telemetry, evidence completeness, and persistence behavior are only partial relative to the complete application design.
2. **PHASE 2 - COMPLETE GOVERNED APPLICATION PATH** begins only after its proto and unresolved-design gates. It first makes the Production Eligibility Controller and common Governed Rule Engine explicit, then implements Phase Velocity ($\omega$), boundary crossover, four-region strategy rules, universe state, candidate ranking, strategy decisions, `ExecutionIntent`, mock execution, and `ExecutionEvent`, together with complete lifecycle, telemetry, and bar accountability.

The plan is proto-first. The existing single authoritative DSE_JEH proto is extended in place before each governed behavior is implemented. It governs all operationally or financially meaningful inbound, internal-processing, outbound, lifecycle, telemetry, rule-outcome, and execution-boundary vocabulary, whether execution is in-process or distributed. Generated Go represents those contracts; Go interfaces implement component mechanics; deterministic Go algorithms implement approved mathematics and state mechanics; `github.com/antonmedv/expr` evaluates governed conditions; typed proto outcomes authorize routing/state transitions; and evidence records every result. No component requires a network hop merely because its operational responsibility is proto-governed.

---

## 3. Authority and Relationship to System Design

The sole architectural authority is the System Design at `DSE_JEH/docs/DSE_JEH_TRANS_SAT_1_SYSTEM_DESIGN_V0_1_091226.md`. This plan may sequence, locate, and test implementation artifacts, but MUST NOT alter decided architecture or close System Design Open Questions without human-approved design resolution.

Every planned artifact traces to a System Design section, named artifact, DEP responsibility, and applicable Open Question (OQ). OQ numbers in this plan mean the numbered questions in System Design §22 **Open Questions and Blocking Decisions**.

| Authority class | Controlling System Design artifacts |
| --- | --- |
| Input and analytical path | §2 Executive Summary; §3 Scope, Authority, and Principles; §4 Overall System Context; §6 One BarEvent Lifecycle; §7 BarEvent Semantics and Ordering; §8 Phase Calculation Inside V0.1; §9.1 Per-Entity Hot State; DEP-01 through DEP-04 |
| Source modes | §14 Online and Offline Stream Modes; §15 The 4,519-Bar Offline Proving Stream; §17 Determinism, Mode Equivalence, and Evidence Identity |
| Dynamic execution | §5 Dynamic Execution Pipeline; §9 Per-Entity and Universe State; §10 Four-Region Circular JEH Rules Engine; §11 Phase Motion and Phase Velocity ($\omega$); §12 Dynamic Allocation and Strategy Behavior; §13 Decision and External Execution Boundary; DEP-05 through DEP-11 |
| Operations and validation | §16 Independent Analytical Reference and Phase Equivalence; §18 Conceptual Component Model; §19 Portability, Persistence, and Failure Isolation; §20 Viewer Boundary; §21 Validation Strategy; §22 Open Questions and Blocking Decisions |

If implementation planning exposes a conflict or missing decision, work stops at the affected boundary, the issue is recorded in §14 of this plan, and the System Design is resolved separately.

---

## 4. Implementation Methodology

The governing order within each phase is:

1. Confirm System Design authority and unresolved blockers.
2. Approve required enums, messages, and genuine service/RPC boundaries.
3. Update the single authoritative proto.
4. Generate Go bindings and stubs with the governed toolchain.
5. Implement internal Go packages against the approved contracts.
6. Build the normal executable.
7. Run unit, component, integration, determinism, and equivalence tests.
8. Produce completion and validation evidence.
9. Obtain separate human authorization before the next phase or deployment step.

Governed runtime contracts MUST NOT first emerge as accidental independent Go types and later be retrofitted into proto. Internal solver buffers, algorithms, locks, queues, and helper interfaces may remain Go implementation details. Governed or financially meaningful states, statuses, events, decisions, outcomes, execution meanings, lifecycle meanings, and evidence MUST NOT exist only as parallel Go vocabulary, even when they remain in-process.

### 4.1 Canonical Implementation Relationship

```text
authoritative proto typed input/evidence
      -> component-scoped context
      -> identified/versioned precompiled expr rule
      -> raw evaluator result
      -> rule-specific typed proto outcome
      -> Go orchestration and authorized state transition
      -> RuleEvaluationEvidence
      -> next governed component or attributable termination
```

For deterministic work the sequence is `typed input -> approved Go algorithm -> typed proto evidence -> governed rule/routing`. For financial execution it is `crossover event -> strategy region -> universe/ranking -> StrategyDecision -> ExecutionIntent -> Executor -> ExecutionEvent`. Implementation MUST NOT collapse these stages.

Future implementation documentation for every authored Go file must identify package purpose, architectural responsibility, System Design cross-reference, inputs, outputs, state ownership, configuration, failures, concurrency assumptions, and non-trivial function behavior. Proto declarations must document semantic purpose, field meaning, identity, units, presence, status lifecycle, and source/causal lineage.

---

## 5. Two-Phase Implementation Model

```text
DSE_JEH_TransSat_1

PHASE 1 - BAR INPUT AND JEH ANALYTICAL PATH
============================================================
                  DSE_JEH_MODE
                       |
                 +-----+-----+
                 |           |
              ONLINE      OFFLINE
                 |           |
       Fin_Feed_Sat_1     Internal
          gRPC stream     Stream Producer
                 |           |
                 +-----+-----+
                       |
                    BarEvent
                       |
                       v
                JEH / Ehlers
                 Phase Solver
                       |
                       v
                 PhaseEvidence

PHASE 2 - DYNAMIC EXECUTION PIPELINE
============================================================
                 PhaseEvidence
                       |
                       v
             Production Eligibility
                   Controller
                       |
             +---------+---------+
             |                   |
       BLOCK + evidence      ELIGIBLE
                                 |
                                 v
                  Phase Motion
                       |
                       v
               Boundary Crossover
                       |
                       v
       JEH Strategy Region / Rules Engine
                       |
                       v
                 Universe State
                       |
                       v
               Candidate Ranking
                       |
                       v
               Strategy Decision
                       |
                       v
                ExecutionIntent
                       |
                       v
          External Execution Boundary
                       |
                       v
        Mock ExecutionEvent during proving
```

These implementation phases do not replace the System Design's DEP-01 through DEP-11 logical model. Phase 1 ends at validated `PhaseEvidence`; Phase Motion is exclusively Phase 2. Phase 2 is one coherent implementation phase and is not subdivided into additional implementation phases.

---

## 6. Single Authoritative Proto Strategy

### 6.1 Reviewed physical contract

Repository convention in `Fin_FeedSat_1` uses Buf v2, proto sources under `api/proto/<package>/v1`, source-relative Go/gRPC generation into `gen`, and a single evolving proto for that runtime. Human review on 2026-09-12 approved the planned physical path and package below. This approval does not pass the System Design §22.3 contract gate or authorize proto creation.

| Item | Reviewed value | Governance |
| --- | --- | --- |
| Sole proto source | `DSE_JEH/api/proto/dse_jeh/v1/DSE_JEH_TransSat_1.proto` | Approved planning value; creation blocked by System Design §22.3 |
| Proto package | `dsejeh.v1` | Approved planning value; creation blocked by System Design §22.3 |
| Proposed Go module | `tramuthus/dse-jeh-transsat-1` rooted at `DSE_JEH` | Deferred until Go implementation authorization |
| Proposed `go_package` | `tramuthus/dse-jeh-transsat-1/gen/dse_jeh/v1;dsejehv1` | Approved planning value contingent on the future module decision |
| Buf module config | `DSE_JEH/buf.yaml` | Match repository Buf v2 convention |
| Generation config | `DSE_JEH/buf.gen.yaml` | Go and gRPC plugins, source-relative paths |
| Generated output | `DSE_JEH/gen/dse_jeh/v1` | Confirm generated-code commit policy before generation |

No other DSE_JEH proto is planned. Phase 1 creates the approved initial declarations; Phase 2 extends this same file without renumbering or reusing existing fields. Compatibility, reserved fields, package versioning, lint, and breaking-change checks become governed build requirements.

### 6.2 Service boundary policy

The upstream ONLINE contract is owned by `Fin_Feed_Sat_1`. Repository evidence establishes the current candidate as `finfeedsat.v1.IngestionService.StreamBars(StreamBarsRequest) returns (stream Bar)`, but binding it is blocked by OQ 2-8 and requires interface review. Its current implementation does not expose accepted sequence, detectable queue loss, finalized-bar guarantee, or resumable position. DSE_JEH MUST import/map only an approved upstream contract at its transport adapter and MUST NOT copy the upstream bar message into its own proto as an ersatz transport API.

The proto MUST describe the complete satellite's operational services and responsibilities, including governed in-process work. This does not require each DEP component to be a separately deployed gRPC server or create an artificial network hop. The JEH Phase Solver, Production Eligibility Controller, Phase Motion Analyzer, Boundary Crossover Detector, Strategy Region / Rules Engine, Universe State Coordinator, and Candidate Ranking Engine may remain in-process Go components while their governed contracts, typed outcomes, and evidence are proto-defined. RPC declarations are added only for approved callable network boundaries.

The same file MUST ultimately cover lifecycle/mode/health, source/subscription state, bar reception/admission, analytical state and JEH evidence, production eligibility, motion, crossover, regions, universe/candidates/ranking, decision, intent, execution event, rule identity/evaluation/evidence, telemetry/activity, diagnostics/degradation, and outbound publication. No additional DSE_JEH proto may be created.

The existing proto has been inspected and implements only the Phase 1 analytical vocabulary. The next authorized contract task must expand this same file; it must not split the contract. Complete operational services may be represented without turning every internal responsibility into a network RPC.

---

## 7. Proto-First Development Workflow

### 7.1 Reviewed proto artifact plan

Names below are planning outcomes from System Design §22. Required declarations remain blocked until their complete semantics are approved. Deferred declarations are omitted from the initial Phase 1 proto rather than guessed.

| Phase | Kind | Proposed name | Purpose | System Design authority | Contract role | Blocking OQ | Approval |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 1 | ENUM | `RuntimeMode` | Represent ONLINE/OFFLINE in governed configuration evidence | §14 Online and Offline Stream Modes | Governed configuration | None; values fixed by design | APPROVED DECLARATION |
| 1 | ENUM | `PhaseStatus` | Distinguish unspecified, initializing, and observable states without a numeric sentinel; do not add an unapproved `INVALID` lifecycle | §8 Phase Calculation Inside V0.1; DEP-04 | Governed evidence | OQ 15 resolved | APPROVED DECLARATION |
| 1 | ENUM / FINDINGS | Admission disposition and input findings | Separate terminal admitted/rejected disposition from potentially coexisting duplicate, conflict, gap, malformed, and out-of-order findings | §7 BarEvent Semantics and Ordering; DEP-01 | Governed evidence | OQ 1, 3-8, 31-33 | BLOCKED; exact shape and precedence require approval |
| 1 | MESSAGE | `SourceProvenance` | Distinguish source observation, producer/provider, source mode, entity order scope, replay policy, and defined source times | §7; §17 Determinism, Mode Equivalence, and Evidence Identity | Governed evidence | OQ 1, 9, 12 | BLOCKED |
| 1 | MESSAGE | `BarEvent` | Common source-independent accepted-bar candidate containing finite `high`/`low` plus approved identity/order/provenance | §4 Overall System Context; §6 One BarEvent Lifecycle; §7; DEP-01 | Governed internal boundary | OQ 1, 12; OQ 16 resolved | BLOCKED |
| 1 | MESSAGE | `BarAdmissionEvidence` | Record deterministic admission identity, terminal disposition, findings, and causal input without state mutation for rejection | §7; §17; DEP-01 | Governed evidence | OQ 1, 31-33 | BLOCKED |
| 1 | MESSAGE | `SolverIdentity` | Identify solver family/name, solver and algorithm/reference versions, median-price input series, initialization requirement, and configuration identity | §8; §17 | Governed evidence | OQ 18 resolved | APPROVED DECLARATION; exact field encoding reviewed with containing evidence |
| 1 | MESSAGE | `PhaseEvidence` | Carry explicit phase status/value, entity order, solver identity, and causal admission lineage | §8; §16; DEP-03/04 | Governed evidence | OQ 1, 12; OQ 15, 18 resolved; OQ 19 affects acceptance | BLOCKED by unresolved identity/lineage |
| Complete app | ENUM / MESSAGE / SERVICE / RPC | Runtime lifecycle, health, activity and operations declarations | Persistent liveness, start/drain/stop, zero-input validity, source health, degradation, activity, and final state | §19 | Governed operations, whether in-process or externally queried | Exact shape remains OQ 49 | REQUIRED BEFORE IMPLEMENTATION OF COMPLETE LIFECYCLE |
| 2 | MESSAGE / ENUM | Production eligibility input, outcome, state and evidence | Govern Rule #1 as an event-flow gate after DEP-04 and before DEP-05 | §8.2; Production Eligibility Controller | Governed internal contract/evidence | Bar-64/bar-65 motion question does not block the eligibility outcome itself | REQUIRED |
| 2 | MESSAGE / ENUM | Rule identity, evaluation status, typed outcome and `RuleEvaluationEvidence` | Attribute every identified/versioned rule evaluation and component-specific outcome | §13.1-13.3 | Governed internal contract/evidence | Exact descriptive names require proto review | REQUIRED |
| 2 | MESSAGE / ENUM | Runtime activity and per-bar processing outcome | Distinguish reception, admission, rejection, initialization, phase eligibility, motion, crossover, hop events, decisions, intents, events, and terminal processing result | §19.1 | Governed telemetry/evidence | Counter/reset and exact outcome taxonomy require review | REQUIRED |
| 2 | ENUM | `JehStrategyRegion` | Represent initializing/unavailable and the canonical persistent regions `DISREGARD`, `ALLOCATE`, `HOLD_AND_TRAIL`, and `LIQUIDATE` | §9; §10; DEP-07 | Governed evidence | OQ 25-28 | Required |
| 2 | ENUM | `DecisionType` | Represent no-action, `ALLOCATE`, `HOLD_AND_TRAIL`, `LIQUIDATE`, or `DISREGARD` decisions as approved | §12; §13; DEP-10 | Governed evidence | OQ 36-44 | Required |
| 2 | ENUM | `ExecutionOutcomeStatus` | Represent mock/external acceptance, rejection, action, failure, and reconciliation | §13 | External/governed outcome | OQ 43, 48 | Required |
| 2 | MESSAGE | `PhaseMotionEvidence` | Record Phase Velocity ($\omega$), signed direction, magnitude, units, validity, and causal phases without freezing unresolved mathematics | §11 Phase Motion and Phase Velocity ($\omega$); DEP-05 | Governed evidence | OQ 20-24, 29-30 | Required after blocker resolution |
| 2 | MESSAGE | `BoundaryCrossoverEvidence` | Record validated circular boundary crossing separately from persistent region membership; represent `HOP-ON` only for a valid 270° crossover into `ALLOCATE`, `HOP-OFF` only for a valid 90° crossover into `LIQUIDATE`, and the 0° transition without inventing a name | §10; DEP-06 | Governed evidence | OQ 25-30 | Required after blocker resolution |
| 2 | MESSAGE | `EntityStrategyState` | Record current/prior strategy region, distinct crossover evidence, candidacy, holding, and freshness | §9; §10; DEP-07/08 | Governed evidence | OQ 25-35, 38 | Required |
| 2 | MESSAGE | `UniverseStateEvidence` | Record latest entity views, holdings, capacity, candidates, and pending work | §9.2 Universe State; DEP-08 | Governed evidence | OQ 34-41 | Required after blocker resolution |
| 2 | MESSAGE | `CandidateRankingEvidence` | Record candidate set, approved score inputs, ordering, ties, and policy identity | §11; §12; DEP-09 | Governed evidence | OQ 20-24, 34-38 | Required after blocker resolution |
| 2 | MESSAGE | `StrategyDecision` | Record decision, reasons, causal state/ranking, and policy identity | §12; §13; DEP-10 | Governed evidence | OQ 36-44 | Required |
| 2 | MESSAGE | `ExecutionIntent` | Governed idempotent request with causal decision, target, validity, and correlation | §13; DEP-11 | External/governed contract | OQ 42, 47 | Required after blocker resolution |
| 2 | MESSAGE | `ExecutionEvent` | Record mock/external outcome and reconciliation lineage | §13 | External/governed contract | OQ 43, 48 | Required after blocker resolution |
| 2 | SERVICE | `ExecutionIntentService` | Boundary for an external execution adapter only if process separation is approved | §13 Decision and External Execution Boundary | External service | OQ 42-43, 47-48 | Required before inclusion |
| 2 | RPC | `StreamExecutionIntents` | Deliver governed intents to the external adapter if streaming is approved | §13 | External RPC | OQ 42-43, 47-48 | Required before inclusion |
| 2 | RPC | `ReportExecutionEvent` | Return outcome/reconciliation evidence if callback RPC is approved | §13 | External RPC | OQ 43, 48 | Required before inclusion |

No service/RPC row is authorization to create that boundary. Human review may retain only messages and use another approved transport. Internal component APIs remain Go interfaces.

---

## 8. PHASE 1 - BAR INPUT AND JEH ANALYTICAL PATH

### 8.1 System Design Authority

- §2 **Executive Summary**
- §3 **Scope, Authority, and Principles**
- §4 **Overall System Context** diagram
- §6 **One BarEvent Lifecycle**
- §7 **BarEvent Semantics and Ordering**
- §8 **Phase Calculation Inside V0.1**
- §9.1 **Per-Entity Hot State**
- §14 **Online and Offline Stream Modes**
- §15 **The 4,519-Bar Offline Proving Stream**
- §16 **Independent Analytical Reference and Phase Equivalence**
- §17 **Determinism, Mode Equivalence, and Evidence Identity**
- §18 **Conceptual Component Model**
- §19 **Portability, Persistence, and Failure Isolation**
- §21 **Validation Strategy**, items 1-4 and 11-14
- §22 OQ 1-19, 31-33, 45-46, 49-50
- DEP-01 **Bar Event Admission**, DEP-02 **Per-Entity Ordered Analytical State**, DEP-03 **JEH / Ehlers Phase Update**, DEP-04 **Circular Phase State**

### 8.2 Scope

Phase 1 delivered the implemented path from bar-stream acquisition through validated `PhaseEvidence`: `DSE_JEH_MODE`; ONLINE and OFFLINE input boundaries; common `BarEvent`; admission; per-entity ordered analytical state; JEH/Ehlers phase calculation; circular phase state; provenance and deterministic identity; build and validation evidence. Its lifecycle, health, telemetry, persistence, stop mechanism, and per-bar outcome evidence are partial and remain work required by the complete target application. Phase Motion is not in Phase 1.

### 8.3 Runtime and Startup

The future host reads `DSE_JEH_MODE` exactly once during startup, accepts only `ONLINE` or `OFFLINE`, validates mode-specific configuration, constructs one producer/input component, wires it to common admission, reports lifecycle status, and owns cancellation, drain, and shutdown. The Dynamic Execution Pipeline never reads the environment variable.

The existing executable is `dse-jeh-transsat-1`, built from `DSE_JEH/cmd/dse-jeh-transsat-1`. The existing `DSE_JEH/scripts/Start-DSEJEHTransSat1.ps1` validates mode/selector basics and launches the built binary; it does not use `go run`. The complete lifecycle requires an approved `Stop-DSEJEHTransSat1.ps1` or equivalent governed process-stop mechanism, proto-defined lifecycle/health evidence, startup dependency ordering, graceful drain/cancel, evidence flush, and final status publication.

ONLINE must remain running while connected and no bars arrive; zero input is valid activity-at-zero, not completion, disconnection, failure, or stop. OFFLINE end-of-selection behavior must be explicitly governed rather than accidentally defining application lifecycle. Existing configuration must be reviewed before normalization; currently implemented names include `DSE_JEH_MODE`, `DSE_JEH_OFFLINE_COLLECTION_RUN_ID`, `DSE_JEH_GRPC_ADDRESS`, `DSE_JEH_SYMBOLS`, `DSE_JEH_FINALIZED_ONLY`, `DSE_JEH_MAX_BARS`, `DSE_JEH_OUTPUT`, `DSE_JEH_REFERENCE_CSV`, and `DSE_JEH_COMPARISON_REPORT`, plus Mongo adapter settings.

### 8.4 ONLINE Input

Planned path:

```text
Fin_Feed_Sat_1 gRPC bar stream
  -> internal ONLINE gRPC Bar Consumer
  -> map to common BarEvent
  -> DEP-01 admission
  -> DEP-02/03/04 analytical path
  -> PhaseEvidence
```

The repository currently exposes the candidate `finfeedsat.v1.IngestionService.StreamBars(StreamBarsRequest) returns (stream Bar)` in `Fin_FeedSat_1/api/proto/fin_feedsat/v1/Fin_FeedSat_1.proto`. Server code supplies per-symbol chronological process-local catch-up followed by live delivery, with bounded queues that can discard an older queued bar. The contract has no accepted sequence, gap/loss notification, finalized-bar guarantee, or resume cursor. This is evidence for interface analysis, not approval of exact binding. Before implementation, OQ 2-8 must resolve service/method compatibility, finalized-bar semantics, ordering, duplicates, gaps/loss, reconnect/resumption, provenance, backpressure, lag, and warm-up. Fin transport types terminate in the ONLINE adapter and do not enter solver or strategy packages.

### 8.5 OFFLINE Input

Planned path:

```text
bar_sequence_db.bar_sequence
  -> Internal Offline Stream Producer
  -> one stored observation at a time
  -> common BarEvent
  -> DEP-01 admission
  -> same DEP-02/03/04 analytical path
```

The verified proving source contains 4,519 observations across multiple collection runs. MongoDB remains behind the producer; the analytical path consumes `BarEvent`, not a database. The producer does no phase, motion, crossover, state, ranking, decision, or intent work and never invokes `phase_angle_series_generator`. It is neither a batch JEH processor nor a separate backtest/strategy engine. Repository evidence establishes entity-scoped order by `(collection_run_id, symbol, generator_sequence_no)` but no unique source-authentic cross-entity total order. OQ 9 therefore requires a human-approved, explicit, versioned deterministic replay-order policy. Any future replay pacing is deferred producer control under OQ 10 and cannot alter event semantics.

### 8.6 Common BarEvent Admission

Both adapters call the same admission API. DEP-01 validates required values, entity and source identity, entity-scoped order, provenance, duplicate/conflict identity, gap/out-of-order policy, and deterministic evidence identity before any state mutation. Rejected input produces evidence and cannot mutate solver state. Terminal admission disposition must be modeled separately from potentially coexisting input findings unless a precedence rule is approved. An undetected missing accepted bar must not be silently crossed when it can change downstream behavior. Solver-required values are resolved by OQ 16; the complete contract and policy remain blocked by OQ 1, 3-8, 12, and 31-33.

### 8.7 Per-Entity Analytical State

DEP-02 owns entity identity, current bar identity, accepted order, initialization, bounded solver state, prior/current circular phase presence, phase validity, lineage, and solver/configuration identity. No valid number, including zero, is a sentinel. OQ 17 bounded checkpoint representation and OQ 45-46 recovery/persistence are deferred from the initial proto; solver state remains internal.

### 8.8 JEH/Ehlers Mathematics

The Go solver MUST reproduce the approved/reference behavior rather than substitute simplified or approximate mathematics. Governing reference artifacts are:

- `Bar_Sequence_Lab/phase_angle_series_generator/docs/PHASE_ANGLE_SERIES_SCIENTIFIC_DEFINITION_V0_1_091126.md` for frozen scientific decisions, input series, initialization, normalization, and validation limits;
- `Bar_Sequence_Lab/phase_angle_series_generator/internal/phase/phase.go`, specifically `MedianPrice`, `SmoothFour`, `NormalizeDegrees`, `DominantCyclePhase`, and `hilbert`, for the exact recurrence, constants, period constraints, phase calculation, and state progression;
- `Bar_Sequence_Lab/phase_angle_series_generator/internal/phase/phase_test.go` for current behavioral vectors and boundaries; and
- its independent TA-Lib `HT_DCPHASE` lineage described by the scientific definition.

Required equivalence includes Median Price

$$
P[n]=\frac{High[n]+Low[n]}{2},
$$

the same four-bar weighted smoothing; parity-separated Hilbert detrender, quadrature, `jI`, and `jQ` paths; exact reference coefficients (currently `0.0962` and `0.5769` in the cited artifacts); homodyne period discrimination; period smoothing and bounds; dominant-cycle projection; phase calculation; 63-observation initialization; first possible observable phase at observation 64; normalization to $[0,360)$; explicit unavailable phase; and legitimate zero degrees.

The reference artifacts, not duplicated prose in this plan, control the complete equations and transfer review. Any intentional deviation requires System Design review and independent validation. The normal Phase 1 path MUST NOT accept precomputed phase in place of bar-driven calculation.

### 8.9 PhaseEvidence

DEP-03/04 produce one attributable initializing or observable result for each admitted bar. `PhaseEvidence` records entity and source-observation identity, entity order, explicit status/presence, normalized angle when observable, solver identity/version, input-series definition, configuration identity, causal admission evidence, and production identity/time as approved. Existing code embeds Rule #1 behavior by withholding `OBSERVABLE` until contiguous valid bar 64; it does not implement the separately governed Production Eligibility Controller or authorize downstream routing.

### 8.10 Concurrency and State Ownership

- One per-entity coordinator owns each entity's ordered solver state; source adapters never mutate it directly.
- Admission routes accepted bars to the responsible owner and serializes mutation for one entity.
- Different entity owners may execute concurrently, subject to bounded queues and host cancellation.
- Duplicate/gap/order disposition occurs before solver mutation.
- ONLINE and OFFLINE invoke the identical admission interface.
- Shutdown stops intake, drains or explicitly cancels admitted work under an approved policy, flushes evidence, and reports final lifecycle state.
- Shared maps, registries, diagnostics, and publishers require single ownership or explicit synchronization; no mutable solver state is shared across entities.

Exact queueing, backpressure, drain, checkpoint, and recovery behavior remains blocked by OQ 7-8, 31-33, 45-46, and 49.

### 8.11 Diagnostics and Health

Structured evidence/logging must include runtime/configuration identity, source mode and provenance, entity, bar identity/order, admission result, solver/status identity, initialization/phase result, latency without changing semantics, error classification, and causal IDs. Secrets and full unnecessary payloads are excluded. Phase 1 health concerns include process lifecycle, selected mode, source connectivity, lag/backpressure where known, per-entity initialization/readiness, solver failures, evidence publication, drain/recovery, and version identity. OQ 49 defers the final lifecycle vocabulary and any `RuntimeStatus`, `RuntimeHealthEvidence`, or external operations service from the initial proto.

### 8.12 Go Module and File Plan

All paths are proposed under a new `DSE_JEH` module and require approval before creation.

| Phase | Package / file group | Proposed path | Responsibility | Inputs / outputs | Proto / internal interfaces | State and concurrency | Authority | Validation / blockers |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 1 | Runtime command | `cmd/dse-jeh-transsat-1` | Composition root, signals, start/drain/stop | Config to runtime exit | Uses approved proto config/health; host interfaces | Owns process lifecycle | §14, §18, §19 | Startup/component tests; OQ 49 |
| 1 | Configuration | `internal/config` | Parse/validate `DSE_JEH_MODE` and mode settings | Environment to immutable config | `RuntimeMode` mapping | Immutable after startup | §14 | Table tests; OQ 2, 9, 49 |
| 1 | Lifecycle | `internal/runtime` | Construct components and coordinate readiness/shutdown | Config and dependencies to lifecycle evidence | Internal producer/admission interfaces | Owns cancellation and drain | §18, §19 | Failure/drain tests; OQ 45, 49 |
| 1 | ONLINE consumer | `internal/input/online` | Consume approved upstream gRPC stream and map bars | Upstream stream to `BarEvent` | Upstream generated client; internal `BarProducer` | Owns stream/reconnect loop, no JEH strategy region | §4, §7, §14 | Contract/component tests; OQ 2-8, 13-15 |
| 1 | OFFLINE producer | `internal/input/offline` | Read stored observations and emit one ordered event at a time | Mongo records to `BarEvent` | Internal `BarProducer`, repository port | Owns replay cursor/pacing only | §14, §15 | 4,519-event/order tests; OQ 9-10 |
| 1 | Mongo adapter | `internal/adapter/mongo` | Implement OFFLINE read port | Query results to source records | Internal repository interface | Owns client/cursor, no solver state | §15, §19 | Read-only integration tests; OQ 9, 45-46 |
| 1 | Common admission | `internal/admission` | Validate, identify, and disposition bars | `BarEvent` to admission evidence/accepted event | `BarEvent`, `BarAdmissionEvidence`; internal `Admitter` | Serial disposition before routing | §6, §7; DEP-01 | Duplicate/gap/order vectors; OQ 1, 31-33 |
| 1 | Entity coordinator | `internal/analytical` | Route and serialize per-entity updates | Accepted event to `PhaseEvidence` | Internal solver/state interfaces | Owns entity registry and mutation routing | §9.1; DEP-02 | Isolation/race/order tests; OQ 17, 45 |
| 1 | Solver | `internal/jeh` | Exact approved JEH/Ehlers phase update | Ordered median prices/state to phase result | Internal `PhaseSolver` | State is entity-owned; no global mutation | §8; DEP-03 | Reference/vector tests; OQ 18-19 |
| 1 | Hilbert state | `internal/jeh/hilbert` | Encapsulate bounded recurrence state and exact reference behavior | One price/update to solver internals | Internal only | One instance per entity | §8, §9.1; DEP-02/03 | Recurrence/state tests; OQ 17-19 |
| 1 | Circular phase | `internal/phase` | Normalize and represent explicit phase presence/status | Solver result to `PhaseEvidence` | `PhaseStatus`, `PhaseEvidence` | Entity update context only | §8; DEP-04 | 0/360/init tests; OQ 19 |
| 1 | Evidence identity | `internal/evidence` | Deterministic IDs and causal/source lineage | Typed identity inputs to IDs/envelopes | Governed evidence messages | Stateless/pure where possible | §17 | Repeatability/conflict tests; OQ 1, 12, 46 |
| 1 | Diagnostics | `internal/diagnostics` | Structured logging, metrics, evidence publication ports | Runtime events to sinks | Internal logger/publisher ports | Non-authoritative; bounded async handling | §19 | Failure/redaction tests; OQ 46, 49 |
| 1 | Health | `internal/health` | Aggregate component status without strategy mutation | Component signals to health evidence | `RuntimeHealthEvidence` | Owns health snapshot | §19 | Lifecycle tests; OQ 49 |
| 1 | Reference comparison | `internal/validation/phasecompare` | Compare independent reference and runtime phase by lineage | Reference/runtime evidence to report rows | No runtime dependency on reference generator | Validation-only | §16, §21 | Equivalence suite; OQ 18-19 |

### 8.13 Build and Executable Plan

After authorization, create the module and Buf configuration, generate bindings, compile tests, and build a normal executable. Proposed commands are `buf lint`, `buf generate`, `go test ./...`, and `go build -o bin/dse-jeh-transsat-1.exe ./cmd/dse-jeh-transsat-1`. Exact automation and binary path require approval. Runtime acceptance invokes the built executable with `DSE_JEH_MODE=OFFLINE` or `DSE_JEH_MODE=ONLINE`; it MUST NOT depend on `go run`.

### 8.14 Validation Plan

1. Contract tests for presence, units, identity, status, and compatibility.
2. Admission vectors for malformed, duplicate, conflict, gap, missing, and out-of-order bars.
3. Per-entity ordering, isolation, concurrent entity processing, and Go race detection.
4. Initialization sequences 1 through 63, first possible observable phase at 64, and legitimate zero-degree behavior.
5. Exact solver component tests derived from cited reference artifacts.
6. Independent comparison: same stored bars to the unchanged reference generator and runtime bar-by-bar solver, joined by approved lineage.
7. Repeated deterministic OFFLINE runs over all verified 4,519 observations.
8. Mode-adapter contract tests proving both modes reach identical admission/analytical APIs.
9. ONLINE integration tests only after OQ 2-8 and 13-15 are resolved and the upstream binding is approved.
10. Failure, cancellation, drain, restart, logging, health, and evidence-identity tests under approved policies.

Numerical tolerance is not invented here; OQ 19 must be approved before phase-equivalence acceptance.

### 8.15 Phase 1 Acceptance Gate

Phase 2 MUST NOT begin until a Phase 1 completion/validation report demonstrates and a human approves:

- deterministic `BarEvent` admission and evidence identity;
- correct entity-scoped ordering and no cross-entity state contamination;
- correct 63-observation initialization and first possible output at observation 64;
- explicit unavailable phase and legitimate zero-degree handling;
- bounded per-entity solver state with race-free ownership;
- `PhaseEvidence` identity, source provenance, solver/configuration lineage, and deterministic repeatability;
- numerical phase equivalence against the independent `phase_angle_series_generator` under an approved OQ 19 tolerance;
- deterministic one-event-at-a-time OFFLINE processing of all 4,519 approved observations;
- no analytical difference caused by OFFLINE source mode;
- approved exact `Fin_Feed_Sat_1` ONLINE binding and an implemented ONLINE consumer with approved ordering, duplicate, gap/loss, reconnect, and backpressure behavior;
- validated ONLINE mapping to the common `BarEvent` boundary and successful processing through the same admission and analytical path used by OFFLINE;
- equivalent `PhaseEvidence` for matching ONLINE and OFFLINE admitted bars and order under the approved comparison policy; and
- documented disposition of unresolved recovery, persistence, and runtime-health issues needed for Phase 1 operation.

Gate failure returns work to Phase 1 or to System Design resolution. It cannot be waived by beginning Phase 2 packages.

---

## 9. PHASE 2 - DYNAMIC EXECUTION PIPELINE

### 9.1 System Design Authority

- §5 **Dynamic Execution Pipeline** and its central diagram
- §6 **One BarEvent Lifecycle**
- §9 **Per-Entity and Universe State**
- §10 **Four-Region Circular JEH Rules Engine**
- §11 **Phase Motion and Phase Velocity ($\omega$)**
- §12 **Dynamic Allocation and Strategy Behavior**
- §13 **Decision and External Execution Boundary**
- §17 **Determinism, Mode Equivalence, and Evidence Identity**
- §18 **Conceptual Component Model**
- §19 **Portability, Persistence, and Failure Isolation**
- §20 **Viewer Boundary**
- §21 **Validation Strategy**, items 5-10 and 12
- §22 OQ 20-50 as applicable
- DEP-05 **Phase Motion Analysis** through DEP-11 **ExecutionIntent Generation**

### 9.2 Scope

Phase 2 consumes validated Phase 1 `PhaseEvidence` and first implements the proto-governed Production Eligibility Controller and common Governed Rule Engine. Only typed production-eligible outcomes route to Phase Motion Analyzer, Boundary Crossover Detector, Four-Region JEH Rules Engine, Universe State Coordinator, Candidate Ranking Engine, Strategy Decision Engine, and `ExecutionIntent` Publisher. Blocked results terminate with evidence and activity accounting. The separately bounded initial executor is mock/simulated only and produces `ExecutionEvent` evidence. No broker or live execution is authorized.

### 9.2.1 Production Eligibility and Bar-64/Bar-65 Gate

The controller implements Rule #1 as event-flow control using only its approved component context. Bars 1 through 63 still update JEH mathematics but map to a typed initializing/blocked outcome and cannot reach DEP-05. Bar 64 is the first possible production-eligible phase. Whether motion/crossover requires prior and current production-eligible phases, making bar 65 earliest, remains unresolved and blocks DEP-05/06 implementation; the plan MUST NOT guess.

### 9.2.2 Common Governed Rule Engine

Before any rule-governed component implementation, extend the single proto with descriptive rule identity/version, evaluation status, typed component outcomes, rule-set/configuration identity, and `RuleEvaluationEvidence`. Then add one common facility containing a Rule Registry, `github.com/antonmedv/expr` compiler, compiled-program cache, evaluator, typed-outcome mapper, and evidence producer.

Rules compile once at startup or approved rule-set load/reload, not per bar. Compilation failure is a governed visible failure. Each component builds its own typed context and exposes only approved dynamic variables; there is no unrestricted global `PipelineContext`. Raw expr results have no global business meaning and MUST map through the owning rule to a typed proto outcome before Go orchestration mutates state or routes evidence.

The Production Eligibility, motion eligibility, crossover, strategy region, universe, ranking eligibility, decision, and intent-eligibility contexts remain separated as specified by System Design §13.3. `expr` governs conditions, eligibility, policy, routing, and outcome selection. JEH/Hilbert recurrence and approved circular delta, velocity, ranking, sizing, and identity calculations remain deterministic Go algorithms where appropriate. No expression may call Alpaca, submit an order, mutate authoritative holding/cash/broker state, bypass decision/intent, or create an `ExecutionEvent`.

Activity contracts must separately represent `bars_received`, `bars_admitted`, `bars_rejected`, `bars_initializing`, `bars_phase_eligible`, `phase_motion_evaluations`, `boundary_crossovers`, `hop_on_events`, `hop_off_events`, `strategy_decisions`, `execution_intents`, and `execution_events`, subject to exact proto naming review. Each bar must finish with typed evidence for rejection, initialization block, motion unavailable, no crossover/no action, or the complete decision-to-execution path.

### 9.3 Scientific and Policy Blockers

| Component | Blocking System Design questions | Required decision before implementation |
| --- | --- | --- |
| Phase Motion Analyzer | OQ 20-24, 29-30 | Signed circular $\Delta\phi$, Phase Velocity ($\omega$), wrap/direction, one- vs multi-bar basis, smoothing, acceleration, reverse motion, large jumps |
| Boundary Crossover Detector | OQ 20, 25-30 | Robust 90, 180, 270, and 360/0 crossings under direction, wrap, and jumps |
| Strategy Region / Rules Engine | OQ 25-33 | Approved crossover effects plus missing/duplicate/out-of-order behavior |
| Universe State Coordinator | OQ 34-35, 38-42 | Staleness, asynchronous snapshot semantics, stale expiry, capacity, holdings, capital, sequencing |
| Candidate Ranking Engine | OQ 20-24, 34-38 | Approved motion evidence, candidate timing, score, tie-breaks, validity, staleness |
| Strategy Decision Engine | OQ 36-44 | Ranking, capacity/holdings, liquidation-before-allocation, mock behavior, trailing policy |
| ExecutionIntent Publisher | OQ 42-43, 47 | Sequencing, idempotency, validity, correlation, and exact contract |
| Mock execution/reconciliation | OQ 43, 45, 48 | Outcome semantics, recovery/reconciliation, exact `ExecutionEvent` contract |
| Production-valid motion start | OQ 51 | Decide whether two eligible observations are required and therefore whether bar 65 is earliest |

Blocked components may receive compile-time interfaces and approved contract scaffolding only if separately authorized; their mathematical or policy behavior MUST NOT be guessed.

### 9.4 Phase Motion

DEP-05 consumes consecutive valid circular phase observations and emits versioned `PhaseMotionEvidence` in degrees per bar. Phase Velocity ($\omega$) is the intended ranking input, but this plan does not define its mathematics. The evidence must represent validity/confidence and abnormal conditions. It cannot use naive positive modulo or conflate velocity with acceleration. Implementation is BLOCKED by OQ 20-24 and 29-30.

### 9.5 Boundary Crossover

DEP-06 consumes circular phase and approved motion context and emits crossover evidence distinct from persistent strategy-region membership. It must represent `HOP-ON` only as a valid 270° crossover into `ALLOCATE`, `HOP-OFF` only as a valid 90° crossover into `LIQUIDATE`, and the 0° transition from `ALLOCATE` into `HOLD & TRAIL` without inventing a name. It must handle wrap, direction, reverse movement, exact boundaries, and large jumps under approved rules. Implementation is BLOCKED by OQ 20 and 25-30.

### 9.6 JEH Strategy-Region Rules

DEP-07 implements the System Design §10 persistent regions/actions and only approved transitions: `DISREGARD`, `ALLOCATE`, `HOLD & TRAIL`, and `LIQUIDATE`, plus explicit initialization/unavailability. Machine-safe contracts may spell `HOLD & TRAIL` as `HOLD_AND_TRAIL`. `HOP-ON` and `HOP-OFF` are crossover events, never persistent states. Same-region persistence creates no false crossover event. Final transition behavior is BLOCKED by OQ 25-33.

### 9.7 Universe State

DEP-08 owns the latest valid per-entity strategy views, holdings, candidates, unavailable entities, capacity/capital representation, freshness, and pending decisions/intents. It must not imply synchronized entity updates. Snapshot, staleness, capacity, holdings, and sequencing are BLOCKED by OQ 34-35 and 38-42.

### 9.8 Candidate Ranking

DEP-09 ranks the approved current `ALLOCATE` candidate set using Phase Velocity ($\omega$) and only other approved evidence under a versioned policy identity. Phase Velocity and score mathematics, tie-breaking, timing, and stale expiry are not defined by this plan. Implementation is BLOCKED by OQ 20-24 and 34-38.

### 9.9 Strategy Decision

DEP-10 evaluates `LIQUIDATE` first, `ALLOCATE`/reallocation when governed capacity is available, `HOLD & TRAIL`, `DISREGARD`, or no action under approved policy. A valid `HOP-OFF` event provides crossover evidence for the `LIQUIDATE` context; it does not itself claim an execution outcome. Freed capital is available for governed reallocation, and exact trailing and capital mathematics remain unresolved. The decision records causal region/crossover/ranking evidence and never claims execution. Implementation is BLOCKED by OQ 36-44.

### 9.10 ExecutionIntent

DEP-11 emits a governed, deterministic, idempotent `ExecutionIntent` only when eligibility and policy permit. Exact fields, expiry, cancellation, correlation, and sequencing require OQ 42 and 47 resolution.

### 9.11 Mock Execution Boundary

Initial proving uses an external in-process or process-separated mock adapter selected by approved boundary design. It receives intents without holding internal strategy locks and emits attributable `ExecutionEvent` outcomes for reconciliation. No live adapter is in scope. OQ 43, 45, and 48 block final behavior and contract.

The approved interface is substitutable: `MockExecutor` is implemented first; a future `AlpacaPaperExecutor` requires separate approval and consumes/produces the same proto-governed intent/event meanings. The rule engine and analytical engine do not depend on Alpaca. No live-money endpoint or order is permitted.

### 9.12 Concurrency and State Ownership

- The per-entity strategy owner serializes that entity's phase-motion, crossover, and state transition.
- A single universe coordinator or equivalent serialized command loop owns mutable universe-wide holdings, candidates, capacity, rankings, and pending decisions.
- Entity results enter the universe coordinator as immutable evidence; entity workers do not mutate universe state directly.
- Candidate-set evaluation and decisions are serialized against a defined universe version to prevent races and inconsistent allocations.
- Publication and external execution occur after committed internal state transitions and outside entity/universe locks.
- Execution latency, rejection, or disconnect cannot block unrelated analytical updates; reconciliation returns as explicit evidence.
- Shutdown drains or explicitly cancels queued entity updates, universe commands, intents, and outcomes under approved OQ 45/49 policy.

### 9.13 Go Module and File Plan

| Phase | Package / file group | Proposed path | Responsibility | Inputs / outputs | Proto / internal interfaces | State and concurrency | Authority | Validation / blockers |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 2 | Motion | `internal/motion` | DEP-05 signed circular motion evidence | `PhaseEvidence` to `PhaseMotionEvidence` | Internal `MotionAnalyzer` | Entity-owned update context | §11; DEP-05 | Vector/property tests; OQ 20-24, 29-30 |
| 2 | Production eligibility | `internal/eligibility` | Explicit Rule #1 event-flow gate | Phase/admission evidence to typed eligibility outcome/evidence | Proto-governed input/outcome; internal controller | Entity update context | §8.2 | Bars 1-65, continuity, invalid-phase tests; OQ 51 affects downstream motion only |
| 2 | Governed rule engine | `internal/rules` | Registry, expr compilation/cache/evaluation, typed outcome mapping and evidence | Component context/rule identity to typed outcome and `RuleEvaluationEvidence` | Proto-governed rule contracts; internal evaluator | Immutable compiled registry after startup unless approved reload | §13.1-13.3 | Compilation, authorization, mapping, evidence tests |
| 2 | Crossover | `internal/crossover` | DEP-06 circular boundary detection | Phase/motion to crossover evidence | Internal `CrossoverDetector` | Stateless or entity context | §10; DEP-06 | Boundary/direction tests; OQ 25-30 |
| 2 | Rules | `internal/strategy` | DEP-07 four-region membership/transitions with crossover events kept distinct | Phase/crossover to entity strategy state | Internal `RulesEngine` | Entity strategy state | §10; DEP-07 | Transition/persistence tests; OQ 25-33 |
| 2 | Universe | `internal/universe` | DEP-08 latest views, holdings, candidates, capacity, pending work | Entity evidence/outcomes to universe versions | Internal command API | Sole owner of mutable universe state | §9.2; DEP-08 | Concurrency/staleness tests; OQ 34-35, 38-42 |
| 2 | Ranking | `internal/ranking` | DEP-09 approved candidate ordering | Universe candidates to ranking evidence | Internal `Ranker` | Pure over immutable snapshot where possible | §11, §12; DEP-09 | Score/tie tests; OQ 20-24, 34-38 |
| 2 | Decisions | `internal/decision` | DEP-10 strategy decisions | State/ranking/capacity to decision | Internal `DecisionEngine` | Runs against serialized universe version | §12, §13; DEP-10 | Workflow tests; OQ 36-44 |
| 2 | Intent publication | `internal/intent` | DEP-11 intent identity and publication | Decision to `ExecutionIntent` | Publisher port; governed message | No external call under strategy locks | §13; DEP-11 | Idempotency/failure tests; OQ 42, 47 |
| 2 | Mock executor | `internal/execution/mock` | Proving-only outcomes and reconciliation | Intent to `ExecutionEvent` | Execution adapter port | Independent worker; deterministic configured behavior | §13 | Outcome/reconciliation tests; OQ 43, 45, 48 |
| 2 | Pipeline orchestration | `internal/pipeline` | Connect DEP-05 through DEP-11 after Phase 1 | Phase evidence to intent/evidence | Internal component interfaces | Coordinates without owning adapter transport | §5, §6, §18 | End-to-end deterministic tests; all Phase 2 blockers |
| 2 | Diagnostics extension | `internal/diagnostics` | Add motion/crossover/state/rank/decision/intent/outcome evidence | Pipeline evidence to approved sinks | Governed messages/publisher ports | Non-causal observer | §17, §19, §20 | Lineage/failure tests; OQ 46-49 |
| 2 | Health extension | `internal/health` | Add universe, ranking, intent, executor, reconciliation status | Component states to health evidence | Existing Phase 1 health contract extension | Health owner only | §19 | Degradation tests; OQ 43, 45, 49 |
| Complete app | Activity/accountability | `internal/telemetry` and pipeline orchestration | Distinct counters and terminal outcome for every received bar | All governed evidence to runtime activity/state | Proto-defined telemetry and processing outcomes | Non-authoritative aggregation over causal evidence | §19.1 | Zero-input, counter-separation, no-silent-drop E2E tests |

### 9.14 Validation Plan

1. Approved signed-motion vectors including 10 to 350, wraps, reverse movement, ties, gaps, and large jumps.
2. Every exact zone boundary in both approved directions and all allowed jump paths.
3. Initial classification, same-region persistence, true exit/crossover/entry, and no repeated `HOP-ON` or `HOP-OFF` event while remaining in a region.
4. Asynchronous entity updates, universe versioning, staleness, candidate timing, ranking ties, capacity, and holdings under approved policies.
5. `LIQUIDATE`-before-`ALLOCATE`, `HOLD & TRAIL`, `DISREGARD`, no-action, and intent idempotency.
6. Mock acceptance, rejection, failure, delayed outcome, duplicate outcome, and reconciliation.
7. Failure isolation and race testing with slow publication/execution.
8. Repeated complete-pipeline deterministic OFFLINE runs over the approved dataset.
9. Full ONLINE/OFFLINE semantic equivalence for matching admitted bars and order.
10. End-to-end causal identity from `BarEvent` through mock `ExecutionEvent`.
11. Every rejection, initialization block, motion-unavailable result, no-crossover/no-action result, and execution path ends with typed attributable evidence.
12. Persistent zero-input operation reports healthy liveness with flat bar activity until explicit stop or a separate health failure.

### 9.15 Phase 2 Acceptance Gate

Completion requires approved blocker resolutions, proto review, passing unit/component/integration/race tests, deterministic complete-pipeline replay, validated crossover/state/ranking behavior, mock execution reconciliation, full evidence lineage, operational health/failure evidence, and a human-approved Phase 2 completion report. Passing this gate does not authorize live execution or production deployment.

---

## 10. Design-to-Implementation Traceability Matrix

| Phase | Implementation artifact | Type | System Design authority | DEP | Requirement | Proto artifact | Go component | Validation artifact | OQ / blocker | Status |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 1 | Mode selection | Runtime | §14 Online and Offline Stream Modes | - | Select exactly one producer above common admission | `RuntimeMode` | `internal/config`, `internal/runtime` | Startup matrix | OQ 49 | PLANNED |
| 1 | ONLINE bar input | Adapter | §4, §7, §14 | DEP-01 boundary | Map approved upstream stream without leaking transport types | Upstream bar contract; DSE `BarEvent` | `internal/input/online` | Contract/integration suite | OQ 2-8, 13-15 | BLOCKED |
| 1 | OFFLINE stream | Adapter | §14, §15 | DEP-01 boundary | Emit 4,519 stored observations one at a time in approved order | `BarEvent`, `SourceProvenance` | `internal/input/offline`, `internal/adapter/mongo` | Replay count/order report | OQ 9-10 | BLOCKED |
| 1 | Bar admission | Go component/evidence | §6, §7, §17 | DEP-01 | Deterministic validation and disposition before mutation | `BarEvent`, `BarAdmissionEvidence`, `BarAdmissionStatus` | `internal/admission` | Admission vectors | OQ 1, 31-33 | BLOCKED |
| 1 | Entity analytical state | Go component | §9.1 | DEP-02 | Isolated ordered bounded state | None beyond evidence contracts | `internal/analytical`, `internal/jeh/hilbert` | Race/isolation/order suite | OQ 17, 45 | BLOCKED |
| 1 | JEH Phase Solver | Go component | §8, §16 | DEP-03 | Reproduce approved reference mathematics from ordered bars | `SolverIdentity`, `PhaseEvidence` | `internal/jeh` | Independent phase-equivalence report | OQ 18-19 | BLOCKED |
| 1 | Circular phase state | Go component/evidence | §8 | DEP-04 | Explicit initialization/value presence and $[0,360)$ | `PhaseStatus`, `PhaseEvidence` | `internal/phase` | Initialization/zero/boundary suite | OQ 19 | PLANNED |
| 1 | Evidence identity | Go component | §17 | DEP-01/03/04 | Deterministic causal and source identity | `SourceProvenance`, Phase 1 evidence | `internal/evidence` | Repeatability report | OQ 1, 12, 46 | BLOCKED |
| 1 | Health/lifecycle | Runtime/evidence | §19 | - | Explicit readiness/degradation/drain | `RuntimeHealthEvidence` | `internal/runtime`, `internal/health` | Lifecycle/failure report | OQ 49 | BLOCKED |
| 2 | Production eligibility | Rule component/evidence | §8.2 | Gate before DEP-05 | Rule #1 typed block/eligible routing | Eligibility context/outcome/evidence; rule evidence | `internal/eligibility`, `internal/rules` | Bars 1-65 and continuity report | OQ 51 for first motion | NOT IMPLEMENTED |
| 2 | Common rule facility | Shared component | §13.1-13.3 | Cross-cutting | Registry, compiled expr, scoped contexts, typed mappings and evidence | Rule identity/outcome/evidence contracts | `internal/rules` | Compile/map/scope/evidence suite | Exact proto names | NOT IMPLEMENTED |
| 2 | Phase motion | Go component/evidence | §11 | DEP-05 | Approved signed motion in degrees/bar | `PhaseMotionEvidence` | `internal/motion` | Motion vector report | OQ 20-24, 29-30 | BLOCKED |
| 2 | Boundary crossover | Go component/evidence | §10 | DEP-06 | Circular crossover distinct from membership | `BoundaryCrossoverEvidence` | `internal/crossover` | Boundary report | OQ 25-30 | BLOCKED |
| 2 | Four-region rules | Go component/evidence | §10 | DEP-07 | Canonical persistent strategy regions and genuine crossover events | `JehStrategyRegion`, `EntityStrategyState` | `internal/strategy` | Region/crossover report | OQ 25-33 | BLOCKED |
| 2 | Universe state | Go component/evidence | §9.2 | DEP-08 | Latest asynchronous views, holdings, capacity, pending work | `UniverseStateEvidence` | `internal/universe` | Staleness/concurrency report | OQ 34-35, 38-42 | BLOCKED |
| 2 | Candidate ranking | Go component/evidence | §11, §12 | DEP-09 | Rank approved current `ALLOCATE` candidates using Phase Velocity ($\omega$) under a versioned policy | `CandidateRankingEvidence` | `internal/ranking` | Ranking validation report | OQ 20-24, 34-38 | BLOCKED |
| 2 | Strategy decision | Go component/evidence | §12, §13 | DEP-10 | `LIQUIDATE`/reallocate/`HOLD & TRAIL`/`DISREGARD` without claiming execution | `DecisionType`, `StrategyDecision` | `internal/decision` | Decision workflow report | OQ 36-44 | BLOCKED |
| 2 | Execution intent | Contract/component | §13 | DEP-11 | Deterministic governed request with causal identity | `ExecutionIntent` | `internal/intent` | Idempotency/publication report | OQ 42, 47 | BLOCKED |
| 2 | Mock execution | Adapter/evidence | §13 | - | Simulated outcome outside strategy locks | `ExecutionEvent`, `ExecutionOutcomeStatus` | `internal/execution/mock` | Mock reconciliation report | OQ 43, 45, 48 | BLOCKED |
| 2 | Complete pipeline | Integration | §5, §6, §17, §21 | DEP-05-11 | Deterministic source-mode-independent behavior | All approved Phase 2 evidence | `internal/pipeline` | Dynamic Execution Pipeline and mode-equivalence reports | OQ 20-50 as applicable | BLOCKED |

Existing implementation status is authoritative as follows: DEP-01, DEP-02, DEP-03, and DEP-04 are implemented and validated; Rule #1 behavior is embedded in current analytical status handling; the explicit Production Eligibility Controller, DEP-05 through DEP-11, common rule engine, complete proto vocabulary, complete lifecycle/telemetry, stop mechanism, and complete execution adapter/event path are not implemented.

### 10.1 Reconciled Implementation Invariants

- The complete target is one independently startable and stoppable persistent built application; startup and stop are governed, `go run` is prohibited, and zero input is a valid running state.
- **Mode convergence:** ONLINE and exactly one selected OFFLINE `collection_run_id` converge at the same `BarEvent` boundary and use one E2E runtime path.
- The single proto is authoritative for governed internal and external vocabulary while in-process implementation does not require artificial RPC hops.
- Rule #1 processes bars 1 through 63 as `INITIALIZING`; bar 64 is the first possible eligible phase; mathematical calculability is not production eligibility; and whether bar 65 is earliest for production-valid motion/crossover remains unresolved.
- `expr` programs are identified/versioned and compiled once per startup or governed load; a raw boolean has no global business meaning and maps to a rule-specific typed proto outcome before Go routing.
- **Every-bar accountability and named metrics:** every received bar ends with attributable typed evidence. Activity separately tracks `bars_received`, `bars_admitted`, `bars_rejected`, `bars_initializing`, `bars_phase_eligible`, `phase_motion_evaluations`, `boundary_crossovers`, `hop_on_events`, `hop_off_events`, `strategy_decisions`, `execution_intents`, and `execution_events`.
- **Event semantic separation:** `HOP_ON` and `HOP_OFF` are crossover events. Strategy region, `StrategyDecision`, `ExecutionIntent`, and `ExecutionEvent` are separate meanings and implementation stages.
- Signed circular delta, Phase Velocity basis/window/smoothing/acceleration, reverse motion, abnormal jumps, 0/90/180/270 crossover mathematics, ranking formula, stale-candidate policy, capacity/holding policy, capital representation, trailing-stop mathematics, quantity sizing, execution reconciliation, and bar-64/bar-65 motion eligibility remain intentionally unresolved.
- This documentation task authorizes no changes to proto, generated code, Go implementation, startup/stop scripts, tests, or broker integration.

The matrix is maintained during authorized work so every contract and component has a direct `System Design -> proto -> Go -> test -> acceptance evidence` chain.

---

## 11. Build and Generated-Code Governance

1. `DSE_JEH` is proposed as one independent Go module; no new workspace-wide build system is introduced.
2. The single proto and Buf configuration follow the existing `Fin_FeedSat_1` source/output convention unless human review changes it.
3. `buf lint` and a breaking-change policy run before generation; `buf generate` is the sole generation route.
4. Generated Go and gRPC output goes only to `DSE_JEH/gen/dse_jeh/v1`. Whether it is committed must be confirmed from repository policy before first generation; generated files are never hand-edited.
5. The normal entrypoint is `cmd/dse-jeh-transsat-1`; `go build` produces a versioned executable under `bin` or another approved artifact directory.
6. Build metadata should include source revision, build time where reproducibility policy permits, proto/schema version, solver version, strategy version, and configuration identity.
7. Configuration is validated before dependencies start. Startup orders logging/health, selected input, admission, and analytical/pipeline components; shutdown stops intake and drains in causal order.
8. Unit/component/integration testing uses normal `go test ./...`; race-sensitive packages additionally use the approved Go race test command.
9. Deterministic validation invokes the built executable and governed harnesses. Runtime implementation and acceptance MUST NOT use ad-hoc `go run`.

---

## 12. Logging, Evidence, and Reports

Planned report location follows repository documentation convention: `DSE_JEH/docs/implementations` for completion reports and `DSE_JEH/docs/validation` for validation reports, subject to human approval before directory creation.

| Phase | Planned output | Minimum content |
| --- | --- | --- |
| 1 | Implementation completion report | Approved contracts, packages, build identity, tests, deviations, unresolved risks |
| 1 | Phase-equivalence report | Reference/runtime lineage join, tolerance decision, mismatch detail, deterministic summary |
| 1 | Deterministic replay report | 4,519-event admission/order/count, phase digest/evidence identity, repeated-run result |
| 1 | ONLINE/OFFLINE input-equivalence report | Matching admitted inputs, provenance differences, and analytical equivalence for both implemented modes |
| 1 | Structured runtime evidence | Startup/mode, admission, solver/phase, health, errors, shutdown, versions, causal IDs |
| 2 | Implementation completion report | DEP-05 through DEP-11 artifacts, approved decisions, tests, deviations |
| 2 | Crossover/state-machine validation report | Boundary vectors, direction/jumps, membership/persistence/transition evidence |
| 2 | Ranking validation report | Candidate snapshots, stale handling, score/ties, policy identity, expected results |
| 2 | Deterministic pipeline report | State/universe/ranking/decision/intent digests and repeatability |
| 2 | Mock execution/reconciliation report | Intent/outcome IDs, acceptance/rejection/failure/recovery cases |
| 2 | Full ONLINE/OFFLINE equivalence report | Complete-pipeline comparison under identical admitted bars and configuration |

Reports are evidence for human review, not self-authorization.

---

## 13. Implementation Sequence

### 13.1 Phase 1 sequence

1. Resolve and approve module, Buf, generated-code, executable, and report-location conventions.
2. Resolve Phase 1 contract blockers and approve the Phase 1 proto artifact plan.
3. Create the single authoritative DSE_JEH proto and Buf configuration.
4. Generate Go bindings/stubs and enforce lint/compatibility checks.
5. Establish the runtime entrypoint, immutable configuration, lifecycle, health, and normal build.
6. Implement common `BarEvent` mapping, identity, and DEP-01 admission.
7. Implement the OFFLINE stream producer and read-only Mongo adapter under the approved order policy.
8. Implement per-entity analytical ownership and bounded Hilbert state.
9. Implement the approved JEH/Ehlers solver from cited reference artifacts.
10. Implement circular state and deterministic `PhaseEvidence`.
11. Implement diagnostics, structured evidence, and health behavior.
12. Validate independent numerical phase equivalence under the approved tolerance.
13. Validate deterministic one-at-a-time processing of the 4,519-observation OFFLINE source.
14. Bind and validate the ONLINE consumer only after exact upstream contract approval.
15. Run the formal Phase 1 Acceptance Gate.
16. Produce Phase 1 completion and validation reports.
17. Obtain explicit human authorization for Phase 2.

OFFLINE-first and ONLINE-later describe implementation order within Phase 1 only. Neither mode is optional at the Phase 1 Acceptance Gate.

### 13.2 Phase 2 sequence

1. Resolve and approve all applicable mathematical, crossover, universe, ranking, decision, and execution blockers.
2. Resolve the bar-64/bar-65 motion/crossover decision and approve component-scoped variables and typed outcomes.
3. Extend the same authoritative proto with complete lifecycle, activity, eligibility, rule, DEP-05 through DEP-11, and execution-boundary vocabulary.
4. Regenerate Go bindings/stubs and run compatibility checks.
5. Implement lifecycle/health, governed start/stop, zero-input operation, and per-bar accountability.
6. Implement the common Rule Registry, expr compiler/cache/evaluator, typed outcome mapper, and rule evidence.
7. Implement the explicit Production Eligibility Controller.
8. Implement DEP-05 Phase Motion Analyzer.
9. Implement DEP-06 Boundary Crossover Detector.
10. Implement DEP-07 Four-Region JEH Rules Engine with distinct crossover-event evidence.
11. Implement DEP-08 Universe State Coordinator.
12. Implement DEP-09 Candidate Ranking Engine.
13. Implement DEP-10 Strategy Decision Engine.
14. Implement DEP-11 `ExecutionIntent` generation/publication.
15. Implement the mock external execution adapter and `ExecutionEvent` reconciliation.
16. Validate deterministic complete-pipeline behavior, full bar accountability, and ONLINE/OFFLINE equivalence.
17. Produce completion and validation reports for human review.

---

## 14. Blocking Decisions and Open Questions

The following gates summarize, but do not replace, System Design §22:

| Gate | System Design OQ | Required resolution |
| --- | --- | --- |
| Common input contract | 1, 12, 16 | OQ 16 resolves finite `high`/`low` as solver inputs; OQ 1/12 still block canonical entity, provenance, source-observation identity, optionality, and timestamp representation |
| ONLINE binding/delivery | 2-8, 13-15 | Exact upstream binding, ordering, duplicates, gaps/loss, reconnect, backpressure, recovery and warm-up |
| OFFLINE stream | 9-10 | OQ 9 requires a versioned deterministic cross-entity replay-order policy; OQ 10 pacing is deferred |
| Mode equivalence | 11 | RESOLVED by comparison over matching admitted bars/order, identity, versions, configuration, and initial state |
| Solver/state | 17-19 | OQ 17 checkpoint representation is deferred; OQ 18 equivalence basis is resolved; OQ 19 production numeric tolerance remains open |
| Motion | 20-24, 29-30 | Signed delta, Phase Velocity ($\omega$) basis, smoothing, acceleration, direction, large jumps |
| Crossovers | 25-30 | Exact 90° `HOP-OFF`, 270° `HOP-ON`, 180°, and unnamed 360°/0° behavior under direction and jumps |
| Admission continuity | 31-33 | Missing, duplicate/conflicting, and out-of-order behavior |
| Universe/ranking | 34-41 | Staleness, snapshots, ranking/ties, candidate timing/expiry, capacity, holdings, capital |
| Decision/execution | 42-44, 47-48 | Sequencing, mock semantics, trailing policy, intent/event contracts |
| Operations | 45-46, 49 | Exact recovery/persistence and lifecycle contract shapes remain unresolved; complete runtime lifecycle, health, zero-input liveness, activity, stop, and evidence responsibilities are mandatory and proto-first |
| Production eligibility | 51 | Decide whether bar 65 is earliest possible production-valid motion/crossover; do not guess |
| Promotion | 50 | DEFERRED; criteria from OFFLINE proving to ONLINE testing remain required before promotion |

System Design §22.3 records the explicit human direction that passed the Phase 1 implementation gate. The implementation uses normalized symbol identity, typed mode-specific provenance, versioned deterministic OFFLINE ordering, conservative admission with mutation only for `ADMITTED`, and the current Fin `StreamBars` contract with bounded reconnect. Upstream ONLINE loss detection, resumable position, and live market validation remain documented limitations rather than claims. Phase 2 gates remain unchanged.

---

## 15. Non-Goals

This reconciliation does not authorize or perform:

- Go implementation, proto modification, code generation, service/RPC implementation, build, or new tests;
- changes to existing ONLINE/OFFLINE, admission, JEH, circular-state, runtime, or script implementation;
- Production Eligibility Controller, rule engine, or DEP-05 through DEP-11 implementation;
- MongoDB mutation or execution of the 4,519-bar replay/phase comparison;
- modification of `Fin_Feed_Sat_1`, `phase_angle_series_generator`, or `DSE_JEH_provers`;
- startup/stop-script changes, ad-hoc runtime execution, broker integration, live execution, or viewer modification;
- repository restructuring, staging, commit, or push; or
- broader architecture outside `DSE_JEH_TransSat_1`.

---

## 16. Change Log

| Version | Date | Change |
| --- | --- | --- |
| V0.1 | 2026-09-12 | Initial implementation plan subordinate to `DSE_JEH_TransSat_1` System Design V0.1. Establishes exactly two implementation phases, one authoritative proto strategy, proto-first workflow, Go package and concurrency ownership plans, independent JEH phase equivalence, formal phase gates, traceability, build governance, and planned evidence reports. |
| V0.1 terminology and gate refinement | 2026-09-12 | Aligned Phase 2 planning with canonical `DISREGARD`, `ALLOCATE`, `HOLD & TRAIL`/`HOLD_AND_TRAIL`, and `LIQUIDATE` persistent regions/actions; treated `HOP-ON` and `HOP-OFF` only as 270° and 90° crossover events; retained Phase Velocity ($\omega$) as the intended ranking input without defining unresolved mathematics; and required implemented, validated ONLINE and OFFLINE paths before the Phase 1 Acceptance Gate can approve Phase 2. |
| V0.1 Phase 1 contract review | 2026-09-12 | Synchronized the plan to the System Design decision register, approved the planned single-proto path/package without authorizing creation, removed deferred runtime-health and operations declarations from the initial proto, and recorded the failed Stage A gate pending identity, provenance, admission, replay-order, and ONLINE continuity decisions. |
| V0.1 Phase 1 implementation | 2026-09-12 | Implemented the single Phase 1 proto and generated binding, Go runtime, Mongo OFFLINE producer, Fin gRPC ONLINE consumer, common admission, entity-isolated JEH solver, PhaseEvidence output, normal executable, deterministic replay and reference comparison, generated-contract integration tests, and validation reports. Phase 2 remains unimplemented. |
| V0.1 complete-application reconciliation | 2026-09-13 | Revised in place for the complete persistent Transformation Satellite: authoritative single-proto coverage for internal and external governed meanings; explicit Production Eligibility Controller; common outcome-based `expr` rule facility with scoped contexts, typed outcomes and evidence; lifecycle/zero-input/activity/accountability work; truthful implementation status; execution separation; and preserved unresolved mathematics/policy including bar 64 versus bar 65. |

---

## 17. Authorization Statement

This implementation plan is **APPROVED** as the V0.1 plan for the complete application architecture.

**DEP-01 through DEP-04 are implemented and validated by existing reports. The explicit Production Eligibility Controller, DEP-05 through DEP-11, common rule engine, complete proto vocabulary, and complete lifecycle/telemetry/execution path are not implemented.**

This in-place reconciliation authorizes no proto, generated-code, Go, startup/stop-script, broker-integration, deployment, or trading changes. Future work follows the proto-first gates in this plan and cannot invent unresolved mathematics or financial policy.