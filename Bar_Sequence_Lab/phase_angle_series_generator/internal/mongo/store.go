// Package mongo owns read-only raw access and idempotent derived-record upserts.
// It never writes to the authoritative raw collection and fails on missing indexes.
package mongo

import (
	"context"
	"fmt"
	"sort"
	"time"

	"bar_sequence_lab/phase_angle_series_generator/internal/config"
	"bar_sequence_lab/phase_angle_series_generator/internal/processing"

	"go.mongodb.org/mongo-driver/v2/bson"
	driver "go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

const (
	UniqueIndex = "ux_phase_run_symbol_sequence_size_solver"
	QueryIndex  = "ix_phase_run_partition_size_solver_symbol_sequence"
)

// Store holds separate raw and derived collection handles on one Mongo client.
type Store struct {
	client *driver.Client
	raw    *driver.Collection
	phase  *driver.Collection
}

// Open connects, pings, and verifies that the explicit initialization script ran.
func Open(ctx context.Context, cfg config.Config) (*Store, error) {
	client, err := driver.Connect(options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		return nil, fmt.Errorf("connect Mongo: %w", err)
	}
	pingCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx, readpref.Primary()); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("ping Mongo: %w", err)
	}
	store := &Store{client: client, raw: client.Database(cfg.MongoDB).Collection(cfg.RawCollection), phase: client.Database(cfg.MongoDB).Collection(cfg.PhaseCollection)}
	if err := store.verifyIndexes(ctx); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, err
	}
	return store, nil
}

// Close releases the Mongo client; it does not mutate either collection.
func (store *Store) Close(ctx context.Context) error { return store.client.Disconnect(ctx) }

func (store *Store) verifyIndexes(ctx context.Context) error {
	cursor, err := store.phase.Indexes().List(ctx)
	if err != nil {
		return fmt.Errorf("list phase indexes (run scripts/Initialize-BarSequencePhaseAngleMongo.js first): %w", err)
	}
	defer cursor.Close(ctx)
	found := map[string]bool{}
	for cursor.Next(ctx) {
		var index struct {
			Name string `bson:"name"`
		}
		if err := cursor.Decode(&index); err != nil {
			return err
		}
		found[index.Name] = true
	}
	if !found[UniqueIndex] || !found[QueryIndex] {
		return fmt.Errorf("required phase indexes missing; run scripts/Initialize-BarSequencePhaseAngleMongo.js explicitly")
	}
	return cursor.Err()
}

// Symbols discovers sorted raw symbols for one run and ingestion partition.
func (store *Store) Symbols(ctx context.Context, runID, partition string) ([]string, error) {
	var symbols []string
	result := store.raw.Distinct(ctx, "symbol", bson.M{"collection_run_id": runID, "partition_id": partition})
	if err := result.Decode(&symbols); err != nil {
		return nil, err
	}
	sort.Strings(symbols)
	return symbols, nil
}

// CountRaw counts authoritative records matching a raw-only filter.
func (store *Store) CountRaw(ctx context.Context, filter bson.M) (int64, error) {
	return store.raw.CountDocuments(ctx, filter)
}

// Prefix reads at most seriesSize bars ordered by generator_sequence_no.
func (store *Store) Prefix(ctx context.Context, runID, partition, symbol string, seriesSize int) ([]processing.RawBar, error) {
	filter := bson.M{"collection_run_id": runID, "partition_id": partition, "symbol": symbol}
	opts := options.Find().SetSort(bson.D{{Key: "generator_sequence_no", Value: 1}}).SetLimit(int64(seriesSize)).SetProjection(bson.M{
		"collection_run_id": 1, "partition_id": 1, "symbol": 1, "generator_sequence_no": 1, "high": 1, "low": 1,
	})
	cursor, err := store.raw.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var bars []processing.RawBar
	if err := cursor.All(ctx, &bars); err != nil {
		return nil, err
	}
	return bars, nil
}

// Upsert writes one derived result by complete analytical identity.
func (store *Store) Upsert(ctx context.Context, record processing.DerivedRecord) error {
	set := bson.M{
		"partition_id": record.PartitionID, "input_series_type": record.InputSeriesType,
		"phase_angle_degrees": record.PhaseAngleDegrees, "phase_observable": record.PhaseObservable,
		"validity_state": record.ValidityState,
	}
	update := bson.M{"$set": set, "$setOnInsert": bson.M{
		"collection_run_id": record.CollectionRunID, "symbol": record.Symbol,
		"generator_sequence_no": record.GeneratorSequenceNo, "series_size": record.SeriesSize,
		"solver_name": record.SolverName, "solver_version": record.SolverVersion,
		"analysis_created_at": record.AnalysisCreatedAt,
	}}
	_, err := store.phase.UpdateOne(ctx, record.Identity(), update, options.UpdateOne().SetUpsert(true))
	return err
}
