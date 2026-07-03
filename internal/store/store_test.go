package store

import (
	"context"
	"path/filepath"
	"testing"
)

func TestOpenAndCreateProject(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "twodev.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	projects := NewProjectStore(db)
	created, err := projects.Create(context.Background(), "onedev/server", "OneDev", "Git server")
	if err != nil {
		t.Fatal(err)
	}
	if created.Path != "onedev/server" {
		t.Fatalf("path = %q", created.Path)
	}

	list, err := projects.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 project, got %d", len(list))
	}
}