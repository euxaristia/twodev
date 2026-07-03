package auth

import (
	"net"
	"net/http"
	"strings"
)

// Guard optionally enforces bearer or basic auth using access tokens.
type Guard struct {
	tokens StaticTokens
}

// NewGuard creates an auth guard. An empty token set disables enforcement.
func NewGuard(tokens StaticTokens) *Guard {
	return &Guard{tokens: tokens}
}

// Enabled reports whether auth is required.
func (g *Guard) Enabled() bool {
	return g != nil && len(g.tokens) > 0
}

// ValidRequest reports whether the request presents an allowed token.
func (g *Guard) ValidRequest(r *http.Request) bool {
	if !g.Enabled() {
		return true
	}
	if token, err := ParseBearer(r.Header.Get("Authorization")); err == nil && g.tokens.Valid(token) {
		return true
	}
	if _, password, ok := r.BasicAuth(); ok && g.tokens.Valid(password) {
		return true
	}
	return false
}

// Middleware wraps a handler with access token auth.
func (g *Guard) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if g.ValidRequest(r) {
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Set("WWW-Authenticate", `Basic realm="twodev"`)
		http.Error(w, "forbidden", http.StatusForbidden)
	})
}

// IsLoopback reports whether the request originated from localhost.
func IsLoopback(r *http.Request) bool {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	host = strings.TrimPrefix(host, "[")
	return host == "127.0.0.1" || host == "::1"
}