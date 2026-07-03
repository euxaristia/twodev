package model

import "time"

// Issue mirrors io.onedev.server.model.Issue core fields.
type Issue struct {
	ID          int64     `json:"id"`
	ProjectID   int64     `json:"projectId"`
	Number      int       `json:"number"`
	Title       string    `json:"title"`
	State       string    `json:"state"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}