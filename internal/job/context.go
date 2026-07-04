package job

import (
	"path/filepath"
	"time"
)

// Status represents build job lifecycle state.
type Status string

const (
	StatusPending   Status = "PENDING"
	StatusRunning   Status = "RUNNING"
	StatusSuccessful Status = "SUCCESSFUL"
	StatusFailed    Status = "FAILED"
	StatusCancelled Status = "CANCELLED"
)

// Context carries runtime metadata for a single job execution.
type Context struct {
	Token         string
	ProjectID     int64
	ProjectPath   string
	BuildNumber   int
	JobName       string
	SubmitSequence int
	WorkDir       string
	Branch        string
	CommitHash    string
	RepoRoot      string
	CloneURL      string
	StartedAt     time.Time
}

// BuildDir returns the workspace directory for this job.
func (c Context) BuildDir(base string) string {
	return filepath.Join(base, c.ProjectPath, "builds", c.JobName, itoa(c.BuildNumber))
}

func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	neg := v < 0
	if neg {
		v = -v
	}
	var buf [20]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}