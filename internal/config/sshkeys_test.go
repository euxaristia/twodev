package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSSHAuthorizedKeysMissingFile(t *testing.T) {
	keys, err := LoadSSHAuthorizedKeys(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if keys != nil {
		t.Fatalf("expected nil keys, got %d", len(keys))
	}
}

func TestLoadSSHAuthorizedKeysFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "conf", "authorized_keys")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	line := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl test@example.com"
	if err := os.WriteFile(path, []byte(line+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	keys, err := LoadSSHAuthorizedKeys(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) != 1 {
		t.Fatalf("keys = %d", len(keys))
	}
}