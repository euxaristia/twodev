package model

import "time"

// Agent mirrors io.onedev.server.model.Agent core fields.
type Agent struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	IPAddress string    `json:"ipAddress,omitempty"`
	OSName    string    `json:"osName,omitempty"`
	OSVersion string    `json:"osVersion,omitempty"`
	OSArch    string    `json:"osArch,omitempty"`
	CPUCount  int       `json:"cpuCount"`
	Online    bool      `json:"online"`
	Paused    bool      `json:"paused"`
	CreatedAt time.Time `json:"createdAt"`
}