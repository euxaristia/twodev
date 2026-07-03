package build

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/euxaristia/twodev/internal/search"
	"github.com/euxaristia/twodev/internal/store"
)

func TestRunnerUpdatesSearchIndexOnStatusChange(t *testing.T) {
	root := t.TempDir()
	db, err := store.Open(filepath.Join(root, "twodev.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	index, err := search.Open(filepath.Join(root, "search.bleve"))
	if err != nil {
		t.Fatal(err)
	}
	defer index.Close()

	projects := store.NewProjectStore(db)
	project, err := projects.Create(context.Background(), "demo/index", "Demo", "")
	if err != nil {
		t.Fatal(err)
	}
	builds := store.NewBuildStore(db)
	build, err := builds.Create(context.Background(), project.ID, "CI", "main", "")
	if err != nil {
		t.Fatal(err)
	}

	runner := NewRunner(db, "", "", search.NewIndexer(db, index), nil)
	if err := runner.updateBuildStatus(context.Background(), project.Path, build, store.BuildStatusSuccessful, true); err != nil {
		t.Fatal(err)
	}

	docs, err := search.NewIndexer(db, index).Search("SUCCESSFUL", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 1 || docs[0].Type != "build" {
		t.Fatalf("docs = %+v", docs)
	}
}