package model

import "time"

// PullRequest mirrors io.onedev.server.model.PullRequest core fields.
type PullRequest struct {
	ID           int64     `json:"id"`
	ProjectID    int64     `json:"projectId"`
	Number       int       `json:"number"`
	Title        string    `json:"title"`
	Status       string    `json:"status"`
	SourceBranch string    `json:"sourceBranch"`
	TargetBranch string    `json:"targetBranch"`
	CreatedAt    time.Time `json:"createdAt"`
}