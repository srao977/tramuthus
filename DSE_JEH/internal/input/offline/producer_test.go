package offline

import (
	"context"
	"testing"

	"tramuthus/dse-jeh-transsat-1/internal/adapter/mongo"
)

type recordingRepository struct {
	collectionRunID string
	observations    []mongo.Observation
}

func (repository *recordingRepository) ReadAll(_ context.Context, collectionRunID string) ([]mongo.Observation, error) {
	repository.collectionRunID = collectionRunID
	var selected []mongo.Observation
	for _, observation := range repository.observations {
		if observation.CollectionRunID == collectionRunID {
			selected = append(selected, observation)
		}
	}
	return selected, nil
}

func TestEventsSelectsConfiguredCollectionRun(t *testing.T) {
	repository := &recordingRepository{observations: []mongo.Observation{
		{CollectionRunID: "20260910T191246Z-1", Symbol: "AAPL", GeneratorSequenceNo: 1},
		{CollectionRunID: "20260911T161623Z-1", Symbol: "AAPL", GeneratorSequenceNo: 1},
		{CollectionRunID: "20260910T190625Z-1", Symbol: "MSFT", GeneratorSequenceNo: 1},
	}}
	events, err := New(repository, "20260911T161623Z-1").Events(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if repository.collectionRunID != "20260911T161623Z-1" {
		t.Fatalf("ReadAll collection run = %q", repository.collectionRunID)
	}
	if len(events) != 1 || events[0].GetProvenance().GetCollectionRunId() != "20260911T161623Z-1" {
		t.Fatalf("Events = %+v", events)
	}
}

func TestEventsRejectsMissingCollectionRun(t *testing.T) {
	repository := &recordingRepository{}
	if _, err := New(repository, " ").Events(context.Background()); err == nil {
		t.Fatal("Events must reject a missing collection run")
	}
	if repository.collectionRunID != "" {
		t.Fatal("Events must not query the repository without a collection run")
	}
}
