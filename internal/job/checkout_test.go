package job

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/euxaristia/twodev/internal/buildspec"
	"github.com/euxaristia/twodev/internal/git"
)

func TestRunCheckoutAndCommand(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	root := t.TempDir()
	repoRoot := filepath.Join(root, "repositories")
	workRoot := filepath.Join(root, "build-work")
	svc := git.NewService("")
	bareDir := filepath.Join(repoRoot, "demo.git")
	if err := svc.InitBareRepo(context.Background(), bareDir); err != nil {
		t.Fatal(err)
	}
	if err := seedCheckoutRepo(context.Background(), svc, root, bareDir); err != nil {
		t.Fatal(err)
	}

	spec, err := buildspec.Parse(`
version: 1
jobs:
- name: ci
  steps:
  - type: CheckoutStep
    name: checkout
  - type: CommandStep
    name: verify
    interpreter:
      type: DefaultInterpreter
      commands: |
        cat readme.txt
`)
	if err != nil {
		t.Fatal(err)
	}

	logger := NewLogger("checkout-test", io.Discard)
	exec := NewExecutorWithRepo(workRoot, repoRoot, logger)
	ctx := Context{
		ProjectPath: "demo",
		BuildNumber: 1,
		JobName:     "ci",
		Branch:      "main",
		RepoRoot:    repoRoot,
	}
	ch := logger.Subscribe()
	if err := exec.RunJob(context.Background(), spec, "ci", ctx); err != nil {
		t.Fatal(err)
	}
	deadline := time.After(5 * time.Second)
	for {
		select {
		case line := <-ch:
			if strings.Contains(line, "checkout-ok") {
				return
			}
		case <-deadline:
			t.Fatal("expected command output after checkout")
		}
	}
}

func seedCheckoutRepo(ctx context.Context, svc *git.Service, root, bareDir string) error {
	work := filepath.Join(root, "work")
	if err := os.MkdirAll(work, 0o755); err != nil {
		return err
	}
	if err := svc.Run(ctx, work, "init"); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(work, "readme.txt"), []byte("checkout-ok"), 0o644); err != nil {
		return err
	}
	if err := svc.Run(ctx, work, "add", "readme.txt"); err != nil {
		return err
	}
	if err := svc.Run(ctx, work, "commit", "-m", "init"); err != nil {
		return err
	}
	if err := svc.Run(ctx, work, "branch", "-M", "main"); err != nil {
		return err
	}
	if err := svc.Run(ctx, work, "remote", "add", "origin", bareDir); err != nil {
		return err
	}
	return svc.Run(ctx, work, "push", "-u", "origin", "main")
}
