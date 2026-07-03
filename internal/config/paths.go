package config

import (
	"os"
	"path/filepath"
)

// Paths holds OneDev-compatible runtime directories under the site folder.
type Paths struct {
	SiteDir      string
	DatabasePath string
	RepoRoot     string
}

// ResolvePaths returns runtime paths for the given site directory.
func ResolvePaths(siteDir string) Paths {
	if siteDir == "" {
		siteDir = "site"
	}
	return Paths{
		SiteDir:      siteDir,
		DatabasePath: filepath.Join(siteDir, "database", "twodev.db"),
		RepoRoot:     filepath.Join(siteDir, "repositories"),
	}
}

// SiteDirFromEnv returns TWODEV_SITE_DIR or the default site folder.
func SiteDirFromEnv() string {
	if dir := os.Getenv("TWODEV_SITE_DIR"); dir != "" {
		return dir
	}
	return "site"
}