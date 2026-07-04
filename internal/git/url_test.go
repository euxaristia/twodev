package git

import "testing"

func TestProjectCloneURL(t *testing.T) {
	tests := []struct {
		name        string
		serverName  string
		httpHost    string
		httpPort    int
		projectPath string
		want        string
	}{
		{
			name:        "server name wins",
			serverName:  "dev.example.com",
			httpHost:    "0.0.0.0",
			httpPort:    6610,
			projectPath: "demo",
			want:        "http://dev.example.com:6610/demo.git",
		},
		{
			name:        "bind all maps to localhost",
			httpHost:    "0.0.0.0",
			httpPort:    8080,
			projectPath: "demo/agent",
			want:        "http://127.0.0.1:8080/demo/agent.git",
		},
		{
			name:        "explicit host",
			httpHost:    "192.168.1.10",
			httpPort:    6610,
			projectPath: "team/app",
			want:        "http://192.168.1.10:6610/team/app.git",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := ProjectCloneURL(tc.serverName, tc.httpHost, tc.httpPort, tc.projectPath); got != tc.want {
				t.Fatalf("ProjectCloneURL() = %q, want %q", got, tc.want)
			}
		})
	}
}