# DSE_TransSat_1 Process Model

**Title:** DSE_TransSat_1 Process Model
**Date:** 2026-09-10
**Version:** V0.1
**Status:** PROPOSED FOR HUMAN REVIEW
**Implementation:** NOT YET AUTHORIZED

## 1. Executive Summary

DSE_TransSat_1 is proposed as a HACCAM Transformer Satellite (`TransSat`) with subscriber type `PRODUCER_CONSUMER`. It consumes authoritative observations and scientific outputs published by Fin_FeedSat_1, maintains independent analytical state for each entity, applies an established John Ehlers Hilbert-transform/dominant-cycle phase method, combines that phase evidence with upstream Price and Volume evidence and portfolio state, and publishes new authoritative Decision Strategy information.

The physical design separates a causal Go worker from a non-causal viewer. `DSE_TransSat_1_worker` owns ingestion, analytical state, phase analysis, strategy transformation, portfolio state, diagnostics, lifecycle, and future DSE gRPC publication. `DSE_TransSat_1_viewer` renders worker-owned state and is never required for science or decisions. A viewer failure or slow viewer must not stop the worker.

Repository inspection establishes that the existing Fin_FeedSat_1 contract already exposes raw bars and three model streams over server-streaming gRPC. Raw OHLCV observations are published by `IngestionService.StreamBars`. `ModelService` separately publishes `DecisionEvent`, `PriceEvent`, and `VolumeEvent`. These streams share `symbol`, `interval_start_unix_ms`, and `market_snapshot_id` lineage, but they do not provide a cross-stream or cross-symbol total order. The DSE inbound boundary must therefore validate, correlate, deduplicate, and sequence observations per entity without changing Fin_FeedSat_1.

This document defines process and ownership boundaries, not implementation. It does not authorize DSE code, protobuf, strategy configuration, phase-zone thresholds, viewer construction, broker execution, or changes to Fin_FeedSat_1.

## 2. Purpose and Scope

This process model establishes the first reviewable architecture for DSE_TransSat_1 within Tramuthus. It defines:

- the observed upstream contracts and runtime behavior;
- the worker/viewer physical boundary;
- causal per-entity processing and state ownership;
- the boundary between established phase mathematics and experimental strategy mathematics;
- candidate, portfolio, and decision state as distinct concepts;
- lifecycle, failure-isolation, publication, and validation expectations; and
- prerequisites and open decisions that must be resolved before implementation.

This is a design artifact only. Proposed names and conceptual output shapes are not protobuf commitments.

## 3. Architectural Identity

**DSE design decision**

| Property | Value |
|---|---|
| Satellite identity | `DSE_TransSat_1` |
| Satellite role | `TransSat` |
| Subscriber type | `PRODUCER_CONSUMER` |
| Upstream producer in the proving topology | `Fin_FeedSat_1` |
| New information owned by DSE | Decision Strategy information and supporting DSE analytical state |

`PRODUCER_CONSUMER` is essential: DSE receives information it does not own, transforms it using DSE-owned state and rules, and publishes a new authoritative information product. It is neither a passive observer nor a relabeling proxy.

## 4. Tramuthus Development Context

The current repository physically separates:

```text
tramuthus/
  Fin_FeedSat_1/
  DSE_TransSat_1_worker/
  DSE_TransSat_1_viewer/
  docs/
```

Fin_FeedSat_1 is the inherited, known-working baseline. The DSE worker and viewer directories are intentionally separate. The current direct connection exists to prove DSE science before introducing HACCAM routing or deployment concerns.

The current proving universe has included AAPL, MSFT, NVDA, and AMZN, but the Fin_FeedSat_1 configuration is dynamic. `FIN_FEEDSAT_SYMBOLS` is parsed as a normalized, deduplicated list, defaults to AAPL, and is bounded by the current Basic-plan limit of 30 symbols in `Fin_FeedSat_1/internal/config/config.go`. Four entities are therefore a proving and presentation choice, not DSE architectural cardinality.

## 5. Relationship to Fin_FeedSat_1

### 5.1 Observed protobuf contract

**Existing repository fact**

