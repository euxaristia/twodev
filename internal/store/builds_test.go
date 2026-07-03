package store

import (
	"context"
	"path/filepath"
	"testing"
)

func TestBuildStoreCreateListGet(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "twodev.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	projects := NewProjectStore(db)
	project, err := projects.Create(context.Background(), "demo/app", "Demo", "")
	if err != nil {
		t.Fatal(err)
	}

	builds := NewBuildStore(db)
	created, err := builds.Create(context.Background(), project.ID, "CI", "main", "abc123")
	if err != nil {
		t.Fatal(err)
	}
	if created.Number != 1 || created.Status != BuildStatusPending {
		t.Fatalf("unexpected build: %+v", created)
	}

	second, err := builds.Create(context.Background(), project.ID, "CI", "main", "def456")
	if err != nil {
		t.Fatal(err)
	}
	if second.Number != 2 {
		t.Fatalf("number = %d, want 2", second.Number)
	}

	list, err := builds.List(context.Background(), project.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 builds, got %d", len(list))
	}

	got, err := builds.Get(context.Background(), project.ID, "CI", 1)
	if err != nil {
		t.Fatal(err)
	}
	if got.CommitHash != "abc123" {
		t.Fatalf("commit = %q", got.CommitHash)
	}

	if err := builds.UpdateStatus(context.Background(), got.ID, BuildStatusSuccessful, true); err != nil {
		t.Fatal(err)
	}
	updated, err := builds.Get(context.Background(), project.ID, "CI", 1)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != BuildStatusSuccessful || updated.FinishDate == nil {
		t.Fatalf("expected finished successful build, got %+v", updated)
	}
}