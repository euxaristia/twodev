package build

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/euxaristia/twodev/internal/git"
	"github.com/euxaristia/twodev/internal/scheduler"
	"github.com/euxaristia/twodev/internal/store"
)

func TestRunnerExecutesQueuedBuild(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	root := t.TempDir()
	db, err := store.Open(filepath.Join(root, "twodev.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	projects := store.NewProjectStore(db)
	project, err := projects.Create(context.Background(), "demo/pipeline", "Demo", "")
	if err != nil {
		t.Fatal(err)
	}

	repoRoot := filepath.Join(root, "repositories")
	workRoot := filepath.Join(root, "build-work")
	svc := git.NewService("")
	bareDir := filepath.Join(repoRoot, "demo/pipeline.git")
	if err := svc.InitBareRepo(context.Background(), bareDir); err != nil {
		t.Fatal(err)
	}
	if err := seedRepoWithBuildspec(context.Background(), svc, root, bareDir); err != nil {
		t.Fatal(err)
	}

	builds := store.NewBuildStore(db)
	created, err := builds.Create(context.Background(), project.ID, "CI", "main", "")
	if err != nil {
		t.Fatal(err)
	}

	runner := NewRunner(db, repoRoot, workRoot, nil)
	req := scheduler.JobRequest{
		ProjectID:   project.ID,
		ProjectPath: project.Path,
		JobName:     created.JobName,
		BuildNumber: created.Number,
	}
	if err := runner.Handle(context.Background(), req); err != nil {
		t.Fatal(err)
	}

	got, err := builds.Get(context.Background(), project.ID, "CI", created.Number)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != store.BuildStatusSuccessful {
		t.Fatalf("status = %q, want %q", got.Status, store.BuildStatusSuccessful)
	}
	if got.FinishDate == nil {
		t.Fatal("expected finish date")
	}
}

func seedRepoWithBuildspec(ctx context.Context, svc *git.Service, root, bareDir string) error {
	work := filepath.Join(root, "work")
	if err := os.MkdirAll(work, 0o755); err != nil {
		return err
	}
	if err := svc.Run(ctx, work, "init"); err != nil {
		return err
	}
	spec := `version: 1
jobs:
- name: CI
  steps:
  - type: CommandStep
    name: echo
    interpreter:
      type: DefaultInterpreter
      commands: |
        echo pipeline-ok
`
	specPath := filepath.Join(work, ".onedev-buildspec.yml")
	if err := os.WriteFile(specPath, []byte(spec), 0o644); err != nil {
		return err
	}
	if err := svc.Run(ctx, work, "add", ".onedev-buildspec.yml"); err != nil {
		return err
	}
	if err := svc.Run(ctx, work, "commit", "-m", "add buildspec"); err != nil {
		return err
	}
	if err := svc.Run(ctx, work, "branch", "-M", "main"); err != nil {
		return err
	}
	if err := svc.Run(ctx, work, "remote", "add", "origin", bareDir); err != nil {
		return err
	}
	if err := svc.Run(ctx, work, "push", "-u", "origin", "main"); err != nil {
		return err
	}
	deadline := time.After(5 * time.Second)
	for {
		_, err := svc.ShowBlob(ctx, bareDir, "main", ".onedev-buildspec.yml")
		if err == nil {
			return nil
		}
		select {
		case <-deadline:
			return err
		case <-time.After(100 * time.Millisecond):
		}
	}
}