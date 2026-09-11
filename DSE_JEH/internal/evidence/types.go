// Package evidence defines storage-neutral records for the DSE_JEH CSV proving stage.
// Inputs are authoritative persisted phase rows; outputs are strategy and decision evidence.
// Configuration identifies the source experiment and candidate strategy version.
// No type in this package calculates Ehlers phase or assigns scientific meaning to a zone.
package evidence

import "time"

const StrategyVersion = "CANDIDATE_ZONES_V0.1"

type Zone string

const (
	ZoneHopOn        Zone = "HOP_ON"
	ZoneMomentumHold Zone = "MOMENTUM_HOLD"
	ZoneHopOff       Zone = "HOP_OFF"
	ZoneDisregard    Zone = "DISREGARD"
)

type EventType string

const (
	EventPhaseBecameObservable EventType = "PHASE_BECAME_OBSERVABLE"
	EventZoneEntered           EventType = "ZONE_ENTERED"
	EventZoneExited            EventType = "ZONE_EXITED"
	EventHopOnCandidate        EventType = "HOP_ON_CANDIDATE"
	EventHopOffCandidate       EventType = "HOP_OFF_CANDIDATE"
	EventDecision              EventType = "DECISION_EVENT"
)

// PhaseEvidence is one admitted CSV source row. PhaseAngleDegrees is nil only
// when the source explicitly reports non-observable initialization evidence.
type PhaseEvidence struct {
	SourceRecordNumber int       `json:"source_record_number"`
	CollectionRunID    string    `json:"collection_run_id"`
	PartitionID        string    `json:"partition_id"`
	Symbol             string    `json:"symbol"`
	GeneratorSequence  int64     `json:"generator_sequence_no"`
	SeriesSize         int       `json:"series_size"`
	SolverName         string    `json:"solver_name"`
	SolverVersion      string    `json:"solver_version"`
	InputSeriesType    string    `json:"input_series_type"`
	PhaseAngleDegrees  *float64  `json:"phase_angle_degrees"`
	PhaseAngleText     string    `json:"phase_angle_text"`
	PhaseObservable    bool      `json:"phase_observable"`
	ValidityState      string    `json:"validity_state"`
	AnalysisCreatedAt  time.Time `json:"analysis_created_at"`
}

// Identity returns the deterministic source-evidence identity used for duplicate checks and event lineage.
func (input PhaseEvidence) Identity() string {
	return input.CollectionRunID + "|" + input.Symbol + "|" + input.PhaseAngleText + "|" +
		input.SolverName + "|" + input.SolverVersion + "|" + input.InputSeriesType + "|" +
		formatInt(input.GeneratorSequence) + "|" + formatInt(int64(input.SeriesSize))
}

// formatInt formats identity components without introducing locale-dependent text.
func formatInt(value int64) string {
	if value == 0 {
		return "0"
	}
	negative := value < 0
	if negative {
		value = -value
	}
	var buffer [20]byte
	position := len(buffer)
	for value > 0 {
		position--
		buffer[position] = byte('0' + value%10)
		value /= 10
	}
	if negative {
		position--
		buffer[position] = '-'
	}
	return string(buffer[position:])
}

// StrategyEvent records a state transition or candidate interpretation. Prior
// values are pointers so a legitimate zero-degree phase is never an init sentinel.
type StrategyEvent struct {
	EventID              string    `json:"event_id"`
	Type                 EventType `json:"event_type"`
	StrategyVersion      string    `json:"strategy_version"`
	SourceIdentity       string    `json:"source_identity"`
	CollectionRunID      string    `json:"collection_run_id"`
	PartitionID          string    `json:"partition_id"`
	Symbol               string    `json:"symbol"`
	GeneratorSequence    int64     `json:"generator_sequence_no"`
	PhaseAngleDegrees    float64   `json:"phase_angle_degrees"`
	PreviousPhaseDegrees *float64  `json:"previous_phase_degrees,omitempty"`
	CurrentZone          Zone      `json:"current_zone"`
	PreviousZone         *Zone     `json:"previous_zone,omitempty"`
	SolverName           string    `json:"solver_name"`
	SolverVersion        string    `json:"solver_version"`
	ValidityState        string    `json:"validity_state"`
	DecisionKind         string    `json:"decision_kind,omitempty"`
	RankingMode          string    `json:"ranking_mode,omitempty"`
}

// EntityState is the current observable strategy state for one independent symbol.
type EntityState struct {
	Symbol               string  `json:"symbol"`
	HasPreviousPhase     bool    `json:"has_previous_phase"`
	PreviousPhaseDegrees float64 `json:"previous_phase_degrees,omitempty"`
	CurrentPhaseDegrees  float64 `json:"current_phase_degrees"`
	PreviousZone         Zone    `json:"previous_zone,omitempty"`
	CurrentZone          Zone    `json:"current_zone"`
	GeneratorSequence    int64   `json:"generator_sequence_no"`
	SourceIdentity       string  `json:"source_identity"`
	SolverName           string  `json:"solver_name"`
	SolverVersion        string  `json:"solver_version"`
	ValidityState        string  `json:"validity_state"`
}

// RunResult is the deterministic output of one complete proving pass.
type RunResult struct {
	Inputs              []PhaseEvidence        `json:"inputs"`
	StrategyEvents      []StrategyEvent        `json:"strategy_events"`
	DecisionEvents      []StrategyEvent        `json:"decision_events"`
	FinalStates         map[string]EntityState `json:"final_states"`
	HopOnCandidates     []string               `json:"hop_on_candidates"`
	RankingMode         string                 `json:"ranking_mode"`
	DuplicateInputCount int                    `json:"duplicate_input_count"`
}
