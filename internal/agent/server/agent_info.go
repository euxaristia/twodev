package server

// AgentInfo describes a connected build agent.
type AgentInfo struct {
	TokenPrefix string `json:"tokenPrefix"`
	Name        string `json:"name,omitempty"`
	OSName      string `json:"osName,omitempty"`
	OSArch      string `json:"osArch,omitempty"`
	CPUCount    int    `json:"cpuCount,omitempty"`
	Online      bool   `json:"online"`
}