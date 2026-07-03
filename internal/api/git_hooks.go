package api

import (
	"encoding/json"
	"net/http"

	"github.com/euxaristia/twodev/internal/auth"
)

func (h *Handler) handleBranchUpdate(w http.ResponseWriter, r *http.Request) {
	if h.guard != nil && h.guard.Enabled() && !h.guard.ValidRequest(r) && !auth.IsLoopback(r) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if h.trigger == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "branch triggers not configured"})
		return
	}
	var req struct {
		ProjectPath string `json:"projectPath"`
		Branch      string `json:"branch"`
		CommitHash  string `json:"commitHash"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if req.ProjectPath == "" || req.Branch == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "projectPath and branch are required"})
		return
	}
	builds, err := h.trigger.BranchUpdate(r.Context(), req.ProjectPath, req.Branch, req.CommitHash)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"builds": builds})
}