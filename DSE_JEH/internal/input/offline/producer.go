package offline

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	dsejehv1 "tramuthus/dse-jeh-transsat-1/gen/dse_jeh/v1"
	"tramuthus/dse-jeh-transsat-1/internal/adapter/mongo"
	"tramuthus/dse-jeh-transsat-1/internal/analytical"
	"tramuthus/dse-jeh-transsat-1/internal/evidence"
)

const ReplayOrderPolicy = "COLLECTION_RUN_SYMBOL_ENTITY_SEQUENCE_PAYLOAD_V1"

type Repository interface {
	ReadAll(context.Context, string) ([]mongo.Observation, error)
}

type Producer struct {
	repository      Repository
	collectionRunID string
}

func New(repository Repository, collectionRunID string) *Producer {
	return &Producer{repository: repository, collectionRunID: collectionRunID}
}

func (producer *Producer) Events(ctx context.Context) ([]*dsejehv1.BarEvent, error) {
	if strings.TrimSpace(producer.collectionRunID) == "" {
		return nil, fmt.Errorf("offline collection_run_id is required")
	}
	observations, err := producer.repository.ReadAll(ctx, producer.collectionRunID)
	if err != nil {
		return nil, err
	}
	events := make([]*dsejehv1.BarEvent, len(observations))
	for index := range observations {
		events[index] = MapObservation(observations[index])
	}
	return events, nil
}

func MapObservation(observation mongo.Observation) *dsejehv1.BarEvent {
	symbol := strings.ToUpper(strings.TrimSpace(observation.Symbol))
	sequence := observation.GeneratorSequenceNo
	sourceObservationID := evidence.ID("offline-observation", observation.CollectionRunID, symbol, strconv.FormatUint(sequence, 10))
	eventID := evidence.ID("bar-event", sourceObservationID, observation.PayloadHash)
	volume := observation.Volume
	eventCount := observation.EventCount
	return &dsejehv1.BarEvent{
		EventId:             eventID,
		EntityId:            symbol,
		Symbol:              symbol,
		Interval:            observation.Interval,
		IntervalStartUnixMs: observation.SourceEventTime.UnixMilli(),
		IntervalEndUnixMs:   observation.SourceEventTime.Add(timeframe(observation.Interval)).UnixMilli(),
		Open:                observation.Open,
		High:                observation.High,
		Low:                 observation.Low,
		Close:               observation.Close,
		Volume:              &volume,
		EventCount:          &eventCount,
		Provenance: &dsejehv1.SourceProvenance{
			SourceMode:          dsejehv1.RuntimeMode_RUNTIME_MODE_OFFLINE,
			SourceId:            observation.SourceID,
			SourceObservationId: sourceObservationID,
			CollectionRunId:     observation.CollectionRunID,
			PartitionId:         observation.PartitionID,
			EntitySequence:      sequence,
			EntitySequenceScope: analytical.Scope(observation.CollectionRunID, symbol),
			ReplayOrderPolicy:   ReplayOrderPolicy,
			SourceTimestamp:     observation.SourceTimestampText,
			SourceEventUnixMs:   observation.SourceEventTime.UnixMilli(),
			ReceivedUnixMs:      observation.ReceivedTime.UnixMilli(),
			IsFinal:             true,
			PayloadHash:         observation.PayloadHash,
		},
	}
}

func timeframe(interval string) time.Duration {
	switch strings.ToLower(strings.TrimSpace(interval)) {
	case "1min", "1m":
		return time.Minute
	case "5min", "5m":
		return 5 * time.Minute
	case "1hour", "1h":
		return time.Hour
	default:
		return 0
	}
}
