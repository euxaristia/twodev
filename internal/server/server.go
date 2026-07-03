package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	agentserver "github.com/euxaristia/twodev/internal/agent/server"
	"github.com/euxaristia/twodev/internal/api"
	"github.com/euxaristia/twodev/internal/auth"
	buildrunner "github.com/euxaristia/twodev/internal/build"
	"github.com/euxaristia/twodev/internal/githttp"
	"github.com/euxaristia/twodev/internal/scheduler"
	"github.com/euxaristia/twodev/internal/sshserver"
	"github.com/euxaristia/twodev/internal/version"
	"golang.org/x/sync/errgroup"
)

// Server is the twodev HTTP entrypoint with composed subsystems.
type Server struct {
	opts        Options
	http        *http.Server
	ssh         *sshserver.Server
	queue       *scheduler.Queue
	worker      *scheduler.Worker
	agentServer *agentserver.Handler
}

// New creates a server from loaded options.
func New(opts Options) *Server {
	queue := scheduler.NewQueue()
	agentRegistry := agentserver.NewRegistry()
	agentHandler := agentserver.NewHandler(opts.AgentTokens, agentRegistry, opts.Logger)
	runner := buildrunner.NewRunner(opts.Database, opts.Paths.RepoRoot, opts.Paths.WorkRoot, opts.Logger)
	dispatcher := buildrunner.NewDispatcher(opts.Database, agentRegistry, runner, opts.Logger)
	worker := scheduler.NewWorker(queue, dispatcher.Handle)

	var sshSrv *sshserver.Server
	if opts.Config.SSHPort != nil {
		var err error
		sshSrv, err = sshserver.New(sshserver.Config{
			Host:     opts.Config.HTTPHost,
			Port:     *opts.Config.SSHPort,
			RepoRoot: opts.Paths.RepoRoot,
			HostKeyPath: filepath.Join(opts.Paths.SiteDir, "conf", "ssh_host_key"),
		})
		if err != nil {
			opts.Logger.Warn("ssh server disabled", "error", err)
			sshSrv = nil
		}
	}

	return &Server{
		opts:        opts,
		queue:       queue,
		worker:      worker,
		ssh:         sshSrv,
		agentServer: agentHandler,
	}
}

// ListenAndServe starts HTTP, optional SSH, and background workers until ctx is canceled.
func (s *Server) ListenAndServe(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handleHealth)

	accessGuard := auth.NewGuard(s.opts.AccessTokens)
	api.NewHandler(s.opts.Database, s.opts.Logger, api.HandlerConfig{
		Queue:    s.queue,
		RepoRoot: s.opts.Paths.RepoRoot,
		HTTPPort: s.opts.Config.HTTPPort,
		Guard:    accessGuard,
		Indexer:  s.opts.Indexer,
	}).Register(mux)
	githttp.NewHandler(s.opts.Paths.RepoRoot, accessGuard).Register(mux)
	mux.Handle("/~server", s.agentServer)

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
			return s.http.Shutdown(shutdownCtx)
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