package sshserver

import (
	"context"
	"fmt"
	"net"

	"golang.org/x/crypto/ssh"
)

// Config configures the embedded SSH server used for git over SSH.
type Config struct {
	Host string
	Port int
}

// Server is a minimal SSH server skeleton for git commands.
type Server struct {
	cfg     Config
	config  *ssh.ServerConfig
}

// New creates an SSH server skeleton.
func New(cfg Config) *Server {
	serverCfg := &ssh.ServerConfig{
		PasswordCallback: func(_ ssh.ConnMetadata, _ []byte) (*ssh.Permissions, error) {
			return nil, fmt.Errorf("password auth not enabled")
		},
	}
	return &Server{cfg: cfg, config: serverCfg}
}

// ListenAndServe accepts SSH connections until ctx is canceled.
func (s *Server) ListenAndServe(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	go func() {
		<-ctx.Done()
		_ = listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return err
		}
		go func() {
			_, _, _, _ = ssh.NewServerConn(conn, s.config)
		}()
	}
}