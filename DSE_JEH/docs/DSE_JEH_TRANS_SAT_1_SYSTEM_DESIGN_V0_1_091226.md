# DSE_JEH_TransSat_1 System Design V0.1

| Document control | Value |
| --- | --- |
| Filename | `DSE_JEH_TRANS_SAT_1_SYSTEM_DESIGN_V0_1_091226.md` |
| Date | 2026-09-12 |
| Version | V0.1 |
| Status | APPROVED |
| Implementation status | PHASE 1 IMPLEMENTED - PENDING HUMAN REVIEW |
| Architectural identity | `DSE_JEH_TransSat_1` |
| Physical documentation location | `DSE_JEH/docs` |
| Relationship to `DSE_JEH_provers` | Experimental proving and regression apparatus; not a TransSat and not operational runtime code |

**Purpose.** Define the proposed architecture for the first concrete runtime instance of the portable John Ehlers-Hilbert Decision Strategy Engine family. This pre-approval revision establishes the **Dynamic Execution Pipeline** as the event-driven inner decision engine. It creates no implementation, package, protobuf, deployment, persistence, or execution authority.

Normative terms `MUST`, `MUST NOT`, `SHOULD`, and `MAY` express design requirements. Strategy interpretations are hypotheses until separately validated; deterministic behavior alone does not establish scientific validity or efficacy.

---

## 1. Naming and Architectural Identity

Strategy runtime instances use `DSE_<StrategyFamily>_TransSat_<Instance>`.

- `DSE_JEH` is the reusable John Ehlers-Hilbert mathematical strategy family.
- `DSE_JEH_TransSat_1` is the first concrete designed runtime instance.
- `DSE_JEH_provers` is independent experimental apparatus, not a TransSat.
- The physical `DSE_JEH` folder is not part of runtime identity.

An instance may process one or many entities. Instance number does not imply one instance per entity. The reusable design MUST NOT embed an instance number, fixed entity population, Finance-only vocabulary, or repository path.

The JEH family concerns phase, angle, circular state, motion, boundary crossing, and strategy interpretation. It MUST remain portable across suitable domains, while each source domain retains authority over the meaning and validity of its source observations and actions. Mathematical portability does not imply semantic equivalence.

---

## 2. Executive Summary

`DSE_JEH_TransSat_1` V0.1 has exactly two bar-stream input modes: **ONLINE** and **OFFLINE**. The host selects one mode at startup through the proposed `DSE_JEH_MODE` environment variable. ONLINE consumes the `Fin_Feed_Sat_1` gRPC bar stream through an internal gRPC bar consumer. OFFLINE uses an internal stream producer to read the verified 4,519 observations in `bar_sequence_db.bar_sequence` and emit them one bar at a time in an approved deterministic replay order.

Both modes map their source observations to the same logical `BarEvent` and converge at the same admission boundary. Whether operating ONLINE or OFFLINE, `DSE_JEH_TransSat_1` receives an ordered stream of bar events through a common internal bar-event boundary. Source selection changes the producer and transport/input component, not JEH analytical or Dynamic Execution Pipeline behavior.

For every admitted `BarEvent`, `DSE_JEH_TransSat_1` advances that entity's ordered analytical state, runs the JEH/Ehlers phase solver, updates circular phase state from $\phi[n-1]$ to $\phi[n]$, derives phase-motion evidence, detects relevant boundary crossovers, advances the four-region rules engine, updates the latest universe state, evaluates `DISREGARD`, `ALLOCATE`, `HOLD & TRAIL`, and `LIQUIDATE` behavior, and produces a strategy decision and `ExecutionIntent` where applicable.

This event-driven loop is the **Dynamic Execution Pipeline**. Actual execution remains outside it in an external execution/action adapter, and only an `ExecutionEvent` can describe an execution outcome.

`PhaseEvidence` remains important, but it is internal analytical evidence produced by the in-boundary phase solver in both modes. The solver and strategy rules are logically distinct and independently testable components inside the V0.1 runtime. No separate offline or backtest strategy implementation is permitted.

---

## 3. Scope, Authority, and Principles

### 3.1 In scope

- ONLINE and OFFLINE bar-stream inputs converging on source-independent `BarEvent` admission
- Startup mode selection through proposed `DSE_JEH_MODE=ONLINE|OFFLINE`
- Internal `Fin_Feed_Sat_1` gRPC bar consumption and internal offline stream production as transport/input boundaries
- Bounded per-entity ordered bar and phase-solver state
- In-runtime JEH/Ehlers phase calculation
- Circular phase, phase motion, crossover, and four-region strategy processing
- Universe candidate/holding state and dynamic candidate comparison
- Strategy decisions, `ExecutionIntent`, and mock `ExecutionEvent` proving boundaries
- Deterministic one-event-at-a-time replay and analytical phase equivalence
- Evidence lineage, versioning, failure isolation, recovery questions, and observational viewing

### 3.2 Governing principles

1. `BarEvent` is the primary V0.1 runtime input; `PhaseEvidence` is internal on that path.
2. Each accepted bar advances only its entity's trajectory.
3. The phase solver and strategy engine are separate internal responsibilities.
4. Initialization, phase observability, strategy eligibility, decision, intent, and execution are distinct.
5. Zero degrees is legitimate and MUST NOT be an initialization sentinel.
6. Strategy-region membership and boundary crossover events are distinct.
7. The universe view uses each entity's latest valid state; it is not a synchronized frame.
8. ONLINE and OFFLINE input MUST feed the same engine logic after common `BarEvent` admission.
9. Persistence, transport, viewing, and execution remain adapters around the mathematical and strategy core.
10. Solver, strategy, ranking, configuration, and evidence identities are versioned and attributable.

## 4. Overall System Context

```mermaid
flowchart TD
  subgraph TS[DSE_JEH_TransSat_1]
    subgraph ON[ONLINE]
      FIN[Fin_Feed_Sat_1]
      GRPC[Internal gRPC Bar Consumer]
      FIN -->|gRPC Bar Stream| GRPC
    end
    subgraph OFF[OFFLINE]
      DB[(bar_sequence_db.bar_sequence<br/>4,519 verified observations)]
      PROD[Internal Offline Stream Producer]
      DB --> PROD
    end
    BAR[Common BarEvent Admission]
    DEP[Dynamic Execution Pipeline]
    INTENT[ExecutionIntent]
    GRPC -->|map/admit| BAR
    PROD -->|one stored bar at a time| BAR
    BAR --> DEP --> INTENT
  end
  EXT[External Execution Adapter]
  INTENT --> EXT
```

