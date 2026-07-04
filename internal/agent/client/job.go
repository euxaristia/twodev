package client

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/euxaristia/twodev/internal/agent/protocol"
	"github.com/euxaristia/twodev/internal/buildspec"
	"github.com/euxaristia/twodev/internal/job"
	"github.com/gorilla/websocket"
)

func (c *Client) handleRequest(ctx context.Context, conn *websocket.Conn, data []byte) error {
	call, err := protocol.DecodeCall(data)
	if err != nil {
		return err
	}
	switch call.Method {
	case protocol.MethodRunJob:
		result, runErr := c.runJob(ctx, conn, call.Payload)
		response := protocol.Call{ID: call.ID}
		if runErr != nil {
			response.Error = runErr.Error()
		} else {
			encoded, err := json.Marshal(result)
			if err != nil {
				response.Error = err.Error()
			} else {
				response.Result = encoded
			}
		}
		return c.sendResponse(conn, response)
	default:
		return c.sendResponse(conn, protocol.Call{ID: call.ID, Error: fmt.Sprintf("unknown method %q", call.Method)})
	}
}

func (c *Client) runJob(ctx context.Context, conn *websocket.Conn, payload json.RawMessage) (protocol.RunJobResult, error) {
	var req protocol.RunJobPayload
	if err := json.Unmarshal(payload, &req); err != nil {
		return protocol.RunJobResult{}, err
	}
	spec, err := buildspec.Parse(req.BuildSpec)
	if err != nil {
		return protocol.RunJobResult{}, err
	}

	workRoot := os.TempDir()
	logger := job.NewLogger(req.Token, os.Stdout)
	loggerSub := logger.Subscribe()
	go c.forwardLogs(conn, req.Token, loggerSub)

	executor := job.NewExecutorWithRepo(workRoot, req.RepoRoot, logger)
	jobCtx := job.Context{
		Token:       req.Token,
		ProjectID:   req.ProjectID,
		ProjectPath: req.ProjectPath,
		BuildNumber: req.BuildNumber,
		JobName:     req.JobName,
		Branch:      req.Branch,
		CommitHash:  req.CommitHash,
		RepoRoot:    req.RepoRoot,
		CloneURL:    req.CloneURL,
		StartedAt:   time.Now().UTC(),
	}
	if err := executor.RunJob(ctx, spec, req.JobName, jobCtx); err != nil {
		return protocol.RunJobResult{}, err
	}
	return protocol.RunJobResult{OK: true}, nil
}

func (c *Client) forwardLogs(conn *websocket.Conn, token string, lines <-chan string) {
	for line := range lines {
		payload := token + "::" + line
		frame, err := protocol.Encode(protocol.Message{Type: protocol.TypeLog, Data: []byte(payload)})
		if err != nil {
			return
		}
		_ = conn.WriteMessage(websocket.BinaryMessage, frame)
	}
}

func (c *Client) sendResponse(conn *websocket.Conn, response protocol.Call) error {
	data, err := protocol.EncodeCall(response)
	if err != nil {
		return err
	}
	frame, err := protocol.Encode(protocol.Message{Type: protocol.TypeResponse, Data: data})
	if err != nil {
		return err
	}
	return conn.WriteMessage(websocket.BinaryMessage, frame)
}

func (c *Client) sendAgentData(conn *websocket.Conn) error {
	info := map[string]any{
		"name":    "twodev-agent",
		"osName":  runtime.GOOS,
		"osArch":  runtime.GOARCH,
		"cpuCount": runtime.NumCPU(),
	}
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	frame, err := protocol.Encode(protocol.Message{Type: protocol.TypeAgentData, Data: data})
	if err != nil {
		return err
	}
	return conn.WriteMessage(websocket.BinaryMessage, frame)
}