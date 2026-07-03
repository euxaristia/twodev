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