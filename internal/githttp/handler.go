package githttp

import (
	"fmt"
	"net/http"
	"strings"
)

// Handler serves git smart HTTP endpoints compatible with OneDev paths.
type Handler struct {
	repoRoot string
}

// NewHandler creates a git HTTP handler rooted at repoRoot.
func NewHandler(repoRoot string) *Handler {
	return &Handler{repoRoot: repoRoot}
}

// Register mounts git routes.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /{project...}/info/refs", h.handleInfoRefs)
	mux.HandleFunc("POST /{project...}/git-upload-pack", h.handleUploadPack)
	mux.HandleFunc("POST /{project...}/git-receive-pack", h.handleReceivePack)
}

func (h *Handler) handleInfoRefs(w http.ResponseWriter, r *http.Request) {
	service := r.URL.Query().Get("service")
	if service == "" {
		http.Error(w, "missing service parameter", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", fmt.Sprintf("application/x-%s-advertisement", strings.TrimPrefix(service, "git-")))
	w.Header().Set("Cache-Control", "no-cache")
	_, _ = w.Write([]byte(fmt.Sprintf("# service=%s\n0000", service)))
}

func (h *Handler) handleUploadPack(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/x-git-upload-pack-result")
	http.Error(w, "git upload-pack not implemented yet", http.StatusNotImplemented)
}

func (h *Handler) handleReceivePack(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/x-git-receive-pack-result")
	http.Error(w, "git receive-pack not implemented yet", http.StatusNotImplemented)
}