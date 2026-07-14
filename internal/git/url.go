package git

import (
	"net"
	"net/url"
	"strconv"
	"strings"
)

// ProjectCloneURL builds an HTTP smart-git clone URL for a project path.
// It returns an empty string when no reachable host is configured (for
// example a bind-all address such as 0.0.0.0), so callers do not hand remote
// agents an unreachable URL.
func ProjectCloneURL(serverName, httpHost string, httpPort int, projectPath string) string {
	host := cloneHost(serverName, httpHost)
	if host == "" {
		return ""
	}
	projectPath = strings.Trim(strings.TrimSpace(projectPath), "/")
	u := url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(host, strconv.Itoa(httpPort)),
		Path:   "/" + projectPath + ".git",
	}
	return u.String()
}

// cloneHost resolves the host to advertise in clone URLs. server_name wins
// when set; otherwise the configured HTTP host is used. Binding addresses
// (0.0.0.0, ::) and empty values yield an empty string because they are not
// reachable from remote agents.
func cloneHost(serverName, httpHost string) string {
	if host := strings.TrimSpace(serverName); host != "" {
		return host
	}
	host := strings.TrimSpace(httpHost)
	if host == "" || host == "0.0.0.0" || host == "::" {
		return ""
	}
	return host
}