The authoritative definition is `Fin_FeedSat_1/api/proto/fin_feedsat/v1/Fin_FeedSat_1.proto`:

- protobuf package: `finfeedsat.v1`;
- `go_package` option: `fin_feedsat_1/gen/fin_feedsat/v1;finfeedsatv1`;
- `IngestionService`: `StreamBars`, `GetBarWindow`, `TriggerGapFill`;
- `MarketFeedService`: `GetFeedHealth`, `GetActiveSource`;
- `OperationsService`: `GetHealth`, `GetReadiness`;
- `ModelService`: `StreamDecisions`, `StreamPriceEvents`, `StreamVolumeEvents`;
- `SemanticService`: `GetTerm`, `ListTerms`, `GetSemanticContract`.

The four streams relevant to initial DSE evidence are server-streaming RPCs:

| RPC | Published information | Symbol filter | Limit |
|---|---|---|---|
| `IngestionService.StreamBars` | `Bar` with OHLCV and quality/lineage | `symbols[]` | `max_bars` |
| `ModelService.StreamDecisions` | adaptive `DecisionEvent` or typed skip | `symbols[]` | `max_events` |
| `ModelService.StreamPriceEvents` | Price Engine `PriceEvent` | `symbols[]` | `max_events` |
| `ModelService.StreamVolumeEvents` | Volume Engine `VolumeEvent` | `symbols[]` | `max_events` |

An empty symbol filter means all configured symbols. `StreamBars` can request finalized bars only. DSE phase input requires an approved observation definition; the likely initial input is a finalized `Bar`, because `Bar` contains numeric OHLC data while `PriceEvent` contains Price Engine interpretation rather than a raw price scalar.

### 5.2 Actual event facts

`Bar` carries `symbol`, `instrument_id`, interval start/end as Unix milliseconds, OHLC, integer volume, provider `source_timestamp`, receipt time, source, quality, final/backfill/source-transition flags, data age, and `market_snapshot_id`.

`DecisionEvent` carries `event_id`, reserved `signal_id`, `decision_id`, `symbol`, bar interval start, `market_snapshot_id`, provider timestamp, `accepted_sequence`, receive/complete/latency timing, model/schema versions, state hashes, and one outcome:

- `Decision`: BUY, SELL, or HOLD plus confidence, adaptive coefficients, path direction, model status, emitter position, rule path, and evidence scalars; or
- `Skip`: typed reason, detail, and model status.

HOLD is an upstream adaptive decision. A skip is not a HOLD and carries no side.

`PriceEvent` carries `event_id`, `symbol`, bar interval start, `market_snapshot_id`, provider timestamp, `accepted_sequence`, latency, status, emission/domain/projection flags, and Price Engine emission, skip, and cockpit information. Emission includes categorical color, trajectory phase, turning tendency, confidence/domain/stability states, current/projected direction, reason codes, and numerical stability diagnostics. Price color is explicitly not BUY/SELL/HOLD.

`VolumeEvent` carries `event_id`, `symbol`, interval start/end, `market_snapshot_id`, provider timestamp, `accepted_sequence`, latency, status, emitted/reason state, and emission or skip. Its quantities (`v_raw`, `v_n`, `v1`, `v2`, interval mean, and predicted next value) carry explicit availability status so numeric zero is distinct from unavailable. Indicator, transition, phase, confidence, and domain state are Volume semantics, not trade actions.

All three model events expose the same primary bar-lineage tuple: entity `symbol`, `market_snapshot_id`, and interval time. Sequence and event identifiers are stream-specific and must not be treated as a universal join key.

### 5.3 Observed stream behavior

**Existing repository fact**

`Fin_FeedSat_1/internal/server/server.go`, `decision.go`, `price.go`, and `volume.go` show these behaviors:

- each request can filter a set of normalized symbols;
- each model stream first sends the last cached event per matching symbol, then live events;
- replay/live overlap is deduplicated by `event_id` within that server call;
- only one latest model event per symbol is retained for reconnect replay;
- raw bar streaming first sends the available in-memory window per selected symbol, then live bars;
- stream cancellation and send errors terminate that RPC;
- model publication uses bounded subscriber channels and does not block science when a subscriber is slow; a full subscriber buffer is logged and the event is dropped for that subscriber.

