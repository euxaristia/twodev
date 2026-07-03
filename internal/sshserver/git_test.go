package sshserver

import "testing"

func TestParseGitCommand(t *testing.T) {
	action, project, ok := parseGitCommand("git-upload-pack '/demo/project.git'")
	if !ok || action != "upload-pack" || project != "/demo/project.git" {
		t.Fatalf("got %q %q %v", action, project, ok)
	}
	action, project, ok = parseGitCommand("git-receive-pack 'demo.git'")
	if !ok || action != "receive-pack" || project != "demo.git" {
		t.Fatalf("got %q %q %v", action, project, ok)
	}
	if _, _, ok := parseGitCommand("ls"); ok {
		t.Fatal("expected unsupported command")
	}
}

func TestResolveRepoDir(t *testing.T) {
	root := t.TempDir()
	if _, err := resolveRepoDir(root, "/demo.git"); err == nil {
		t.Fatal("expected missing repo error")
	}
}