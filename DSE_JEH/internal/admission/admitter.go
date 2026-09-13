package admission

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"

	dsejehv1 "tramuthus/dse-jeh-transsat-1/gen/dse_jeh/v1"
	"tramuthus/dse-jeh-transsat-1/internal/evidence"
)

type entityState struct {
	lastSequence uint64
	identities   map[uint64]string
	seenEvents   map[string]uint64
}

type Admitter struct {
	mu       sync.Mutex
	entities map[string]*entityState
}

func New() *Admitter {
	return &Admitter{entities: make(map[string]*entityState)}
}

func (admitter *Admitter) Admit(event *dsejehv1.BarEvent) *dsejehv1.BarAdmissionEvidence {
	admitter.mu.Lock()
	defer admitter.mu.Unlock()

	result := &dsejehv1.BarAdmissionEvidence{
		BarEventId: event.GetEventId(),
		EntityId:   event.GetEntityId(),
		Status:     dsejehv1.BarAdmissionStatus_BAR_ADMISSION_STATUS_INVALID,
	}
	if event.GetProvenance() != nil {
		result.EntitySequence = event.GetProvenance().GetEntitySequence()
	}
	result.AdmissionId = evidence.ID("admission", result.BarEventId)
	if reason := validate(event); reason != "" {
		result.Reason = reason
		return result
	}

	scope := event.GetProvenance().GetEntitySequenceScope()
	state := admitter.entities[scope]
	if state == nil {
		state = &entityState{identities: make(map[uint64]string), seenEvents: make(map[string]uint64)}
		admitter.entities[scope] = state
	}
	sequence := event.GetProvenance().GetEntitySequence()
	if existing, ok := state.identities[sequence]; ok {
		if existing == event.GetEventId() {
			result.Status = dsejehv1.BarAdmissionStatus_BAR_ADMISSION_STATUS_DUPLICATE
			result.Reason = "event identity was already admitted at this entity sequence"
		} else {
			result.Status = dsejehv1.BarAdmissionStatus_BAR_ADMISSION_STATUS_CONFLICT
			result.Reason = "different event identity already occupies this entity sequence"
		}
		return result
	}
	if previousSequence, ok := state.seenEvents[event.GetEventId()]; ok {
		result.Status = dsejehv1.BarAdmissionStatus_BAR_ADMISSION_STATUS_DUPLICATE
		result.Reason = fmt.Sprintf("event identity was already admitted at entity sequence %d", previousSequence)
		return result
	}
	if state.lastSequence > 0 && sequence < state.lastSequence {
		result.Status = dsejehv1.BarAdmissionStatus_BAR_ADMISSION_STATUS_OUT_OF_ORDER
		result.Reason = fmt.Sprintf("sequence %d is before admitted sequence %d", sequence, state.lastSequence)
		return result
	}
	if sequence != state.lastSequence+1 {
		result.Status = dsejehv1.BarAdmissionStatus_BAR_ADMISSION_STATUS_GAP
		result.Reason = fmt.Sprintf("sequence %d does not follow admitted sequence %d", sequence, state.lastSequence)
		return result
	}
	state.lastSequence = sequence
	state.identities[sequence] = event.GetEventId()
	state.seenEvents[event.GetEventId()] = sequence
	result.Status = dsejehv1.BarAdmissionStatus_BAR_ADMISSION_STATUS_ADMITTED
	return result
}

func validate(event *dsejehv1.BarEvent) string {
	if event == nil {
		return "bar event is nil"
	}
	if strings.TrimSpace(event.GetEventId()) == "" || strings.TrimSpace(event.GetEntityId()) == "" {
		return "event_id and entity_id are required"
	}
	if event.GetProvenance() == nil || event.GetProvenance().GetEntitySequenceScope() == "" || event.GetProvenance().GetEntitySequence() == 0 {
		return "provenance, entity sequence scope, and positive entity sequence are required"
	}
	if math.IsNaN(event.GetHigh()) || math.IsInf(event.GetHigh(), 0) || math.IsNaN(event.GetLow()) || math.IsInf(event.GetLow(), 0) {
		return "high and low must be finite"
	}
	if event.GetHigh() < event.GetLow() {
		return "high must be greater than or equal to low"
	}
	for name, value := range map[string]float64{"open": event.GetOpen(), "close": event.GetClose()} {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return name + " must be finite: " + strconv.FormatFloat(value, 'g', -1, 64)
		}
	}
	return ""
}
