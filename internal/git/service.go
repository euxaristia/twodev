package git

import (
	"bytes"
	"context"
	"fmt"
	"io"
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

// GitPath returns the configured git executable path.
func (s *Service) GitPath() string {
	return s.gitPath
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

// AdvertiseRefs returns pkt-line ref advertisement for the given service.
// protocol is the Git-Protocol header value (e.g. "version=2"); pass empty for default.
func (s *Service) AdvertiseRefs(ctx context.Context, repoDir, service, protocol string) ([]byte, error) {
	command, err := gitRPCCommand(service)
	if err != nil {
		return nil, err
	}
	args := []string{"-C", repoDir, command, "--stateless-rpc", "--advertise-refs", "."}
	out, err := s.outputWithEnv(ctx, "", protocolEnv(protocol), args...)
	if err != nil {
		return nil, err
	}
	return AdvertiseService(service, []byte(out))
}

// UploadPack runs stateless RPC upload-pack, streaming request and response.
func (s *Service) UploadPack(ctx context.Context, repoDir, protocol string, request io.Reader, response io.Writer) error {
	return s.statelessRPC(ctx, repoDir, "upload-pack", protocol, request, response)
}

// ReceivePack runs stateless RPC receive-pack, streaming request and response.
func (s *Service) ReceivePack(ctx context.Context, repoDir, protocol string, request io.Reader, response io.Writer) error {
	return s.statelessRPC(ctx, repoDir, "receive-pack", protocol, request, response)
}

// Run executes git with args in dir.
func (s *Service) Run(ctx context.Context, dir string, args ...string) error {
	return s.run(ctx, dir, args...)
}

// ShowBlob reads a file from a bare or normal repository at ref.
func (s *Service) ShowBlob(ctx context.Context, repoDir, ref, path string) ([]byte, error) {
	spec := fmt.Sprintf("%s:%s", ref, path)
	out, err := s.output(ctx, "", "-C", repoDir, "show", spec)
	if err != nil {
		return nil, err
	}
	return []byte(out), nil
}

// InitBareRepo creates a bare repository at dir.
func (s *Service) InitBareRepo(ctx context.Context, dir string) error {
	if err := os.MkdirAll(filepath.Dir(dir), 0o755); err != nil {
		return err
	}
	return s.run(ctx, "", "init", "--bare", dir)
}

func gitRPCCommand(service string) (string, error) {
	switch service {
	case "git-upload-pack", "upload-pack":
		return "upload-pack", nil
	case "git-receive-pack", "receive-pack":
		return "receive-pack", nil
	default:
		return "", fmt.Errorf("unknown git service %q", service)
	}
}

func (s *Service) statelessRPC(ctx context.Context, repoDir, command, protocol string, request io.Reader, response io.Writer) error {
	cmd := exec.CommandContext(ctx, s.gitPath, "-C", repoDir, command, "--stateless-rpc", ".")
	cmd.Stdin = request
	cmd.Stdout = response
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if env := protocolEnv(protocol); len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}
	if err := cmd.Run(); err != nil {
		if msg := strings.TrimSpace(stderr.String()); msg != "" {
			return fmt.Errorf("%s: %s", command, msg)
		}
		return err
	}
	return nil
}

func protocolEnv(protocol string) []string {
	protocol = strings.TrimSpace(protocol)
	if protocol == "" {
		return nil
	}
	return []string{"GIT_PROTOCOL=" + protocol}
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
	return s.outputWithEnv(ctx, dir, nil, args...)
}

func (s *Service) outputWithEnv(ctx context.Context, dir string, env []string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, s.gitPath, args...)
	cmd.Dir = dir
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		if msg := strings.TrimSpace(stderr.String()); msg != "" {
			return "", fmt.Errorf("%w: %s", err, msg)
		}
		return "", err
	}
	return string(out), nil
}