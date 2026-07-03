package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/euxaristia/twodev/internal/agent/protocol"
	"github.com/gorilla/websocket"
)

const dispatchTimeout = 30 * time.Minute

type agentConn struct {
	token   string
	conn    *websocket.Conn
	pending map[string]chan protocol.Call
	info    AgentInfo
}

// Registry tracks connected agents and dispatches JSON RPC calls.
type Registry struct {
	mu     sync.RWMutex
	agents map[string]*agentConn
}

// NewRegistry creates an empty agent registry.
func NewRegistry() *Registry {
	return &Registry{agents: make(map[string]*agentConn)}
}

// Register adds an agent connection.
func (r *Registry) Register(token string, conn *websocket.Conn) *agentConn {
	prefix := token
	if len(prefix) > 8 {
		prefix = prefix[:8]
	}
	ac := &agentConn{
		token:   token,
		conn:    conn,
		pending: make(map[string]chan protocol.Call),
		info: AgentInfo{
			TokenPrefix: prefix,
			Online:      true,
		},
	}
	r.mu.Lock()
	r.agents[token] = ac
	r.mu.Unlock()
	return ac
}

// SetInfo updates metadata reported by an agent.
func (r *Registry) SetInfo(token string, info AgentInfo) {
	r.mu.RLock()
	ac, ok := r.agents[token]
	r.mu.RUnlock()
	if !ok {
		return
	}
	info.TokenPrefix = ac.info.TokenPrefix
	info.Online = true
	ac.info = info
}

// List returns currently connected agents.
func (r *Registry) List() []AgentInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]AgentInfo, 0, len(r.agents))
	for _, ac := range r.agents {
		out = append(out, ac.info)
	}
	return out
}

// Unregister removes an agent connection.
func (r *Registry) Unregister(token string) {
	r.mu.Lock()
	ac, ok := r.agents[token]
	delete(r.agents, token)
	r.mu.Unlock()
	if !ok {
		return
	}
	for id, ch := range ac.pending {
		close(ch)
		delete(ac.pending, id)
	}
}

// Connected returns the number of online agents.
func (r *Registry) Connected() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.agents)
}

// Complete resolves a pending call with a response from the agent.
func (r *Registry) Complete(token string, response protocol.Call) {
	r.mu.RLock()
	ac, ok := r.agents[token]
	r.mu.RUnlock()
	if !ok {
		return
	}
	ch, ok := ac.pending[response.ID]
	if !ok {
		return
	}
	select {
	case ch <- response:
	default:
	}
}

// Dispatch sends a call to the first connected agent and waits for a response.
func (r *Registry) Dispatch(ctx context.Context, call protocol.Call) (protocol.Call, error) {
	ac, err := r.pickAgent()
	if err != nil {
		return protocol.Call{}, err
	}
	return ac.dispatch(ctx, call)
}

func (r *Registry) pickAgent() (*agentConn, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, ac := range r.agents {
		return ac, nil
	}
	return nil, fmt.Errorf("no agents connected")
}

func (ac *agentConn) dispatch(ctx context.Context, call protocol.Call) (protocol.Call, error) {
	ch := make(chan protocol.Call, 1)
	ac.pending[call.ID] = ch
	defer func() { delete(ac.pending, call.ID) }()

	data, err := protocol.EncodeCall(call)
	if err != nil {
		return protocol.Call{}, err
	}
	frame, err := protocol.Encode(protocol.Message{Type: protocol.TypeRequest, Data: data})
	if err != nil {
		return protocol.Call{}, err
	}
	if err := ac.conn.WriteMessage(websocket.BinaryMessage, frame); err != nil {
		return protocol.Call{}, err
	}

	timeout, cancel := context.WithTimeout(ctx, dispatchTimeout)
	defer cancel()
	select {
	case <-timeout.Done():
		return protocol.Call{}, timeout.Err()
	case response := <-ch:
		if response.Error != "" {
			return response, fmt.Errorf("%s", response.Error)
		}
		return response, nil
	}
}