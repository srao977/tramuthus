package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"bar_sequence_lab/generator/internal/config"
	"bar_sequence_lab/generator/internal/runtime"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	engine, err := runtime.NewEngine(cfg)
	if err != nil {
		log.Fatalf("create engine: %v", err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := engine.Run(ctx, runtime.NewAlpacaSource(cfg)); err != nil {
		log.Fatalf("generator: %v", err)
	}
}