The mode distinction ends at `BarEvent`. The TransSat owns phase calculation and JEH strategy truth for admitted bars. The ONLINE consumer and OFFLINE producer own source access, mapping, and provenance. The execution adapter owns domain-specific action attempts and outcomes.

---

## 5. Dynamic Execution Pipeline

The **Dynamic Execution Pipeline** is the event-driven inner decision engine of `DSE_JEH_TransSat_1`. It continuously reacts to accepted bar events. It is not a batch phase classifier, CSV processor, viewer, precomputed-`PhaseEvidence` consumer, polling process, broker adapter, or downstream-only execution stage.

```mermaid
flowchart TD
  B[BAR EVENT n]
  O[DEP-02 Per-Entity Ordered Bar State]
  S[DEP-03 JEH / Ehlers Phase Solver]
  C[DEP-04 Circular Phase State<br/>phi n-1 to phi n]
  X[DEP-06 Boundary Crossover Detector]
  M[DEP-05 Phase Motion / Phase Velocity omega]
  R[DEP-07 STRATEGY REGION / RULES ENGINE]
  D0[DISREGARD<br/>180° to 270°]
  HO[HOP-ON crossover event<br/>valid 270° boundary crossing]
  D1[ALLOCATE<br/>270° to 360°<br/>rank candidates by omega]
  Z[0° boundary crossover]
  D2[HOLD & TRAIL<br/>0° to 90°<br/>dynamic trailing behavior]
  HF[HOP-OFF crossover event<br/>valid 90° boundary crossing]
  D3[LIQUIDATE<br/>90° to 180°<br/>freed capital available for reallocation]
  U[DEP-08 Universe Candidate / Holding State]
  K[DEP-09 Candidate Comparison / Ranking Policy]
  D[DEP-10 Dynamic Strategy Decision]
  I[DEP-11 ExecutionIntent]
  A[External Execution Adapter]
  E[ExecutionEvent]
  B --> O --> S --> C
  C --> X
  C --> M
  X --> R
  M --> R
  R --> D0
  R --> HO --> D1
  R --> Z --> D2
  R --> HF --> D3
  D0 --> U
  D1 --> U
  D2 --> U
  D3 --> U
  U --> K --> D --> I --> A --> E
```

### 5.1 Design-local responsibilities

| ID | Logical responsibility |
| --- | --- |
| DEP-01 | Bar Event Admission: validate identity, required values, lineage, order, duplicate/conflict status, and admission disposition |
| DEP-02 | Per-Entity Ordered Analytical State: advance only the addressed entity and retain bounded sufficient solver state |
| DEP-03 | JEH / Ehlers Phase Update: calculate phase from the accepted ordered bar trajectory |
| DEP-04 | Circular Phase State: represent initialization, observability, validity, $\phi[n-1]$, and $\phi[n]$ without sentinels |
| DEP-05 | Phase Motion Analysis: derive versioned directional motion evidence in degrees per bar |
| DEP-06 | Boundary Crossover Detection: distinguish robust circular crossover events from strategy-region membership |
| DEP-07 | Strategy Region / Rules Engine: apply the governed four-region strategy interpretation and distinguish persistent strategy-region membership from crossover events; `HOP-ON` and `HOP-OFF` are not persistent states |
| DEP-08 | Universe Candidate / Holding State: maintain latest per-entity strategy region, allocation candidates, holdings, liquidation context, capacity, and pending work |
| DEP-09 | Candidate Comparison / Ranking: select among current eligible `ALLOCATE` candidates using an approved Phase Velocity ($\omega$) policy |
| DEP-10 | Strategy Decision Generation: decide whether to allocate, hold/trail, liquidate, disregard, or take no action with causal evidence |
| DEP-11 | ExecutionIntent Generation: publish an idempotent governed request without claiming execution |

These identifiers are local to this design. They are not repository package, service, type, or protobuf names and MUST NOT become implementation names without separate approval.

---

## 6. One BarEvent Lifecycle

```mermaid
sequenceDiagram
  participant SRC as Bar source adapter
  participant ADM as DEP-01 Admission
  participant ENT as Per-entity coordinator
  participant SOL as JEH phase solver
  participant ANA as Motion and crossover analysis
  participant SM as State-machine rules
  participant UNI as Universe coordinator
  participant DEC as Decision engine
  participant PUB as Intent publisher
  SRC->>ADM: new BarEvent(entity S, n)
  ADM->>ENT: accepted ordered bar
  ENT->>SOL: update bounded solver state
  SOL-->>ENT: PhaseEvidence / initialization evidence
  ENT->>ANA: phi n-1, phi n, validity
  ANA-->>SM: motion evidence and crossover evidence
  SM-->>UNI: entity state transition or persistence
  UNI->>DEC: latest asynchronous universe view
  DEC-->>PUB: decision and optional ExecutionIntent
```

For one newly accepted bar:

1. Admit the event and establish deterministic identity and source lineage.
2. Advance only that entity's ordered bar/solver trajectory.
3. Run the JEH/Ehlers phase update.
4. Record initialization or current observable phase $\phi[n]$.
5. Retain/use prior observable phase $\phi[n-1]$ where valid.
6. Derive phase-motion evidence and detect circular boundary crossover evidence.
7. Update the entity's four-region JEH rules state, preserving any distinct crossover event.
8. Update universe candidate, holding, unavailable, and pending state.
9. Evaluate liquidation first, then current candidate reallocation, hold, or disregard behavior.
10. Produce a strategy decision and optional `ExecutionIntent`.
11. Leave action and outcome production to the external adapter.

Malformed input, missing prerequisites, or an unobservable phase MUST produce explicit non-action evidence and MUST NOT corrupt another entity.

---

## 7. BarEvent Semantics and Ordering

The fundamental event is **NEW ACCEPTED BAR FOR ENTITY S**. The JEH mathematics requires only finite `high` and `low` values to derive Median Price. Open, close, volume, timestamps, interval, and provider fields are not solver inputs. The common event still requires separately approved typed identity, ordering, provenance, and validation fields; those concerns MUST NOT be inferred from the mathematical minimum.

