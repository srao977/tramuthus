package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	ServiceName         = "Bar_Sequence_Lab generator"
	MaxSymbolsBasicPlan = 30
	IEXStreamURL        = "wss://stream.data.alpaca.markets/v2/iex"
	TestStreamURL       = "wss://stream.data.alpaca.markets/v2/test"
	HeartbeatInterval   = time.Second
	HeartbeatMaxRTT     = 1500 * time.Millisecond
	HeartbeatMaxMisses  = 3
	ReconnectBase       = 100 * time.Millisecond
	ReconnectCap        = 30 * time.Second
	StreamReadIdle      = 90 * time.Second
	DefaultBufferCap    = 256
	DefaultFlushCount   = 50
	DefaultFlushTime    = 200 * time.Millisecond
	DefaultMetricsEvery = 5 * time.Second
)

type Config struct {
	Feed              string
	StreamURL         string
	APIKey            string
	APISecret         string
	DataDir           string
	SelectedGroups    []string
	GroupA            []string
	GroupB            []string
	GroupC            []string
	SymbolToPartition map[string]string
	SubscribeSymbols  []string
	BufferCapacity    int
	FlushCount        int
	FlushInterval     time.Duration
	MetricsInterval   time.Duration
	MaxBars           int
	Duration          time.Duration
	MongoEnabled      bool
	MongoURI          string
	MongoDB           string
	MongoCollection   string
}

func Load() (Config, error) {
	feed := strings.ToLower(strings.TrimSpace(environmentOrDefault("BAR_SEQ_LAB_FEED", "test")))
	if feed != "iex" && feed != "test" {
		return Config{}, fmt.Errorf("BAR_SEQ_LAB_FEED must be iex or test, got %q", feed)
	}

	selected, err := ParseGroups(environmentOrDefault("BAR_SEQ_LAB_SELECTED_GROUPS", ""))
	if err != nil {
		return Config{}, err
	}
	if len(selected) == 0 {
		return Config{}, fmt.Errorf("BAR_SEQ_LAB_SELECTED_GROUPS is required (one or more of A B C); no default of A+B+C")
	}

	groupA := splitSymbols(environmentOrDefault("BAR_SEQ_LAB_GROUP_A", defaultGroupA(feed)))
	groupB := splitSymbols(environmentOrDefault("BAR_SEQ_LAB_GROUP_B", defaultGroupB(feed)))
	groupC := splitSymbols(environmentOrDefault("BAR_SEQ_LAB_GROUP_C", defaultGroupC(feed)))

	cfg := Config{
		Feed:            feed,
		StreamURL:       streamURL(feed),
		APIKey:          trimCredential(firstEnv("ALPACA_API_KEY")),
		APISecret:       trimCredential(firstEnv("ALPACA_API_SECRET", "ALPACA_SECRET_KEY")),
		DataDir:         environmentOrDefault("BAR_SEQ_LAB_DATA_DIR", "data"),
		SelectedGroups:  selected,
		GroupA:          groupA,
		GroupB:          groupB,
		GroupC:          groupC,
		BufferCapacity:  envInt("BAR_SEQ_LAB_BUFFER_CAP", DefaultBufferCap),
		FlushCount:      envInt("BAR_SEQ_LAB_FLUSH_COUNT", DefaultFlushCount),
		FlushInterval:   envDuration("BAR_SEQ_LAB_FLUSH_INTERVAL", DefaultFlushTime),
		MetricsInterval: envDuration("BAR_SEQ_LAB_METRICS_INTERVAL", DefaultMetricsEvery),
		MaxBars:         envInt("BAR_SEQ_LAB_MAX_BARS", 0),
		Duration:        envDuration("BAR_SEQ_LAB_DURATION", 0),
		MongoEnabled:    envBool("BAR_SEQ_LAB_MONGO_ENABLED", false),
		MongoURI:        strings.TrimSpace(os.Getenv("BAR_SEQ_LAB_MONGO_URI")),
		MongoDB:         strings.TrimSpace(os.Getenv("BAR_SEQ_LAB_MONGO_DB")),
		MongoCollection: environmentOrDefault("BAR_SEQ_LAB_MONGO_COLLECTION", "bar_sequence"),
	}
	if cfg.BufferCapacity <= 0 {
		return Config{}, fmt.Errorf("BAR_SEQ_LAB_BUFFER_CAP must be > 0")
	}
	if cfg.FlushCount <= 0 {
		return Config{}, fmt.Errorf("BAR_SEQ_LAB_FLUSH_COUNT must be > 0")
	}
	if cfg.FlushInterval <= 0 {
		return Config{}, fmt.Errorf("BAR_SEQ_LAB_FLUSH_INTERVAL must be > 0")
	}
	if cfg.APIKey == "" || cfg.APISecret == "" {
		return Config{}, fmt.Errorf("ALPACA_API_KEY and ALPACA_API_SECRET (or ALPACA_SECRET_KEY) are required")
	}
	if cfg.MongoEnabled {
		if cfg.MongoURI == "" {
			return Config{}, fmt.Errorf("BAR_SEQ_LAB_MONGO_URI is required when BAR_SEQ_LAB_MONGO_ENABLED=true")
		}
		if cfg.MongoDB == "" {
			return Config{}, fmt.Errorf("BAR_SEQ_LAB_MONGO_DB is required when BAR_SEQ_LAB_MONGO_ENABLED=true")
		}
		if cfg.MongoCollection == "" {
			return Config{}, fmt.Errorf("BAR_SEQ_LAB_MONGO_COLLECTION must not be empty")
		}
	}

	mapping, subscribe, err := freezeAssignment(cfg)
	if err != nil {
		return Config{}, err
	}
	cfg.SymbolToPartition = mapping
	cfg.SubscribeSymbols = subscribe
	if len(cfg.SubscribeSymbols) > MaxSymbolsBasicPlan {
		return Config{}, fmt.Errorf("symbol cap exceeded: %d > %d", len(cfg.SubscribeSymbols), MaxSymbolsBasicPlan)
	}
	return cfg, nil
}

