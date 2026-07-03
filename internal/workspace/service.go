package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Spec describes a developer workspace provisioned by an agent.
type Spec struct {
	ProjectID       int64
	WorkspaceNumber int
	ProjectPath     string
}

// Service tracks workspace directories under the site folder.
type Service struct {
	mu      sync.Mutex
	siteDir string
	active  map[string]Spec
}

// NewService creates a workspace service.
func NewService(siteDir string) *Service {
	return &Service{siteDir: siteDir, active: make(map[string]Spec)}
}

// Provision creates a workspace directory.
func (s *Service) Provision(spec Spec) (string, error) {
	key := fmt.Sprintf("%d:%d", spec.ProjectID, spec.WorkspaceNumber)
	dir := filepath.Join(s.siteDir, "workspaces", spec.ProjectPath, fmt.Sprintf("%d", spec.WorkspaceNumber))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	s.mu.Lock()
	s.active[key] = spec
	s.mu.Unlock()
	return dir, nil
}

// Delete removes a workspace directory.
func (s *Service) Delete(spec Spec) error {
	key := fmt.Sprintf("%d:%d", spec.ProjectID, spec.WorkspaceNumber)
	dir := filepath.Join(s.siteDir, "workspaces", spec.ProjectPath, fmt.Sprintf("%d", spec.WorkspaceNumber))
	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	s.mu.Lock()
	delete(s.active, key)
	s.mu.Unlock()
	return nil
}