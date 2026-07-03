package sshserver

import (
	"context"
	"fmt"
	"io"
	"net"
	"path/filepath"

	"github.com/euxaristia/twodev/internal/git"
	"golang.org/x/crypto/ssh"
)

// Config configures the embedded SSH server used for git over SSH.
type Config struct {
	Host        string
	Port        int
	RepoRoot    string
	HostKeyPath string
}

// Server serves git commands over SSH.
type Server struct {
	cfg    Config
	config *ssh.ServerConfig
	git    *git.Service
}

// New creates an SSH git server.
func New(cfg Config) (*Server, error) {
	hostKeyPath := cfg.HostKeyPath
	if hostKeyPath == "" && cfg.RepoRoot != "" {
		hostKeyPath = filepath.Join(filepath.Dir(cfg.RepoRoot), "conf", "ssh_host_key")
	}
	signer, err := loadOrCreateHostKey(hostKeyPath)
	if err != nil {
		return nil, fmt.Errorf("load host key: %w", err)
	}
	serverCfg := &ssh.ServerConfig{
		PublicKeyCallback: func(_ ssh.ConnMetadata, _ ssh.PublicKey) (*ssh.Permissions, error) {
			return &ssh.Permissions{}, nil
		},
	}
	serverCfg.AddHostKey(signer)
	return &Server{cfg: cfg, config: serverCfg, git: git.NewService("")}, nil
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
		go s.handleConn(ctx, conn)
	}
}

func (s *Server) handleConn(ctx context.Context, raw net.Conn) {
	sshConn, chans, reqs, err := ssh.NewServerConn(raw, s.config)
	if err != nil {
		return
	}
	defer sshConn.Close()
	go ssh.DiscardRequests(reqs)
	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			_ = newChannel.Reject(ssh.UnknownChannelType, "unsupported channel type")
			continue
		}
		channel, requests, err := newChannel.Accept()
		if err != nil {
			continue
		}
		go s.handleSession(ctx, channel, requests)
	}
}

func (s *Server) handleSession(ctx context.Context, channel ssh.Channel, requests <-chan *ssh.Request) {
	defer channel.Close()
	for req := range requests {
		switch req.Type {
		case "exec":
			err := handleGitSession(ctx, s.git, s.cfg.RepoRoot, req.Payload, channel, channel)
			if err != nil {
				_, _ = io.WriteString(channel, fmt.Sprintf("ERR %v\n", err))
			}
			_ = req.Reply(err == nil, nil)
			return
		default:
			_ = req.Reply(false, nil)
		}
	}
}