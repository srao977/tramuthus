package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type ModelMode string

const (
	ModelOff      ModelMode = "off"
	ModelAdaptive ModelMode = "adaptive"
)

type PricingMode string

const (
	PricingOff  PricingMode = "off"
	PricingExpm PricingMode = "expm"
)

const (
	DefaultGRPCPort           = "50051"
	DefaultSatelliteID        = "Fin_FeedSat_1"
	DefaultSatelliteRole      = "FeedSat"
	DefaultSubscriberType     = "PRODUCER"
	DefaultDomainPlane        = "Finance"
	DefaultSource             = "alpaca"
	DefaultFeed               = "iex"
	DefaultInterval           = "1Min"
	MaxSymbolsBasicPlan       = 30
	IEXStreamURL              = "wss://stream.data.alpaca.markets/v2/iex"
	TestStreamURL             = "wss://stream.data.alpaca.markets/v2/test"
	DataRESTURL               = "https://data.alpaca.markets"
	HeartbeatInterval         = time.Second
	HeartbeatMaxRTT           = 1500 * time.Millisecond
	HeartbeatMaxMisses        = 3
	ReconnectBase             = 100 * time.Millisecond
	ReconnectCap              = 30 * time.Second
	StreamReadIdle            = 90 * time.Second
	WindowLimit               = 64
	SubscriberQueue           = 16
	ConsumerQueue             = 2
	DefaultModelMode          = ModelOff
	DefaultPricingMode        = PricingOff
	DefaultModelDeadline      = 200 * time.Millisecond
	MaxModelDeadline          = 2 * time.Second
	DefaultStageTransitionLog = "./stage_transitions.txt"
	EnvStageTransitionLog     = "FIN_FEEDSAT_STAGE_TRANSITION_LOG"
)

type Config struct {
	SatelliteID        string
	SatelliteRole      string
	SubscriberType     string
	DomainPlane        string
	GRPCPort           string
	Source             string
	Feed               string
	StreamURL          string
	DataREST           string
	APIKey             string
	APISecret          string
	Symbols            []string
	CSVPath            string
	Interval           string
	Model              ModelMode
	Pricing            PricingMode
	ModelDeadline      time.Duration
	StageTransitionLog string
}

func Load() (Config, error) {
	satelliteID := environmentOrDefault("FIN_FEEDSAT_ID", DefaultSatelliteID)
	satelliteRole := environmentOrDefault("FIN_FEEDSAT_ROLE", DefaultSatelliteRole)
	subscriberType := environmentOrDefault("FIN_FEEDSAT_SUBSCRIBER_TYPE", DefaultSubscriberType)
	domainPlane := environmentOrDefault("FIN_FEEDSAT_DOMAIN_PLANE", DefaultDomainPlane)

	source := strings.ToLower(environmentOrDefault("FIN_FEEDSAT_SOURCE", DefaultSource))
	feed := strings.ToLower(environmentOrDefault("FIN_FEEDSAT_FEED", DefaultFeed))
	symbols := splitSymbols(environmentOrDefault("FIN_FEEDSAT_SYMBOLS", "AAPL"))
	if feed == "test" && environmentOrDefault("FIN_FEEDSAT_SYMBOLS", "") == "" {
		symbols = []string{"FAKEPACA"}
	}
	if len(symbols) == 0 {
		return Config{}, fmt.Errorf("FIN_FEEDSAT_SYMBOLS must contain at least one symbol")
	}
	if len(symbols) > MaxSymbolsBasicPlan {
		return Config{}, fmt.Errorf("Basic plan allows at most %d symbols, got %d", MaxSymbolsBasicPlan, len(symbols))
	}

	mode, err := ParseModelMode(environmentOrDefault("FIN_FEEDSAT_MODEL", string(DefaultModelMode)))
	if err != nil {
		return Config{}, err
	}
	pricing, err := ParsePricingMode(environmentOrDefault("FIN_FEEDSAT_PRICING", string(DefaultPricingMode)))
	if err != nil {
		return Config{}, err
	}
	if err := ValidatePricingRequiresAdaptive(pricing, mode); err != nil {
		return Config{}, err
	}
	deadline, err := ParseModelDeadline(environmentOrDefault("FIN_FEEDSAT_MODEL_DEADLINE", DefaultModelDeadline.String()))
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		SatelliteID:        satelliteID,
		SatelliteRole:      satelliteRole,
		SubscriberType:     subscriberType,
		DomainPlane:        domainPlane,
		GRPCPort:           environmentOrDefault("GRPC_PORT", DefaultGRPCPort),
		Source:             source,
		Feed:               feed,
		StreamURL:          streamURL(feed),
		DataREST:           environmentOrDefault("ALPACA_DATA_REST", DataRESTURL),
		APIKey:             trimCredential(firstEnv("ALPACA_API_KEY")),
		APISecret:          trimCredential(firstEnv("ALPACA_API_SECRET", "ALPACA_SECRET_KEY")),
		Symbols:            symbols,
		CSVPath:            environmentOrDefault("FIN_FEEDSAT_CSV_PATH", "AAPL_1min_firstratedata.csv"),
		Interval:           environmentOrDefault("FIN_FEEDSAT_INTERVAL", DefaultInterval),
		Model:              mode,
		Pricing:            pricing,
		ModelDeadline:      deadline,
		StageTransitionLog: environmentOrDefault(EnvStageTransitionLog, DefaultStageTransitionLog),
	}
	if source == "alpaca" && (cfg.APIKey == "" || cfg.APISecret == "") {
		return Config{}, fmt.Errorf("ALPACA_API_KEY and ALPACA_API_SECRET (or ALPACA_SECRET_KEY) are required for source=alpaca")
	}
	if source != "alpaca" && source != "csv" {
		return Config{}, fmt.Errorf("FIN_FEEDSAT_SOURCE must be alpaca or csv")
	}
	return cfg, nil
}

