package issue

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/euxaristia/twodev/internal/model"
)

// Service provides issue CRUD backed by SQLite.
type Service struct {
	db *sql.DB
}

// NewService creates an issue service.
func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

// Create opens a new issue in a project.
func (s *Service) Create(ctx context.Context, projectID int64, title, state, description string) (model.Issue, error) {
	var number int
	if err := s.db.QueryRowContext(ctx,
		`SELECT COALESCE(MAX(number), 0) + 1 FROM issues WHERE project_id = ?`, projectID,
	).Scan(&number); err != nil {
		return model.Issue{}, err
	}
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO issues(project_id, number, title, state, description) VALUES (?, ?, ?, ?, ?)`,
		projectID, number, title, state, description,
	)
	if err != nil {
		return model.Issue{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return model.Issue{}, err
	}
	return model.Issue{
		ID:          id,
		ProjectID:   projectID,
		Number:      number,
		Title:       title,
		State:       state,
		Description: description,
		CreatedAt:   time.Now().UTC(),
	}, nil
}

// List returns issues for a project.
func (s *Service) List(ctx context.Context, projectID int64) ([]model.Issue, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, project_id, number, title, state, description, created_at
		 FROM issues WHERE project_id = ? ORDER BY number DESC`, projectID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var issues []model.Issue
	for rows.Next() {
		var issue model.Issue
		var created string
		if err := rows.Scan(&issue.ID, &issue.ProjectID, &issue.Number, &issue.Title, &issue.State, &issue.Description, &created); err != nil {
			return nil, err
		}
		issue.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", created)
		issues = append(issues, issue)
	}
	return issues, rows.Err()
}

// Get returns an issue by project and number.
func (s *Service) Get(ctx context.Context, projectID int64, number int) (model.Issue, error) {
	var issue model.Issue
	var created string
	err := s.db.QueryRowContext(ctx,
		`SELECT id, project_id, number, title, state, description, created_at
		 FROM issues WHERE project_id = ? AND number = ?`, projectID, number,
	).Scan(&issue.ID, &issue.ProjectID, &issue.Number, &issue.Title, &issue.State, &issue.Description, &created)
	if err == sql.ErrNoRows {
		return model.Issue{}, fmt.Errorf("issue not found")
	}
	if err != nil {
		return model.Issue{}, err
	}
	issue.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", created)
	return issue, nil
}