`Fin_FeedSat_1/internal/modelhost/host.go` creates one worker and independent Adaptive and Volume engines per configured symbol, plus an optional Price engine. A symbol worker processes accepted bars serially. Cross-symbol processing and the three published model streams do not establish one global causal order.

### 5.4 Existing consumer evidence

The in-repository consumer is `Fin_FeedSat_1/cmd/fin-feedsat-ingest-client/main.go`. It:

- creates a native Go gRPC connection with `grpc.NewClient`;
- reuses generated clients from `fin_feedsat_1/gen/fin_feedsat/v1`;
- sends symbol-filtered requests;
- receives server-streamed bars or decisions in a blocking `Recv` loop; and
- treats receive/connect failures as fatal for that CLI invocation.

No reconnect loop is implemented in this CLI.

The prompt identifies a successful Next.js viewer pattern, but that viewer is not present in this repository. Repository documentation states that the Adaptive Pipeline viewer lives in a separate `Fin_FeedSat_1-dashboard` repository, uses its own protobuf copy, and creates a `ModelService` client. Therefore its Next.js server boundary, browser transport, reconnect policy, and exact generated-client reuse cannot be verified from Tramuthus. `DSE_TransSat_1_viewer` is currently empty. This is a recorded evidence gap, not permission to invent a replacement pattern.

## 6. Scientific Ownership Boundary

Fin_FeedSat_1 owns every observation and scientific output it publishes. DSE may retain upstream values and provenance as evidence but must not relabel them as DSE-owned science.

DSE owns only new information created by its transformation, including:

- DSE phase/cycle state produced by the selected established method;
- DSE candidate assessments;
- DSE portfolio state;
- DSE strategy recommendations and decisions;
- DSE readiness, diagnostics, evidence associations, and transition records.

The authoritative DSE output is Decision Strategy state. Phase values, chart geometry, and presentation labels alone are not the DSE product.

## 7. Physical Component Model

### 7.1 DSE_TransSat_1_worker

The worker is the causal runtime and future independent Go module. Its conceptual responsibilities are:

- native Fin_FeedSat_1 gRPC clients and connection lifecycle;
- inbound validation, correlation, deduplication, and per-entity sequencing;
- per-entity input and phase state;
- the established phase solver and phase diagnostics;
- evidence assembly without loss of upstream provenance;
- Strategy Maker and experimental Hop-On/Hop-Off transformation;
- portfolio and active-position state;
- authoritative DSE state, diagnostics, and logs; and
- a future outbound DSE protobuf/gRPC server.

The worker operates without the viewer. It must not depend on a browser session, Next.js process, or viewer acknowledgement to advance science.

### 7.2 DSE_TransSat_1_viewer

The viewer is a non-causal Next.js/TypeScript browser application planned for PC, tablet, mobile, and PWA-friendly use. Its likely deployment boundary is:

```text
Browser/PWA -> HTTPS -> Cloudflare .com -> Railway Next.js
            -> server-side native gRPC -> DSE worker
```

The Next.js server may later bridge worker updates to the browser through SSE or WebSocket. This document does not select either transport. The choice is an implementation concern constrained by reconnection, ordering, fan-out, hosting, and operational evidence.

## 8. DSE End-to-End Information Flow

```text
Fin_FeedSat_1 gRPC
  |-- finalized Bars (raw price observations and lineage)
  |-- DecisionEvents
  |-- PriceEvents
  `-- VolumeEvents
          |
          v
DSE inbound validation, correlation, and per-entity sequencing
          |
          v
Per-Entity Input State
  |                       \
  v                        v
Established Phase Solver   Upstream Price/Volume/Decision evidence
  |                        /
  +-----------------------+
          |
          v
Evidence Assembly -> Strategy Maker -> Strategy Transformation
          |
          v
Portfolio / Active State -> Decision Strategy State
          |
          v
