package auth

import (
	"net/http"
)

// TokenValidator validates bearer tokens.
type TokenValidator interface {
	Valid(token string) bool
}

// StaticTokens validates against a fixed token set.
type StaticTokens map[string]struct{}

// Valid reports whether token is allowed.
func (s StaticTokens) Valid(token string) bool {
	_, ok := s[token]
	return ok
}

// RequireBearer wraps handlers with bearer token auth.
func RequireBearer(validator TokenValidator, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := ParseBearer(r.Header.Get("Authorization"))
		if err != nil || !validator.Valid(token) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}