package githttp

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/euxaristia/twodev/internal/auth"
	"github.com/euxaristia/twodev/internal/git"
)

// Handler serves git smart HTTP endpoints compatible with OneDev paths.
type Handler struct {
	repoRoot string
	git      *git.Service
	guard    *auth.Guard
}

// NewHandler creates a git HTTP handler rooted at repoRoot.
func NewHandler(repoRoot string, guard *auth.Guard) *Handler {
	return &Handler{repoRoot: repoRoot, git: git.NewService(""), guard: guard}
}

// Register mounts git routes. Project paths may contain slashes.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /{path...}", h.routeGET)
	mux.HandleFunc("POST /{path...}", h.routePOST)
}

func (h *Handler) routeGET(w http.ResponseWriter, r *http.Request) {
	project, action, ok := splitGitPath(r.PathValue("path"))
	if !ok || action != "info-refs" {
		http.NotFound(w, r)
		return
	}
	if !h.authorized(w, r) {
		return
	}
	h.handleInfoRefs(w, r, project)
}

func (h *Handler) routePOST(w http.ResponseWriter, r *http.Request) {
	project, action, ok := splitGitPath(r.PathValue("path"))
	if !ok {
		http.NotFound(w, r)
		return
	}
	if !h.authorized(w, r) {
		return
	}
	switch action {
	case "upload-pack":
		h.serveRPC(w, r, project, "upload-pack", h.git.UploadPack)
	case "receive-pack":
		h.serveRPC(w, r, project, "receive-pack", h.git.ReceivePack)
	default:
		http.NotFound(w, r)
	}
}

func splitGitPath(full string) (project, action string, ok bool) {
	full = strings.TrimPrefix(full, "/")
	switch {
	case strings.HasSuffix(full, "/info/refs"):
		return strings.TrimSuffix(full, "/info/refs"), "info-refs", true
	case strings.HasSuffix(full, "/git-upload-pack"):
		return strings.TrimSuffix(full, "/git-upload-pack"), "upload-pack", true
	case strings.HasSuffix(full, "/git-receive-pack"):
		return strings.TrimSuffix(full, "/git-receive-pack"), "receive-pack", true
	default:
		return "", "", false
	}
}

func (h *Handler) handleInfoRefs(w http.ResponseWriter, r *http.Request, project string) {
	service := r.URL.Query().Get("service")
	if service != "git-upload-pack" && service != "git-receive-pack" {
		http.Error(w, "missing or invalid service parameter", http.StatusBadRequest)
		return
	}
	repoDir, err := h.repoDir(project)
	if err != nil {
		http.Error(w, "repository not found", http.StatusNotFound)
		return
	}
	body, err := h.git.AdvertiseRefs(r.Context(), repoDir, service, r.Header.Get("Git-Protocol"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", fmt.Sprintf("application/x-%s-advertisement", service))
	w.Header().Set("Cache-Control", "no-cache")
	_, _ = w.Write(body)
}

type rpcFunc func(context.Context, string, string, io.Reader, io.Writer) error

func (h *Handler) serveRPC(w http.ResponseWriter, r *http.Request, project, name string, fn rpcFunc) {
	repoDir, err := h.repoDir(project)
	if err != nil {
		http.Error(w, "repository not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", fmt.Sprintf("application/x-git-%s-result", name))
	w.WriteHeader(http.StatusOK)
	if err := fn(r.Context(), repoDir, r.Header.Get("Git-Protocol"), r.Body, w); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "git %s failed: %v\n", name, err)
	}
}

func (h *Handler) authorized(w http.ResponseWriter, r *http.Request) bool {
	if h.guard == nil || !h.guard.Enabled() || h.guard.ValidRequest(r) {
		return true
	}
	w.Header().Set("WWW-Authenticate", `Basic realm="twodev"`)
	http.Error(w, "forbidden", http.StatusForbidden)
	return false
}

func (h *Handler) repoDir(project string) (string, error) {
	project = strings.TrimSuffix(strings.TrimSpace(project), ".git")
	if project == "" {
		return "", fmt.Errorf("empty project path")
	}
	dir := filepath.Join(h.repoRoot, project+".git")
	if _, err := os.Stat(dir); err != nil {
		return "", err
	}
	return dir, nil
}