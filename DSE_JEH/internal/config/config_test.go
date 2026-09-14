package config

import (
	"strconv"
	"testing"
	"time"

	dsejehv1 "tramuthus/dse-jeh-transsat-1/gen/dse_jeh/v1"
)

func TestLoadModes(t *testing.T) {
	t.Setenv("DSE_JEH_MODE", "OFFLINE")
	t.Setenv("DSE_JEH_OFFLINE_COLLECTION_RUN_ID", "20260911T161623Z-1")
	t.Setenv("DSE_JEH_RUN_TYPE", "RUN_A")
	t.Setenv("DSE_JEH_RISK_R", "0")
	t.Setenv("DSE_JEH_SERVER_ADDRESS", "")
	t.Setenv("DSE_JEH_STATUS_INTERVAL", "")
	t.Setenv("DSE_JEH_REMAIN_RUNNING", "")
	cfg, err := Load()
	if err != nil || cfg.Mode != dsejehv1.RuntimeMode_RUNTIME_MODE_OFFLINE || cfg.RunType != "RUN_A" || cfg.CollectionRunID != "20260911T161623Z-1" || cfg.MongoReservoirCollection != "capital_reservoir_events" || cfg.ServerAddress != "127.0.0.1:50052" || cfg.StatusInterval != 10*time.Second || !cfg.RemainRunning {
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
	t.Setenv("DSE_JEH_PIPELINE_RUN_ID", " bounded-run-a ")
	t.Setenv("DSE_JEH_SERVER_ADDRESS", "127.0.0.1:0")
	t.Setenv("DSE_JEH_STATUS_INTERVAL", "250ms")
	t.Setenv("DSE_JEH_REMAIN_RUNNING", "false")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.PipelineRunID != "bounded-run-a" || cfg.ServerAddress != "127.0.0.1:0" || cfg.StatusInterval != 250*time.Millisecond || cfg.RemainRunning {
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

func TestLoadAcceptsIndependentRunTypeAndRisk(t *testing.T) {
	for _, test := range []struct {
		runType string
		risk    string
	}{
		{runType: "RUN_C", risk: "1"},
		{runType: "RUN_D", risk: "0.20"},
		{runType: "RUN_E", risk: "0.20"},
		{runType: "RESERVOIR_VALIDATION_001", risk: "0.35"},
	} {
		t.Run(test.runType, func(t *testing.T) {
			t.Setenv("DSE_JEH_MODE", "OFFLINE")
			t.Setenv("DSE_JEH_OFFLINE_COLLECTION_RUN_ID", "20260911T161623Z-1")
			t.Setenv("DSE_JEH_RUN_TYPE", test.runType)
			t.Setenv("DSE_JEH_RISK_R", test.risk)
			cfg, err := Load()
			if err != nil {
				t.Fatal(err)
			}
			expectedRisk, err := strconv.ParseFloat(test.risk, 64)
			if err != nil {
				t.Fatal(err)
			}
			if string(cfg.RunType) != test.runType || cfg.Risk != expectedRisk {
				t.Fatalf("Load run_type=%q risk=%v", cfg.RunType, cfg.Risk)
			}
		})
	}
}

func TestLoadRejectsInvalidRunType(t *testing.T) {
	t.Setenv("DSE_JEH_MODE", "OFFLINE")
	t.Setenv("DSE_JEH_OFFLINE_COLLECTION_RUN_ID", "20260911T161623Z-1")
	t.Setenv("DSE_JEH_RUN_TYPE", "run_c")
	if _, err := Load(); err == nil {
		t.Fatal("Load must reject run_type outside the governed naming rule")
	}
}

func TestLoadRejectsRiskOutsideDomain(t *testing.T) {
	for _, risk := range []string{"-0.01", "1.01", "NaN"} {
		t.Run(risk, func(t *testing.T) {
			t.Setenv("DSE_JEH_MODE", "OFFLINE")
			t.Setenv("DSE_JEH_OFFLINE_COLLECTION_RUN_ID", "20260911T161623Z-1")
			t.Setenv("DSE_JEH_RUN_TYPE", "RUN_C")
			t.Setenv("DSE_JEH_RISK_R", risk)
			if _, err := Load(); err == nil {
				t.Fatalf("Load must reject Risk R=%s", risk)
			}
		})
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