DSE gRPC Server -> Viewer and future consumers
```

No viewer operation appears on the causal path.

## 9. Inbound Information Consumption

**DSE design decision**

The worker will consume existing Fin_FeedSat_1 protobuf/gRPC without changing that contract. Initial design responsibilities are:

1. Observe Fin_FeedSat_1 health/readiness before declaring upstream readiness.
2. Subscribe only to configured entities and explicitly select finalized-bar semantics for phase input.
3. Maintain independent stream lifecycle and health for Bars, Decisions, Price, and Volume.
4. Validate entity, finality, quality, lineage, status, and monotonic per-entity observation progression.
5. Deduplicate reconnect replay and avoid advancing the solver twice for one observation.
6. Correlate sibling evidence primarily by `symbol` and `market_snapshot_id`, with interval time as a checked lineage attribute.
7. Treat absent, maturing, skipped, disabled, late, or failed sibling evidence explicitly; never manufacture an upstream event.
8. Preserve original upstream identifiers and statuses in DSE evidence references.

The current CLI proves native Go gRPC construction and generated-client use, but not resilient reconnection. DSE connection backoff, jitter, resubscription, checkpoints, and stale-state policy remain implementation design work.

## 10. Per-Entity State Model

The worker supports `0..N` configured candidate entities. Each entity has an isolated state partition containing conceptually:

- latest accepted input lineage and observation sequence;
- raw observation history required by the phase solver;
- phase solver internal state and warm-up count;
- latest valid phase, dominant period, and phase movement;
- latest correlated upstream Decision, Price, and Volume evidence;
- evidence completeness/freshness state;
- candidate eligibility and trajectory assessment; and
- entity-local errors, gaps, discontinuity, and diagnostics.

One entity's invalid observation or solver failure must not mutate another entity's state. Cross-entity comparison happens only after eligible entity snapshots are assembled.

Analytical time advances on accepted sequential observations, not merely wall-clock passage. If no applicable observation arrives for two hours, phase state remains at the last accepted observation. The next observation advances state only after interval/finality/quality/gap semantics pass validation. Asynchronous arrival alone does not prove scientific correctness.

## 11. Established Ehlers Phase/Cycle Analysis

**Established/external method**

The phase solver is to use a recognized John Ehlers Hilbert-transform/dominant-cycle phase methodology. Its conceptual path is:

```text
price observation -> smoothing -> detrending -> Hilbert transform
                  -> in-phase I and quadrature Q
                  -> dominant-cycle estimation
                  -> dominant period -> dominant phase
