// Package config resolves batch inputs from flags and Lab-safe environment defaults.
// Invalid SERIES_SIZE or missing identifiers fail before any Mongo connection or write.
package config

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

const DefaultCollectionRunID = "20260910T191246Z-1"

// Config identifies one analytical experiment and its Mongo boundaries.
type Config struct {
	SeriesSize      int
	CollectionRunID string
	MongoURI        string
	MongoDB         string
	RawCollection   string
	PhaseCollection string
}

// Parse accepts explicit batch flags. Environment values configure Mongo only;
// SERIES_SIZE remains an explicit required command-line experiment parameter.
func Parse(args []string) (Config, error) {
	fs := flag.NewFlagSet("phase-angle-series-generator", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	cfg := Config{}
	fs.IntVar(&cfg.SeriesSize, "series-size", 0, "number of ordered bars supplied to the solver")
	fs.StringVar(&cfg.CollectionRunID, "collection-run-id", DefaultCollectionRunID, "raw collection run identifier")
	fs.StringVar(&cfg.MongoURI, "mongo-uri", env("BAR_SEQ_LAB_MONGO_URI", "mongodb://127.0.0.1:27017"), "MongoDB URI")
	fs.StringVar(&cfg.MongoDB, "mongo-db", env("BAR_SEQ_LAB_MONGO_DB", "bar_sequence_db"), "MongoDB database")
	fs.StringVar(&cfg.RawCollection, "raw-collection", env("BAR_SEQ_LAB_RAW_COLLECTION", "bar_sequence"), "authoritative raw collection")
	fs.StringVar(&cfg.PhaseCollection, "phase-collection", env("BAR_SEQ_LAB_PHASE_COLLECTION", "bar_sequence_phase_angle_series"), "derived phase collection")
	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}
	if cfg.SeriesSize <= 0 {
		return Config{}, fmt.Errorf("series-size must be greater than zero")
	}
	if strings.TrimSpace(cfg.CollectionRunID) == "" {
		return Config{}, fmt.Errorf("collection-run-id must not be empty")
	}
	if cfg.RawCollection == cfg.PhaseCollection {
		return Config{}, fmt.Errorf("raw and phase collections must be different")
	}
	return cfg, nil
}

func env(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}
