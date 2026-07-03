package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/euxaristia/twodev/internal/store"
)

func TestIssueAndBuildAPI(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "twodev.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	projects := store.NewProjectStore(db)
	project, err := projects.Create(context.Background(), "demo/api", "Demo", "")
	if err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	NewHandler(db, nil, nil).Register(mux)
	id := strconv.FormatInt(project.ID, 10)

	createIssue := httptest.NewRequest(http.MethodPost, "/~api/twodev/projects/"+id+"/issues", bytes.NewBufferString(`{"title":"Bug","state":"Open"}`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, createIssue)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create issue status = %d body=%s", rec.Code, rec.Body.String())
	}

	listIssues := httptest.NewRequest(http.MethodGet, "/~api/twodev/projects/"+id+"/issues", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, listIssues)
	if rec.Code != http.StatusOK {
		t.Fatalf("list issues status = %d", rec.Code)
	}
	var issues []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &issues); err != nil {
		t.Fatal(err)
	}
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}

	createBuild := httptest.NewRequest(http.MethodPost, "/~api/twodev/projects/"+id+"/builds", bytes.NewBufferString(`{"jobName":"CI","branch":"main"}`))
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, createBuild)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create build status = %d body=%s", rec.Code, rec.Body.String())
	}
}