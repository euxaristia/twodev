package api

import (
	"fmt"
	"io"
	"net/http"
)

func readYAMLBody(r *http.Request) (string, error) {
	defer r.Body.Close()
	body, err := io.ReadAll(io.LimitReader(r.Body, 8<<20))
	if err != nil {
		return "", err
	}
	if len(body) == 0 {
		return "", fmt.Errorf("request body is required")
	}
	return string(body), nil
}