package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/euxaristia/twodev/internal/model"
)

// ProjectStore persists projects.
type ProjectStore struct {
	db *sql.DB
}

// NewProjectStore creates a project store.
func NewProjectStore(db *sql.DB) *ProjectStore {
	return &ProjectStore{db: db}
}

// List returns all projects ordered by path.
func (s *ProjectStore) List(ctx context.Context) ([]model.Project, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, path, name, description, created_at FROM projects ORDER BY path`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []model.Project
	for rows.Next() {
		var p model.Project
		var created string
		if err := rows.Scan(&p.ID, &p.Path, &p.Name, &p.Description, &created); err != nil {
			return nil, err
		}
		p.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", created)
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

// Create inserts a project and returns it.
func (s *ProjectStore) Create(ctx context.Context, path, name, description string) (model.Project, error) {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO projects(path, name, description) VALUES (?, ?, ?)`,
		path, name, description,
	)
	if err != nil {
		return model.Project{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return model.Project{}, err
	}
	return model.Project{
		ID:          id,
		Path:        path,
		Name:        name,
		Description: description,
		CreatedAt:   time.Now().UTC(),
	}, nil
}

// GetByPath returns a project by path.
func (s *ProjectStore) GetByPath(ctx context.Context, path string) (model.Project, error) {
	var p model.Project
	var created string
	err := s.db.QueryRowContext(ctx,
		`SELECT id, path, name, description, created_at FROM projects WHERE path = ?`, path,
	).Scan(&p.ID, &p.Path, &p.Name, &p.Description, &created)
	if err == sql.ErrNoRows {
		return model.Project{}, fmt.Errorf("project not found")
	}
	if err != nil {
		return model.Project{}, err
	}
	p.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", created)
	return p, nil
}