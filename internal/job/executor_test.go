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

