package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/euxaristia/twodev/internal/auth"
)

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