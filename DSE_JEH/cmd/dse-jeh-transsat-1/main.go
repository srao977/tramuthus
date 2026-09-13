package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"tramuthus/dse-jeh-transsat-1/internal/app"
	"tramuthus/dse-jeh-transsat-1/internal/config"
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
	stopFile := strings.TrimSpace(os.Getenv("DSE_JEH_STOP_FILE"))
	if stopFile != "" {
		_ = os.Remove(stopFile)
		go watchStopFile(ctx, cancel, stopFile, os.Getpid())
		defer os.Remove(stopFile)
	}
	application, err := app.New(cfg, app.Options{Output: os.Stdout})
	if err != nil {
		return err
	}
	return application.Run(ctx)
}

func watchStopFile(ctx context.Context, cancel context.CancelFunc, path string, processID int) {
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	want := strconv.Itoa(processID)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			content, err := os.ReadFile(path)
			if err == nil && strings.TrimSpace(string(content)) == want {
				cancel()
				return
			}
		}
	}
}
