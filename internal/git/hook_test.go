package git

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallPostReceiveHook(t *testing.T) {
	root := t.TempDir()
	svc := NewService("")
	if err := svc.InitBareRepo(context.Background(), root); err != nil {
		t.Fatal(err)
	}
	if err := svc.InstallPostReceiveHook(root, "demo/project", 6610); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, "hooks", "post-receive"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "demo/project") {
		t.Fatal("expected project path in hook")
	}
	if !strings.Contains(content, "6610") {
		t.Fatal("expected port in hook")
	}
	if !strings.Contains(content, "branch-update") {
		t.Fatal("expected branch-update endpoint in hook")
	}
}