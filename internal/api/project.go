package api

import (
	"fmt"
	"net/http"
	"strconv"
)

func projectIDFromRequest(r *http.Request) (int64, error) {
	raw := r.PathValue("id")
	if raw == "" {
		return 0, fmt.Errorf("project id is required")
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid project id")
	}
	return id, nil
}

func issueNumberFromRequest(r *http.Request) (int, error) {
	raw := r.PathValue("number")
	number, err := strconv.Atoi(raw)
	if err != nil || number <= 0 {
		return 0, fmt.Errorf("invalid issue number")
	}
	return number, nil
}

func pullNumberFromRequest(r *http.Request) (int, error) {
	return issueNumberFromRequest(r)
}

func buildNumberFromRequest(r *http.Request) (int, error) {
	raw := r.PathValue("number")
	number, err := strconv.Atoi(raw)
	if err != nil || number <= 0 {
		return 0, fmt.Errorf("invalid build number")
	}
	return number, nil
}

func jobNameFromRequest(r *http.Request) (string, error) {
	job := r.PathValue("job")
	if job == "" {
		return "", fmt.Errorf("job name is required")
	}
	return job, nil
}