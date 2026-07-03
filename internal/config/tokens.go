package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/euxaristia/twodev/internal/auth"
)

func checkTokenFilePermissions(path string) error {
	if runtime.GOOS == "windows" {
		return nil
	}
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.Mode().Perm()&0o077 != 0 {
		return fmt.Errorf("%s must be owner-readable only (mode %o)", path, info.Mode().Perm())
	}
	return nil
}

// LoadAgentTokens reads bearer tokens for the /~server agent websocket.
// TWODEV_AGENT_TOKENS (comma-separated) takes precedence over site/conf/agent-tokens.txt.
func LoadAgentTokens(siteDir string) (auth.StaticTokens, error) {
	tokens := auth.StaticTokens{}
	if raw := os.Getenv("TWODEV_AGENT_TOKENS"); raw != "" {
		for part := range strings.SplitSeq(raw, ",") {
			token := strings.TrimSpace(part)
			if token != "" {
				tokens[token] = struct{}{}
			}
		}
		return tokens, nil
	}

	path := filepath.Join(siteDir, "conf", "agent-tokens.txt")
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return tokens, nil
	}
	if err != nil {
		return nil, err
	}
	if err := checkTokenFilePermissions(path); err != nil {
		_ = file.Close()
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		tokens[line] = struct{}{}
	}
	return tokens, scanner.Err()
}