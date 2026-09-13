package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"tramuthus/dse-jeh-transsat-1/internal/config"
	runtimeapp "tramuthus/dse-jeh-transsat-1/internal/runtime"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	stats, err := runtimeapp.Run(ctx, cfg)
	if err != nil {
		return err
	}
	encoded, err := json.Marshal(stats)
	if err != nil {
		return fmt.Errorf("marshal runtime summary: %w", err)
	}
	fmt.Println(string(encoded))
	return nil
}
