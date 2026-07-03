package search

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/euxaristia/twodev/internal/model"
	"github.com/euxaristia/twodev/internal/store"
)

// Indexer maintains the Bleve search index for projects and issues.
type Indexer struct {
	index    *Index
	projects *store.ProjectStore
	db       *sql.DB
}

// NewIndexer creates a search indexer.
func NewIndexer(db *sql.DB, index *Index) *Indexer {
	return &Indexer{
		index:    index,
		projects: store.NewProjectStore(db),
		db:       db,
	}
}

// IndexProject indexes a project document.
func (i *Indexer) IndexProject(project model.Project) error {
	if i == nil || i.index == nil {
		return nil
	}
	return i.index.IndexDocument(Document{
		ID:      projectDocID(project.ID),
		Type:    "project",
		Project: project.Path,
		Title:   project.Name,
		Body:    project.Description,
	})
}

// IndexIssue indexes an issue document.
func (i *Indexer) IndexIssue(projectPath string, issue model.Issue) error {
	if i == nil || i.index == nil {
		return nil
	}
	return i.index.IndexDocument(Document{
		ID:      issueDocID(issue.ProjectID, issue.Number),
		Type:    "issue",
		Project: projectPath,
		Title:   issue.Title,
		Body:    issue.Description,
	})
}

// Search returns matching documents for a query string.
func (i *Indexer) Search(query string, limit int) ([]Document, error) {
	if i == nil || i.index == nil {
		return nil, nil
	}
	return i.index.SearchDocuments(query, limit)
}

// RebuildAll reindexes projects and issues from the database.
func (i *Indexer) RebuildAll(ctx context.Context) error {
	if i == nil || i.index == nil {
		return nil
	}
	projects, err := i.projects.List(ctx)
	if err != nil {
		return err
	}
	for _, project := range projects {
		if err := i.IndexProject(project); err != nil {
			return err
		}
		rows, err := i.db.QueryContext(ctx,
			`SELECT id, project_id, number, title, state, description FROM issues WHERE project_id = ?`,
			project.ID,
		)
		if err != nil {
			return err
		}
		for rows.Next() {
			var issue model.Issue
			if err := rows.Scan(&issue.ID, &issue.ProjectID, &issue.Number, &issue.Title, &issue.State, &issue.Description); err != nil {
				rows.Close()
				return err
			}
			if err := i.IndexIssue(project.Path, issue); err != nil {
				rows.Close()
				return err
			}
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return err
		}
	}
	return nil
}

func projectDocID(id int64) string {
	return fmt.Sprintf("project:%d", id)
}

func issueDocID(projectID int64, number int) string {
	return fmt.Sprintf("issue:%d:%s", projectID, strconv.Itoa(number))
}