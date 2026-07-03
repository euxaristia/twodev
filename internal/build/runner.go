package build

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/euxaristia/twodev/internal/buildspec"
	"github.com/euxaristia/twodev/internal/git"
	"github.com/euxaristia/twodev/internal/job"
	"github.com/euxaristia/twodev/internal/scheduler"
	"github.com/euxaristia/twodev/internal/store"
)

// Runner executes queued builds using the local job executor.
type Runner struct {
	builds   *store.BuildStore
	repoRoot string
	workRoot string
	git      *git.Service
	logger   *slog.Logger
}

// NewRunner creates a build runner.
func NewRunner(db *sql.DB, repoRoot, workRoot string, logger *slog.Logger) *Runner {
	if logger == nil {
		logger = slog.Default()
	}
	return &Runner{
		builds:   store.NewBuildStore(db),
		repoRoot: repoRoot,
		workRoot: workRoot,
		git:      git.NewService(""),
		logger:   logger,
	}
}

// Handle processes a scheduled job request.
func (r *Runner) Handle(ctx context.Context, req scheduler.JobRequest) error {
	build, err := r.builds.Get(ctx, req.ProjectID, req.JobName, req.BuildNumber)
	if err != nil {
		return err
	}

	if err := r.builds.UpdateStatus(ctx, build.ID, store.BuildStatusRunning, false); err != nil {
		return err
	}

	spec, err := r.loadBuildSpec(ctx, req.ProjectPath, build.Branch)
	if err != nil {
		r.failBuild(ctx, build.ID, err)
		return err
	}

	token := fmt.Sprintf("build-%d-%s-%d", req.ProjectID, req.JobName, req.BuildNumber)
	jobLogger := job.NewLogger(token, os.Stdout)
	executor := job.NewExecutorWithRepo(r.workRoot, r.repoRoot, jobLogger)
	jobCtx := job.Context{
		Token:       token,
		ProjectID:   req.ProjectID,
		ProjectPath: req.ProjectPath,
		BuildNumber: req.BuildNumber,
		JobName:     req.JobName,
		Branch:      build.Branch,
		CommitHash:  build.CommitHash,
		RepoRoot:    r.repoRoot,
		StartedAt:   time.Now().UTC(),
	}

	runErr := executor.RunJob(ctx, spec, req.JobName, jobCtx)
	status := store.BuildStatusSuccessful
	if runErr != nil {
		status = store.BuildStatusFailed
		r.logger.Error("build failed", "project", req.ProjectPath, "job", req.JobName, "build", req.BuildNumber, "error", runErr)
	} else {
		r.logger.Info("build finished", "project", req.ProjectPath, "job", req.JobName, "build", req.BuildNumber, "status", status)
	}
	if err := r.builds.UpdateStatus(ctx, build.ID, status, true); err != nil {
		return err
	}
	return runErr
}

func (r *Runner) loadBuildSpec(ctx context.Context, projectPath, branch string) (*buildspec.BuildSpec, error) {
	content, err := r.loadBuildSpecRaw(ctx, projectPath, branch)
	if err != nil {
		return nil, err
	}
	return buildspec.Parse(content)
}

func (r *Runner) loadBuildSpecRaw(ctx context.Context, projectPath, branch string) (string, error) {
	repoDir := filepath.Join(r.repoRoot, projectPath+".git")
	ref := strings.TrimSpace(branch)
	if ref == "" {
		ref = "HEAD"
	}
	content, err := r.git.ShowBlob(ctx, repoDir, ref, buildspec.BlobPath)
	if err != nil {
		return "", fmt.Errorf("load %s from %s: %w", buildspec.BlobPath, projectPath, err)
	}
	return string(content), nil
}

func (r *Runner) failBuild(ctx context.Context, buildID int64, cause error) {
	r.logger.Error("build aborted", "buildId", buildID, "error", cause)
	_ = r.builds.UpdateStatus(ctx, buildID, store.BuildStatusFailed, true)
}