```

Implementation must select and cite a precise authoritative algorithm/version, initialization convention, coefficient set, angle convention, period bounds, and numerical behavior. This document does not define an approximation and does not authorize implementation.

**Firm boundary:** Volume does not modify `I`, `Q`, smoothing, detrending, period estimation, or phase. In particular, no design may hide a formula such as `Q = I * volume_modifier` inside the established solver. Any future volume-modified phase proposal is a separate experimental model requiring an explicit specification and independent validation.

## 12. Phase Warm-Up and Validity

The worker must distinguish at least:

| State | Meaning |
|---|---|
| Process alive | Runtime is executing; no claim about upstream or science |
| Upstream connected | Required gRPC channels are connected; no claim about sufficient data |
| Phase initialized | Solver state exists and has begun accumulation |
| Phase valid | Required sequential history and numerical validity checks pass |
| Strategy ready | Required candidate, evidence, configuration, and portfolio prerequisites pass |

Before validity, phase output is diagnostic only and cannot silently enter strategy as valid evidence. Entity-level diagnostics must conceptually include entity identity, event/bar lineage, smoothed price, dominant period, dominant phase, phase delta/movement, warm-up state, and valid/not-valid state. Final protobuf names are deferred.

Warm-up resets and recovery after gaps, restarts, configuration changes, or entity reactivation require explicit policy and deterministic tests.

## 13. Price / Volume / Phase Evidence Relationship

Phase, Price, Volume, and upstream Decision information are sibling evidence at the Strategy Maker boundary:

```text
Established phase/cycle evidence ---+
Fin_FeedSat_1 Price evidence --------+--> Strategy Maker
Fin_FeedSat_1 Volume evidence -------+
Fin_FeedSat_1 Decision evidence -----+
Portfolio state --------------------+
```

Upstream semantic domains remain intact:

- a raw finalized `Bar` supplies the approved numeric price observation for phase analysis;
- `PriceEvent` supplies Price Engine trajectory and stability interpretation, not raw price and not a trade action;
- `VolumeEvent` supplies independent activity/confirmation evidence, not phase modification and not a trade action;
- `DecisionEvent` supplies Fin_FeedSat_1 adaptive BUY/SELL/HOLD or skip evidence, not a DSE strategy decision.

Evidence assembly must preserve source event IDs, lineage, statuses, and missingness. It must not coerce unavailable quantities to zero or interpret absent events as neutral evidence.

## 14. Strategy Maker

**Experimental strategy**

The Strategy Maker is the boundary where DSE-specific experimentation begins. It consumes immutable snapshots of eligible candidate state, established phase evidence, upstream Price/Volume/Decision evidence, current portfolio state, and versioned strategy configuration.

Its output is a proposed Decision Strategy state with an action, target where applicable, evidence/reason, configuration identity, and input lineage. The transformation must be deterministic for the same ordered inputs, prior portfolio state, and configuration. Diagnostic calculations may explain a result but may not silently alter it.

Strategy configuration is expected eventually to express strategy identity, candidate universe, phase zones, transition conditions, confirmation requirements, portfolio constraints, and rotation mode. No schema or JSON file is authorized here.

## 15. Hop-On / Hop-Off Transformation Model

The initial strategy concept is rotational:

1. assess each eligible candidate trajectory;
2. assess whether the active trajectory remains supportable;
3. compare eligible alternatives under configured evidence rules;
4. recommend holding, entering, exiting, or rotating; and
5. record why the recommendation changed or remained stable.

Conceptual actions include `HOLD`, `HOP_ON`, `HOP_OFF`, and `HOP_FROM_A_TO_B`. These names are **PROPOSED**, not frozen enums.

Exact phase-zone boundaries are open. Earlier concepts around 0, 90, 180, and 270 degrees are inconsistent and are not authoritative. Zone definitions, angle orientation, wrap behavior, hysteresis, confirmation duration, and transition thresholds belong to versioned strategy configuration and require experimental validation.

## 16. Candidate State

Candidate State is analytical state for one potential entity. It includes conceptually:

- scientific entity ID;
- observation and evidence lineage;
- phase, phase movement, period, validity, and freshness;
- upstream Price, Volume, and Decision evidence with status;
- trajectory assessment;
- eligibility and exclusion reason; and
- comparable strategy features derived under one strategy version.

Candidate State does not say what the portfolio owns and does not itself command a transition.

## 17. Portfolio State

Portfolio State records the strategy's current conceptual occupancy independently of candidate analysis:

- active entity, or no active entity;
- entry/activation lineage and context;
- active-since observation state;
- previous active entity;
- pending transition state, if the approved model later requires one; and
- strategy interaction mode.

Portfolio State is DSE strategy state, not proof of a broker position. Reconciliation with external execution is outside the initial model.

## 18. Decision Strategy State

Decision Strategy State records what the Strategy Maker currently recommends:

- proposed action;
- source and target entities where applicable;
- decision lineage and strategy/configuration identity;
- evidence snapshot references and reason;
- readiness/validity status; and
- relationship to the previous recommendation and portfolio state.

Candidate State, Portfolio State, and Decision Strategy State must remain separate types conceptually and in any later contract design.

## 19. DSE Authoritative Outputs

DSE-owned authoritative information includes:

- per-entity DSE analytical validity and phase/cycle state where publication is approved;
- candidate eligibility and strategy assessment;
- portfolio/active state maintained by DSE;
- current Decision Strategy recommendation;
- strategy transition records and evidence references;
- worker lifecycle, upstream connectivity, and strategy readiness; and
- diagnostics required to reproduce and evaluate decisions.

Publication must distinguish authoritative decision state from diagnostics and presentation aliases.

## 20. Outbound Publication Model

The worker will eventually host a DSE-owned protobuf/gRPC server. Its conceptual contract must support:

- current snapshots and streaming updates;
- entity analytical and phase/cycle diagnostics where appropriate;
- candidate strategy state;
- active portfolio state;
- Decision Strategy action and target;
- evidence/reason and source lineage;
- readiness and lifecycle state; and
- version identities needed for reproducibility.

Exact package, services, RPCs, messages, fields, enums, replay depth, persistence, and compatibility policy are deferred to a dedicated interface-design task. No Fin_FeedSat_1 protobuf change is required or authorized.

Slow or failed downstream subscribers must not block analytical or strategy processing. The future contract must define snapshot/replay and resynchronization semantics rather than relying on an unbounded live stream.

## 21. Viewer Boundary

The viewer may subscribe, cache for presentation, filter, sort, animate, explain, and collect non-authoritative human interaction. It must not:

- calculate dominant phase or period;
- modify Hilbert `I` or `Q`;
- derive candidate rankings;
- decide Hop-On or Hop-Off;
- determine portfolio ownership;
- calculate Decision Strategy; or
- become the system of record for DSE state.

The viewer renders worker-produced authoritative state. Its absence cannot change a worker result.

## 22. Four-Wave Visualization Model

A current viewer configuration may present four simultaneous trajectories:

- Alpha Wave;
- Beta Signal;
- Gamma Trend; and
- Delta Cycle.

The primary visual questions are: Where am I? What alternatives exist? What does the engine recommend? It should distinguish current/active trajectory, alternatives, proposed target, Active Position, Engine Signal, Rotation Mode, and Portfolio Wave Vector.

The proposed target must never be visually confusable with the currently active position. Four is a presentation selection over `0..N` worker candidates, not a worker limit.

## 23. Visual Identity vs Entity Identity

Scientific identity and presentation identity are separate:

```text
entity_id = AAPL       # scientific identity
visual_role = Gamma    # session/configuration presentation assignment
```

Gamma must never replace AAPL in lineage, state keys, diagnostics, or decisions. A different session may map Gamma to another entity without changing the entity's scientific state.

## 24. Live / Learn / Game Presentation Modes

LIVE, LEARN, and GAME are presentation/application modes over the same worker-owned science:

- **LIVE:** show actual analytical, portfolio, and recommendation state; later expose approved Auto/Manual interaction.
- **LEARN:** replay or slow presentation, explain state, and progressively disclose mathematics.
- **GAME:** optionally conceal a recommendation, record a human selection, and later compare human choice, DSE choice, and subsequent outcome.

These modes must not instantiate different scientific engines or change authoritative results for identical inputs and configuration.

## 25. Human Explanation Layers

The primary interface should communicate current ride, trajectory direction, strengthening/weakening, alternatives, HOLD, and HOP without requiring Hilbert-transform knowledge.

Deeper layers may expose phase, dominant period, movement, upstream evidence, readiness, lineage, and eventually detailed mathematical diagnostics. The governing principle is: **Hide complexity, not information.** Explanations must be traceable to worker-published evidence rather than independently recomputed in the viewer.

## 26. Auto vs Manual Strategy Interaction

Auto Engine and Manual Hop are future interaction modes, not approved execution semantics.

- A **DSE recommendation** is the worker's authoritative strategy information.
- A **human selection** is an external interaction that may later be recorded separately.
- An **actual trade execution** is a broker/execution concern and is not implied by either item.

Future design must specify who may accept, reject, or override a recommendation; how that affects conceptual Portfolio State; and how external execution acknowledgement is represented. This document makes no such choice.

## 27. Lifecycle / Health / Failure Isolation

Lifecycle must expose component and per-entity status rather than one ambiguous healthy flag. At minimum it must distinguish process liveness, upstream connectivity, stream status, entity warm-up, phase validity, evidence readiness, strategy readiness, and outbound service readiness.

Failure boundaries:

- viewer unavailable: worker science and publication continue;
- slow downstream subscriber: isolate or drop/disconnect according to a bounded policy; never block science;
- one entity invalid or failed: mark that entity unavailable/degraded without corrupting other entities;
- Fin_FeedSat_1 disconnected: preserve committed state, mark evidence stale/degraded, reconnect under a later policy, and fabricate no events;
- replay duplicate: do not advance analytical state twice;
- observed gap or regression: stop that entity's valid advancement until the approved recovery policy is satisfied;
- one optional evidence stream unavailable: expose partial evidence and apply only an explicitly approved readiness policy;
- worker restart: do not claim continuity unless state and input position are restored under an approved persistence/replay design.

The current Fin_FeedSat_1 model streams replay only the latest event per symbol, so they cannot alone reconstruct DSE history after extended downtime.

## 28. Cardinality and Multi-Entity Behavior

The process model supports `0..N` candidates. A zero-candidate configuration can be alive but cannot be strategy-ready. Each entity advances independently in causal observation order. Cross-entity ranking consumes stable candidate snapshots at an explicitly defined comparison point; arrival order must not bias rankings accidentally.

The initial four-entity experiment does not justify premature worker-pool complexity. Start with a simple, inspectable concurrency model that preserves per-entity serialization and isolate scale changes behind measured requirements.

## 29. Current Direct Tramuthus Topology

```text
Fin_FeedSat_1
    | existing protobuf/gRPC
    v