Repository evidence establishes that an OFFLINE observation is identified at least by `(collection_run_id, symbol, generator_sequence_no)`. Its `generator_sequence_no` is scoped to one symbol within one collection run. The stored source also preserves source timestamp text, received and persisted times, payload hash, provider message type, and duplicate/regression indicators. ONLINE currently exposes both `symbol` and `instrument_id`, uses `(symbol, interval_start)` for deduplication, and supplies no accepted sequence or resume cursor. The canonical DSE_JEH entity identity and cross-mode source-observation identity therefore remain human decisions.

Bars can arrive asynchronously:

```text
AAPL, AAPL, NVDA, SPY, AAPL, QQQ, NVDA
```

There is no required synchronized 30-symbol frame. `generator_sequence_no` represents accepted arrival order for one entity and is not a global market clock. The universe-level pipeline observes the latest valid state currently known for every entity.

Common admission requires explicit policy for duplicates, conflicts, gaps, missing bars, and out-of-order bars. No policy may silently rewrite accepted history or fabricate bars. These conditions can coexist and MUST NOT be collapsed into one mutually exclusive admission enum unless a precedence rule is approved. A terminal admission disposition and detected input conditions are distinct concepts. Rejected input MUST NOT mutate solver state.

The stored OFFLINE evidence cannot reconstruct a unique source-authentic cross-entity total arrival order: partition writers persist concurrently, `received_time` can tie, and `persisted_time` can be batch-assigned. OFFLINE replay therefore preserves authoritative entity-scoped order and requires an explicit versioned deterministic replay-order policy that distinguishes replay order from source order. Bars MUST NOT be globally sorted by `generator_sequence_no`.

The repository's current ONLINE candidate is `finfeedsat.v1.IngestionService.StreamBars(StreamBarsRequest) returns (stream Bar)`. Its implementation sends per-symbol chronological catch-up from process-local windows and then live accepted events, but does not provide a global order, accepted sequence, gap marker, resume cursor, or loss notification. Its bounded subscriber queue can discard the oldest queued bar for a slow consumer. Catch-up/live handoff can repeat a bar, and same-interval partial updates may be replaced or suppressed. These are observed interface semantics, not approval to bind DSE_JEH to that RPC. Ordering, finalized-bar meaning, duplicate/conflict handling, loss detection, reconnect/resumption, warm-up, backpressure, and lag policy MUST be approved or added upstream before ONLINE implementation. The Dynamic Execution Pipeline MUST NOT silently continue across an undetected missing accepted bar if doing so could alter solver state or evidence.

---

## 8. Phase Calculation Inside V0.1

The primary path is:

```text
BarEvent -> JEH phase solver -> PhaseEvidence -> Dynamic Execution Pipeline state logic
```

The JEH/Ehlers phase solver is inside the V0.1 TransSat runtime boundary. `PhaseEvidence` is therefore an internal analytical evidence/state concept on this path, not the sole primary external contract. The phase solver and strategy rules remain logically separate so solver equivalence, strategy behavior, version replacement, and failures can be tested independently.

The current approved-reference behavior uses ordered median price, $P[n]=(High[n]+Low[n])/2$, normalized observable phase in $[0,360)$, and the current solver/configuration's 63-observation lookback. This does not make median price or 63 observations a universal JEH law.

For the proven current configuration:

- sequence 1..63: `INITIALIZING`;
- sequence 64: first possible observable and production-eligible phase;
- all initializing bars are processed and advance analytical state;
- unavailable phase is represented explicitly, never as 0 degrees; and
- production eligibility permits a Phase Angle to enter future Dynamic Execution processing but does not itself authorize a strategy action.

No valid strategy action may be generated from unobservable phase.

### 8.1 Rule #1 - 63-Contiguous-Bar Eligibility

Each symbol owns an independent causal bar sequence and JEH state. A symbol's first 63 admitted contiguous valid bars update JEH state and emit `INITIALIZING` without an eligible Phase Angle. The next valid contiguous bar, sequence 64, updates JEH state and may emit `OBSERVABLE`; its Phase Angle is then eligible to enter Dynamic Execution processing. Every later admitted contiguous bar follows the same observable path.

Continuity is sequence integrity within the symbol's sequence scope, not elapsed wall-clock time. Irregular or long intervals between successive valid bars do not reset analytical history. Duplicate, conflicting, gapped, out-of-order, and invalid candidates retain their deterministic admission disposition and do not mutate JEH state. No missing bar is synthesized or interpolated. If causal sequence integrity is not restored, later candidates cannot advance the solver or regain production eligibility.

---

## 9. Per-Entity and Universe State

### 9.1 Per-entity hot state

Each entity conceptually maintains:

- entity identity and source lineage;
- ordered bar and bounded sufficient solver state;
- initialization state and current bar identity;
- previous accepted analytical state;
- explicit $\phi[n-1]$ and $\phi[n]$ presence;
- phase observability and validity;
- previous and current JEH strategy region;
- most recent boundary crossover;
- phase-motion evidence and its validity/confidence;
- candidate and holding status where applicable; and
- solver, strategy, ranking, and configuration version identity.

Hot state MUST NOT require unbounded bar history when bounded sufficient solver state can preserve equivalent behavior. Durable evidence may retain history outside hot state.

### 9.2 Universe state

Universe state is distinct from per-entity analytical state and includes current holdings, `ALLOCATE` candidates, `HOLD & TRAIL` entities, `LIQUIDATE` entities, `DISREGARD` entities, initializing/unavailable entities, candidate ranking state, execution capacity/capital state where applicable, and pending decisions/intents.

```mermaid
flowchart LR
  subgraph Entities[Independent per-entity state]
    A[Entity A<br/>latest bar, solver, phase,<br/>motion, crossover, JEH strategy region]
    B[Entity B<br/>latest bar, solver, phase,<br/>motion, crossover, JEH strategy region]
    C[Entity C<br/>latest bar, solver, phase,<br/>motion, crossover, JEH strategy region]
  end
  U[UniverseState<br/>latest valid entity views]
  H[Holdings and capacity]
  K[Candidate set and ranking]
  P[Pending decisions and intents]
  A --> U
  B --> U
  C --> U
  U --> H
  U --> K
  U --> P
```

Universe state is not a claim that all entities were updated simultaneously. Each entity view needs freshness/staleness evidence. Snapshot semantics, stale-state eligibility, and stale candidate expiry remain open.

---

## 10. Four-Region Circular JEH Rules Engine

The System Design adopts the canonical region/action vocabulary from the supplied Hop On Hop Off strategy artifact. The artifact itself is not stored in this repository, so no repository path is asserted. The four persistent strategy regions/actions are:

