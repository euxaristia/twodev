package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/euxaristia/twodev/internal/server"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	if err := run(logger); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	opts, err := server.LoadOptionsFromEnv(logger)
	if err != nil {
		return err
	}
	defer opts.Database.Close()
	if opts.SearchIndex != nil {
		defer opts.SearchIndex.Close()
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	err = server.New(opts).ListenAndServe(ctx)
	if err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	return nil
}