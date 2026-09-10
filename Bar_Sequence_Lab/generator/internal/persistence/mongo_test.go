package persistence

import (
	"context"
	"testing"
	"time"

	"bar_sequence_lab/generator/internal/config"
	"bar_sequence_lab/generator/internal/types"
)

func TestOpenMongoPingLocalIfEnabled(t *testing.T) {
	uri := "mongodb://127.0.0.1:27017"
	cfg := config.Config{
		MongoEnabled:    true,
		MongoURI:        uri,
		MongoDB:         "bar_sequence_db",
		MongoCollection: "bar_sequence",
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	store, err := OpenMongo(ctx, cfg)
	if err != nil {
		t.Skipf("local mongo not reachable (connectivity check only; no dummy inserts): %v", err)
	}
	defer store.Close(context.Background())
	if store.Database() != "bar_sequence_db" || store.Collection() != "bar_sequence" {
		t.Fatalf("db/coll %s %s", store.Database(), store.Collection())
	}
	if store.Host() != "127.0.0.1:27017" {
		t.Fatalf("host %s", store.Host())
	}
}

func TestMongoWriterNilIsNoop(t *testing.T) {
	var w *MongoWriter
	if err := w.WriteBatch(context.Background(), []types.Observation{{Symbol: "AAPL"}}); err != nil {
		t.Fatal(err)
	}
}
