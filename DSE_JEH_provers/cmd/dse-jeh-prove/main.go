// Command dse-jeh-prove runs the bounded CSV-driven DSE_JEH proving stage.
// Input is authoritative persisted phase CSV filtered by explicit experiment identity.
// Outputs are deterministic local strategy/decision JSONL, replay metadata, and a summary.
// It does not calculate phase, velocity, profitability, orders, portfolio state, or P&L.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"tramuthus/dse-jeh/internal/evidence"
	"tramuthus/dse-jeh/internal/inputcsv"
	"tramuthus/dse-jeh/internal/persistence"
	"tramuthus/dse-jeh/internal/proving"
)

// main reports proving failures through a non-zero process exit code.
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "DSE_JEH proving failed:", err)
		os.Exit(1)
	}
}

// run validates configuration, proves replay equality, writes local evidence, and reports counts.
func run() error {
	csvPath := flag.String("csv", "", "path to authoritative phase-angle CSV proving fixture")
	collectionRunID := flag.String("collection-run-id", "20260911T161623Z-1", "source collection_run_id")
	seriesSize := flag.Int("series-size", 120, "source phase experiment series_size")
	solverName := flag.String("solver-name", "EHLERS_DOMINANT_CYCLE_PHASE", "source solver_name")
	solverVersion := flag.String("solver-version", "V0.1", "source solver_version")
	inputSeriesType := flag.String("input-series-type", "MEDIAN_PRICE", "source input_series_type")
	outputRoot := flag.String("output", "output", "local deterministic proving evidence directory")
	flag.Parse()

	if *csvPath == "" {
		return fmt.Errorf("-csv is required")
	}
	absoluteCSV, err := filepath.Abs(*csvPath)
	if err != nil {
		return fmt.Errorf("resolve CSV path: %w", err)
	}
	experiment := inputcsv.Experiment{
		CollectionRunID: *collectionRunID, SeriesSize: *seriesSize,
		SolverName: *solverName, SolverVersion: *solverVersion, InputSeriesType: *inputSeriesType,
	}
	result, err := proving.Run(absoluteCSV, experiment)
	if err != nil {
		return err
	}
	summary := proving.Summarize(result.Evidence)
	sourceDigest, err := fileSHA256(absoluteCSV)
	if err != nil {
		return err
	}
	runID := "csv-prove-" + result.Digest[:16]
	outputDirectory, err := filepath.Abs(filepath.Join(*outputRoot, runID))
	if err != nil {
		return fmt.Errorf("resolve output directory: %w", err)
	}
	metadata := persistence.RunMetadata{
		RunID: runID, Mode: "CSV_PROVE_AND_REPLAY", SourceCSV: absoluteCSV,
		CollectionRunID: experiment.CollectionRunID, SeriesSize: experiment.SeriesSize,
		SolverName: experiment.SolverName, SolverVersion: experiment.SolverVersion,
		InputSeriesType: experiment.InputSeriesType, StrategyVersion: evidence.StrategyVersion,
		RankingMode: result.Evidence.RankingMode, SourceCSVSHA256: sourceDigest,
		InputCount:         len(result.Evidence.Inputs),
		StrategyEventCount: len(result.Evidence.StrategyEvents), DecisionEventCount: len(result.Evidence.DecisionEvents),
		DuplicateInputCount: result.Evidence.DuplicateInputCount, OutputDigest: result.Digest,
		ReplayDigest: result.ReplayDigest, ReplayMatched: result.ReplayMatch,
	}
	if err := writeEvidence(outputDirectory, metadata, result.Evidence); err != nil {
		return err
	}

	fmt.Printf("RUN_ID=%s\n", runID)
	fmt.Printf("CSV=%s\n", absoluteCSV)
	fmt.Printf("EXPERIMENT=%s|%d|%s|%s|%s\n", experiment.CollectionRunID, experiment.SeriesSize, experiment.SolverName, experiment.SolverVersion, experiment.InputSeriesType)
	fmt.Printf("INPUTS=%d SYMBOLS=%d OBSERVABLE=%d INITIALIZING=%d\n", summary.InputCount, summary.SymbolCount, summary.ObservableCount, summary.InitializingCount)
	fmt.Printf("INITIALIZING_ONLY_SYMBOLS=%v\n", summary.InitializingOnlySymbols)
	fmt.Printf("ZONES=MOMENTUM_HOLD:%d,HOP_OFF:%d,DISREGARD:%d,HOP_ON:%d\n", summary.ZoneCounts[evidence.ZoneMomentumHold], summary.ZoneCounts[evidence.ZoneHopOff], summary.ZoneCounts[evidence.ZoneDisregard], summary.ZoneCounts[evidence.ZoneHopOn])
	fmt.Printf("ZONE_TRANSITIONS=%d HOP_ON_EVENTS=%d HOP_OFF_EVENTS=%d DECISION_EVENTS=%d\n", summary.ZoneTransitionCount, summary.HopOnCandidateEvents, summary.HopOffCandidateEvents, summary.DecisionEvents)
	fmt.Printf("RANKING_MODE=%s CURRENT_HOP_ON_CANDIDATES=%v\n", result.Evidence.RankingMode, summary.HopOnCandidateSet)
	fmt.Printf("REPLAY_MATCH=%t DIGEST=%s\n", result.ReplayMatch, result.Digest)
	fmt.Printf("LOCAL_EVIDENCE=%s\n", outputDirectory)
	return nil
}

// writeEvidence separates admitted input, strategy, decision, final state, and run metadata files.
func writeEvidence(directory string, metadata persistence.RunMetadata, result evidence.RunResult) error {
	if err := persistence.WriteJSONL(filepath.Join(directory, "admitted_phase_input_evidence.jsonl"), result.Inputs); err != nil {
		return err
	}
	if err := persistence.WriteJSONL(filepath.Join(directory, "strategy_state_evidence.jsonl"), result.StrategyEvents); err != nil {
		return err
	}
	if err := persistence.WriteJSONL(filepath.Join(directory, "decision_evidence.jsonl"), result.DecisionEvents); err != nil {
		return err
	}
	if err := persistence.WriteJSON(filepath.Join(directory, "final_entity_states.json"), result.FinalStates); err != nil {
		return err
	}
	return persistence.WriteJSON(filepath.Join(directory, "run_metadata.json"), metadata)
}

// fileSHA256 identifies the exact immutable CSV fixture used by a proving run.
func fileSHA256(path string) (string, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("hash CSV fixture: %w", err)
	}
	digest := sha256.Sum256(contents)
	return hex.EncodeToString(digest[:]), nil
}