func ParseGroups(value string) ([]string, error) {
	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == ' ' || r == ';'
	})
	seen := map[string]struct{}{}
	var groups []string
	for _, part := range parts {
		g := strings.ToUpper(strings.TrimSpace(part))
		if g == "" {
			continue
		}
		if g != "A" && g != "B" && g != "C" {
			return nil, fmt.Errorf("invalid group %q; allowed: A B C", part)
		}
		if _, ok := seen[g]; ok {
			continue
		}
		seen[g] = struct{}{}
		groups = append(groups, g)
	}
	order := []string{"A", "B", "C"}
	var selected []string
	for _, g := range order {
		if _, ok := seen[g]; ok {
			selected = append(selected, g)
		}
	}
	return selected, nil
}

func freezeAssignment(cfg Config) (map[string]string, []string, error) {
	mapping := map[string]string{}
	var subscribe []string
	for _, g := range cfg.SelectedGroups {
		symbols := symbolsFor(cfg, g)
		if len(symbols) == 0 {
			return nil, nil, fmt.Errorf("selected group %s is empty", g)
		}
		for _, symbol := range symbols {
			if existing, ok := mapping[symbol]; ok {
				return nil, nil, fmt.Errorf("duplicate symbol %s across selected groups %s and %s", symbol, existing, g)
			}
			mapping[symbol] = g
			subscribe = append(subscribe, symbol)
		}
	}
	return mapping, subscribe, nil
}

func symbolsFor(cfg Config, group string) []string {
	switch group {
	case "A":
		return cfg.GroupA
	case "B":
		return cfg.GroupB
	case "C":
		return cfg.GroupC
	default:
		return nil
	}
}

func SourceID(feed string) string {
	if feed == "test" {
		return "ALPACA_TEST"
	}
	return "ALPACA_IEX"
}

func StartupLines(cfg Config) []string {
	lines := []string{
		fmt.Sprintf("service=%s", ServiceName),
		"version=phase-1.5-jsonl+mongo",
		"environment=lab/local",
		fmt.Sprintf("alpaca feed=%s stream_host=%s", cfg.Feed, hostOnly(cfg.StreamURL)),
		persistenceStartupLine(cfg),
		fmt.Sprintf("data_dir=%s", cfg.DataDir),
		fmt.Sprintf("selected_groups=%s", strings.Join(cfg.SelectedGroups, ",")),
		fmt.Sprintf("symbol_count=%d", len(cfg.SubscribeSymbols)),
		fmt.Sprintf("subscribe=%s", strings.Join(cfg.SubscribeSymbols, ",")),
	}
	for _, g := range cfg.SelectedGroups {
		lines = append(lines, fmt.Sprintf("group_%s=%s", g, strings.Join(symbolsFor(cfg, g), ",")))
	}
	lines = append(lines,
		fmt.Sprintf("buffer_capacity=%d (A1-A3/B1-B3/C1-C3 as selected)", cfg.BufferCapacity),
		fmt.Sprintf("flush_count=%d flush_interval=%s", cfg.FlushCount, cfg.FlushInterval),
		"dropped_valid target=0",
	)
	return lines
}

func persistenceStartupLine(cfg Config) string {
	if !cfg.MongoEnabled {
		return "persistence=jsonl (Mongo disabled; JSONL audit active)"
	}
	return fmt.Sprintf("persistence=jsonl+mongo mongo_host=%s mongo_db=%s mongo_collection=%s",
		MongoHost(cfg.MongoURI), cfg.MongoDB, cfg.MongoCollection)
}

func MongoHost(uri string) string {
	trimmed := strings.TrimSpace(uri)
	trimmed = strings.TrimPrefix(trimmed, "mongodb+srv://")
	trimmed = strings.TrimPrefix(trimmed, "mongodb://")
	if at := strings.LastIndex(trimmed, "@"); at >= 0 {
		trimmed = trimmed[at+1:]
	}
	if i := strings.IndexAny(trimmed, "/?"); i >= 0 {
		trimmed = trimmed[:i]
	}
	return trimmed
}

func hostOnly(url string) string {
	trimmed := strings.TrimPrefix(url, "wss://")
	trimmed = strings.TrimPrefix(trimmed, "https://")
	if i := strings.Index(trimmed, "/"); i >= 0 {
		return trimmed[:i]
	}
	return trimmed
}

func defaultGroupA(feed string) string {
	if feed == "test" {
		return "FAKEPACA"
	}
	return "AAPL"
}

func defaultGroupB(feed string) string {
	if feed == "test" {
		return ""
	}
	return "MSFT"
}

func defaultGroupC(feed string) string {
	if feed == "test" {
		return ""
	}
	return "NVDA"
}

func streamURL(feed string) string {
	if feed == "test" {
		return TestStreamURL
	}
	return IEXStreamURL
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

func envBool(name string, fallback bool) bool {
	raw := strings.ToLower(strings.TrimSpace(os.Getenv(name)))
	if raw == "" {
		return fallback
	}
	switch raw {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
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

func envInt(name string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return n
}

func envDuration(name string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return fallback
	}
	return d
}