| Half-open phase interval | Canonical strategy region/action | Current strategy interpretation |
| --- | --- | --- |
| $0 \le \phi < 90$ | `HOLD & TRAIL` | Maintain the holding and trail dynamically; exact trailing-stop policy remains unresolved |
| $90 \le \phi < 180$ | `LIQUIDATE` | Liquidation context; freed capital becomes available for governed reallocation |
| $180 \le \phi < 270$ | `DISREGARD` | Not an allocation candidate |
| $270 \le \phi < 360$ | `ALLOCATE` | Allocation-candidate region; rank candidates by Phase Velocity ($\omega$) under an approved policy |

```mermaid
stateDiagram-v2
  state "HOLD & TRAIL" as HOLD_AND_TRAIL
  [*] --> INITIALIZING
  INITIALIZING --> DISREGARD: first observable phase in 180-270
  INITIALIZING --> ALLOCATE: first observable phase in 270-360
  INITIALIZING --> HOLD_AND_TRAIL: first observable phase in 0-90
  INITIALIZING --> LIQUIDATE: first observable phase in 90-180
  DISREGARD --> ALLOCATE: valid 270° crossover / HOP-ON event
  ALLOCATE --> HOLD_AND_TRAIL: valid 0° boundary crossover
  HOLD_AND_TRAIL --> LIQUIDATE: valid 90° crossover / HOP-OFF event
  LIQUIDATE --> DISREGARD: validated cross 180°
  DISREGARD: 180-270 DISREGARD
  ALLOCATE: 270-360 ALLOCATE / rank by Phase Velocity
  HOLD_AND_TRAIL: 0-90 HOLD & TRAIL / dynamic trailing behavior
  LIQUIDATE: 90-180 LIQUIDATE / freed capital for reallocation
```

The wheel is one full circular phase space, interpreted clockwise from top as `HOLD & TRAIL`, `LIQUIDATE`, `DISREGARD`, then `ALLOCATE`. Terms such as trough, peak, accelerating upward, or rolling over are **JEH strategy interpretations**, not universal scientific facts unless independently validated.

**Strategy-region membership is not a boundary crossover event.** `HOP-ON` is the named event for a valid 270° boundary crossover in the approved direction into `ALLOCATE`. `HOP-OFF` is the named event for a valid 90° boundary crossover in the approved direction into `LIQUIDATE`. Neither is a persistent strategy state. Five successive bars in `ALLOCATE` represent persistent membership, not five `HOP-ON` events. First observable classification, same-region persistence, region exit, boundary crossover, region entry, candidate status, decision, and intent require distinct evidence.

The transition from `ALLOCATE` into `HOLD & TRAIL` uses the 0° boundary crossover; no additional event name is defined. Source trigger concepts such as $\phi[n]\ge270$ and $\phi[n-1]<270$, or $\phi[n]\ge90$ and $\phi[n-1]<90$, express intended directional cases only. They MUST NOT be applied blindly to circular wraps, reverse movement, exact-boundary samples, jitter, or large jumps. The names `HOP-ON` and `HOP-OFF` establish event semantics, not detection mathematics. Robust 270°, 90°, 180°, and 360°/0° crossover algorithms remain blocking questions.

---

## 11. Phase Motion and Phase Velocity ($\omega$)

```mermaid
flowchart LR
  P[phi n-1 and phi n] --> M[Phase Motion Analyzer]
  M --> E[PhaseMotionEvidence<br/>Phase Velocity omega<br/>direction, magnitude, validity,<br/>degrees per bar]
  E --> R[Candidate Ranking Policy]
```

The pipeline requires phase-motion evidence because multiple entities may occupy `ALLOCATE`. The canonical intended ranking quantity is **Phase Velocity ($\omega$)**, measured in **degrees per bar**, because V0.1 is driven by accepted bar index rather than wall-clock frequency. Phase Velocity is an intended ranking input; its exact definition remains unresolved.

The exact mathematics are not approved. In particular, the naive positive-modulo expression

$$
\omega = (\phi[n]-\phi[n-1]) \bmod 360
$$

MUST NOT be frozen as the definition. A transition from 10 degrees to 350 degrees may represent approximately -20 degrees of signed circular movement, not +340 degrees.

Explicit counterexample: $\phi[n-1] = 10°$ and $\phi[n] = 350°$ may represent approximately $-20°$ of signed circular movement; naive positive modulo produces $+340°$.

The approved design must resolve signed circular difference, direction, wrap handling, 180-degree ties, abnormal jumps, initialization, gaps, confidence/validity, smoothing, one-bar versus multi-bar estimation, and whether acceleration is separately required. Velocity MUST NOT be called acceleration. Historical candidate values or rankings from source analysis are examples, not evidence of acceleration, optimal allocation, or efficacy.

---

## 12. Dynamic Allocation and Strategy Behavior

Candidate comparison and ranking are part of the Dynamic Execution Pipeline. The first intended ranking signal is Phase Velocity ($\omega$), but its definition and the ranking formula, timing, ties, staleness, and eligibility are unresolved.

The current intended workflow is:

1. **LIQUIDATE first.** A valid 90° `HOP-OFF` crossover places the entity in the `LIQUIDATE` region/action context. The Strategy Decision determines whether liquidation should be requested; if eligible, it generates an `ExecutionIntent`. Freed capital becomes available for governed reallocation under still-unresolved sequencing and capital rules.
2. **ALLOCATE / reallocate.** When approved capacity/capital is available, inspect the current non-stale `ALLOCATE` candidate universe, rank candidates using the approved Phase Velocity ($\omega$) policy, and generate allocation intent only when the Strategy Decision permits it.
3. **HOLD & TRAIL.** Held entities in `HOLD & TRAIL` remain held under the strategy-region model and are intended to trail stops dynamically. Exact trailing-stop mathematics are not approved and remain open.
4. **DISREGARD.** Entities in `DISREGARD` are not allocation candidates under the current region model. Remaining in `DISREGARD` does not generate a `HOP-ON` event.

The design does not yet approve single versus multiple holdings, capital representation, capacity, replacement, candidate-set timing, or liquidation/allocation transaction semantics.

---

## 13. Decision and External Execution Boundary

```mermaid
flowchart LR
  J[JEH Strategy Region] --> T[Strategy Transition / Crossover Evidence]
  T --> G[Execution Eligibility]
  G --> D[Strategy Decision]
  D --> I[ExecutionIntent]
  I --> A[External Execution Adapter]
  A --> E[ExecutionEvent]
  E --> R[Reconciliation and evidence]
```

