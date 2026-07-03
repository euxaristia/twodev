package server

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/euxaristia/twodev/internal/auth"
	"github.com/euxaristia/twodev/internal/config"
	"github.com/euxaristia/twodev/internal/store"
)

// LoadConfigFromEnv loads server.properties from env or the OneDev default path.
func LoadConfigFromEnv() (config.Server, error) {
	path := os.Getenv("TWODEV_SERVER_PROPERTIES")
	if path == "" {
		path = "server-product/system/conf/server.properties"
	}
	return config.LoadServer(path)
}

// Options holds everything required to run the twodev server.
type Options struct {
	Config      config.Server
	Paths       config.Paths
	Database    *sql.DB
	AgentTokens auth.StaticTokens
	Logger      *slog.Logger
}

// LoadOptionsFromEnv loads server.properties, opens the database, and resolves site paths.
func LoadOptionsFromEnv(logger *slog.Logger) (Options, error) {
	if logger == nil {
		logger = slog.Default()
	}

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		return Options{}, err
	}

	paths := config.ResolvePaths(config.SiteDirFromEnv())
	if err := ensurePaths(paths); err != nil {
		return Options{}, err
	}

	db, err := store.Open(paths.DatabasePath)
	if err != nil {
		return Options{}, fmt.Errorf("open database: %w", err)
	}

	tokens, err := config.LoadAgentTokens(paths.SiteDir)
	if err != nil {
		_ = db.Close()
		return Options{}, fmt.Errorf("load agent tokens: %w", err)
	}
	if len(tokens) == 0 {
		logger.Warn("no agent tokens configured; set TWODEV_AGENT_TOKENS or site/conf/agent-tokens.txt")
	}

	return Options{
		Config:      cfg,
		Paths:       paths,
		Database:    db,
		AgentTokens: tokens,
		Logger:      logger,
	}, nil
}

func ensurePaths(paths config.Paths) error {
	dirs := []string{
		filepath.Dir(paths.DatabasePath),
		paths.RepoRoot,
		filepath.Join(paths.SiteDir, "conf"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create %s: %w", dir, err)
		}
	}
	return nil
}