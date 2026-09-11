// Command phase-angle-series-generator runs one finite Bar Sequence Lab experiment.
// It reads ordered raw Mongo bars, supplies only the requested prefix to a fresh
// phase solver, upserts derived evidence, reconciles counts, and exits nonzero on
// configuration, read, scientific, write, or raw-data-integrity failure.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"bar_sequence_lab/phase_angle_series_generator/internal/config"
	storemongo "bar_sequence_lab/phase_angle_series_generator/internal/mongo"
	"bar_sequence_lab/phase_angle_series_generator/internal/phase"
	"bar_sequence_lab/phase_angle_series_generator/internal/processing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type totals struct {
	selected, produced, observable, initializing, upserts, failures int64
}

type rawSnapshot map[string]int64

func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	cfg, err := config.Parse(args)
	if err != nil {
		return err
	}
	printConfiguration(cfg)
	store, err := storemongo.Open(ctx, cfg)
	if err != nil {
		return err
	}
	defer func() {
		closeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := store.Close(closeCtx); err != nil {
			fmt.Fprintf(os.Stderr, "Mongo close warning: %v\n", err)
		}
	}()

	before, err := snapshotRaw(ctx, store, cfg.CollectionRunID)
	if err != nil {
		return fmt.Errorf("raw snapshot before analysis: %w", err)
	}
	fmt.Printf("Raw before     : %d observations\n\n", before["TOTAL"])

	overall := totals{}
	writeFailed := false
	for _, partition := range []string{"A", "B", "C"} {
		symbols, err := store.Symbols(ctx, cfg.CollectionRunID, partition)
		if err != nil {
			return fmt.Errorf("discover partition %s symbols: %w", partition, err)
		}
		fmt.Printf("Partition %s\n  Discovered symbols: %v\n", partition, symbols)
		group := totals{}
		for _, symbol := range symbols {
			filter := bson.M{"collection_run_id": cfg.CollectionRunID, "partition_id": partition, "symbol": symbol}
			available, err := store.CountRaw(ctx, filter)
			if err != nil {
				return fmt.Errorf("count %s/%s: %w", partition, symbol, err)
			}
			bars, err := store.Prefix(ctx, cfg.CollectionRunID, partition, symbol, cfg.SeriesSize)
			if err != nil {
				return fmt.Errorf("read %s/%s prefix: %w", partition, symbol, err)
			}
			records, err := processing.Analyze(bars, cfg.SeriesSize, time.Now().UTC())
			if err != nil {
				return fmt.Errorf("analyze %s/%s: %w", partition, symbol, err)
			}
			symbolTotals := totals{selected: int64(len(bars)), produced: int64(len(records))}
			for _, record := range records {
				if record.PhaseObservable {
					symbolTotals.observable++
				} else {
					symbolTotals.initializing++
				}
				if err := store.Upsert(ctx, record); err != nil {
					symbolTotals.failures++
					writeFailed = true
					fmt.Fprintf(os.Stderr, "  WRITE FAILURE %s sequence=%d: %v\n", symbol, record.GeneratorSequenceNo, err)
				} else {
					symbolTotals.upserts++
				}
			}
			if available < int64(cfg.SeriesSize) {
				fmt.Printf("  %-8s available=%d requested=%d selected=%d SHORT SEQUENCE produced=%d observable=%d initializing=%d upserts=%d failures=%d\n",
					symbol, available, cfg.SeriesSize, symbolTotals.selected, symbolTotals.produced, symbolTotals.observable, symbolTotals.initializing, symbolTotals.upserts, symbolTotals.failures)
			} else {
				fmt.Printf("  %-8s available=%d requested=%d selected=%d produced=%d observable=%d initializing=%d upserts=%d failures=%d\n",
					symbol, available, cfg.SeriesSize, symbolTotals.selected, symbolTotals.produced, symbolTotals.observable, symbolTotals.initializing, symbolTotals.upserts, symbolTotals.failures)
			}
			add(&group, symbolTotals)
		}
		fmt.Printf("  %s totals: selected=%d produced=%d observable=%d initializing=%d upserts=%d failures=%d\n\n",
			partition, group.selected, group.produced, group.observable, group.initializing, group.upserts, group.failures)
		add(&overall, group)
	}

	after, err := snapshotRaw(ctx, store, cfg.CollectionRunID)
	if err != nil {
		return fmt.Errorf("raw snapshot after analysis: %w", err)
	}
	fmt.Printf("Overall totals : selected=%d produced=%d observable=%d initializing=%d upserts=%d failures=%d\n", overall.selected, overall.produced, overall.observable, overall.initializing, overall.upserts, overall.failures)
	fmt.Printf("Raw after      : %d observations\n", after["TOTAL"])
	if !equalSnapshots(before, after) {
		return fmt.Errorf("raw Mongo counts changed during analysis: before=%v after=%v", before, after)
	}
	fmt.Println("Raw protection : PASS (total, partition, and symbol counts unchanged)")
	if writeFailed {
		return fmt.Errorf("one or more derived Mongo upserts failed")
	}
	return nil
}

func printConfiguration(cfg config.Config) {
	fmt.Println("Bar Sequence Lab - Phase Angle Series Generator")
	fmt.Printf("Collection Run  : %s\n", cfg.CollectionRunID)
	fmt.Printf("SERIES_SIZE     : %d\n", cfg.SeriesSize)
	fmt.Printf("Mongo DB        : %s\n", cfg.MongoDB)
	fmt.Printf("Raw Collection  : %s\n", cfg.RawCollection)
	fmt.Printf("Phase Collection: %s\n", cfg.PhaseCollection)
	fmt.Printf("Solver          : %s\n", phase.SolverName)
	fmt.Printf("Solver Version  : %s\n", phase.SolverVersion)
	fmt.Printf("Input Series    : %s\n", phase.InputSeriesType)
	fmt.Printf("Lookback        : %d bars\n\n", phase.Lookback)
}

func snapshotRaw(ctx context.Context, store *storemongo.Store, runID string) (rawSnapshot, error) {
	snapshot := rawSnapshot{}
	total, err := store.CountRaw(ctx, bson.M{"collection_run_id": runID})
	if err != nil {
		return nil, err
	}
	snapshot["TOTAL"] = total
	for _, partition := range []string{"A", "B", "C"} {
		count, err := store.CountRaw(ctx, bson.M{"collection_run_id": runID, "partition_id": partition})
		if err != nil {
			return nil, err
		}
		snapshot["PARTITION:"+partition] = count
		symbols, err := store.Symbols(ctx, runID, partition)
		if err != nil {
			return nil, err
		}
		for _, symbol := range symbols {
			count, err := store.CountRaw(ctx, bson.M{"collection_run_id": runID, "partition_id": partition, "symbol": symbol})
			if err != nil {
				return nil, err
			}
			snapshot["SYMBOL:"+partition+":"+symbol] = count
		}
	}
	return snapshot, nil
}

func equalSnapshots(left, right rawSnapshot) bool {
	if len(left) != len(right) {
		return false
	}
	for key, value := range left {
		if right[key] != value {
			return false
		}
	}
	return true
}

func add(destination *totals, source totals) {
	destination.selected += source.selected
	destination.produced += source.produced
	destination.observable += source.observable
	destination.initializing += source.initializing
	destination.upserts += source.upserts
	destination.failures += source.failures
}
