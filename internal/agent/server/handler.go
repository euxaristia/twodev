package server

import (
	"log/slog"
	"net/http"
	"sync"

	"github.com/euxaristia/twodev/internal/agent/protocol"
	"github.com/euxaristia/twodev/internal/auth"
	"github.com/euxaristia/twodev/internal/version"
	"github.com/gorilla/websocket"
)

// AgentRegistry tracks connected agents.
type AgentRegistry struct {
	mu     sync.RWMutex
	agents map[string]*websocket.Conn
}

// NewAgentRegistry creates an empty registry.
func NewAgentRegistry() *AgentRegistry {
	return &AgentRegistry{agents: make(map[string]*websocket.Conn)}
}

// Handler serves the legacy /~server websocket endpoint.
type Handler struct {
	tokens   auth.TokenValidator
	registry *AgentRegistry
	logger   *slog.Logger
	upgrader websocket.Upgrader
}

// NewHandler creates an agent websocket handler.
func NewHandler(tokens auth.TokenValidator, logger *slog.Logger) *Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{
		tokens:   tokens,
		registry: NewAgentRegistry(),
		logger:   logger,
		upgrader: websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }},
	}
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

	h.registry.mu.Lock()
	h.registry.agents[token] = conn
	h.registry.mu.Unlock()
	defer func() {
		h.registry.mu.Lock()
		delete(h.registry.agents, token)
		h.registry.mu.Unlock()
	}()

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
			h.logger.Info("agent connected", "token_prefix", token[:min(8, len(token))])
		case protocol.TypeLog:
			h.logger.Info("job log", "payload", string(msg.Data))
		default:
			h.logger.Debug("agent message", "type", msg.Type.String())
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}