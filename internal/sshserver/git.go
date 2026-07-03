package sshserver

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/euxaristia/twodev/internal/git"
)

func parseGitCommand(command string) (action, project string, ok bool) {
	command = strings.TrimSpace(command)
	switch {
	case strings.HasPrefix(command, "git-upload-pack "):
		return "upload-pack", unquoteGitArg(command[len("git-upload-pack "):]), true
	case strings.HasPrefix(command, "git-receive-pack "):
		return "receive-pack", unquoteGitArg(command[len("git-receive-pack "):]), true
	default:
		return "", "", false
	}
}

func unquoteGitArg(arg string) string {
	arg = strings.TrimSpace(arg)
	if len(arg) >= 2 && arg[0] == '\'' && arg[len(arg)-1] == '\'' {
		return strings.TrimSuffix(strings.TrimPrefix(arg, "'"), "'")
	}
	return strings.Trim(arg, "'\"")
}

func resolveRepoDir(repoRoot, project string) (string, error) {
	project = strings.TrimPrefix(strings.TrimSpace(project), "/")
	project = strings.TrimSuffix(project, ".git")
	if project == "" {
		return "", fmt.Errorf("empty project path")
	}
	dir := filepath.Join(repoRoot, project+".git")
	if _, err := os.Stat(dir); err != nil {
		return "", err
	}
	return dir, nil
}

func runGitCommand(ctx context.Context, gitSvc *git.Service, repoDir, command string, in io.Reader, out io.Writer) error {
	cmd := exec.CommandContext(ctx, gitSvc.GitPath(), "-C", repoDir, command)
	cmd.Stdin = in
	cmd.Stdout = out
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if msg := strings.TrimSpace(stderr.String()); msg != "" {
			return fmt.Errorf("%s: %s", command, msg)
		}
		return err
	}
	return nil
}

func handleGitSession(ctx context.Context, gitSvc *git.Service, repoRoot string, payload []byte, in io.Reader, out io.Writer) error {
	command := string(payload)
	if len(payload) >= 4 {
		command = string(payload[4:])
	}
	action, project, ok := parseGitCommand(command)
	if !ok {
		return fmt.Errorf("unsupported command %q", command)
	}
	dir, err := resolveRepoDir(repoRoot, project)
	if err != nil {
		return err
	}
	return runGitCommand(ctx, gitSvc, dir, action, in, out)
}