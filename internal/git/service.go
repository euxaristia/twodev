package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CloneOptions configures a repository clone.
type CloneOptions struct {
	URL        string
	Branch     string
	Depth      int
	WithLFS    bool
	WithSubmodules bool
	DestDir    string
}

// Service wraps git CLI operations used by twodev.
type Service struct {
	gitPath string
}

// NewService creates a git service using the given git executable.
func NewService(gitPath string) *Service {
	if strings.TrimSpace(gitPath) == "" {
		gitPath = "git"
	}
	return &Service{gitPath: gitPath}
}

// Clone clones a repository into destDir.
func (s *Service) Clone(ctx context.Context, opts CloneOptions) error {
	if err := os.MkdirAll(filepath.Dir(opts.DestDir), 0o755); err != nil {
		return err
	}
	args := []string{"clone"}
	if opts.Depth > 0 {
		args = append(args, "--depth", fmt.Sprintf("%d", opts.Depth))
	}
	if opts.Branch != "" {
		args = append(args, "--branch", opts.Branch)
	}
	if opts.WithSubmodules {
		args = append(args, "--recurse-submodules")
	}
	args = append(args, opts.URL, opts.DestDir)
	if err := s.run(ctx, "", args...); err != nil {
		return err
	}
	if opts.WithLFS {
		return s.run(ctx, opts.DestDir, "lfs", "pull")
	}
	return nil
}

// Checkout checks out ref inside repoDir.
func (s *Service) Checkout(ctx context.Context, repoDir, ref string) error {
	return s.run(ctx, repoDir, "checkout", ref)
}

// Version returns the installed git version string.
func (s *Service) Version(ctx context.Context) (string, error) {
	out, err := s.output(ctx, "", "version")
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(out, "git version ") {
		return strings.TrimSpace(strings.TrimPrefix(out, "git version ")), nil
	}
	return strings.TrimSpace(out), nil
}

func (s *Service) run(ctx context.Context, dir string, args ...string) error {
	cmd := exec.CommandContext(ctx, s.gitPath, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (s *Service) output(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, s.gitPath, args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}