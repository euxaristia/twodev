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
	if paths.WorkRoot != filepath.Join("site", "build-work") {
		t.Fatalf("work root = %q", paths.WorkRoot)
	}
	if paths.SearchIndex != filepath.Join("site", "index", "search.bleve") {
		t.Fatalf("search index = %q", paths.SearchIndex)
	}
}