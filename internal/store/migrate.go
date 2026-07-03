package store

import (
	"database/sql"
	"embed"
	"fmt"
	"sort"
	"strings"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

func migrate(db *sql.DB) error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version TEXT PRIMARY KEY NOT NULL,
		applied_at TEXT NOT NULL DEFAULT (datetime('now'))
	)`); err != nil {
		return err
	}

	entries, err := migrationFS.ReadDir("migrations")
	if err != nil {
		return err
	}
	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)

	for _, file := range files {
		version := strings.TrimSuffix(file, ".sql")
		var exists int
		if err := db.QueryRow(`SELECT COUNT(1) FROM schema_migrations WHERE version = ?`, version).Scan(&exists); err != nil {
			return err
		}
		if exists > 0 {
			continue
		}
		sqlBytes, err := migrationFS.ReadFile("migrations/" + file)
		if err != nil {
			return err
		}
		if _, err := db.Exec(string(sqlBytes)); err != nil {
			return fmt.Errorf("apply %s: %w", file, err)
		}
		if _, err := db.Exec(`INSERT INTO schema_migrations(version) VALUES (?)`, version); err != nil {
			return err
		}
	}
	return nil
}