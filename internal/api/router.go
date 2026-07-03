package api

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"

	buildtrigger "github.com/euxaristia/twodev/internal/build"
	"github.com/euxaristia/twodev/internal/buildspec"
	"github.com/euxaristia/twodev/internal/git"
	"github.com/euxaristia/twodev/internal/issue"
	"github.com/euxaristia/twodev/internal/pullrequest"
	"github.com/euxaristia/twodev/internal/scheduler"
	"github.com/euxaristia/twodev/internal/store"
	"github.com/euxaristia/twodev/internal/version"
)

// Handler serves twodev JSON APIs under /~api/twodev/.
type Handler struct {
	projects *store.ProjectStore
	builds   *store.BuildStore
	issues   *issue.Service
	pulls    *pullrequest.Service
	queue    *scheduler.Queue
	trigger  *buildtrigger.Trigger
	repoRoot string
	httpPort int
	git      *git.Service
	logger   *slog.Logger
}

// HandlerConfig configures optional git and build trigger integration.
type HandlerConfig struct {
	Queue    *scheduler.Queue
	RepoRoot string
	HTTPPort int
}

// NewHandler creates an API handler.
func NewHandler(db *sql.DB, logger *slog.Logger, cfg HandlerConfig) *Handler {
	if logger == nil {
		logger = slog.Default()
	}
	h := &Handler{
		projects: store.NewProjectStore(db),
		builds:   store.NewBuildStore(db),
		issues:   issue.NewService(db),
		pulls:    pullrequest.NewService(db),
		queue:    cfg.Queue,
		repoRoot: cfg.RepoRoot,
		httpPort: cfg.HTTPPort,
		logger:   logger,
	}
	if cfg.Queue != nil && cfg.RepoRoot != "" {
		h.trigger = buildtrigger.NewTrigger(db, cfg.RepoRoot, cfg.Queue, logger)
		h.git = git.NewService("")
	}
	return h
}

// Register mounts routes on mux.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /~api/twodev/version", h.handleVersion)
	mux.HandleFunc("POST /~api/twodev/buildspec/validate", h.handleValidateBuildSpec)
	mux.HandleFunc("GET /~api/twodev/projects", h.handleListProjects)
	mux.HandleFunc("POST /~api/twodev/projects", h.handleCreateProject)
	mux.HandleFunc("GET /~api/twodev/projects/{id}", h.handleGetProject)

	mux.HandleFunc("GET /~api/twodev/projects/{id}/issues", h.handleListIssues)
	mux.HandleFunc("POST /~api/twodev/projects/{id}/issues", h.handleCreateIssue)
	mux.HandleFunc("GET /~api/twodev/projects/{id}/issues/{number}", h.handleGetIssue)

	mux.HandleFunc("GET /~api/twodev/projects/{id}/pulls", h.handleListPulls)
	mux.HandleFunc("POST /~api/twodev/projects/{id}/pulls", h.handleCreatePull)
	mux.HandleFunc("GET /~api/twodev/projects/{id}/pulls/{number}", h.handleGetPull)

	mux.HandleFunc("GET /~api/twodev/projects/{id}/builds", h.handleListBuilds)
	mux.HandleFunc("POST /~api/twodev/projects/{id}/builds", h.handleCreateBuild)
	mux.HandleFunc("GET /~api/twodev/projects/{id}/builds/{job}/{number}", h.handleGetBuild)

	mux.HandleFunc("POST /~api/twodev/git/branch-update", h.handleBranchUpdate)
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
	if err := h.initProjectRepo(r, project.Path); err != nil {
		h.logger.Error("init project repo failed", "path", project.Path, "error", err)
	}
	writeJSON(w, http.StatusCreated, project)
}

func (h *Handler) handleGetProject(w http.ResponseWriter, r *http.Request) {
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
	writeJSON(w, http.StatusOK, project)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}