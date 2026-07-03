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

// IndexPull indexes a pull request document.
func (i *Indexer) IndexPull(projectPath string, pr model.PullRequest) error {
	if i == nil || i.index == nil {
		return nil
	}
	body := fmt.Sprintf("%s -> %s (%s)", pr.SourceBranch, pr.TargetBranch, pr.Status)
	return i.index.IndexDocument(Document{
		ID:      pullDocID(pr.ProjectID, pr.Number),
		Type:    "pull",
		Project: projectPath,
		Title:   pr.Title,
		Body:    body,
	})
}

// IndexBuild indexes a build document.
func (i *Indexer) IndexBuild(projectPath string, build model.Build) error {
	if i == nil || i.index == nil {
		return nil
	}
	body := fmt.Sprintf("%s #%d %s", build.JobName, build.Number, build.Status)
	if build.Branch != "" {
		body += " " + build.Branch
	}
	if build.CommitHash != "" {
		body += " " + build.CommitHash
	}
	return i.index.IndexDocument(Document{
		ID:      buildDocID(build.ProjectID, build.JobName, build.Number),
		Type:    "build",
		Project: projectPath,
		Title:   build.JobName,
		Body:    body,
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

		pullRows, err := i.db.QueryContext(ctx,
			`SELECT id, project_id, number, title, status, source_branch, target_branch FROM pull_requests WHERE project_id = ?`,
			project.ID,
		)
		if err != nil {
			return err
		}
		for pullRows.Next() {
			var pr model.PullRequest
			if err := pullRows.Scan(&pr.ID, &pr.ProjectID, &pr.Number, &pr.Title, &pr.Status, &pr.SourceBranch, &pr.TargetBranch); err != nil {
				pullRows.Close()
				return err
			}
			if err := i.IndexPull(project.Path, pr); err != nil {
				pullRows.Close()
				return err
			}
		}
		pullRows.Close()
		if err := pullRows.Err(); err != nil {
			return err
		}

		buildRows, err := i.db.QueryContext(ctx,
			`SELECT id, project_id, job_name, number, status, branch, commit_hash FROM builds WHERE project_id = ?`,
			project.ID,
		)
		if err != nil {
			return err
		}
		for buildRows.Next() {
			var build model.Build
			var branch, commit sql.NullString
			if err := buildRows.Scan(&build.ID, &build.ProjectID, &build.JobName, &build.Number, &build.Status, &branch, &commit); err != nil {
				buildRows.Close()
				return err
			}
			if branch.Valid {
				build.Branch = branch.String
			}
			if commit.Valid {
				build.CommitHash = commit.String
			}
			if err := i.IndexBuild(project.Path, build); err != nil {
				buildRows.Close()
				return err
			}
		}
		buildRows.Close()
		if err := buildRows.Err(); err != nil {
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

func pullDocID(projectID int64, number int) string {
	return fmt.Sprintf("pull:%d:%s", projectID, strconv.Itoa(number))
}

func buildDocID(projectID int64, jobName string, number int) string {
	return fmt.Sprintf("build:%d:%s:%s", projectID, jobName, strconv.Itoa(number))
}