The pipeline decides what should be requested. It does not place orders or claim outcomes. `ExecutionIntent` requires deterministic identity, target, requested action, causal decision, strategy/configuration lineage, idempotency/correlation identity, and validity semantics in the final contract. `ExecutionEvent` separately records acceptance, rejection, action/fill, cancellation, failure, or reconciliation.

The first proving executor is **MOCK / SIMULATED EXECUTION ONLY**: no broker, capital, live order, or Alpaca submission. The proof chain is `bar event -> phase -> motion -> crossover -> state -> ranking -> decision -> ExecutionIntent -> mock ExecutionEvent`. No external call may hold a per-entity or universe strategy-state lock.

---

## 14. Online and Offline Stream Modes

```mermaid
flowchart TD
  MODE[DSE_JEH_MODE]
  ONLINE[ONLINE]
  OFFLINE[OFFLINE]
  FIN[Fin_Feed_Sat_1<br/>gRPC bar stream]
  CON[Internal gRPC Bar Consumer]
  DB[(bar_sequence_db.bar_sequence)]
  PROD[Internal Offline Stream Producer]
  BAR[Common BarEvent Admission]
  DEP[Same Dynamic Execution Pipeline]
  MODE --> ONLINE --> FIN --> CON -->|map/admit| BAR
  MODE --> OFFLINE --> DB --> PROD -->|one bar at a time| BAR
  BAR --> DEP
```

The proposed startup environment variable accepts exactly:

```text
DSE_JEH_MODE=ONLINE
DSE_JEH_MODE=OFFLINE
```

The host/startup layer reads `DSE_JEH_MODE` and selects one producer/input component. The Dynamic Execution Pipeline MUST NOT read this variable or branch on source mode.

**ONLINE** maps bars from the `Fin_Feed_Sat_1` gRPC bar stream through an internal gRPC consumer into the common `BarEvent` boundary. The transport is not JEH mathematics, and the pipeline MUST NOT depend on Fin protobuf types or gRPC semantics. This design neither invents nor freezes a Fin service or protobuf contract.

The inspected candidate binding is `finfeedsat.v1.IngestionService.StreamBars`; the request is `finfeedsat.v1.StreamBarsRequest` and the server-streamed response is `finfeedsat.v1.Bar`. The DSE_JEH adapter would require adapter-local identity mapping, sequencing/continuity validation, and duplicate handling. The candidate is not approved for binding because its current contract cannot expose queue loss, guarantee finalized bars, or resume from an acknowledged position.

**OFFLINE** uses an internal stream producer to read `bar_sequence_db.bar_sequence` in an approved deterministic order and emit one stored observation at a time into the common `BarEvent` boundary. OFFLINE remains a stream, not a batch strategy processor. The producer MUST NOT calculate phase, phase motion, crossover, JEH strategy region, ranking, decisions, or `ExecutionIntent`, and MUST NOT invoke `phase_angle_series_generator`.

For equivalent admitted bars and ordering, both modes execute identical solver, motion, crossover, state, universe, ranking, decision, and intent logic. Source provenance may differ; strategy behavior must not. A separate offline or backtest strategy implementation is prohibited. Optional future OFFLINE pacing is a producer/runtime concern and MUST NOT change bar interpretation or pipeline behavior.

---

## 15. The 4,519-Bar Offline Proving Stream

The initial OFFLINE proving source is the existing `bar_sequence` collection in `bar_sequence_db`. On 2026-09-12, a read-only local `countDocuments({})` query independently verified **4,519 bar observations**. This is a proving dataset size, not an architectural constant or the Dynamic Execution Pipeline input contract.

```mermaid
flowchart TD
  DB[(bar_sequence_db.bar_sequence<br/>4,519 verified observations)]
  O[Deterministic cross-entity replay ordering policy]
  R[Internal Offline Stream Producer]
  E[One accepted BarEvent at a time]
  D[Dynamic Execution Pipeline]
  M[Mock decisions, intents, and ExecutionEvents]
  DB --> O --> R --> E --> D --> M
```

The OFFLINE producer must preserve entity-scoped `generator_sequence_no`. Available timestamp and persistence fields do not reconstruct a unique source-authentic total accepted-arrival order, so the chosen deterministic replay policy must be explicit, deterministic, versioned, and carried in replay/evidence configuration. It is incorrect to treat equal sequence numbers as frames or sort all entities globally by `generator_sequence_no`.

Bars 1 through 63 for an entity still enter the pipeline and advance solver state. They are not discarded merely because phase is initializing.

---

## 16. Independent Analytical Reference and Phase Equivalence

`Bar_Sequence_Lab/phase_angle_series_generator` is an **independent analytical reference / batch phase generator**. It calculates phase-angle series from collected sequences. It is not the Dynamic Execution Pipeline, is not moved or modified by this design, is not invoked as a batch application by the runtime, and its application structure is not copied into the runtime.

```mermaid
flowchart TD
  DB[(Same stored bars)]
  GEN[PATH A - ANALYTICAL REFERENCE<br/>phase_angle_series_generator]
  REF[Reference phase series]
  REP[PATH B - OFFLINE RUNTIME PROVING<br/>Internal Offline Stream Producer]
  BAR[One BarEvent at a time]
  DEP[Dynamic Execution Pipeline]
  RUN[Runtime phase series<br/>state evidence, decisions, intents]
  CMP[PHASE COMPARISON]
  DB --> GEN --> REF --> CMP
  DB --> REP --> BAR --> DEP --> RUN --> CMP
```

For matching observations, runtime phase behavior should equal the approved reference behavior by entity/symbol, collection run or equivalent source lineage, `generator_sequence_no`, solver identity/version, and input-series definition. The current reference test uses absolute tolerance `1e-9`; that is repository evidence, not an approved DSE_JEH production tolerance. Exact DSE_JEH numeric tolerance/precision remains open. Required mathematical equivalence does not imply code reuse.

The existing phase CSV and `DSE_JEH_provers` remain regression/reference evidence for `precomputed PhaseEvidence -> zone/transition` behavior. They are not the runtime input architecture and are not discarded or promoted as the runtime.

---

## 17. Determinism, Mode Equivalence, and Evidence Identity

For the same admitted bar observations and order, ONLINE and OFFLINE MUST produce equivalent results given the same solver version, strategy version, configuration, ranking policy, and initial state:

