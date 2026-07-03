package pullrequest

import (
	"context"
	"database/sql"
	"fmt"
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

// List returns pull requests for a project.
func (s *Service) List(ctx context.Context, projectID int64) ([]model.PullRequest, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, project_id, number, title, status, source_branch, target_branch, created_at
		 FROM pull_requests WHERE project_id = ? ORDER BY number DESC`, projectID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prs []model.PullRequest
	for rows.Next() {
		var pr model.PullRequest
		var created string
		if err := rows.Scan(&pr.ID, &pr.ProjectID, &pr.Number, &pr.Title, &pr.Status, &pr.SourceBranch, &pr.TargetBranch, &created); err != nil {
			return nil, err
		}
		pr.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", created)
		prs = append(prs, pr)
	}
	return prs, rows.Err()
}

// Get returns a pull request by project and number.
func (s *Service) Get(ctx context.Context, projectID int64, number int) (model.PullRequest, error) {
	var pr model.PullRequest
	var created string
	err := s.db.QueryRowContext(ctx,
		`SELECT id, project_id, number, title, status, source_branch, target_branch, created_at
		 FROM pull_requests WHERE project_id = ? AND number = ?`, projectID, number,
	).Scan(&pr.ID, &pr.ProjectID, &pr.Number, &pr.Title, &pr.Status, &pr.SourceBranch, &pr.TargetBranch, &created)
	if err == sql.ErrNoRows {
		return model.PullRequest{}, fmt.Errorf("pull request not found")
	}
	if err != nil {
		return model.PullRequest{}, err
	}
	pr.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", created)
	return pr, nil
}