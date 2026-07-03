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
	if keys.Enforce {
		t.Fatal("expected enforce=false for missing file")
	}
}

func TestLoadSSHAuthorizedKeysFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "conf", "authorized_keys")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	line := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl test@example.com"
	if err := os.WriteFile(path, []byte(line+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	keys, err := LoadSSHAuthorizedKeys(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !keys.Enforce || len(keys.Keys) != 1 {
		t.Fatalf("keys = %+v", keys)
	}
}

func TestLoadSSHAuthorizedKeysEmptyFileDeniesAll(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "conf", "authorized_keys")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("# no keys\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	keys, err := LoadSSHAuthorizedKeys(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !keys.Enforce || len(keys.Keys) != 0 {
		t.Fatalf("expected enforced empty keys, got %+v", keys)
	}
}