package admission

import (
	"testing"

	dsejehv1 "tramuthus/dse-jeh-transsat-1/gen/dse_jeh/v1"
)

func TestAdmissionStatuses(t *testing.T) {
	admitter := New()
	first := event("one", 1, 10)
	if got := admitter.Admit(first).GetStatus(); got != dsejehv1.BarAdmissionStatus_BAR_ADMISSION_STATUS_ADMITTED {
		t.Fatalf("first status = %v", got)
	}
	if got := admitter.Admit(first).GetStatus(); got != dsejehv1.BarAdmissionStatus_BAR_ADMISSION_STATUS_DUPLICATE {
		t.Fatalf("duplicate status = %v", got)
	}
	if got := admitter.Admit(event("conflict", 1, 11)).GetStatus(); got != dsejehv1.BarAdmissionStatus_BAR_ADMISSION_STATUS_CONFLICT {
		t.Fatalf("conflict status = %v", got)
	}
	if got := admitter.Admit(event("gap", 3, 12)).GetStatus(); got != dsejehv1.BarAdmissionStatus_BAR_ADMISSION_STATUS_GAP {
		t.Fatalf("gap status = %v", got)
	}
	if got := admitter.Admit(event("two", 2, 12)).GetStatus(); got != dsejehv1.BarAdmissionStatus_BAR_ADMISSION_STATUS_ADMITTED {
		t.Fatalf("second status = %v", got)
	}
	if got := admitter.Admit(event("two", 3, 12)).GetStatus(); got != dsejehv1.BarAdmissionStatus_BAR_ADMISSION_STATUS_DUPLICATE {
		t.Fatalf("reconnect duplicate status = %v", got)
	}
	if got := admitter.Admit(event("late", 1, 12)).GetStatus(); got != dsejehv1.BarAdmissionStatus_BAR_ADMISSION_STATUS_CONFLICT {
		t.Fatalf("late status = %v", got)
	}
}

func event(id string, sequence uint64, high float64) *dsejehv1.BarEvent {
	return &dsejehv1.BarEvent{
		EventId:  id,
		EntityId: "AAPL",
		High:     high,
		Low:      9,
		Provenance: &dsejehv1.SourceProvenance{
			EntitySequence:      sequence,
			EntitySequenceScope: "run|AAPL",
		},
	}
}
