package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/euxaristia/twodev/internal/server"
	"github.com/euxaristia/twodev/internal/sshserver"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	opts, err := server.LoadOptionsFromEnv(logger)
	if err != nil {
		logger.Error("load server options", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if opts.Config.SSHPort != nil {
		go func() {
			sshSrv := sshserver.New(sshserver.Config{
				Host: opts.Config.HTTPHost,
				Port: *opts.Config.SSHPort,
			})
			if err := sshSrv.ListenAndServe(ctx); err != nil && ctx.Err() == nil {
				logger.Error("ssh server stopped", "error", err)
			}
		}()
	}

	if err := server.New(opts).ListenAndServe(ctx); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}