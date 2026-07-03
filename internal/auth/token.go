package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
)

// BearerPrefix is the authorization header prefix used by OneDev agents.
const BearerPrefix = "Bearer "

// GenerateToken creates a random bearer token.
func GenerateToken(n int) (string, error) {
	if n <= 0 {
		n = 32
	}
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

// ParseBearer extracts the bearer token from an Authorization header.
func ParseBearer(header string) (string, error) {
	if header == "" {
		return "", fmt.Errorf("missing authorization header")
	}
	if !strings.HasPrefix(header, BearerPrefix) {
		return "", fmt.Errorf("expected bearer token")
	}
	token := strings.TrimSpace(strings.TrimPrefix(header, BearerPrefix))
	if token == "" {
		return "", fmt.Errorf("empty bearer token")
	}
	return token, nil
}