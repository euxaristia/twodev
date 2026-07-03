package api

import (
	"net/http"
)

func (h *Handler) handleListAgents(w http.ResponseWriter, _ *http.Request) {
	if h.agents == nil {
		writeJSON(w, http.StatusOK, []any{})
		return
	}
	writeJSON(w, http.StatusOK, h.agents.List())
}