- phase and phase-motion evidence;
- crossover and state-transition evidence;
- universe-state evolution and rankings;
- decisions and `ExecutionIntent` records; and
- mock `ExecutionEvent` records.

Evidence identity must include enough typed and versioned information to distinguish entity, source observation, source mode/provenance, entity order, replay/order policy where applicable, solver, strategy, zone model, phase-motion policy, ranking policy, configuration, causal predecessor, and output kind. IDs must support idempotency, mode-equivalence comparison, reconciliation, and conflict detection without making transport types, file layout, or database keys part of strategy mathematics.

---

## 18. Conceptual Component Model

```text
DSE_JEH_TransSat_1
|
+-- Runtime / Startup Configuration
|     +-- DSE_JEH_MODE = ONLINE | OFFLINE
+-- ONLINE Bar Input
|     +-- Fin_Feed_Sat_1 gRPC Bar Consumer
+-- OFFLINE Bar Input
|     +-- Internal Bar Sequence Stream Producer
+-- Common BarEvent Admission
+-- Per-Entity Analytical State Coordinator
+-- JEH Phase Solver
+-- Circular Phase State Manager
+-- Phase Motion Analyzer
+-- Boundary Crossover Detector
+-- Strategy Region / Rules Engine
+-- Universe State Coordinator
+-- Candidate Ranking Engine
+-- Strategy Decision Engine
+-- ExecutionIntent Publisher
+-- Execution Adapter Interface
+-- Evidence / Diagnostics
+-- optional Persistence Adapter
```

These are logical design components only. They do not authorize packages, services, files, or contracts.

---

## 19. Portability, Persistence, and Failure Isolation

The Dynamic Execution Pipeline MUST NOT depend mathematically on MongoDB, CSV, phase CSV, proving folders, Next.js, a Bar Sequence Lab path, a broker SDK, Alpaca, a fixed 30-symbol population, a fixed 4,519-bar dataset, `collection_run_id`, or `SERIES_SIZE`.

For OFFLINE proving, MongoDB is the stored source used by the internal stream producer. The DSE_JEH engine consumes `BarEvent` streams, not a database. Persistence is an optional infrastructure adapter and does not define JEH mathematics. Checkpoint and recovery semantics remain open.

Failures must be isolated:

| Failure | Required response |
| --- | --- |
| Malformed, duplicate, conflicting, missing, or out-of-order bar | Apply explicit deterministic admission policy; do not silently mutate history |
| One entity's solver/state failure | Degrade/quarantine that entity without corrupting others |
| Ranking or decision failure | Preserve accepted analytical evidence and surface failure |
| Persistence/publication/viewer failure | Surface health degradation; do not alter strategy truth |
| Execution adapter latency/failure | Preserve strategy state; do not claim execution; do not block unrelated entities |
| Restart/version mismatch | Refuse unsafe restoration or reconstruct under an approved policy |

Runtime health should distinguish liveness, input health, per-entity initialization/readiness, universe readiness, persistence/publication health, execution-adapter health, recovery/replay state, and active version identities.

---

## 20. Viewer Boundary

The viewer is observational only and MUST NOT calculate authoritative phase, motion, crossover, state, ranking, decision, intent, or execution truth. A future viewer may display polar universe state, entity phase/motion/state, crossovers, candidates, ranking, holdings, decisions, `ExecutionIntent`, and `ExecutionEvent`. Viewer failure or latency cannot affect processing.

Polar visualization represents one circular phase space. A rectangular chart's 360/0 wrap is not automatically a scientific discontinuity.

---

## 21. Validation Strategy

The first separately authorized validation plan should cover:

1. Every `BarEvent` admission state, identity combination, duplicate, conflict, gap, and out-of-order case.
2. Entity isolation and asynchronous sequences without synchronized frames.
3. Current 63-observation initialization behavior, first possible phase at 64, and legitimate zero degrees.
4. Phase equivalence against `phase_angle_series_generator` with approved tolerance.
5. Signed motion, wrap, reverse motion, large jump, and abnormal/invalid vectors.
6. Canonical strategy-region membership versus initial entry, persistence, exit, named `HOP-ON`/`HOP-OFF` events, other crossover evidence, and entry.
7. Every boundary, including 90, 180, 270, and 360/0 in both movement directions.
8. Stale universe members, candidate-set timing, ranking ties, and capacity behavior once approved.
9. Liquidation-before-allocation ordering and intent idempotency.
10. Decision, intent, mock execution, and reconciliation non-equivalence.
11. Deterministic 4,519-event replay and repeated evidence identity/digest comparison.
12. ONLINE/OFFLINE semantic equivalence, source-specific recovery, and component failure isolation.
13. ONLINE ordering, duplicate/gap/loss detection, reconnect/resumption, backpressure, and lag behavior once contracted.
14. ONLINE startup with insufficient phase history and any approved solver warm-up mechanism.

Promotion requires explicit acceptance criteria. The dataset and existing prover are regression evidence, not proof of profitability, optimality, or production readiness.

---

## 22. Open Questions and Blocking Decisions

No answer is invented where evidence does not yet exist. `RESOLVED` below closes the design question stated, not implementation authorization.

### 22.1 Phase 1 contract decision register

