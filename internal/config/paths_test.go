package config

import (
	"path/filepath"
	"testing"
)

func TestResolvePathsDefaults(t *testing.T) {
	paths := ResolvePaths("")
	if paths.SiteDir != "site" {
		t.Fatalf("site dir = %q", paths.SiteDir)
	}
	if paths.DatabasePath != filepath.Join("site", "database", "twodev.db") {
		t.Fatalf("database path = %q", paths.DatabasePath)
	}
	if paths.RepoRoot != filepath.Join("site", "repositories") {
		t.Fatalf("repo root = %q", paths.RepoRoot)
	}
}