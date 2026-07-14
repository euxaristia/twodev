package job

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/euxaristia/twodev/internal/buildspec"
	"github.com/euxaristia/twodev/internal/git"
	"github.com/euxaristia/twodev/internal/githttp"
)

func TestRunCheckoutViaHTTPCloneURL(t *testing.T) {
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

	mux := http.NewServeMux()
	githttp.NewHandler(repoRoot, nil).Register(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	cloneURL := server.URL + "/demo.git"
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

	logger := NewLogger("checkout-http-test", io.Discard)
	exec := NewExecutorWithRepo(workRoot, "", logger)
	ctx := Context{
		ProjectPath: "demo",
		BuildNumber: 1,
		JobName:     "ci",
		Branch:      "main",
		CloneURL:    cloneURL,
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
			t.Fatal("expected command output after HTTP checkout")
		}
	}
}
