package api

import (
	"encoding/json"
	"net/http"

	"github.com/euxaristia/twodev/internal/scheduler"
)

func (h *Handler) handleListBuilds(w http.ResponseWriter, r *http.Request) {
	projectID, err := projectIDFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if _, err := h.projects.GetByID(r.Context(), projectID); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	builds, err := h.builds.List(r.Context(), projectID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, builds)
}

func (h *Handler) handleCreateBuild(w http.ResponseWriter, r *http.Request) {
	projectID, err := projectIDFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	project, err := h.projects.GetByID(r.Context(), projectID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	var req struct {
		JobName    string `json:"jobName"`
		Branch     string `json:"branch"`
		CommitHash string `json:"commitHash"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if req.JobName == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "jobName is required"})
		return
	}
	created, err := h.builds.Create(r.Context(), projectID, req.JobName, req.Branch, req.CommitHash)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if h.queue != nil {
		h.queue.Enqueue(scheduler.JobRequest{
			ProjectID:   projectID,
			ProjectPath: project.Path,
			JobName:     created.JobName,
			BuildNumber: created.Number,
		})
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *Handler) handleGetBuild(w http.ResponseWriter, r *http.Request) {
	projectID, err := projectIDFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	job, err := jobNameFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	number, err := buildNumberFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	got, err := h.builds.Get(r.Context(), projectID, job, number)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, got)
}