| OQ | Classification | Decision, evidence, and reason | Initial proto consequence |
| --- | --- | --- | --- |
| OQ 1 | REQUIRES HUMAN DECISION | The mathematical minimum and OFFLINE identity are evidenced, but canonical entity identity, cross-mode observation identity, optionality, timestamp representation, and admission shape are not approved. | **BLOCKER:** do not freeze `BarEvent`, `SourceProvenance`, or admission evidence. |
| OQ 2 | REQUIRES HUMAN DECISION | Repository evidence identifies `finfeedsat.v1.IngestionService.StreamBars(StreamBarsRequest) returns (stream Bar)`. Existing code establishes only candidate behavior; approval must determine whether DSE_JEH binds it or requires an upstream revision. | Does not require Fin types in the DSE proto; blocks an approved ONLINE mapping. |
| OQ 3 | REQUIRES HUMAN DECISION | Current code provides chronological catch-up within each symbol, separate symbol traversal, then live delivery. It provides no global order or accepted sequence. | Define order scope independently; do not claim stronger ONLINE ordering. |
| OQ 4 | REQUIRES HUMAN DECISION | Fin deduplicates by `(symbol, interval_start)`, may replace/suppress partials, and can repeat a bar across catch-up/live handoff. DSE_JEH duplicate identity and handling are not approved. | **BLOCKER:** duplicate/conflict findings and disposition cannot be frozen. |
| OQ 5 | REQUIRES HUMAN DECISION | The stream carries no predecessor sequence, gap marker, or slow-consumer loss notification. | **BLOCKER:** continuity claims require an upstream contract or approved fail-closed adapter policy. |
| OQ 6 | REQUIRES HUMAN DECISION | No resume token or cursor exists; reconnect creates a new subscription over current process-local windows. | Resume fields MUST NOT be invented; recovery remains blocked. |
| OQ 7 | REQUIRES HUMAN DECISION | Current subscriber queues are bounded and lossy, dropping the oldest queued bar before retrying the newest. | **BLOCKER:** approve loss/backpressure behavior before ONLINE analytical use. |
| OQ 8 | REQUIRES HUMAN DECISION | Current catch-up may help prime from a retained window, but restart, loss, duplicates, and state compatibility are not guaranteed. | Recovery/warm-up fields remain undefined. |
| OQ 9 | REQUIRES HUMAN DECISION | Repository evidence proves no unique source-authentic cross-entity total order. Entity-local `(collection_run_id, symbol, generator_sequence_no)` order is authoritative; a versioned deterministic replay policy must be selected. | **BLOCKER:** represent entity order now only after replay-policy identity and evidence placement are approved. |
| OQ 10 | DEFERRED - NOT REQUIRED FOR INITIAL PHASE 1 PROTO | Pacing is producer control and cannot alter event semantics. No current evidence requires it. | Omit pacing declarations. |
| OQ 11 | RESOLVED BY EXISTING SYSTEM DESIGN | Equivalence compares outputs for the same admitted bars and order under the same solver/configuration/initial state and causal lineage. | Evidence must support typed joins; exact acceptance tolerance remains OQ 19. |
| OQ 12 | REQUIRES HUMAN DECISION | OFFLINE and ONLINE expose different identity schemes; source identity, runtime identity, evidence identity, entity identity, and order-policy identity must remain distinct. | **BLOCKER:** `SourceProvenance` identity fields cannot be frozen. |
| OQ 13 | REQUIRES HUMAN DECISION | A current Fin subscriber can receive a retained per-symbol window, but this is not a guaranteed compatible warm-up contract. | No warm-up source declaration is approved. |
| OQ 14 | RESOLVED BY EXISTING SYSTEM DESIGN | A fresh solver requires 63 ordered observations before observation 64 can first expose phase, unless compatible state restoration is separately approved. | Initialization count belongs in solver identity/configuration, not an availability sentinel. |
| OQ 15 | RESOLVED BY EXISTING SYSTEM DESIGN | Insufficient history yields explicit `INITIALIZING`; each admitted bar advances state and no strategy action is allowed. | `PhaseStatus` requires `UNSPECIFIED`, `INITIALIZING`, and `OBSERVABLE`; phase value requires presence. |
| OQ 16 | RESOLVED FROM REPOSITORY EVIDENCE | `phase.go` computes Median Price solely from finite `High` and `Low`: `P[n]=(High[n]+Low[n])/2`. Other bar fields are provenance/admission concerns, not JEH inputs. | Solver-value fields are `high` and `low`; this does not resolve the complete `BarEvent`. |
| OQ 17 | DEFERRED - NOT REQUIRED FOR INITIAL PHASE 1 PROTO | `phase.go` demonstrates bounded recurrence state, but checkpoint serialization is not approved and need not cross the initial contract. | Keep solver state internal; add no checkpoint message. |
| OQ 18 | RESOLVED FROM REPOSITORY EVIDENCE | The scientific definition, `phase.go`, and frozen vectors in `phase_test.go` establish the independent TA-Lib-compatible equivalence basis and exact current mathematics. | `SolverIdentity` must distinguish family/name, version, algorithm/reference version, input series, initialization requirement, and configuration identity; build identity belongs on production evidence if approved. |
| OQ 19 | REQUIRES HUMAN DECISION | The reference test uses absolute tolerance `1e-9`, but no DSE_JEH production precision/tolerance is approved. | Does not require a tolerance field in analytical evidence; blocks final equivalence acceptance. |
| OQ 31 | REQUIRES HUMAN DECISION | OFFLINE entity-sequence gaps are detectable; ONLINE gaps are not reliably detectable. The reference processor tolerates gaps, which is not authority for runtime admission. | **BLOCKER:** missing-predecessor finding and terminal effect are not approved. |
| OQ 32 | REQUIRES HUMAN DECISION | Generator, Fin, and prover behaviors differ for exact and conflicting duplicates. No DSE_JEH policy controls. | **BLOCKER:** duplicate/conflict identity, findings, and mutation rules are not fully defined beyond rejected input not mutating state. |
| OQ 33 | REQUIRES HUMAN DECISION | Generator preserves source-time regressions, Fin inserts/replaces chronologically, and the prover rejects non-increasing new positions. | **BLOCKER:** out-of-order finding and disposition are not approved. |
| OQ 45 | DEFERRED - NOT REQUIRED FOR INITIAL PHASE 1 PROTO | Checkpoint storage, compatibility, restart, and reconciliation mechanisms are not approved. | Omit checkpoint/recovery declarations; fresh initialization remains valid. |
| OQ 46 | DEFERRED - NOT REQUIRED FOR INITIAL PHASE 1 PROTO | DSE_JEH durability, retention, and publication boundaries are not approved. | Define no persistence API or service. |
| OQ 49 | DEFERRED - NOT REQUIRED FOR INITIAL PHASE 1 PROTO | The design identifies health dimensions but no lifecycle state machine or external operations consumer. | Omit `RuntimeStatus`, `RuntimeHealthEvidence`, and operations service/RPC. |
| OQ 50 | DEFERRED - NOT REQUIRED FOR INITIAL PHASE 1 PROTO | Promotion criteria require later validated OFFLINE evidence and safe ONLINE delivery semantics. | No promotion declaration belongs in the proto. |

### 22.2 Phase 2 open questions preserved unchanged

