package config

import (
	"os"
	"strings"
	"testing"
)

func TestParseGroupsNoDefault(t *testing.T) {
	got, err := ParseGroups("")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("expected no default groups, got %v", got)
	}
}

func TestParseGroupsInvalid(t *testing.T) {
	_, err := ParseGroups("A D")
	if err == nil {
		t.Fatal("expected invalid group error")
	}
}

func TestParseGroupsCombinations(t *testing.T) {
	got, err := ParseGroups("C A B A")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(got, ",") != "A,B,C" {
		t.Fatalf("got %v", got)
	}
}

func TestLoadDuplicateSymbolAcrossSelectedGroups(t *testing.T) {
	t.Setenv("BAR_SEQ_LAB_MONGO_ENABLED", "false")
	t.Setenv("ALPACA_API_KEY", "k")
	t.Setenv("ALPACA_API_SECRET", "s")
	t.Setenv("BAR_SEQ_LAB_FEED", "test")
	t.Setenv("BAR_SEQ_LAB_SELECTED_GROUPS", "A,B")
	t.Setenv("BAR_SEQ_LAB_GROUP_A", "FAKEPACA")
	t.Setenv("BAR_SEQ_LAB_GROUP_B", "FAKEPACA")
	t.Setenv("BAR_SEQ_LAB_GROUP_C", "")
	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "duplicate symbol") {
		t.Fatalf("expected duplicate symbol, got %v", err)
	}
}

func TestLoadEmptySelectedGroup(t *testing.T) {
	t.Setenv("BAR_SEQ_LAB_MONGO_ENABLED", "false")
	t.Setenv("ALPACA_API_KEY", "k")
	t.Setenv("ALPACA_API_SECRET", "s")
	t.Setenv("BAR_SEQ_LAB_FEED", "test")
	t.Setenv("BAR_SEQ_LAB_SELECTED_GROUPS", "B")
	t.Setenv("BAR_SEQ_LAB_GROUP_A", "FAKEPACA")
	t.Setenv("BAR_SEQ_LAB_GROUP_B", "")
	t.Setenv("BAR_SEQ_LAB_GROUP_C", "")
	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "empty") {
		t.Fatalf("expected empty group, got %v", err)
	}
}

func TestLoadSelectedAOnly(t *testing.T) {
	t.Setenv("BAR_SEQ_LAB_MONGO_ENABLED", "false")
	t.Setenv("ALPACA_API_KEY", "k")
	t.Setenv("ALPACA_API_SECRET", "s")
	t.Setenv("BAR_SEQ_LAB_FEED", "test")
	t.Setenv("BAR_SEQ_LAB_SELECTED_GROUPS", "A")
	t.Setenv("BAR_SEQ_LAB_GROUP_A", "FAKEPACA")
	t.Setenv("BAR_SEQ_LAB_GROUP_B", "MSFT")
	t.Setenv("BAR_SEQ_LAB_GROUP_C", "NVDA")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.SubscribeSymbols) != 1 || cfg.SubscribeSymbols[0] != "FAKEPACA" {
		t.Fatalf("subscribe=%v", cfg.SubscribeSymbols)
	}
	if _, ok := cfg.SymbolToPartition["MSFT"]; ok {
		t.Fatal("inactive group B must not be in dispatcher map")
	}
}

func TestStartupLinesDoNotLeakSecrets(t *testing.T) {
	cfg := Config{
		Feed:             "iex",
		StreamURL:        IEXStreamURL,
		APIKey:           "SUPERSECRETKEY",
		APISecret:        "SUPERSECRETSECRET",
		DataDir:          "data",
		SelectedGroups:   []string{"A"},
		GroupA:           []string{"AAPL"},
		SubscribeSymbols: []string{"AAPL"},
		BufferCapacity:   8,
		FlushCount:       2,
		FlushInterval:    DefaultFlushTime,
	}
	for _, line := range StartupLines(cfg) {
		if strings.Contains(line, "SUPERSECRET") {
			t.Fatalf("secret leaked: %s", line)
		}
	}
}

