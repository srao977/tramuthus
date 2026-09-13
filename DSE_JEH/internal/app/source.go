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

func productionSource(cfg config.Config) Source {
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
