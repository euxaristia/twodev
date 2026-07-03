package git

import (
	"fmt"
	"os"
	"path/filepath"
)

// InstallPostReceiveHook writes a hook that notifies twodev of branch pushes.
func (s *Service) InstallPostReceiveHook(repoDir, projectPath string, httpPort int) error {
	hookDir := filepath.Join(repoDir, "hooks")
	if err := os.MkdirAll(hookDir, 0o755); err != nil {
		return err
	}
	script := fmt.Sprintf(`#!/bin/sh
# twodev post-receive hook
while read oldrev newrev refname; do
  case "$refname" in
    refs/heads/*) branch="${refname#refs/heads/}" ;;
    *) continue ;;
  esac
  if [ "$newrev" = "0000000000000000000000000000000000000000" ]; then
    continue
  fi
  curl -sf -X POST -H "Content-Type: application/json" \
    -d "{\"projectPath\":\"%s\",\"branch\":\"$branch\",\"commitHash\":\"$newrev\"}" \
    "http://127.0.0.1:%d/~api/twodev/git/branch-update" >/dev/null 2>&1 || true
done
`, projectPath, httpPort)
	hookPath := filepath.Join(hookDir, "post-receive")
	if err := os.WriteFile(hookPath, []byte(script), 0o755); err != nil {
		return err
	}
	return nil
}