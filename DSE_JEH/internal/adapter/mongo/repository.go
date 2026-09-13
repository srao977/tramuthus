package mongo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	driver "go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

type Observation struct {
	CollectionRunID     string    `bson:"collection_run_id"`
	PartitionID         string    `bson:"partition_id"`
	Symbol              string    `bson:"symbol"`
	GeneratorSequenceNo uint64    `bson:"generator_sequence_no"`
	SourceEventTime     time.Time `bson:"source_event_time"`
	SourceTimestampText string    `bson:"source_timestamp_text"`
	ReceivedTime        time.Time `bson:"received_time"`
	PersistedTime       time.Time `bson:"persisted_time,omitempty"`
	Interval            string    `bson:"interval"`
	Open                float64   `bson:"open"`
	High                float64   `bson:"high"`
	Low                 float64   `bson:"low"`
	Close               float64   `bson:"close"`
	Volume              uint64    `bson:"volume"`
	EventCount          uint32    `bson:"event_count,omitempty"`
	SourceID            string    `bson:"source_id"`
	AlpacaMessageType   string    `bson:"alpaca_message_type"`
	PayloadHash         string    `bson:"payload_hash"`
}

type Repository struct {
	client     *driver.Client
	collection *driver.Collection
}

func Open(ctx context.Context, uri, database, collection string) (*Repository, error) {
	client, err := driver.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("connect mongo: %w", err)
	}
	pingContext, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	if err := client.Ping(pingContext, readpref.Primary()); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("ping mongo: %w", err)
	}
	return &Repository{client: client, collection: client.Database(database).Collection(collection)}, nil
}

func (repository *Repository) ReadAll(ctx context.Context, collectionRunID string) ([]Observation, error) {
	collectionRunID = strings.TrimSpace(collectionRunID)
	if collectionRunID == "" {
		return nil, fmt.Errorf("collection_run_id is required")
	}
	filter := bson.D{{Key: "collection_run_id", Value: collectionRunID}}
	findOptions := options.Find().SetSort(bson.D{
		{Key: "collection_run_id", Value: 1},
		{Key: "symbol", Value: 1},
		{Key: "generator_sequence_no", Value: 1},
		{Key: "payload_hash", Value: 1},
	})
	cursor, err := repository.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("find observations: %w", err)
	}
	defer cursor.Close(ctx)
	var observations []Observation
	if err := cursor.All(ctx, &observations); err != nil {
		return nil, fmt.Errorf("decode observations: %w", err)
	}
	return observations, nil
}

func (repository *Repository) Close(ctx context.Context) error {
	if repository == nil || repository.client == nil {
		return nil
	}
	return repository.client.Disconnect(ctx)
}
