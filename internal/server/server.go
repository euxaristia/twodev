package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	agentserver "github.com/euxaristia/twodev/internal/agent/server"
	"github.com/euxaristia/twodev/internal/api"
	"github.com/euxaristia/twodev/internal/githttp"
	"github.com/euxaristia/twodev/internal/scheduler"
	"github.com/euxaristia/twodev/internal/sshserver"
	"github.com/euxaristia/twodev/internal/version"
	"golang.org/x/sync/errgroup"
)

// Server is the twodev HTTP entrypoint with composed subsystems.
type Server struct {
	opts   Options
	http   *http.Server
	ssh    *sshserver.Server
	queue  *scheduler.Queue
	worker *scheduler.Worker
}

// New creates a server from loaded options.
func New(opts Options) *Server {
	queue := scheduler.NewQueue()
	worker := scheduler.NewWorker(queue, func(ctx context.Context, req scheduler.JobRequest) error {
		opts.Logger.Info(
			"job scheduled (executor wiring pending)",
			"project", req.ProjectPath,
			"job", req.JobName,
			"build", req.BuildNumber,
		)
		return nil
	})

	var sshSrv *sshserver.Server
	if opts.Config.SSHPort != nil {
		sshSrv = sshserver.New(sshserver.Config{
			Host: opts.Config.HTTPHost,
			Port: *opts.Config.SSHPort,
		})
	}

	return &Server{
		opts:   opts,
		queue:  queue,
		worker: worker,
		ssh:    sshSrv,
	}
}

// ListenAndServe starts HTTP, optional SSH, and background workers until ctx is canceled.
func (s *Server) ListenAndServe(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handleHealth)

	api.NewHandler(s.opts.Database, s.opts.Logger).Register(mux)
	githttp.NewHandler(s.opts.Paths.RepoRoot).Register(mux)
	mux.Handle("/~server", agentserver.NewHandler(s.opts.AgentTokens, s.opts.Logger))

	addr := fmt.Sprintf("%s:%d", s.opts.Config.HTTPHost, s.opts.Config.HTTPPort)
	s.http = &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		s.opts.Logger.Info("twodev listening", "addr", addr, "version", version.Version)
		errCh := make(chan error, 1)
		go func() {
			errCh <- s.http.ListenAndServe()
		}()

		select {
		case <-ctx.Done():
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			return errors.Join(ctx.Err(), s.http.Shutdown(shutdownCtx))
		case err := <-errCh:
			if errors.Is(err, http.ErrServerClosed) {
				return nil
			}
			return err
		}
	})

	if s.ssh != nil {
		g.Go(func() error {
			s.opts.Logger.Info("ssh listening", "addr", fmt.Sprintf("%s:%d", s.opts.Config.HTTPHost, *s.opts.Config.SSHPort))
			if err := s.ssh.ListenAndServe(ctx); err != nil && ctx.Err() == nil {
				return err
			}
			return nil
		})
	}

	g.Go(func() error {
		if err := s.worker.Run(ctx); err != nil && ctx.Err() == nil {
			return err
		}
		return nil
	})

	return g.Wait()
}

// JobQueue exposes the in-memory scheduler for git hooks and triggers.
func (s *Server) JobQueue() *scheduler.Queue {
	return s.queue
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}