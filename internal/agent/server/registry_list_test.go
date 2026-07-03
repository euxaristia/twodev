package server

import "testing"

func TestRegistryList(t *testing.T) {
	registry := NewRegistry()
	ac := &agentConn{
		token: "agent-token-long",
		info: AgentInfo{
			TokenPrefix: "agent-to",
			Online:      true,
		},
	}
	registry.mu.Lock()
	registry.agents["agent-token-long"] = ac
	registry.mu.Unlock()

	registry.SetInfo("agent-token-long", AgentInfo{Name: "twodev-agent", OSName: "linux"})
	list := registry.List()
	if len(list) != 1 || list[0].Name != "twodev-agent" {
		t.Fatalf("list = %+v", list)
	}
}