package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	agentserver "github.com/euxaristia/twodev/internal/agent/server"
	"github.com/euxaristia/twodev/internal/store"
)

func TestListAgentsAPIEmpty(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "twodev.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mux := http.NewServeMux()
	NewHandler(db, nil, HandlerConfig{Agents: agentserver.NewRegistry()}).Register(mux)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/~api/twodev/agents", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var agents []any
	if err := json.Unmarshal(rec.Body.Bytes(), &agents); err != nil {
		t.Fatal(err)
	}
	if len(agents) != 0 {
		t.Fatalf("expected empty list, got %d", len(agents))
	}
}