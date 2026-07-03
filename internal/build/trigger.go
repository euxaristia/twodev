package build

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/euxaristia/twodev/internal/buildspec"
	"github.com/euxaristia/twodev/internal/git"
	"github.com/euxaristia/twodev/internal/model"
	"github.com/euxaristia/twodev/internal/scheduler"
	"github.com/euxaristia/twodev/internal/store"
)

// Trigger matches buildspec branch-update triggers and enqueues builds.
type Trigger struct {
	projects *store.ProjectStore
	builds   *store.BuildStore
	repoRoot string
	git      *git.Service
	queue    *scheduler.Queue
	logger   *slog.Logger
}

// NewTrigger creates a branch-update trigger service.
func NewTrigger(db *sql.DB, repoRoot string, queue *scheduler.Queue, logger *slog.Logger) *Trigger {
	if logger == nil {
		logger = slog.Default()
	}
	return &Trigger{
		projects: store.NewProjectStore(db),
		builds:   store.NewBuildStore(db),
		repoRoot: repoRoot,
		git:      git.NewService(""),
		queue:    queue,
		logger:   logger,
	}
}

// BranchUpdate creates and enqueues builds for jobs matching a branch push.
func (t *Trigger) BranchUpdate(ctx context.Context, projectPath, branch, commitHash string) ([]model.Build, error) {
	project, err := t.projects.GetByPath(ctx, projectPath)
	if err != nil {
		return nil, err
	}

	spec, err := t.loadBuildSpec(ctx, projectPath, branch)
	if err != nil {
		return nil, err
	}

	jobNames := buildspec.JobsForBranchUpdate(spec, branch, projectPath)
	if len(jobNames) == 0 {
		return nil, nil
	}

	created := make([]model.Build, 0, len(jobNames))
	for _, jobName := range jobNames {
		build, err := t.builds.Create(ctx, project.ID, jobName, branch, commitHash)
		if err != nil {
			return created, err
		}
		if t.queue != nil {
			t.queue.Enqueue(scheduler.JobRequest{
				ProjectID:   project.ID,
				ProjectPath: project.Path,
				JobName:     build.JobName,
				BuildNumber: build.Number,
			})
		}
		created = append(created, build)
		t.logger.Info("build triggered by push", "project", projectPath, "branch", branch, "job", jobName, "build", build.Number)
	}
	return created, nil
}

func (t *Trigger) loadBuildSpec(ctx context.Context, projectPath, branch string) (*buildspec.BuildSpec, error) {
	repoDir := filepath.Join(t.repoRoot, projectPath+".git")
	content, err := t.git.ShowBlob(ctx, repoDir, branch, buildspec.BlobPath)
	if err != nil {
		return nil, fmt.Errorf("load %s from %s@%s: %w", buildspec.BlobPath, projectPath, branch, err)
	}
	return buildspec.Parse(string(content))
}