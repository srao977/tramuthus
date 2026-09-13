package config

import (
	"testing"

	dsejehv1 "tramuthus/dse-jeh-transsat-1/gen/dse_jeh/v1"
)

func TestLoadModes(t *testing.T) {
	t.Setenv("DSE_JEH_MODE", "OFFLINE")
	t.Setenv("DSE_JEH_OFFLINE_COLLECTION_RUN_ID", "20260911T161623Z-1")
	cfg, err := Load()
	if err != nil || cfg.Mode != dsejehv1.RuntimeMode_RUNTIME_MODE_OFFLINE || cfg.CollectionRunID != "20260911T161623Z-1" {
		t.Fatalf("Load OFFLINE = %+v, %v", cfg, err)
	}
	t.Setenv("DSE_JEH_MODE", "ONLINE")
	t.Setenv("DSE_JEH_SYMBOLS", "aapl, msft")
	cfg, err = Load()
	if err != nil || cfg.Mode != dsejehv1.RuntimeMode_RUNTIME_MODE_ONLINE || len(cfg.Symbols) != 2 {
		t.Fatalf("Load ONLINE = %+v, %v", cfg, err)
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