func ParseModelMode(value string) (ModelMode, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", string(ModelOff):
		return ModelOff, nil
	case string(ModelAdaptive):
		return ModelAdaptive, nil
	default:
		return "", fmt.Errorf("FIN_FEEDSAT_MODEL must be off or adaptive, got %q", value)
	}
}

func ParsePricingMode(value string) (PricingMode, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", string(PricingOff):
		return PricingOff, nil
	case string(PricingExpm):
		return PricingExpm, nil
	default:
		return "", fmt.Errorf("FIN_FEEDSAT_PRICING must be off or expm, got %q", value)
	}
}

func ValidatePricingRequiresAdaptive(pricing PricingMode, model ModelMode) error {
	if pricing == PricingExpm && model != ModelAdaptive {
		return fmt.Errorf("FIN_FEEDSAT_PRICING=expm requires FIN_FEEDSAT_MODEL=adaptive")
	}
	return nil
}

func ParseModelDeadline(value string) (time.Duration, error) {
	deadline, err := time.ParseDuration(strings.TrimSpace(value))
	if err != nil {
		return 0, fmt.Errorf("FIN_FEEDSAT_MODEL_DEADLINE: %w", err)
	}
	if deadline <= 0 || deadline > MaxModelDeadline {
		return 0, fmt.Errorf("FIN_FEEDSAT_MODEL_DEADLINE must be > 0 and <= %s, got %s", MaxModelDeadline, deadline)
	}
	return deadline, nil
}

func streamURL(feed string) string {
	switch feed {
	case "test":
		return TestStreamURL
	default:
		return IEXStreamURL
	}
}

func splitSymbols(value string) []string {
	parts := strings.Split(value, ",")
	symbols := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		symbol := strings.ToUpper(strings.TrimSpace(part))
		if symbol == "" {
			continue
		}
		if _, duplicate := seen[symbol]; duplicate {
			continue
		}
		seen[symbol] = struct{}{}
		symbols = append(symbols, symbol)
	}
	return symbols
}

func trimCredential(value string) string {
	return strings.Trim(strings.TrimSpace(value), `"'`)
}

func firstEnv(names ...string) string {
	for _, name := range names {
		if value := os.Getenv(name); value != "" {
			return value
		}
	}
	return ""
}

func environmentOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