func TestMissingSelectedGroups(t *testing.T) {
	t.Setenv("BAR_SEQ_LAB_MONGO_ENABLED", "false")
	t.Setenv("ALPACA_API_KEY", "k")
	t.Setenv("ALPACA_API_SECRET", "s")
	os.Unsetenv("BAR_SEQ_LAB_SELECTED_GROUPS")
	t.Setenv("BAR_SEQ_LAB_SELECTED_GROUPS", "")
	_, err := Load()
	if err == nil {
		t.Fatal("expected missing selected groups")
	}
}

func TestLoadMongoEnabledRequiresURIAndDB(t *testing.T) {
	t.Setenv("ALPACA_API_KEY", "k")
	t.Setenv("ALPACA_API_SECRET", "s")
	t.Setenv("BAR_SEQ_LAB_FEED", "test")
	t.Setenv("BAR_SEQ_LAB_SELECTED_GROUPS", "A")
	t.Setenv("BAR_SEQ_LAB_GROUP_A", "FAKEPACA")
	t.Setenv("BAR_SEQ_LAB_MONGO_ENABLED", "true")
	t.Setenv("BAR_SEQ_LAB_MONGO_URI", "")
	t.Setenv("BAR_SEQ_LAB_MONGO_DB", "")
	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "BAR_SEQ_LAB_MONGO_URI") {
		t.Fatalf("expected mongo uri required, got %v", err)
	}
}

func TestLoadMongoEnabledReadsCollectionDefault(t *testing.T) {
	t.Setenv("ALPACA_API_KEY", "k")
	t.Setenv("ALPACA_API_SECRET", "s")
	t.Setenv("BAR_SEQ_LAB_FEED", "test")
	t.Setenv("BAR_SEQ_LAB_SELECTED_GROUPS", "A")
	t.Setenv("BAR_SEQ_LAB_GROUP_A", "FAKEPACA")
	t.Setenv("BAR_SEQ_LAB_MONGO_ENABLED", "true")
	t.Setenv("BAR_SEQ_LAB_MONGO_URI", "mongodb://user:secret@127.0.0.1:27017")
	t.Setenv("BAR_SEQ_LAB_MONGO_DB", "bar_sequence_db")
	os.Unsetenv("BAR_SEQ_LAB_MONGO_COLLECTION")
	t.Setenv("BAR_SEQ_LAB_MONGO_COLLECTION", "")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.MongoEnabled || cfg.MongoDB != "bar_sequence_db" || cfg.MongoCollection != "bar_sequence" {
		t.Fatalf("mongo cfg %+v", cfg)
	}
	for _, line := range StartupLines(cfg) {
		if strings.Contains(line, "secret") || strings.Contains(line, "user:secret") {
			t.Fatalf("mongo secret leaked: %s", line)
		}
	}
}

func TestMongoHostStripsCredentials(t *testing.T) {
	got := MongoHost("mongodb://user:p@ss@127.0.0.1:27017/bar_sequence_db")
	if got != "127.0.0.1:27017" {
		t.Fatalf("got %q", got)
	}
}

func TestLoadTargetBarsPerSymbol(t *testing.T) {
	t.Setenv("BAR_SEQ_LAB_MONGO_ENABLED", "false")
	t.Setenv("ALPACA_API_KEY", "k")
	t.Setenv("ALPACA_API_SECRET", "s")
	t.Setenv("BAR_SEQ_LAB_FEED", "test")
	t.Setenv("BAR_SEQ_LAB_SELECTED_GROUPS", "A")
	t.Setenv("BAR_SEQ_LAB_GROUP_A", "FAKEPACA")
	t.Setenv("BAR_SEQ_LAB_TARGET_BARS_PER_SYMBOL", "120")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.TargetBarsPerSymbol != 120 {
		t.Fatalf("target=%d", cfg.TargetBarsPerSymbol)
	}
}
