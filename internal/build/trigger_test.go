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

func TestTriggerBranchUpdate(t *testing.T) {
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
	project, err := projects.Create(context.Background(), "demo/trigger", "Demo", "")
	if err != nil {
		t.Fatal(err)
	}

	repoRoot := filepath.Join(root, "repositories")
	svc := git.NewService("")
	bareDir := filepath.Join(repoRoot, "demo/trigger.git")
	if err := svc.InitBareRepo(context.Background(), bareDir); err != nil {
		t.Fatal(err)
	}
	if err := seedTriggeredRepo(context.Background(), svc, root, bareDir); err != nil {
		t.Fatal(err)
	}

	queue := scheduler.NewQueue()
	sub := queue.Subscribe()
	trigger := NewTrigger(db, repoRoot, queue, nil)
	builds, err := trigger.BranchUpdate(context.Background(), project.Path, "main", "deadbeef")
	if err != nil {
		t.Fatal(err)
	}
	if len(builds) != 1 {
		t.Fatalf("builds = %d, want 1", len(builds))
	}
	if builds[0].JobName != "CI" {
		t.Fatalf("job = %q", builds[0].JobName)
	}

	select {
	case job := <-sub:
		if job.JobName != "CI" {
			t.Fatalf("enqueued job = %q", job.JobName)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("expected enqueue")
	}
}

func seedTriggeredRepo(ctx context.Context, svc *git.Service, root, bareDir string) error {
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
      commands: echo ok
  triggers:
  - type: BranchUpdateTrigger
    branches: main
    projects: demo/trigger
`
	if err := os.WriteFile(filepath.Join(work, ".onedev-buildspec.yml"), []byte(spec), 0o644); err != nil {
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
	return svc.Run(ctx, work, "push", "-u", "origin", "main")
}