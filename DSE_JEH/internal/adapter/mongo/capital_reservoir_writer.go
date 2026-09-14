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

	"tramuthus/dse-jeh-transsat-1/internal/execution"
)

type CapitalReservoirWriter struct {
	client     *driver.Client
	collection *driver.Collection
}

func OpenCapitalReservoirWriter(ctx context.Context, uri, database, collection string) (*CapitalReservoirWriter, error) {
	if strings.TrimSpace(database) == "" || strings.TrimSpace(collection) == "" {
		return nil, fmt.Errorf("capital reservoir database and collection are required")
	}
	client, err := driver.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("connect capital reservoir mongo: %w", err)
	}
	pingContext, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	if err := client.Ping(pingContext, readpref.Primary()); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("ping capital reservoir mongo: %w", err)
	}
	databaseHandle := client.Database(database)
	names, err := databaseHandle.ListCollectionNames(pingContext, bson.D{{Key: "name", Value: collection}})
	if err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("inspect capital reservoir collection %s.%s: %w", database, collection, err)
	}
	if len(names) != 1 {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("required capital reservoir collection %s.%s is unavailable; run the approved setup script first", database, collection)
	}
	return &CapitalReservoirWriter{client: client, collection: databaseHandle.Collection(collection)}, nil
}

func (writer *CapitalReservoirWriter) Write(ctx context.Context, event execution.CapitalReservoirEvent) error {
	if err := event.Validate(); err != nil {
		return err
	}
	if _, err := writer.collection.InsertOne(ctx, event); err != nil {
		return fmt.Errorf("insert capital reservoir event: %w", err)
	}
	return nil
}

func (writer *CapitalReservoirWriter) Close(ctx context.Context) error {
	if writer == nil || writer.client == nil {
		return nil
	}
	return writer.client.Disconnect(ctx)
}
