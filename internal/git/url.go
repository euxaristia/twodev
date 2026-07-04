package git

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

// ProjectCloneURL builds an HTTP smart-git clone URL for a project path.
func ProjectCloneURL(serverName, httpHost string, httpPort int, projectPath string) string {
	host := cloneHost(serverName, httpHost)
	projectPath = strings.Trim(strings.TrimSpace(projectPath), "/")
	base := fmt.Sprintf("http://%s", net.JoinHostPort(host, strconv.Itoa(httpPort)))
	if projectPath == "" {
		return base + "/.git"
	}
	return base + "/" + projectPath + ".git"
}

func cloneHost(serverName, httpHost string) string {
	if host := strings.TrimSpace(serverName); host != "" {
		return host
	}
	host := strings.TrimSpace(httpHost)
	if host == "" || host == "0.0.0.0" {
		return "127.0.0.1"
	}
	return host
}