package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/euxaristia/twodev/internal/agent/protocol"
	"github.com/euxaristia/twodev/internal/auth"
	"github.com/gorilla/websocket"
)

func TestRegistryDispatchRunJob(t *testing.T) {
	tokens := auth.StaticTokens{"agent-token": {}}
	registry := NewRegistry()
	handler := NewHandler(tokens, registry, nil)

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

	if _, _, err := conn.ReadMessage(); err != nil {
		t.Fatal(err)
	}
	info, _ := json.Marshal(map[string]string{"name": "test-agent"})
	if err := conn.WriteMessage(websocket.BinaryMessage, mustFrame(protocol.TypeAgentData, info)); err != nil {
		t.Fatal(err)
	}

	go func() {
		_, frame, err := conn.ReadMessage()
		if err != nil {
			return
		}
		msg, err := protocol.Decode(frame)
		if err != nil || msg.Type != protocol.TypeRequest {
			return
		}
		call, err := protocol.DecodeCall(msg.Data)
		if err != nil {
			return
		}
		result, _ := json.Marshal(protocol.RunJobResult{OK: true})
		response, _ := protocol.EncodeCall(protocol.Call{ID: call.ID, Result: result})
		_ = conn.WriteMessage(websocket.BinaryMessage, mustFrame(protocol.TypeResponse, response))
	}()

	payload, _ := json.Marshal(protocol.RunJobPayload{Token: "t1", JobName: "CI"})
	call := protocol.Call{ID: "test-1", Method: protocol.MethodRunJob, Payload: payload}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	response, err := registry.Dispatch(ctx, call)
	if err != nil {
		t.Fatal(err)
	}
	var result protocol.RunJobResult
	if err := json.Unmarshal(response.Result, &result); err != nil {
		t.Fatal(err)
	}
	if !result.OK {
		t.Fatal("expected ok result")
	}
}

func mustFrame(typ protocol.Type, data []byte) []byte {
	frame, err := protocol.Encode(protocol.Message{Type: typ, Data: data})
	if err != nil {
		panic(err)
	}
	return frame
}