DSE_TransSat_1_worker
    | future DSE protobuf/gRPC
    v
DSE_TransSat_1_viewer
```

This direct topology is the proving environment. DSE must consume the existing contract unchanged.

## 30. Future HACCAM Topology

```text
Fin_FeedSat_1 -> HACCAM -> DSE_TransSat_1 -> HACCAM
                                            |-> viewer
                                            `-> future consumers
```

HACCAM may later govern identity, routing, policy, lifecycle, and delivery metadata. It must not redefine DSE phase or strategy science.

A future equivalence test is required: given the same ordered Fin_FeedSat_1 information sequence and DSE configuration, direct and federated delivery must produce behaviorally equivalent DSE analytical and strategy results, except for explicitly modeled transport/lifecycle metadata.

## 31. Proposed DSE Process Decomposition

The following identifiers are **PROPOSED**:

| Process | Name | Responsibility | Owned state/output |
|---|---|---|---|
| DSE-01 | Inbound Information Consumption | Connect, subscribe, validate, deduplicate, correlate, and sequence Fin_FeedSat_1 information | Stream lifecycle, checkpoints, upstream lineage |
| DSE-02 | Per-Entity State Coordination | Isolate and coordinate entity input and analytical state | Entity partitions and readiness |
| DSE-03 | Phase / Cycle Analysis | Apply the selected established Ehlers method | Solver state and phase diagnostics |
| DSE-04 | Evidence Assembly | Form provenance-preserving evidence snapshots | Correlated candidate evidence |
| DSE-05 | Strategy Maker | Apply versioned experimental decision rules | Proposed strategy result and explanation |
| DSE-06 | Portfolio State | Maintain conceptual active/candidate/transition state | Authoritative DSE portfolio state |
| DSE-07 | Decision Strategy Publication | Serve DSE-owned snapshots and streams | Published Decision Strategy information |
| DSE-08 | Health / Lifecycle | Report liveness, connectivity, warm-up, validity, readiness, and service state | Lifecycle state |
| DSE-09 | Diagnostics / Evidence | Record phase diagnostics, input lineage, transitions, and proving evidence | Reproducibility and audit records |

