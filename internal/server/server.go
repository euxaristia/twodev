package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/euxaristia/twodev/internal/config"
	"github.com/euxaristia/twodev/internal/version"
)

// Server is the twodev HTTP entrypoint.
type Server struct {
	cfg    config.Server
	logger *slog.Logger
	http   *http.Server
}

// New creates a server from server.properties-backed config.
func New(cfg config.Server, logger *slog.Logger) *Server {
	if logger == nil {
		logger = slog.Default()
	}
	return &Server{cfg: cfg, logger: logger}
}

// ListenAndServe starts the HTTP server until ctx is canceled.
func (s *Server) ListenAndServe(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealth)
	mux.HandleFunc("GET /~api/twodev/version", s.handleVersion)

	addr := fmt.Sprintf("%s:%d", s.cfg.HTTPHost, s.cfg.HTTPPort)
	s.http = &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
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

// LoadConfigFromEnv loads server.properties from env or the OneDev default path.
func LoadConfigFromEnv() (config.Server, error) {
	path := os.Getenv("TWODEV_SERVER_PROPERTIES")
	if path == "" {
		path = "server-product/system/conf/server.properties"
	}
	return config.LoadServer(path)
}