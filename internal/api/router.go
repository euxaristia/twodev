package api

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/euxaristia/twodev/internal/buildspec"
	"github.com/euxaristia/twodev/internal/store"
	"github.com/euxaristia/twodev/internal/version"
)

// Handler serves twodev JSON APIs under /~api/twodev/.
type Handler struct {
	projects *store.ProjectStore
	logger   *slog.Logger
}

// NewHandler creates an API handler.
func NewHandler(db *sql.DB, logger *slog.Logger) *Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{projects: store.NewProjectStore(db), logger: logger}
}

// Register mounts routes on mux.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /~api/twodev/version", h.handleVersion)
	mux.HandleFunc("POST /~api/twodev/buildspec/validate", h.handleValidateBuildSpec)
	mux.HandleFunc("GET /~api/twodev/projects", h.handleListProjects)
	mux.HandleFunc("POST /~api/twodev/projects", h.handleCreateProject)
}

func (h *Handler) handleVersion(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"name":    version.Name,
		"version": version.Version,
	})
}

func (h *Handler) handleValidateBuildSpec(w http.ResponseWriter, r *http.Request) {
	spec, err := readYAMLBody(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	parsed, err := buildspec.Parse(spec)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, buildspec.Summarize(parsed))
}

func (h *Handler) handleListProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := h.projects.List(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, projects)
}

func (h *Handler) handleCreateProject(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path        string `json:"path"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	project, err := h.projects.Create(r.Context(), req.Path, req.Name, req.Description)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, project)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}