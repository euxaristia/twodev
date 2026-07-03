package build

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/euxaristia/twodev/internal/agent/protocol"
	agentserver "github.com/euxaristia/twodev/internal/agent/server"
	"github.com/euxaristia/twodev/internal/scheduler"
	"github.com/euxaristia/twodev/internal/store"
)

// Dispatcher routes queued builds to connected agents, falling back to local execution.
type Dispatcher struct {
	builds   *store.BuildStore
	registry *agentserver.Registry
	local    *Runner
	logger   *slog.Logger
}

// NewDispatcher creates a build dispatcher.
func NewDispatcher(db *sql.DB, registry *agentserver.Registry, local *Runner, logger *slog.Logger) *Dispatcher {
	if logger == nil {
		logger = slog.Default()
	}
	return &Dispatcher{
		builds:   store.NewBuildStore(db),
		registry: registry,
		local:    local,
		logger:   logger,
	}
}

// Handle processes a scheduled job request.
func (d *Dispatcher) Handle(ctx context.Context, req scheduler.JobRequest) error {
	if d.registry != nil && d.registry.Connected() > 0 {
		if err := d.dispatchToAgent(ctx, req); err != nil {
			d.logger.Warn("agent dispatch failed, running locally", "error", err)
			return d.local.Handle(ctx, req)
		}
		return nil
	}
	return d.local.Handle(ctx, req)
}

func (d *Dispatcher) dispatchToAgent(ctx context.Context, req scheduler.JobRequest) error {
	build, err := d.builds.Get(ctx, req.ProjectID, req.JobName, req.BuildNumber)
	if err != nil {
		return err
	}
	if err := d.builds.UpdateStatus(ctx, build.ID, store.BuildStatusRunning, false); err != nil {
		return err
	}

	specYAML, err := d.local.loadBuildSpecRaw(ctx, req.ProjectPath, build.Branch)
	if err != nil {
		d.local.failBuild(ctx, build.ID, err)
		return err
	}

	token := fmt.Sprintf("build-%d-%s-%d", req.ProjectID, req.JobName, req.BuildNumber)
	payload, err := json.Marshal(protocol.RunJobPayload{
		Token:       token,
		ProjectID:   req.ProjectID,
		ProjectPath: req.ProjectPath,
		JobName:     req.JobName,
		BuildNumber: req.BuildNumber,
		BuildSpec:   specYAML,
		Branch:      build.Branch,
		CommitHash:  build.CommitHash,
		RepoRoot:    d.local.repoRoot,
	})
	if err != nil {
		d.local.failBuild(ctx, build.ID, err)
		return err
	}

	call := protocol.Call{
		ID:      "job-" + strconv.FormatInt(time.Now().UnixNano(), 10),
		Method:  protocol.MethodRunJob,
		Payload: payload,
	}
	response, err := d.registry.Dispatch(ctx, call)
	status := store.BuildStatusSuccessful
	if err != nil {
		status = store.BuildStatusFailed
		d.logger.Error("agent build failed", "project", req.ProjectPath, "job", req.JobName, "build", req.BuildNumber, "error", err)
	} else {
		var result protocol.RunJobResult
		if response.Result != nil {
			_ = json.Unmarshal(response.Result, &result)
		}
		if !result.OK {
			status = store.BuildStatusFailed
		}
		d.logger.Info("agent build finished", "project", req.ProjectPath, "job", req.JobName, "build", req.BuildNumber, "status", status)
	}
	if updateErr := d.builds.UpdateStatus(ctx, build.ID, status, true); updateErr != nil {
		return updateErr
	}
	return err
}