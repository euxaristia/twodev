package api

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/euxaristia/twodev/internal/scheduler"
	"github.com/euxaristia/twodev/internal/store"
)

func TestCreateBuildEnqueuesJob(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "twodev.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	projects := store.NewProjectStore(db)
	project, err := projects.Create(context.Background(), "demo/queue", "Demo", "")
	if err != nil {
		t.Fatal(err)
	}

	queue := scheduler.NewQueue()
	sub := queue.Subscribe()
	mux := http.NewServeMux()
	NewHandler(db, nil, queue).Register(mux)

	id := strconv.FormatInt(project.ID, 10)
	req := httptest.NewRequest(http.MethodPost, "/~api/twodev/projects/"+id+"/builds", bytes.NewBufferString(`{"jobName":"CI","branch":"main"}`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create build status = %d body=%s", rec.Code, rec.Body.String())
	}

	select {
	case job := <-sub:
		if job.ProjectID != project.ID {
			t.Fatalf("project id = %d, want %d", job.ProjectID, project.ID)
		}
		if job.ProjectPath != project.Path {
			t.Fatalf("project path = %q, want %q", job.ProjectPath, project.Path)
		}
		if job.JobName != "CI" || job.BuildNumber != 1 {
			t.Fatalf("unexpected job: %+v", job)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("expected enqueue notification")
	}
}