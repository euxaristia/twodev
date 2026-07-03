package api

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/euxaristia/twodev/internal/auth"
	"github.com/euxaristia/twodev/internal/store"
)

func TestAPIRequiresAccessTokenWhenConfigured(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "twodev.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mux := http.NewServeMux()
	NewHandler(db, nil, HandlerConfig{Guard: auth.NewGuard(auth.StaticTokens{"secret": {}})}).Register(mux)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/~api/twodev/version", nil))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/~api/twodev/version", nil)
	req.Header.Set("Authorization", "Bearer secret")
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}