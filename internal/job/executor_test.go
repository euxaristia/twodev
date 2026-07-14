package job

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/euxaristia/twodev/internal/buildspec"
)

func TestRunCommandStep(t *testing.T) {
	if os.Getenv("CI") == "true" && os.Getenv("RUN_JOB_EXECUTOR_TEST") != "1" {
		t.Skip("optional integration test")
	}

	spec, err := buildspec.Parse(`
version: 1
jobs:
- name: test
  steps:
  - type: CommandStep
    name: echo
    interpreter:
      type: DefaultInterpreter
      commands: |
        echo hello-twodev
`)
	if err != nil {
		t.Fatal(err)
	}

	root := t.TempDir()
	logger := NewLogger("test-token", io.Discard)
	exec := NewExecutor(root, logger)
	ctx := Context{ProjectPath: "demo", BuildNumber: 1, JobName: "test"}
	ch := logger.Subscribe()
	if err := exec.RunJob(context.Background(), spec, "test", ctx); err != nil {
		t.Fatal(err)
	}
	deadline := time.After(2 * time.Second)
	found := false
	for !found {
		select {
		case line := <-ch:
			if strings.Contains(line, "hello-twodev") {
				found = true
			}
		case <-deadline:
			t.Fatal("expected command output in logs")
		}
	}

	workDir := filepath.Join(root, "demo", "builds", "test", "1", "work")
	if _, err := os.Stat(workDir); err != nil {
		t.Fatalf("work dir missing: %v", err)
	}
}

func TestCheckoutSource(t *testing.T) {
	logger := NewLogger("checkout-source-test", io.Discard)

	t.Run("clone url wins", func(t *testing.T) {
		e := NewExecutorWithRepo(t.TempDir(), "/ignored/repo/root", logger)
		got, err := e.checkoutSource(Context{CloneURL: "http://h/demo.git"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "http://h/demo.git" {
			t.Fatalf("got %q, want clone url", got)
		}
	})

	t.Run("repo root from context", func(t *testing.T) {
		root := t.TempDir()
		repoDir := filepath.Join(root, "demo.git")
		if err := os.MkdirAll(repoDir, 0o755); err != nil {
			t.Fatal(err)
		}
		e := NewExecutorWithRepo(t.TempDir(), "", logger)
		got, err := e.checkoutSource(Context{RepoRoot: root, ProjectPath: "demo"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != repoDir {
			t.Fatalf("got %q, want %q", got, repoDir)
		}
	})

	t.Run("repo root falls back to executor", func(t *testing.T) {
		root := t.TempDir()
		repoDir := filepath.Join(root, "demo.git")
		if err := os.MkdirAll(repoDir, 0o755); err != nil {
			t.Fatal(err)
		}
		e := NewExecutorWithRepo(t.TempDir(), root, logger)
		got, err := e.checkoutSource(Context{ProjectPath: "demo"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != repoDir {
			t.Fatalf("got %q, want %q", got, repoDir)
		}
	})

	t.Run("missing repo root errors", func(t *testing.T) {
		e := NewExecutorWithRepo(t.TempDir(), "", logger)
		if _, err := e.checkoutSource(Context{ProjectPath: "demo"}); err == nil {
			t.Fatal("expected error when no repo root or clone url")
		}
	})

	t.Run("missing repo dir errors", func(t *testing.T) {
		e := NewExecutorWithRepo(t.TempDir(), t.TempDir(), logger)
		if _, err := e.checkoutSource(Context{ProjectPath: "demo"}); err == nil {
			t.Fatal("expected error when local repo dir missing")
		}
	})
}
