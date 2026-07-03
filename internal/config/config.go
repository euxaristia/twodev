package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	DefaultHTTPHost     = "0.0.0.0"
	DefaultHTTPPort     = 6610
	DefaultSSHPort      = 6611
	DefaultClusterPort  = 5710
	DefaultAgentVersion = "0.1.0"
)

// Server holds OneDev-compatible server.properties settings.
type Server struct {
	HTTPHost    string
	HTTPPort    int
	SSHPort     *int
	ServerName  string
	ClusterIP   string
	ClusterPort int
}

// Agent holds OneDev-compatible agent.properties settings.
type Agent struct {
	ServerURL  string
	Token      string
	Name       string
	GitPath    string
	DockerPath string
}

// LoadServer reads server.properties from path.
func LoadServer(path string) (Server, error) {
	values, err := loadProperties(path)
	if err != nil {
		return Server{}, err
	}

	httpPort, err := getInt(values, "http_port", DefaultHTTPPort)
	if err != nil {
		return Server{}, fmt.Errorf("parse http_port: %w", err)
	}
	clusterPort, err := getInt(values, "cluster_port", DefaultClusterPort)
	if err != nil {
		return Server{}, fmt.Errorf("parse cluster_port: %w", err)
	}

	cfg := Server{
		HTTPHost:    getString(values, "http_host", DefaultHTTPHost),
		HTTPPort:    httpPort,
		ServerName:  getString(values, "server_name", ""),
		ClusterIP:   getString(values, "cluster_ip", ""),
		ClusterPort: clusterPort,
	}
	if raw, ok := values["ssh_port"]; ok && strings.TrimSpace(raw) != "" {
		port, err := strconv.Atoi(strings.TrimSpace(raw))
		if err != nil {
			return Server{}, fmt.Errorf("parse ssh_port: %w", err)
		}
		cfg.SSHPort = &port
	}
	return cfg, nil
}

// LoadAgent reads agent.properties from path.
func LoadAgent(path string) (Agent, error) {
	values, err := loadProperties(path)
	if err != nil {
		return Agent{}, err
	}

	cfg := Agent{
		ServerURL:  strings.TrimRight(getString(values, "serverUrl", ""), "/"),
		Token:      getString(values, "agentToken", ""),
		Name:       getString(values, "agentName", ""),
		GitPath:    getString(values, "gitPath", "git"),
		DockerPath: getString(values, "dockerPath", "docker"),
	}
	if cfg.ServerURL == "" {
		return Agent{}, fmt.Errorf("serverUrl is required")
	}
	if cfg.Token == "" {
		return Agent{}, fmt.Errorf("agentToken is required")
	}
	return cfg, nil
}

func loadProperties(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	values := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		values[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return values, nil
}

func getString(values map[string]string, key, fallback string) string {
	if value, ok := values[key]; ok {
		return value
	}
	return fallback
}

func getInt(values map[string]string, key string, fallback int) (int, error) {
	raw, ok := values[key]
	if !ok || strings.TrimSpace(raw) == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0, err
	}
	return value, nil
}