These are logical process boundaries, not a requirement for nine binaries, goroutine pools, or deployment units.

## 32. Explicit Non-Responsibilities

DSE_TransSat_1 does not initially own or implement:

- Fin_FeedSat_1 science;
- Fin_FeedSat_1 Price mathematics;
- Fin_FeedSat_1 Volume mathematics;
- Fin_FeedSat_1 protobuf or generated code;
- market data acquisition or provider recovery;
- broker, order, capital-allocation, or trade execution;
- public web hosting concerns;
- Cloudflare routing semantics;
- Railway deployment semantics;
- HACCAM routing/governance implementation;
- viewer-side scientific calculations;
- a new phase approximation; or
- hidden fusion that changes upstream or established phase semantics.

## 33. Open Decisions / TBDs

1. Select and cite the exact Ehlers algorithm/version, trusted reference implementation or dataset, coefficients, initialization, phase convention, and numerical tolerances.
2. Define the approved phase input price from finalized `Bar` fields and the interval/quality/backfill/source-transition rules.
3. Define DSE reconnect, backoff, deduplication, checkpoint, persistence, replay, and gap-recovery policy. Latest-event replay is insufficient for full reconstruction.
4. Define evidence correlation timing: required versus optional streams, lateness window, incomplete snapshots, and when a candidate snapshot becomes comparable.
5. Define phase warm-up length, valid/invalid transitions, stale-state behavior, and reset semantics.
6. Define experimentally validated phase zones, direction convention, wrap handling, hysteresis, confirmation, and Hop thresholds.
7. Define candidate ranking, tie-breaking, eligibility, portfolio constraints, and deterministic transition rules.
8. Define the future DSE protobuf package, service, messages, RPCs, versioning, snapshot/replay, and compatibility policy.
9. Decide whether the worker directly reuses the existing generated Go package through a module dependency or generates a client from the unchanged upstream contract in a controlled build process.
10. Inspect the external `Fin_FeedSat_1-dashboard` repository before claiming its Next.js-to-browser bridge or reconnect design as a proven reusable pattern.
11. Resolve the risk that external viewer documentation says it maintains its own proto copy; contract drift controls are not visible in this repository.
12. Define Auto/Manual authority, human override recording, and the boundary to any future execution system.
13. Select SSE or WebSocket for the future browser bridge using deployment and operational evidence.
14. Define durable diagnostic/audit storage and retention.
15. Determine whether `signal_id`, currently present but not populated in observed DecisionEvent mapping, has any DSE lineage role.

