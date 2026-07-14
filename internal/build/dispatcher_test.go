package build

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/euxaristia/twodev/internal/agent/protocol"
	agentserver "github.com/euxaristia/twodev/internal/agent/server"
	"github.com/euxaristia/twodev/internal/auth"
	"github.com/euxaristia/twodev/internal/config"
	"github.com/euxaristia/twodev/internal/git"
	"github.com/euxaristia/twodev/internal/scheduler"
	"github.com/euxaristia/twodev/internal/store"
	"github.com/gorilla/websocket"
)

func TestDispatcherUsesAgentWhenConnected(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	root := t.TempDir()
	db, err := store.Open(filepath.Join(root, "twodev.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	projects := store.NewProjectStore(db)
	project, err := projects.Create(context.Background(), "demo/agent", "Demo", "")
	if err != nil {
		t.Fatal(err)
	}
	builds := store.NewBuildStore(db)
	created, err := builds.Create(context.Background(), project.ID, "CI", "main", "")
	if err != nil {
		t.Fatal(err)
	}

	tokens := auth.StaticTokens{"agent-token": {}}
	registry := agentserver.NewRegistry()
	handler := agentserver.NewHandler(tokens, registry, nil)
	server := httptest.NewServer(handler)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, http.Header{
		"Authorization": []string{"Bearer agent-token"},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	_, _, _ = conn.ReadMessage()

	var seenCloneURL string
	go func() {
		for {
			_, frame, err := conn.ReadMessage()
			if err != nil {
				return
			}
			msg, err := protocol.Decode(frame)
			if err != nil || msg.Type != protocol.TypeRequest {
				continue
			}
			call, err := protocol.DecodeCall(msg.Data)
			if err != nil {
				continue
			}
			var payload protocol.RunJobPayload
			if err := json.Unmarshal(call.Payload, &payload); err == nil {
				seenCloneURL = payload.CloneURL
			}
			result, _ := json.Marshal(protocol.RunJobResult{OK: true})
			response, _ := protocol.EncodeCall(protocol.Call{ID: call.ID, Result: result})
			_ = conn.WriteMessage(websocket.BinaryMessage, mustProtocolFrame(protocol.TypeResponse, response))
		}
	}()

	repoRoot := filepath.Join(root, "repositories")
	workRoot := filepath.Join(root, "build-work")
	svc := git.NewService("")
	bareDir := filepath.Join(repoRoot, "demo/agent.git")
	if err := svc.InitBareRepo(context.Background(), bareDir); err != nil {
		t.Fatal(err)
	}
	if err := seedDispatcherRepo(context.Background(), svc, root, bareDir); err != nil {
		t.Fatal(err)
	}
	runner := NewRunner(db, repoRoot, workRoot, nil, nil)
	dispatcher := NewDispatcher(db, registry, runner, config.Server{HTTPHost: "127.0.0.1", HTTPPort: 6610}, nil)
	req := scheduler.JobRequest{
		ProjectID:   project.ID,
		ProjectPath: project.Path,
		JobName:     created.JobName,
		BuildNumber: created.Number,
	}
	if err := dispatcher.Handle(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	wantCloneURL := git.ProjectCloneURL("", "127.0.0.1", 6610, project.Path)
	if seenCloneURL != wantCloneURL {
		t.Fatalf("clone URL = %q, want %q", seenCloneURL, wantCloneURL)
	}

	got, err := builds.Get(context.Background(), project.ID, "CI", created.Number)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != store.BuildStatusSuccessful {
		t.Fatalf("status = %q", got.Status)
	}
}

func seedDispatcherRepo(ctx context.Context, svc *git.Service, root, bareDir string) error {
	work := filepath.Join(root, "work")
	if err := os.MkdirAll(work, 0o755); err != nil {
		return err
	}
	if err := svc.Run(ctx, work, "init"); err != nil {
		return err
	}
	spec := `version: 1
jobs:
- name: CI
  steps:
  - type: CommandStep
    name: echo
    interpreter:
      type: DefaultInterpreter
      commands: echo agent-dispatch
`
	if err := os.WriteFile(filepath.Join(work, ".onedev-buildspec.yml"), []byte(spec), 0o644); err != nil {
		return err
	}
	if err := svc.Run(ctx, work, "add", ".onedev-buildspec.yml"); err != nil {
		return err
	}
	if err := svc.Run(ctx, work, "commit", "-m", "add buildspec"); err != nil {
		return err
	}
	if err := svc.Run(ctx, work, "branch", "-M", "main"); err != nil {
		return err
	}
	if err := svc.Run(ctx, work, "remote", "add", "origin", bareDir); err != nil {
		return err
	}
	return svc.Run(ctx, work, "push", "-u", "origin", "main")
}

func mustProtocolFrame(typ protocol.Type, data []byte) []byte {
	frame, err := protocol.Encode(protocol.Message{Type: typ, Data: data})
	if err != nil {
		panic(err)
	}
	return frame
}