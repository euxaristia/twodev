package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAgentTokensFromEnv(t *testing.T) {
	t.Setenv("TWODEV_AGENT_TOKENS", "alpha, beta ,gamma")
	tokens, err := LoadAgentTokens("site")
	if err != nil {
		t.Fatal(err)
	}
	for _, token := range []string{"alpha", "beta", "gamma"} {
		if !tokens.Valid(token) {
			t.Fatalf("expected token %q", token)
		}
	}
}

func TestLoadAgentTokensFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "conf", "agent-tokens.txt")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("# comment\ntoken-one\n\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TWODEV_AGENT_TOKENS", "")

	tokens, err := LoadAgentTokens(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !tokens.Valid("token-one") {
		t.Fatal("expected token-one from file")
	}
}

func TestLoadAccessTokensFromEnv(t *testing.T) {
	t.Setenv("TWODEV_ACCESS_TOKENS", "api-one,api-two")
	tokens, err := LoadAccessTokens("site")
	if err != nil {
		t.Fatal(err)
	}
	for _, token := range []string{"api-one", "api-two"} {
		if !tokens.Valid(token) {
			t.Fatalf("expected token %q", token)
		}
	}
}