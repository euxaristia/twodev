package search

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/euxaristia/twodev/internal/model"
	"github.com/euxaristia/twodev/internal/store"
)

func TestIndexerProjectAndIssue(t *testing.T) {
	root := t.TempDir()
	db, err := store.Open(filepath.Join(root, "twodev.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	index, err := Open(filepath.Join(root, "search.bleve"))
	if err != nil {
		t.Fatal(err)
	}
	defer index.Close()

	projects := store.NewProjectStore(db)
	project, err := projects.Create(context.Background(), "demo/search", "Search Demo", "full text")
	if err != nil {
		t.Fatal(err)
	}

	indexer := NewIndexer(db, index)
	if err := indexer.IndexProject(project); err != nil {
		t.Fatal(err)
	}
	issue := model.Issue{
		ProjectID:   project.ID,
		Number:      1,
		Title:       "Memory leak",
		Description: "heap grows without bound",
	}
	if err := indexer.IndexIssue(project.Path, issue); err != nil {
		t.Fatal(err)
	}

	docs, err := indexer.Search("heap", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 1 || docs[0].Type != "issue" {
		t.Fatalf("docs = %+v", docs)
	}
}

func TestIndexerPullAndBuild(t *testing.T) {
	root := t.TempDir()
	db, err := store.Open(filepath.Join(root, "twodev.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	index, err := Open(filepath.Join(root, "search.bleve"))
	if err != nil {
		t.Fatal(err)
	}
	defer index.Close()

	projects := store.NewProjectStore(db)
	project, err := projects.Create(context.Background(), "demo/search2", "Search", "")
	if err != nil {
		t.Fatal(err)
	}

	indexer := NewIndexer(db, index)
	pr := model.PullRequest{
		ProjectID:    project.ID,
		Number:       1,
		Title:        "Add metrics",
		Status:       "OPEN",
		SourceBranch: "feature/metrics",
		TargetBranch: "main",
	}
	if err := indexer.IndexPull(project.Path, pr); err != nil {
		t.Fatal(err)
	}
	build := model.Build{
		ProjectID: project.ID,
		JobName:   "CI",
		Number:    3,
		Status:    "SUCCESSFUL",
		Branch:    "main",
	}
	if err := indexer.IndexBuild(project.Path, build); err != nil {
		t.Fatal(err)
	}

	docs, err := indexer.Search("metrics", 10)
	if err != nil || len(docs) != 1 || docs[0].Type != "pull" {
		t.Fatalf("pull search = %+v err=%v", docs, err)
	}
	docs, err = indexer.Search("SUCCESSFUL", 10)
	if err != nil || len(docs) != 1 || docs[0].Type != "build" {
		t.Fatalf("build search = %+v err=%v", docs, err)
	}
}