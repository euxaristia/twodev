package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadServer(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "server.properties")
	content := `# comment
http_host=127.0.0.1
http_port=8080
ssh_port=8022
cluster_port=6000
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadServer(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.HTTPHost != "127.0.0.1" {
		t.Fatalf("http_host = %q, want 127.0.0.1", cfg.HTTPHost)
	}
	if cfg.HTTPPort != 8080 {
		t.Fatalf("http_port = %d, want 8080", cfg.HTTPPort)
	}
	if cfg.SSHPort == nil || *cfg.SSHPort != 8022 {
		t.Fatalf("ssh_port = %v, want 8022", cfg.SSHPort)
	}
	if cfg.ClusterPort != 6000 {
		t.Fatalf("cluster_port = %d, want 6000", cfg.ClusterPort)
	}
}

func TestLoadServerRejectsMalformedPort(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "server.properties")
	if err := os.WriteFile(path, []byte("http_port=not-a-number\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadServer(path); err == nil {
		t.Fatal("expected error for malformed http_port")
	}
}

func TestLoadAgentRequiresFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "agent.properties")
	if err := os.WriteFile(path, []byte("agentToken=abc\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadAgent(path); err == nil {
		t.Fatal("expected error when serverUrl is missing")
	}
}