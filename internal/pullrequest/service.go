package pullrequest

import (
	"context"
	"database/sql"
	"time"

	"github.com/euxaristia/twodev/internal/model"
)

// Service provides pull request persistence.
type Service struct {
	db *sql.DB
}

// NewService creates a pull request service.
func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

// Create opens a pull request.
func (s *Service) Create(ctx context.Context, projectID int64, title, status, sourceBranch, targetBranch string) (model.PullRequest, error) {
	var number int
	if err := s.db.QueryRowContext(ctx,
		`SELECT COALESCE(MAX(number), 0) + 1 FROM pull_requests WHERE project_id = ?`, projectID,
	).Scan(&number); err != nil {
		return model.PullRequest{}, err
	}
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO pull_requests(project_id, number, title, status, source_branch, target_branch)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		projectID, number, title, status, sourceBranch, targetBranch,
	)
	if err != nil {
		return model.PullRequest{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return model.PullRequest{}, err
	}
	return model.PullRequest{
		ID:           id,
		ProjectID:    projectID,
		Number:       number,
		Title:        title,
		Status:       status,
		SourceBranch: sourceBranch,
		TargetBranch: targetBranch,
		CreatedAt:    time.Now().UTC(),
	}, nil
}