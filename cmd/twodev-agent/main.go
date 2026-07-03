package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/euxaristia/twodev/internal/agent/client"
	"github.com/euxaristia/twodev/internal/config"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	path := os.Getenv("TWODEV_AGENT_PROPERTIES")
	if path == "" {
		path = "conf/agent.properties"
	}

	cfg, err := config.LoadAgent(path)
	if err != nil {
		logger.Error("load agent config", "error", err, "path", path)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := client.New(cfg, logger).Run(ctx); err != nil && ctx.Err() == nil {
		logger.Error("agent stopped", "error", err)
		os.Exit(1)
	}
}