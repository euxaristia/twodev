package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/euxaristia/twodev/internal/model"
)

const (
	BuildStatusPending   = "PENDING"
	BuildStatusRunning   = "RUNNING"
	BuildStatusSuccessful = "SUCCESSFUL"
	BuildStatusFailed    = "FAILED"
	BuildStatusCancelled = "CANCELLED"
)

// BuildStore persists CI builds.
type BuildStore struct {
	db *sql.DB
}

// NewBuildStore creates a build store.
func NewBuildStore(db *sql.DB) *BuildStore {
	return &BuildStore{db: db}
}

// Create inserts a build and returns it with an assigned build number.
func (s *BuildStore) Create(ctx context.Context, projectID int64, jobName, branch, commitHash string) (model.Build, error) {
	var number int
	if err := s.db.QueryRowContext(ctx,
		`SELECT COALESCE(MAX(number), 0) + 1 FROM builds WHERE project_id = ? AND job_name = ?`,
		projectID, jobName,
	).Scan(&number); err != nil {
		return model.Build{}, err
	}

	now := time.Now().UTC()
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO builds(project_id, job_name, number, status, branch, commit_hash, submit_date)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		projectID, jobName, number, BuildStatusPending, branch, commitHash, now.Format("2006-01-02 15:04:05"),
	)
	if err != nil {
		return model.Build{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return model.Build{}, err
	}
	return model.Build{
		ID:         id,
		ProjectID:  projectID,
		JobName:    jobName,
		Number:     number,
		Status:     BuildStatusPending,
		Branch:     branch,
		CommitHash: commitHash,
		SubmitDate: now,
	}, nil
}

// List returns builds for a project, newest first.
func (s *BuildStore) List(ctx context.Context, projectID int64) ([]model.Build, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, project_id, job_name, number, status, branch, commit_hash, submit_date, finish_date
		 FROM builds WHERE project_id = ? ORDER BY submit_date DESC`, projectID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var builds []model.Build
	for rows.Next() {
		build, err := scanBuild(rows)
		if err != nil {
			return nil, err
		}
		builds = append(builds, build)
	}
	return builds, rows.Err()
}

// Get returns a build by project, job name, and number.
func (s *BuildStore) Get(ctx context.Context, projectID int64, jobName string, number int) (model.Build, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, project_id, job_name, number, status, branch, commit_hash, submit_date, finish_date
		 FROM builds WHERE project_id = ? AND job_name = ? AND number = ?`,
		projectID, jobName, number,
	)
	build, err := scanBuild(row)
	if err == sql.ErrNoRows {
		return model.Build{}, fmt.Errorf("build not found")
	}
	return build, err
}

// UpdateStatus sets build status and optionally finish time.
func (s *BuildStore) UpdateStatus(ctx context.Context, buildID int64, status string, finished bool) error {
	if finished {
		_, err := s.db.ExecContext(ctx,
			`UPDATE builds SET status = ?, finish_date = ? WHERE id = ?`,
			status, time.Now().UTC().Format("2006-01-02 15:04:05"), buildID,
		)
		return err
	}
	_, err := s.db.ExecContext(ctx, `UPDATE builds SET status = ? WHERE id = ?`, status, buildID)
	return err
}

type buildRow interface {
	Scan(dest ...any) error
}

func scanBuild(row buildRow) (model.Build, error) {
	var build model.Build
	var branch, commit, submit, finish sql.NullString
	if err := row.Scan(
		&build.ID, &build.ProjectID, &build.JobName, &build.Number, &build.Status,
		&branch, &commit, &submit, &finish,
	); err != nil {
		return model.Build{}, err
	}
	if branch.Valid {
		build.Branch = branch.String
	}
	if commit.Valid {
		build.CommitHash = commit.String
	}
	if submit.Valid {
		build.SubmitDate, _ = time.Parse("2006-01-02 15:04:05", submit.String)
	}
	if finish.Valid {
		t, err := time.Parse("2006-01-02 15:04:05", finish.String)
		if err == nil {
			build.FinishDate = &t
		}
	}
	return build, nil
}