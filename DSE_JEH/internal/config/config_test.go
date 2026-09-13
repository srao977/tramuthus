package config

import (
	"testing"
	"time"

	dsejehv1 "tramuthus/dse-jeh-transsat-1/gen/dse_jeh/v1"
)

func TestLoadModes(t *testing.T) {
	t.Setenv("DSE_JEH_MODE", "OFFLINE")
	t.Setenv("DSE_JEH_OFFLINE_COLLECTION_RUN_ID", "20260911T161623Z-1")
	cfg, err := Load()
	if err != nil || cfg.Mode != dsejehv1.RuntimeMode_RUNTIME_MODE_OFFLINE || cfg.CollectionRunID != "20260911T161623Z-1" || cfg.ServerAddress != "127.0.0.1:50052" || cfg.StatusInterval != 10*time.Second || !cfg.RemainRunning {
		t.Fatalf("Load OFFLINE = %+v, %v", cfg, err)
	}
	t.Setenv("DSE_JEH_MODE", "ONLINE")
	t.Setenv("DSE_JEH_SYMBOLS", "aapl, msft")
	cfg, err = Load()
	if err != nil || cfg.Mode != dsejehv1.RuntimeMode_RUNTIME_MODE_ONLINE || len(cfg.Symbols) != 2 {
		t.Fatalf("Load ONLINE = %+v, %v", cfg, err)
	}
}

func TestLoadServerRuntimeConfiguration(t *testing.T) {
	t.Setenv("DSE_JEH_MODE", "ONLINE")
	t.Setenv("DSE_JEH_SYMBOLS", "AAPL")
	t.Setenv("DSE_JEH_SERVER_ADDRESS", "127.0.0.1:0")
	t.Setenv("DSE_JEH_STATUS_INTERVAL", "250ms")
	t.Setenv("DSE_JEH_REMAIN_RUNNING", "false")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ServerAddress != "127.0.0.1:0" || cfg.StatusInterval != 250*time.Millisecond || cfg.RemainRunning {
		t.Fatalf("server runtime config = %+v", cfg)
	}
}

func TestLoadRejectsInvalidStatusInterval(t *testing.T) {
	t.Setenv("DSE_JEH_MODE", "ONLINE")
	t.Setenv("DSE_JEH_SYMBOLS", "AAPL")
	t.Setenv("DSE_JEH_STATUS_INTERVAL", "never")
	if _, err := Load(); err == nil {
		t.Fatal("Load must reject an invalid status interval")
	}
}

func TestLoadRejectsUnsupportedMode(t *testing.T) {
	t.Setenv("DSE_JEH_MODE", "TEST")
	if _, err := Load(); err == nil {
		t.Fatal("Load must reject unsupported mode")
	}
}

func TestLoadRequiresOfflineCollectionRun(t *testing.T) {
	t.Setenv("DSE_JEH_MODE", "OFFLINE")
	t.Setenv("DSE_JEH_OFFLINE_COLLECTION_RUN_ID", "")
	t.Setenv("DSE_JEH_COLLECTION_RUN_ID", "")
	if _, err := Load(); err == nil {
		t.Fatal("Load must reject OFFLINE mode without a collection run")
	}
}

func TestLoadRejectsConflictingCollectionRuns(t *testing.T) {
	t.Setenv("DSE_JEH_MODE", "OFFLINE")
	t.Setenv("DSE_JEH_OFFLINE_COLLECTION_RUN_ID", "new-run")
	t.Setenv("DSE_JEH_COLLECTION_RUN_ID", "legacy-run")
	if _, err := Load(); err == nil {
		t.Fatal("Load must reject conflicting collection-run configuration")
	}
}

func TestLoadOnlineIgnoresOfflineCollectionRunConfiguration(t *testing.T) {
	t.Setenv("DSE_JEH_MODE", "ONLINE")
	t.Setenv("DSE_JEH_SYMBOLS", "AAPL")
	t.Setenv("DSE_JEH_OFFLINE_COLLECTION_RUN_ID", "new-run")
	t.Setenv("DSE_JEH_COLLECTION_RUN_ID", "legacy-run")
	if _, err := Load(); err != nil {
		t.Fatalf("Load ONLINE must ignore OFFLINE collection-run configuration: %v", err)
	}
}
