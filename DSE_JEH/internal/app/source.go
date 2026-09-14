package app

import (
	"context"
	"fmt"

	dsejehv1 "tramuthus/dse-jeh-transsat-1/gen/dse_jeh/v1"
	"tramuthus/dse-jeh-transsat-1/internal/adapter/mongo"
	"tramuthus/dse-jeh-transsat-1/internal/config"
	"tramuthus/dse-jeh-transsat-1/internal/input/offline"
	"tramuthus/dse-jeh-transsat-1/internal/input/online"
)

type Source interface {
	Run(context.Context, func(*dsejehv1.BarEvent) error) error
}

type sourceFunc func(context.Context, func(*dsejehv1.BarEvent) error) error

func (function sourceFunc) Run(ctx context.Context, emit func(*dsejehv1.BarEvent) error) error {
	return function(ctx, emit)
}

func productionSource(cfg config.Config, terminal interface{ Event(string, string, ...any) }) Source {
	switch cfg.Mode {
	case dsejehv1.RuntimeMode_RUNTIME_MODE_ONLINE:
		consumer := online.New(cfg.GRPCAddress, cfg.Symbols, cfg.MaxBars, cfg.FinalizedOnly)
		return sourceFunc(consumer.Stream)
	case dsejehv1.RuntimeMode_RUNTIME_MODE_OFFLINE:
		return sourceFunc(func(ctx context.Context, emit func(*dsejehv1.BarEvent) error) error {
			repository, err := mongo.Open(ctx, cfg.MongoURI, cfg.MongoDatabase, cfg.MongoCollection)
			if err != nil {
				return err
			}
			defer repository.Close(context.Background())
			events, err := offline.New(repository, cfg.CollectionRunID).Events(ctx)
			if err != nil {
				return err
			}
			if len(events) == 0 {
				return fmt.Errorf("supplied collection_run_id %s contains no usable records", cfg.CollectionRunID)
			}
			symbols := make(map[string]struct{})
			minimumSequence := events[0].GetProvenance().GetEntitySequence()
			maximumSequence := minimumSequence
			for _, event := range events {
				symbols[event.GetEntityId()] = struct{}{}
				sequence := event.GetProvenance().GetEntitySequence()
				if sequence < minimumSequence {
					minimumSequence = sequence
				}
				if sequence > maximumSequence {
					maximumSequence = sequence
				}
			}
			terminal.Event("SOURCE VERIFY", "collection_run_id=%s | records=%d | distinct_symbols=%d | sequence_range=%d..%d | FROZEN", cfg.CollectionRunID, len(events), len(symbols), minimumSequence, maximumSequence)
			for _, event := range events {
				if err := emit(event); err != nil {
					return err
				}
			}
			return nil
		})
	default:
		return sourceFunc(func(context.Context, func(*dsejehv1.BarEvent) error) error {
			return fmt.Errorf("unsupported runtime mode %s", cfg.Mode)
		})
	}
}
