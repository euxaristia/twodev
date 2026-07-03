package api

import (
	"net/http"
	"path/filepath"
)

func (h *Handler) initProjectRepo(r *http.Request, projectPath string) error {
	if h.git == nil || h.repoRoot == "" || h.httpPort == 0 {
		return nil
	}
	repoDir := filepath.Join(h.repoRoot, projectPath+".git")
	if err := h.git.InitBareRepo(r.Context(), repoDir); err != nil {
		return err
	}
	return h.git.InstallPostReceiveHook(repoDir, projectPath, h.httpPort)
}