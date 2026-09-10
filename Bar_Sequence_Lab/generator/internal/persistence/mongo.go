package persistence

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"bar_sequence_lab/generator/internal/config"
	"bar_sequence_lab/generator/internal/types"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

type MongoStats struct {
	Partition string
	Attempted uint64
	Persisted uint64
	Failed    uint64
	Pending   int
}

// Store is a shared Mongo client/pool for one database+collection.
// Per-partition writers reuse this store; InsertMany calls may proceed concurrently.
type Store struct {
	client     *mongo.Client
	coll       *mongo.Collection
	database   string
	collection string
	host       string
}

type MongoWriter struct {
	store     *Store
	partition string

	attempted atomic.Uint64
	persisted atomic.Uint64
	failed    atomic.Uint64
}

func OpenMongo(ctx context.Context, cfg config.Config) (*Store, error) {
	if !cfg.MongoEnabled {
		return nil, fmt.Errorf("mongo not enabled")
	}
	client, err := mongo.Connect(options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		return nil, fmt.Errorf("mongo connect: %w", err)
	}
	pingCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx, readpref.Primary()); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("mongo ping: %w", err)
	}
	return &Store{
		client:     client,
		coll:       client.Database(cfg.MongoDB).Collection(cfg.MongoCollection),
		database:   cfg.MongoDB,
		collection: cfg.MongoCollection,
		host:       config.MongoHost(cfg.MongoURI),
	}, nil
}

func (s *Store) Close(ctx context.Context) error {
	if s == nil || s.client == nil {
		return nil
	}
	return s.client.Disconnect(ctx)
}

func (s *Store) Host() string       { return s.host }
func (s *Store) Database() string   { return s.database }
func (s *Store) Collection() string { return s.collection }

func (s *Store) Writer(partition string) *MongoWriter {
	return &MongoWriter{store: s, partition: partition}
}

func (w *MongoWriter) WriteBatch(ctx context.Context, obs []types.Observation) error {
	if w == nil || w.store == nil {
		return nil
	}
	if len(obs) == 0 {
		return nil
	}
	w.attempted.Add(uint64(len(obs)))
	docs := make([]any, len(obs))
	for i := range obs {
		docs[i] = obs[i]
	}
	writeCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	_, err := w.store.coll.InsertMany(writeCtx, docs)
	if err != nil {
		w.failed.Add(uint64(len(obs)))
		if IsDuplicateKey(err) {
			return fmt.Errorf("mongo duplicate-key on existing index (raw arrival evidence must not be suppressed): %w", err)
		}
		return err
	}
	w.persisted.Add(uint64(len(obs)))
	return nil
}

func (w *MongoWriter) Snapshot(pending int) MongoStats {
	if w == nil {
		return MongoStats{}
	}
	return MongoStats{
		Partition: w.partition,
		Attempted: w.attempted.Load(),
		Persisted: w.persisted.Load(),
		Failed:    w.failed.Load(),
		Pending:   pending,
	}
}

func IsDuplicateKey(err error) bool {
	if err == nil {
		return false
	}
	var we mongo.WriteException
	if errors.As(err, &we) {
		for _, e := range we.WriteErrors {
			if e.Code == 11000 {
				return true
			}
		}
	}
	var bwe mongo.BulkWriteException
	if errors.As(err, &bwe) {
		for _, e := range bwe.WriteErrors {
			if e.Code == 11000 {
				return true
			}
		}
	}
	return strings.Contains(strings.ToLower(err.Error()), "duplicate key")
}

// RedactedURIHost is a helper for tests; production uses config.MongoHost.
func RedactedURIHost(uri string) string {
	u, err := url.Parse(uri)
	if err != nil {
		return config.MongoHost(uri)
	}
	return u.Host
}
