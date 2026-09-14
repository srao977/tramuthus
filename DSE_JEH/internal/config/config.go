package config

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	dsejehv1 "tramuthus/dse-jeh-transsat-1/gen/dse_jeh/v1"
	"tramuthus/dse-jeh-transsat-1/internal/execution"
)

type Config struct {
	Mode                     dsejehv1.RuntimeMode
	MongoURI                 string
	MongoDatabase            string
	MongoCollection          string
	MongoTraceCollection     string
	MongoReservoirCollection string
	CollectionRunID          string
	PipelineRunID            string
	RunType                  execution.RunType
	StartingCapital          float64
	AllocationPct            float64
	Risk                     float64
	GRPCAddress              string
	ServerAddress            string
	StatusInterval           time.Duration
	RemainRunning            bool
	Symbols                  []string
	MaxBars                  uint32
	FinalizedOnly            bool
	OutputPath               string
	ReferenceCSV             string
	ComparisonReport         string
}

func Load() (Config, error) {
	cfg := Config{
		MongoURI:                 env("BAR_SEQ_LAB_MONGO_URI", "mongodb://127.0.0.1:27017"),
		MongoDatabase:            env("BAR_SEQ_LAB_MONGO_DB", "bar_sequence_db"),
		MongoCollection:          env("BAR_SEQ_LAB_MONGO_COLLECTION", "bar_sequence"),
		MongoTraceCollection:     env("DSE_JEH_TRACE_COLLECTION", "h_h_stage_emit_values"),
		MongoReservoirCollection: env("DSE_JEH_CAPITAL_RESERVOIR_COLLECTION", "capital_reservoir_events"),
		PipelineRunID:            strings.TrimSpace(os.Getenv("DSE_JEH_PIPELINE_RUN_ID")),
		RunType:                  execution.RunType(strings.TrimSpace(os.Getenv("DSE_JEH_RUN_TYPE"))),
		GRPCAddress:              env("DSE_JEH_GRPC_ADDRESS", "127.0.0.1:50051"),
		ServerAddress:            env("DSE_JEH_SERVER_ADDRESS", "127.0.0.1:50052"),
		RemainRunning:            envBool("DSE_JEH_REMAIN_RUNNING", true),
		Symbols:                  split(os.Getenv("DSE_JEH_SYMBOLS")),
		FinalizedOnly:            envBool("DSE_JEH_FINALIZED_ONLY", true),
		OutputPath:               env("DSE_JEH_OUTPUT", "exports/phase_evidence.jsonl"),
		ReferenceCSV:             strings.TrimSpace(os.Getenv("DSE_JEH_REFERENCE_CSV")),
		ComparisonReport:         env("DSE_JEH_COMPARISON_REPORT", "exports/phase_equivalence.json"),
	}
	var err error
	if cfg.StartingCapital, err = envFloat("DSE_JEH_STARTING_CAPITAL", 100_000); err != nil {
		return Config{}, err
	}
	if cfg.AllocationPct, err = envFloat("DSE_JEH_ALLOCATION_PCT", 1); err != nil {
		return Config{}, err
	}
	if cfg.Risk, err = envFloat("DSE_JEH_RISK_R", 0); err != nil {
		return Config{}, err
	}
	if cfg.StartingCapital <= 0 {
		return Config{}, fmt.Errorf("DSE_JEH_STARTING_CAPITAL must be positive")
	}
	if cfg.AllocationPct < 0 || cfg.AllocationPct > 1 {
		return Config{}, fmt.Errorf("DSE_JEH_ALLOCATION_PCT must be in [0,1]")
	}
	if math.IsNaN(cfg.Risk) || math.IsInf(cfg.Risk, 0) || cfg.Risk < 0 || cfg.Risk > 1 {
		return Config{}, fmt.Errorf("DSE_JEH_RISK_R must be finite and in [0,1]")
	}
	statusInterval, err := time.ParseDuration(env("DSE_JEH_STATUS_INTERVAL", "10s"))
	if err != nil || statusInterval <= 0 {
		return Config{}, fmt.Errorf("DSE_JEH_STATUS_INTERVAL must be a positive duration")
	}
	cfg.StatusInterval = statusInterval
	switch strings.ToUpper(strings.TrimSpace(os.Getenv("DSE_JEH_MODE"))) {
	case "ONLINE":
		cfg.Mode = dsejehv1.RuntimeMode_RUNTIME_MODE_ONLINE
	case "OFFLINE":
		cfg.Mode = dsejehv1.RuntimeMode_RUNTIME_MODE_OFFLINE
		collectionRunID, err := offlineCollectionRunID()
		if err != nil {
			return Config{}, err
		}
		cfg.CollectionRunID = collectionRunID
	default:
		return Config{}, fmt.Errorf("DSE_JEH_MODE must be ONLINE or OFFLINE")
	}
	if value := strings.TrimSpace(os.Getenv("DSE_JEH_MAX_BARS")); value != "" {
		parsed, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return Config{}, fmt.Errorf("DSE_JEH_MAX_BARS: %w", err)
		}
		cfg.MaxBars = uint32(parsed)
	}
	if cfg.Mode == dsejehv1.RuntimeMode_RUNTIME_MODE_ONLINE && len(cfg.Symbols) == 0 {
		return Config{}, fmt.Errorf("DSE_JEH_SYMBOLS is required in ONLINE mode")
	}
	if cfg.Mode == dsejehv1.RuntimeMode_RUNTIME_MODE_OFFLINE && cfg.CollectionRunID == "" {
		return Config{}, fmt.Errorf("DSE_JEH_OFFLINE_COLLECTION_RUN_ID is required in OFFLINE mode")
	}
	if cfg.Mode == dsejehv1.RuntimeMode_RUNTIME_MODE_OFFLINE {
		if err := execution.ValidateRunType(cfg.RunType); err != nil {
			return Config{}, fmt.Errorf("DSE_JEH_RUN_TYPE: %w", err)
		}
	}
	if strings.TrimSpace(cfg.ServerAddress) == "" {
		return Config{}, fmt.Errorf("DSE_JEH_SERVER_ADDRESS is required")
	}
	return cfg, nil
}

func offlineCollectionRunID() (string, error) {
	configured := strings.TrimSpace(os.Getenv("DSE_JEH_OFFLINE_COLLECTION_RUN_ID"))
	legacy := strings.TrimSpace(os.Getenv("DSE_JEH_COLLECTION_RUN_ID"))
	if configured != "" && legacy != "" && configured != legacy {
		return "", fmt.Errorf("DSE_JEH_OFFLINE_COLLECTION_RUN_ID conflicts with DSE_JEH_COLLECTION_RUN_ID")
	}
	if configured != "" {
		return configured, nil
	}
	return legacy, nil
}

func env(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}

func envBool(name string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envFloat(name string, fallback float64) (float64, error) {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", name, err)
	}
	return parsed, nil
}

func split(value string) []string {
	var values []string
	for _, item := range strings.Split(value, ",") {
		if item = strings.ToUpper(strings.TrimSpace(item)); item != "" {
			values = append(values, item)
		}
	}
	return values
}