## 34. Initial Validation Strategy

No validation implementation is authorized by this document. Recommended progression:

1. Prove a DSE client can consume unchanged Fin_FeedSat_1 bar, Decision, Price, and Volume streams.
2. Validate the selected Ehlers solver independently against a trusted implementation or reference dataset.
3. Verify independent analytical state and failure isolation per entity.
4. Emit and inspect raw phase diagnostics before enabling strategy decisions.
5. Observe the four current proving entities without encoding four as a limit.
6. Verify warm-up, validity, stale, gap, reconnect, and reset behavior.
7. Add deterministic Strategy Maker logic under an identified configuration.
8. Log every proposed HOLD/HOP transition with complete evidence lineage.
9. Compare recommendations with subsequent observations under a predeclared evaluation method.
10. Introduce and verify portfolio-state transitions independently of broker execution.
11. Publish authoritative DSE state to the viewer boundary.
12. Validate desktop, tablet, and mobile presentation without permitting viewer-side science.
13. Demonstrate direct-versus-HACCAM behavioral equivalence when a federated path exists.
14. Only later evaluate capital allocation or execution integration.

Concurrency optimization is deferred until measurements show a need.

## 35. Implementation Preconditions

Implementation may begin only after human review authorizes a scoped increment and resolves or explicitly defers its blocking decisions. At minimum:

- this process model is accepted or revised;
- the precise phase method and validation oracle are approved;
- observation, ordering, correlation, reconnect, and gap semantics are specified;
- warm-up and readiness rules are specified;
- an initial deterministic strategy experiment and configuration boundary are approved;
- state ownership and persistence expectations are decided for the increment;
- the inbound reuse/build approach is selected without changing Fin_FeedSat_1;
- the DSE outbound contract is designed separately before protobuf creation; and
- acceptance tests and non-regression checks are defined.

No implementation precondition authorizes modifications to Fin_FeedSat_1. Any discovered upstream contract deficiency remains an open issue for separate governance.

## 36. Change Log

| Date | Version | Status | Change |
|---|---|---|---|
| 2026-09-10 | V0.1 | PROPOSED FOR HUMAN REVIEW | Initial repository-grounded DSE_TransSat_1 process model; implementation not authorized. |
