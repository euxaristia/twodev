package githttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/euxaristia/twodev/internal/git"
)

func TestInfoRefsUploadPack(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	root := t.TempDir()
	repoDir := filepath.Join(root, "demo.git")
	svc := git.NewService("")
	if err := svc.InitBareRepo(context.Background(), repoDir); err != nil {
		t.Fatal(err)
	}
	work := filepath.Join(root, "work")
	if err := seedBareRepo(context.Background(), svc, work, repoDir); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	NewHandler(root, nil).Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/demo.git/info/refs?service=git-upload-pack", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/x-git-upload-pack-advertisement" {
		t.Fatalf("content-type = %q", ct)
	}
	if !strings.Contains(rec.Body.String(), "service=git-upload-pack") {
		t.Fatal("expected service advertisement")
	}
}

func seedBareRepo(ctx context.Context, svc *git.Service, workDir, bareDir string) error {
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return err
	}
	if err := svc.Run(ctx, workDir, "init"); err != nil {
		return err
	}
	if err := svc.Run(ctx, workDir, "commit", "--allow-empty", "-m", "init"); err != nil {
		return err
	}
	if err := svc.Run(ctx, workDir, "branch", "-M", "main"); err != nil {
		return err
	}
	if err := svc.Run(ctx, workDir, "remote", "add", "origin", bareDir); err != nil {
		return err
	}
	return svc.Run(ctx, workDir, "push", "-u", "origin", "main")
}