package model

import "time"

// Project mirrors io.onedev.server.model.Project core fields.
type Project struct {
	ID          int64     `json:"id"`
	Path        string    `json:"path"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}