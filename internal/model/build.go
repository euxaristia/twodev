package model

import "time"

// Build mirrors io.onedev.server.model.Build core fields.
type Build struct {
	ID         int64     `json:"id"`
	ProjectID  int64     `json:"projectId"`
	JobName    string    `json:"jobName"`
	Number     int       `json:"number"`
	Status     string    `json:"status"`
	Branch     string    `json:"branch,omitempty"`
	CommitHash string    `json:"commitHash,omitempty"`
	SubmitDate time.Time `json:"submitDate"`
	FinishDate *time.Time `json:"finishDate,omitempty"`
}