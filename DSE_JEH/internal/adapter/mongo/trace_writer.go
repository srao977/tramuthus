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

type TraceWriter struct {
	client     *driver.Client
	collection *driver.Collection
}

func OpenTraceWriter(ctx context.Context, uri, database, collection string) (*TraceWriter, error) {
	if strings.TrimSpace(database) == "" || strings.TrimSpace(collection) == "" {
		return nil, fmt.Errorf("trace database and collection are required")
	}
	client, err := driver.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("connect trace mongo: %w", err)
	}
	pingContext, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	if err := client.Ping(pingContext, readpref.Primary()); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("ping trace mongo: %w", err)
	}
	databaseHandle := client.Database(database)
	names, err := databaseHandle.ListCollectionNames(pingContext, bson.D{{Key: "name", Value: collection}})
	if err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("inspect trace collection %s.%s: %w", database, collection, err)
	}
	if len(names) != 1 {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("required trace collection %s.%s is unavailable; run the approved setup verification first", database, collection)
	}
	return &TraceWriter{client: client, collection: databaseHandle.Collection(collection)}, nil
}

func (writer *TraceWriter) Write(ctx context.Context, record execution.TraceRecord) (bson.ObjectID, error) {
	if err := record.ValidateRunType(); err != nil {
		return bson.NilObjectID, err
	}
	result, err := writer.collection.InsertOne(ctx, record)
	if err != nil {
		return bson.NilObjectID, fmt.Errorf("insert stage/emit trace: %w", err)
	}
	insertedID, ok := result.InsertedID.(bson.ObjectID)
	if !ok || insertedID.IsZero() {
		return bson.NilObjectID, fmt.Errorf("insert stage/emit trace returned a non-ObjectId identifier")
	}
	return insertedID, nil
}

func (writer *TraceWriter) Close(ctx context.Context) error {
	if writer == nil || writer.client == nil {
		return nil
	}
	return writer.client.Disconnect(ctx)
}
