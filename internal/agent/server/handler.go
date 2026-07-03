package server

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/euxaristia/twodev/internal/agent/protocol"
	"github.com/euxaristia/twodev/internal/auth"
	"github.com/euxaristia/twodev/internal/version"
	"github.com/gorilla/websocket"
)

// Handler serves the legacy /~server websocket endpoint.
type Handler struct {
	tokens   auth.TokenValidator
	registry *Registry
	logger   *slog.Logger
	upgrader websocket.Upgrader
}

// NewHandler creates an agent websocket handler.
func NewHandler(tokens auth.TokenValidator, registry *Registry, logger *slog.Logger) *Handler {
	if logger == nil {
		logger = slog.Default()
	}
	if registry == nil {
		registry = NewRegistry()
	}
	return &Handler{
		tokens:   tokens,
		registry: registry,
		logger:   logger,
		upgrader: websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }},
	}
}

// Registry exposes connected agents for job dispatch.
func (h *Handler) Registry() *Registry {
	return h.registry
}

// ServeHTTP upgrades connections and speaks the OneDev agent protocol.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	token, err := auth.ParseBearer(r.Header.Get("Authorization"))
	if err != nil || !h.tokens.Valid(token) {
		http.Error(w, "A valid agent token is expected", http.StatusForbidden)
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("websocket upgrade failed", "error", err)
		return
	}
	defer conn.Close()

	update, err := protocol.Encode(protocol.Message{Type: protocol.TypeUpdate, Data: []byte(version.Version)})
	if err != nil {
		return
	}
	if err := conn.WriteMessage(websocket.BinaryMessage, update); err != nil {
		return
	}

	ac := h.registry.Register(token, conn)
	defer h.registry.Unregister(token)

	for {
		_, frame, err := conn.ReadMessage()
		if err != nil {
			return
		}
		msg, err := protocol.Decode(frame)
		if err != nil {
			return
		}
		switch msg.Type {
		case protocol.TypeAgentData:
			var info struct {
				Name string `json:"name"`
			}
			if err := json.Unmarshal(msg.Data, &info); err == nil && info.Name != "" {
				h.logger.Info("agent connected", "name", info.Name, "token_prefix", token[:min(8, len(token))])
			} else {
				h.logger.Info("agent connected", "token_prefix", token[:min(8, len(token))])
			}
		case protocol.TypeResponse:
			response, err := protocol.DecodeCall(msg.Data)
			if err != nil {
				h.logger.Error("invalid agent response", "error", err)
				continue
			}
			h.registry.Complete(token, response)
		case protocol.TypeLog:
			h.logger.Info("job log", "payload", string(msg.Data))
		default:
			h.logger.Debug("agent message", "type", msg.Type.String())
		}
		_ = ac
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}