// Package strategy implements candidate zone membership and entity-local transitions.
// Inputs are admitted phase evidence; outputs are strategy and decision evidence.
// Candidate boundaries and strategy version are explicit configuration semantics.
// This package does not calculate phase, velocity, profitability, orders, or allocation.
package strategy

import (
	"fmt"
	"sort"

	"tramuthus/dse-jeh/internal/evidence"
)

const ProvingOnlyOrder = "PROVING_ONLY_ORDER"

// Classify applies the candidate half-open zones to an authoritative normalized angle.
func Classify(phase float64) (evidence.Zone, error) {
	if phase < 0 || phase >= 360 {
		return "", fmt.Errorf("phase must be normalized to [0,360): %v", phase)
	}
	switch {
	case phase < 90:
		return evidence.ZoneMomentumHold, nil
	case phase < 180:
		return evidence.ZoneHopOff, nil
	case phase < 270:
		return evidence.ZoneDisregard, nil
	default:
		return evidence.ZoneHopOn, nil
	}
}

type entityRuntime struct {
	state        evidence.EntityState
	lastSequence int64
	seen         map[int64]string
}

type Engine struct {
	entities       map[string]*entityRuntime
	strategyEvents []evidence.StrategyEvent
	decisionEvents []evidence.StrategyEvent
	duplicates     int
}

// NewEngine creates empty independent per-entity proving state.
func NewEngine() *Engine { return &Engine{entities: make(map[string]*entityRuntime)} }

// Apply advances only the addressed entity. Exact duplicate inputs are ignored;
// conflicting duplicates and non-increasing new positions are rejected.
func (engine *Engine) Apply(input evidence.PhaseEvidence) error {
	runtime := engine.entities[input.Symbol]
	if runtime == nil {
		runtime = &entityRuntime{seen: make(map[int64]string)}
		engine.entities[input.Symbol] = runtime
	}
	identity := input.Identity()
	if previousIdentity, exists := runtime.seen[input.GeneratorSequence]; exists {
		if previousIdentity != identity {
			return fmt.Errorf("conflicting duplicate for %s sequence %d", input.Symbol, input.GeneratorSequence)
		}
		engine.duplicates++
		return nil
	}
	if input.GeneratorSequence <= runtime.lastSequence {
		return fmt.Errorf("non-increasing sequence for %s: %d after %d", input.Symbol, input.GeneratorSequence, runtime.lastSequence)
	}
	runtime.seen[input.GeneratorSequence] = identity
	runtime.lastSequence = input.GeneratorSequence
	if !input.PhaseObservable {
		return nil
	}
	if input.PhaseAngleDegrees == nil {
		return fmt.Errorf("observable input %s sequence %d has no phase", input.Symbol, input.GeneratorSequence)
	}
	zone, err := Classify(*input.PhaseAngleDegrees)
	if err != nil {
		return err
	}
	if runtime.state.ValidityState == "" {
		runtime.state = stateFrom(input, zone, nil)
		engine.emit(input, evidence.EventPhaseBecameObservable, zone, nil, nil, "")
		engine.emit(input, evidence.EventZoneEntered, zone, nil, nil, "")
		engine.emitCandidate(input, zone, nil, nil)
		return nil
	}
	previousPhase := runtime.state.CurrentPhaseDegrees
	previousZone := runtime.state.CurrentZone
	runtime.state = stateFrom(input, zone, &runtime.state)
	if zone == previousZone {
		return nil
	}
	engine.emit(input, evidence.EventZoneExited, zone, &previousZone, &previousPhase, "")
	engine.emit(input, evidence.EventZoneEntered, zone, &previousZone, &previousPhase, "")
	engine.emitCandidate(input, zone, &previousZone, &previousPhase)
	return nil
}

// stateFrom creates current entity state while preserving explicit prior observable state.
func stateFrom(input evidence.PhaseEvidence, zone evidence.Zone, previous *evidence.EntityState) evidence.EntityState {
	state := evidence.EntityState{
		Symbol: input.Symbol, CurrentPhaseDegrees: *input.PhaseAngleDegrees, CurrentZone: zone,
		GeneratorSequence: input.GeneratorSequence, SourceIdentity: input.Identity(),
		SolverName: input.SolverName, SolverVersion: input.SolverVersion, ValidityState: input.ValidityState,
	}
	if previous != nil {
		state.HasPreviousPhase = true
		state.PreviousPhaseDegrees = previous.CurrentPhaseDegrees
		state.PreviousZone = previous.CurrentZone
	}
	return state
}

// emitCandidate records candidate and decision evidence only on entry into a candidate zone.
func (engine *Engine) emitCandidate(input evidence.PhaseEvidence, zone evidence.Zone, previousZone *evidence.Zone, previousPhase *float64) {
	var candidateType evidence.EventType
	var decisionKind string
	switch zone {
	case evidence.ZoneHopOn:
		candidateType, decisionKind = evidence.EventHopOnCandidate, "HOP_ON_CANDIDATE_RECORDED"
	case evidence.ZoneHopOff:
		candidateType, decisionKind = evidence.EventHopOffCandidate, "HOP_OFF_CANDIDATE_RECORDED"
	default:
		return
	}
	engine.emit(input, candidateType, zone, previousZone, previousPhase, "")
	engine.emit(input, evidence.EventDecision, zone, previousZone, previousPhase, decisionKind)
}

// emit creates deterministic event evidence linked to one admitted CSV source row.
func (engine *Engine) emit(input evidence.PhaseEvidence, eventType evidence.EventType, zone evidence.Zone, previousZone *evidence.Zone, previousPhase *float64, decisionKind string) {
	event := evidence.StrategyEvent{
		EventID: input.Identity() + "|" + string(eventType), Type: eventType,
		StrategyVersion: evidence.StrategyVersion, SourceIdentity: input.Identity(),
		CollectionRunID: input.CollectionRunID, PartitionID: input.PartitionID, Symbol: input.Symbol,
		GeneratorSequence: input.GeneratorSequence, PhaseAngleDegrees: *input.PhaseAngleDegrees,
		PreviousPhaseDegrees: previousPhase, CurrentZone: zone, PreviousZone: previousZone,
		SolverName: input.SolverName, SolverVersion: input.SolverVersion, ValidityState: input.ValidityState,
		DecisionKind: decisionKind,
	}
	if eventType == evidence.EventDecision {
		event.RankingMode = ProvingOnlyOrder
		engine.decisionEvents = append(engine.decisionEvents, event)
	} else {
		engine.strategyEvents = append(engine.strategyEvents, event)
	}
}

// Result returns copied deterministic evidence and a symbol-ascending current
// HOP_ON candidate set. That ordering is explicitly non-scientific.
func (engine *Engine) Result(inputs []evidence.PhaseEvidence) evidence.RunResult {
	states := make(map[string]evidence.EntityState, len(engine.entities))
	var candidates []string
	for symbol, runtime := range engine.entities {
		if runtime.state.ValidityState == "" {
			continue
		}
		states[symbol] = runtime.state
		if runtime.state.CurrentZone == evidence.ZoneHopOn {
			candidates = append(candidates, symbol)
		}
	}
	sort.Strings(candidates)
	return evidence.RunResult{
		Inputs:         append([]evidence.PhaseEvidence(nil), inputs...),
		StrategyEvents: append([]evidence.StrategyEvent(nil), engine.strategyEvents...),
		DecisionEvents: append([]evidence.StrategyEvent(nil), engine.decisionEvents...),
		FinalStates:    states, HopOnCandidates: candidates, RankingMode: ProvingOnlyOrder,
		DuplicateInputCount: engine.duplicates,
	}
}