20. What is the exact signed circular $\Delta\phi$ definition, including tie behavior?
21. What is the exact Phase Velocity ($\omega$) definition in degrees per bar?
22. Does $\omega$ use one-bar or multi-bar estimation?
23. Is phase motion smoothed; if so, under what separately identified policy?
24. Is angular acceleration needed as a distinct quantity?
25. What are robust 90-degree `HOP-OFF` crossover semantics?
26. What are robust 180-degree crossover semantics?
27. What are robust 270-degree `HOP-ON` crossover semantics?
28. What are robust 360/0 crossover semantics?
29. How is reverse phase movement interpreted and processed?
30. How are large or abnormal phase jumps classified?
34. How is entity-state staleness measured and exposed?
35. What are universe snapshot semantics under asynchronous arrivals?
36. What candidate ranking mathematics, tie-breaking, and validity rules are approved?
37. When is the candidate set sampled relative to an incoming event and decisions?
38. When and how do stale candidates expire?
39. How is allocation/execution capacity defined?
40. Are single or multiple simultaneous holdings supported?
41. How is capital state represented without coupling strategy mathematics to an executor?
42. What atomicity and reconciliation govern liquidation-before-allocation sequencing?
43. What are mock execution acceptance, rejection, fill, latency, and failure semantics?
44. Is a trailing-stop policy part of V0.1, and what mathematics would govern it?
47. What is the exact versioned `ExecutionIntent` contract?
48. What is the exact versioned `ExecutionEvent` contract?

### 22.3 Stage A proto gate

**PASSED BY EXPLICIT HUMAN IMPLEMENTATION DIRECTION ON 2026-09-12.** Phase 1 uses normalized symbol as canonical entity identity; typed source provenance keeps ONLINE and OFFLINE observation identities distinct; OFFLINE replay uses versioned `(collection_run_id, symbol, generator_sequence_no, payload_hash)` ordering; and admission conservatively rejects invalid, duplicate, conflicting, missing-predecessor, and out-of-order candidates without mutating solver state. ONLINE binds the current `finfeedsat.v1.IngestionService.StreamBars` contract with bounded reconnect and documents the upstream continuity limitations. These deterministic implementation choices do not claim stronger upstream guarantees or resolve Phase 2 semantics.

---

## 23. Traceability and Scientific Caution

| Source | Design use | Authority limit |
| --- | --- | --- |
| Supplied Hop On Hop Off strategy artifact (not stored in this repository) | Canonical `DISREGARD`, `ALLOCATE`, `HOLD & TRAIL`, and `LIQUIDATE` region/action names; `HOP-ON` and `HOP-OFF` crossover-event names; Phase Velocity ranking intent; trailing and reallocation intent | Terminology is authoritative for this design; crossover, velocity, ranking, trailing-stop, and capital mathematics remain unresolved |
| `phase_angle_series_generator` docs and implementation | Median-price solver behavior, 63-bar initialization, normalized phase, independent reference path | Standalone batch generator; not runtime or strategy authority |
| `DSE_JEH_provers` | PhaseEvidence admission, zone/transition regression evidence, deterministic proving patterns | Precomputed-phase apparatus; not runtime input architecture or efficacy evidence |
| `bar_sequence_db.bar_sequence` | Initial realtime-equivalent replay source; 4,519 observations verified 2026-09-12 | Dataset and MongoDB are not architectural dependencies |
| This document | Proposed V0.1 runtime boundary and requirements | Human review required; implementation not authorized |

The source strategy's polar interpretation motivates the four states and dynamic behavior. It does not establish that a particular phase means a universal trough/peak, that one-bar motion proves acceleration, or that historical $\omega$ examples prove optimal allocation.

---

## 24. Non-Goals

This revision does not authorize or perform:

- Go implementation, gRPC client implementation, or protobuf creation/modification;
- modification of `Fin_Feed_Sat_1`;
- implementation of the Internal Offline Stream Producer or Dynamic Execution Pipeline;
- replay execution;
- modification of `phase_angle_series_generator`, Bar Sequence Lab, or `DSE_JEH_provers`;
- MongoDB data modification or repository restructuring;
- live broker, Alpaca, real-capital, or live-order execution;
- trailing-stop implementation;
- final phase-velocity, crossover, or candidate-ranking mathematics;
- a separate backtest strategy;
- viewer changes, deployment, commit, or push.

---

## 25. Change Log

| Version | Date | Change |
| --- | --- | --- |
| V0.1 | 2026-09-12 | Initial proposed system design for `DSE_JEH_TransSat_1`. |
| V0.1 pre-approval revision | 2026-09-12 | Revised in place before approval to establish the Dynamic Execution Pipeline as the central inner decision engine; make `BarEvent` the common engine input; place JEH/Ehlers phase calculation, event-driven crossover processing, phase-motion analysis, four-region rules, universe state, and ranking inside the V0.1 runtime; and retain `phase_angle_series_generator` as an independent analytical reference. Refined before approval to establish two V0.1 bar-stream modes selected by proposed `DSE_JEH_MODE`: ONLINE through an internal `Fin_Feed_Sat_1` gRPC bar consumer and OFFLINE through an internal stream producer reading the verified 4,519-bar `bar_sequence` collection. Both modes converge at common `BarEvent` admission and use identical Dynamic Execution Pipeline behavior; DSE_JEH consumes bar streams rather than databases; ONLINE delivery semantics remain open interface-design questions. |
| V0.1 terminology refinement | 2026-09-12 | Aligned the Dynamic Execution Pipeline with the supplied Hop On Hop Off strategy artifact: established `DISREGARD`, `ALLOCATE`, `HOLD & TRAIL`, and `LIQUIDATE` as the four canonical persistent strategy region/action terms; defined `HOP-ON` and `HOP-OFF` as 270° and 90° crossover events rather than states; retained Phase Velocity ($\omega$) as the intended ranking quantity; and preserved unresolved crossover, velocity, ranking, trailing-stop, and capital-allocation mathematics. |
| V0.1 Phase 1 contract review | 2026-09-12 | Recorded repository evidence for source contracts, ordering, JEH inputs, and ONLINE delivery; classified Phase 1 OQs; and held the initial proto at the Stage A gate pending approved identity, provenance, admission, replay-order, and ONLINE continuity decisions. |
| V0.1 Phase 1 implementation authorization | 2026-09-12 | Human direction authorized deterministic Phase 1 implementation choices, the single authoritative proto, generated bindings, ONLINE and OFFLINE adapters, common admission, JEH analytical processing, executable validation, and reports while preserving all Phase 2 boundaries. |

---

## 26. Authorization Statement

This document is **APPROVED** as the Phase 1 architectural authority.

**Phase 1 is implemented and pending human review of validation evidence.**

This approval and the explicit 2026-09-12 implementation direction authorize Phase 1 contract, runtime, test, replay-validation, and reporting work only. They do not authorize Phase 2 implementation, deployment, execution integration, or trading.