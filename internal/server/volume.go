package server

import (
	finfeedsatv1 "fin_feedsat_1/gen/fin_feedsat/v1"
	"fin_feedsat_1/internal/config"
	"fin_feedsat_1/internal/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// StreamVolumeEvents publishes Host-produced VolumeEvents. It does not ingest
// Bars or run Volume science. A slow or cancelled client cannot block ModelHost.
func (s *Server) StreamVolumeEvents(request *finfeedsatv1.StreamVolumeEventsRequest, stream finfeedsatv1.ModelService_StreamVolumeEventsServer) error {
	if s.volumes == nil {
		return status.Error(codes.FailedPrecondition, "volume is not wired")
	}
	wanted, err := normalizeSymbols(request.GetSymbols())
	if err != nil {
		return err
	}

	id, events := s.volumes.SubscribeVolumeEvents(config.SubscriberQueue)
	defer s.volumes.UnsubscribeVolumeEvents(id)

	var sent uint32
	seen := make(map[string]struct{})
	for _, ev := range s.volumes.LastVolumeEvents() {
		if len(wanted) > 0 && !wanted[ev.Lineage.Symbol] {
			continue
		}
		if err := stream.Send(toProtoVolumeEvent(ev)); err != nil {
			return err
		}
		if ev.EventID != "" {
			seen[ev.EventID] = struct{}{}
		}
		sent++
		if request.GetMaxEvents() > 0 && sent >= request.GetMaxEvents() {
			return nil
		}
	}

	for {
		select {
		case <-stream.Context().Done():
			return status.FromContextError(stream.Context().Err()).Err()
		case ev, ok := <-events:
			if !ok {
				return nil
			}
			if ev.EventID != "" {
				if _, dup := seen[ev.EventID]; dup {
					delete(seen, ev.EventID)
					continue
				}
			}
			if len(wanted) > 0 && !wanted[ev.Lineage.Symbol] {
				continue
			}
			if err := stream.Send(toProtoVolumeEvent(ev)); err != nil {
				return err
			}
			sent++
			if request.GetMaxEvents() > 0 && sent >= request.GetMaxEvents() {
				return nil
			}
		}
	}
}

// toProtoVolumeEvent maps validated domain Volume Output to the canonical
// protobuf contract. It does not recompute Volume science.
func toProtoVolumeEvent(ev domain.VolumeEvent) *finfeedsatv1.VolumeEvent {
	out := &finfeedsatv1.VolumeEvent{
		EventId:             ev.EventID,
		Symbol:              ev.Lineage.Symbol,
		IntervalStartUnixMs: unixMilli(ev.Lineage.IntervalStart),
		IntervalEndUnixMs:   unixMilli(ev.Lineage.IntervalEnd),
		MarketSnapshotId:    ev.Lineage.MarketSnapshotID,
		SourceTimestamp:     ev.Lineage.SourceTimestamp,
		AcceptedSequence:    uint64(ev.AcceptedSequence),
		Status:              toProtoVolumeStatus(ev.Status),
		Emitted:             ev.Status == domain.VolumeStatusAvailable,
		Reason:              ev.Reason,
		Emission:            toProtoVolumeEmission(ev),
	}
	if skip := toProtoVolumeSkip(ev.Status, ev.Reason); skip != nil {
		out.Skip = skip
	}
	return out
}

func toProtoVolumeEmission(ev domain.VolumeEvent) *finfeedsatv1.VolumeEmission {
	return &finfeedsatv1.VolumeEmission{
		VRaw:            toProtoVolumeQuantity(ev.VRaw),
		VN:              toProtoVolumeQuantity(ev.VN),
		V1:              toProtoVolumeQuantity(ev.V1),
		V2:              toProtoVolumeQuantity(ev.V2),
		IntervalMeanVn:  toProtoVolumeQuantity(ev.IntervalMeanVN),
		PredictedNextVN: toProtoVolumeQuantity(ev.PredictedNextVN),
		RawColor:        toProtoVolumeIndicator(ev.RawColor),
		Indicator:       toProtoVolumeIndicator(ev.Indicator),
		Transition:      toProtoVolumeTransition(ev.Transition),
		Phase:           toProtoVolumePhase(ev.Phase),
		Confidence:      toProtoVolumeConfidence(ev.Confidence),
		DomainState:     toProtoVolumeDomainState(ev.DomainState),
	}
}

func toProtoVolumeQuantity(q domain.VolumeQuantity) *finfeedsatv1.VolumeQuantity {
	out := &finfeedsatv1.VolumeQuantity{Status: toProtoVolumeQuantityStatus(q.Status)}
	if q.Status == domain.VolumeQtyAvailable {
		value := q.Value
		out.Value = &value
	}
	return out
}

func toProtoVolumeSkip(status, reason string) *finfeedsatv1.VolumeSkip {
	switch status {
	case domain.VolumeStatusMaturing:
		return &finfeedsatv1.VolumeSkip{
			Reason: finfeedsatv1.VolumeSkipReason_VOLUME_SKIP_REASON_MATURING,
			Detail: reason,
		}
	case domain.VolumeStatusInvalid:
		return &finfeedsatv1.VolumeSkip{
			Reason: finfeedsatv1.VolumeSkipReason_VOLUME_SKIP_REASON_INVALID,
			Detail: reason,
		}
	case domain.VolumeStatusError:
		return &finfeedsatv1.VolumeSkip{
			Reason: finfeedsatv1.VolumeSkipReason_VOLUME_SKIP_REASON_ENGINE_ERROR,
			Detail: reason,
		}
	default:
		return nil
	}
}

func toProtoVolumeStatus(value string) finfeedsatv1.VolumeStatus {
	switch value {
	case domain.VolumeStatusMaturing:
		return finfeedsatv1.VolumeStatus_VOLUME_STATUS_MATURING
	case domain.VolumeStatusAvailable:
		return finfeedsatv1.VolumeStatus_VOLUME_STATUS_AVAILABLE
	case domain.VolumeStatusInvalid:
		return finfeedsatv1.VolumeStatus_VOLUME_STATUS_INVALID
	case domain.VolumeStatusError:
		return finfeedsatv1.VolumeStatus_VOLUME_STATUS_ENGINE_ERROR
	default:
		return finfeedsatv1.VolumeStatus_VOLUME_STATUS_UNSPECIFIED
	}
}

func toProtoVolumeQuantityStatus(value string) finfeedsatv1.VolumeQuantityStatus {
	switch value {
	case domain.VolumeQtyInsufficient:
		return finfeedsatv1.VolumeQuantityStatus_VOLUME_QUANTITY_STATUS_INSUFFICIENT
	case domain.VolumeQtyAvailable:
		return finfeedsatv1.VolumeQuantityStatus_VOLUME_QUANTITY_STATUS_AVAILABLE
	case domain.VolumeQtyUndefined:
		return finfeedsatv1.VolumeQuantityStatus_VOLUME_QUANTITY_STATUS_UNDEFINED
	default:
		return finfeedsatv1.VolumeQuantityStatus_VOLUME_QUANTITY_STATUS_UNSPECIFIED
	}
}

func toProtoVolumeIndicator(value string) finfeedsatv1.VolumeIndicator {
	switch value {
	case domain.VolumeColorGreen:
		return finfeedsatv1.VolumeIndicator_VOLUME_INDICATOR_GREEN
	case domain.VolumeColorAmber:
		return finfeedsatv1.VolumeIndicator_VOLUME_INDICATOR_AMBER
	case domain.VolumeColorRed:
		return finfeedsatv1.VolumeIndicator_VOLUME_INDICATOR_RED
	default:
		return finfeedsatv1.VolumeIndicator_VOLUME_INDICATOR_UNSPECIFIED
	}
}

func toProtoVolumeTransition(value string) finfeedsatv1.VolumeTransition {
	switch value {
	case "STABLE":
		return finfeedsatv1.VolumeTransition_VOLUME_TRANSITION_STABLE
	case "PENDING_GREEN":
		return finfeedsatv1.VolumeTransition_VOLUME_TRANSITION_PENDING_GREEN
	case "PENDING_AMBER":
		return finfeedsatv1.VolumeTransition_VOLUME_TRANSITION_PENDING_AMBER
	case "PENDING_RED":
		return finfeedsatv1.VolumeTransition_VOLUME_TRANSITION_PENDING_RED
	case "CONFIRMED_GREEN":
		return finfeedsatv1.VolumeTransition_VOLUME_TRANSITION_CONFIRMED_GREEN
	case "CONFIRMED_AMBER":
		return finfeedsatv1.VolumeTransition_VOLUME_TRANSITION_CONFIRMED_AMBER
	case "CONFIRMED_RED":
		return finfeedsatv1.VolumeTransition_VOLUME_TRANSITION_CONFIRMED_RED
	default:
		return finfeedsatv1.VolumeTransition_VOLUME_TRANSITION_UNSPECIFIED
	}
}

func toProtoVolumePhase(value string) finfeedsatv1.VolumePhase {
	switch value {
	case domain.VolumePhaseStationary:
		return finfeedsatv1.VolumePhase_VOLUME_PHASE_ACTIVITY_STATIONARY
	case domain.VolumePhaseIncreasingAccelerating:
		return finfeedsatv1.VolumePhase_VOLUME_PHASE_ACTIVITY_INCREASING_ACCELERATING
	case domain.VolumePhaseIncreasingDecelerating:
		return finfeedsatv1.VolumePhase_VOLUME_PHASE_ACTIVITY_INCREASING_DECELERATING
	case domain.VolumePhaseDecreasingAccelerating:
		return finfeedsatv1.VolumePhase_VOLUME_PHASE_ACTIVITY_DECREASING_ACCELERATING
	case domain.VolumePhaseDecreasingDecelerating:
		return finfeedsatv1.VolumePhase_VOLUME_PHASE_ACTIVITY_DECREASING_DECELERATING
	default:
		return finfeedsatv1.VolumePhase_VOLUME_PHASE_UNSPECIFIED
	}
}

func toProtoVolumeConfidence(value string) finfeedsatv1.VolumeConfidence {
	if value == domain.VolumeConfidenceHigh {
		return finfeedsatv1.VolumeConfidence_VOLUME_CONFIDENCE_HIGH
	}
	return finfeedsatv1.VolumeConfidence_VOLUME_CONFIDENCE_UNSPECIFIED
}

func toProtoVolumeDomainState(value string) finfeedsatv1.VolumeDomainState {
	if value == domain.VolumeDomainCausalLocal {
		return finfeedsatv1.VolumeDomainState_VOLUME_DOMAIN_STATE_CAUSAL_LOCAL_VOLUME
	}
	return finfeedsatv1.VolumeDomainState_VOLUME_DOMAIN_STATE_UNSPECIFIED
}
