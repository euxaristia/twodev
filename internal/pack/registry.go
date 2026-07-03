package pack

import (
	"fmt"
	"path/filepath"
	"sync"
)

// Package describes a published package artifact.
type Package struct {
	Project string
	Type    string
	Name    string
	Version string
}

// Registry stores package metadata in memory until blob storage is ported.
type Registry struct {
	mu       sync.RWMutex
	siteRoot string
	packages []Package
}

// NewRegistry creates a package registry rooted at site dir.
func NewRegistry(siteRoot string) *Registry {
	return &Registry{siteRoot: siteRoot}
}

// Publish records a package version.
func (r *Registry) Publish(pkg Package) error {
	if pkg.Project == "" || pkg.Type == "" || pkg.Name == "" || pkg.Version == "" {
		return fmt.Errorf("incomplete package metadata")
	}
	r.mu.Lock()
	r.packages = append(r.packages, pkg)
	r.mu.Unlock()
	return nil
}

// List returns packages for a project and type.
func (r *Registry) List(project, typ string) []Package {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []Package
	for _, pkg := range r.packages {
		if pkg.Project == project && pkg.Type == typ {
			out = append(out, pkg)
		}
	}
	return out
}

// BlobPath returns the on-disk blob path for a package version.
func (r *Registry) BlobPath(pkg Package) string {
	return filepath.Join(r.siteRoot, "packs", pkg.Project, pkg.Type, pkg.Name, pkg.Version)
}