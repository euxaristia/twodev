package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	agentserver "github.com/euxaristia/twodev/internal/agent/server"
	"github.com/euxaristia/twodev/internal/api"
	"github.com/euxaristia/twodev/internal/auth"
	"github.com/euxaristia/twodev/internal/config"
	"github.com/euxaristia/twodev/internal/githttp"
	"github.com/euxaristia/twodev/internal/store"
	"github.com/euxaristia/twodev/internal/version"
)

// Options configures the twodev HTTP server.
type Options struct {
	Config      config.Server
	Logger      *slog.Logger
	Database    *sql.DB
	AgentTokens auth.TokenValidator
	SiteDir     string
}

// Server is the twodev HTTP entrypoint.
type Server struct {
	cfg    config.Server
	logger *slog.Logger
	db     *sql.DB
	tokens auth.TokenValidator
	site   string
	http   *http.Server
}

// New creates a server from options.
func New(opts Options) *Server {
	if opts.Logger == nil {
		opts.Logger = slog.Default()
	}
	return &Server{
		cfg:    opts.Config,
		logger: opts.Logger,
		db:     opts.Database,
		tokens: opts.AgentTokens,
		site:   opts.SiteDir,
	}
}

// ListenAndServe starts the HTTP server until ctx is canceled.
func (s *Server) ListenAndServe(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealth)

	if s.db != nil {
		api.NewHandler(s.db, s.logger).Register(mux)
	} else {
		mux.HandleFunc("GET /~api/twodev/version", s.handleVersion)
	}

	if s.tokens != nil {
		mux.Handle("GET /~server", agentserver.NewHandler(s.tokens, s.logger))
	}

	if s.site != "" {
		githttp.NewHandler(filepath.Join(s.site, "projects")).Register(mux)
	}

	addr := fmt.Sprintf("%s:%d", s.cfg.HTTPHost, s.cfg.HTTPPort)
	s.http = &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		s.logger.Info("twodev listening", "addr", addr, "version", version.Version)
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
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleVersion(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"name":    version.Name,
		"version": version.Version,
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// LoadOptionsFromEnv loads server options from environment and defaults.
func LoadOptionsFromEnv(logger *slog.Logger) (Options, error) {
	path := os.Getenv("TWODEV_SERVER_PROPERTIES")
	if path == "" {
		path = "server-product/system/conf/server.properties"
	}
	cfg, err := config.LoadServer(path)
	if err != nil {
		return Options{}, err
	}

	opts := Options{Config: cfg, Logger: logger}
	siteDir := os.Getenv("TWODEV_SITE_DIR")
	if siteDir == "" {
		siteDir = "server-product/system/site"
	}
	opts.SiteDir = siteDir

	dbPath := os.Getenv("TWODEV_DATABASE")
	if dbPath == "" {
		dbPath = filepath.Join(siteDir, "twodev.db")
	}
	if os.Getenv("TWODEV_DISABLE_DATABASE") != "1" {
		db, err := store.Open(dbPath)
		if err != nil {
			return Options{}, err
		}
		opts.Database = db
	}

	if token := os.Getenv("TWODEV_AGENT_TOKEN"); token != "" {
		opts.AgentTokens = auth.StaticTokens{token: {}}
	}

	return opts, nil
}