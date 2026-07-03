package client

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/euxaristia/twodev/internal/agent/protocol"
	"github.com/euxaristia/twodev/internal/config"
	"github.com/euxaristia/twodev/internal/version"
	"github.com/gorilla/websocket"
)

// Client connects to a OneDev server using the legacy agent websocket protocol.
type Client struct {
	cfg    config.Agent
	dialer *websocket.Dialer
	logger *slog.Logger
}

// New creates an agent client from agent.properties.
func New(cfg config.Agent, logger *slog.Logger) *Client {
	if logger == nil {
		logger = slog.Default()
	}
	return &Client{
		cfg: cfg,
		dialer: &websocket.Dialer{
			HandshakeTimeout: 10 * time.Second,
		},
		logger: logger,
	}
}

// Run maintains a websocket connection to the configured server.
func (c *Client) Run(ctx context.Context) error {
	url, err := serverWebSocketURL(c.cfg.ServerURL)
	if err != nil {
		return err
	}

	header := http.Header{}
	header.Set("Authorization", "Bearer "+c.cfg.Token)

	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		conn, resp, err := c.dialer.DialContext(ctx, url, header)
		if err != nil {
			if resp != nil {
				c.logger.Error("websocket handshake failed", "status", resp.Status)
			}
			c.logger.Error("connect failed, retrying", "error", err)
			if wait(ctx, 5*time.Second) != nil {
				return ctx.Err()
			}
			continue
		}

		c.logger.Info("connected to server", "url", c.cfg.ServerURL)
		sessionErr := c.serveSession(ctx, conn)
		_ = conn.Close()
		if sessionErr != nil && ctx.Err() != nil {
			return ctx.Err()
		}
		if sessionErr != nil {
			c.logger.Error("session ended", "error", sessionErr)
		}
		if wait(ctx, 5*time.Second) != nil {
			return ctx.Err()
		}
	}
}

func (c *Client) serveSession(ctx context.Context, conn *websocket.Conn) error {
	// OneDev agents send AGENT_DATA after receiving UPDATE with matching version.
	// Full registration still requires Java-serialized AgentData for the legacy server.
	registered := false

	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		_ = conn.SetReadDeadline(time.Now().Add(protocol.SocketIdleTimeoutMs * time.Millisecond))

		_, frame, err := conn.ReadMessage()
		if err != nil {
			return err
		}

		msg, err := protocol.Decode(frame)
		if err != nil {
			return err
		}

		switch msg.Type {
		case protocol.TypeUpdate:
			serverVersion := string(msg.Data)
			c.logger.Info("server agent version", "version", serverVersion)
			if serverVersion == version.Version {
				if !registered {
					c.logger.Warn(
						"skipping AGENT_DATA registration until Java serialization bridge is implemented",
						"twodev_version", version.Version,
					)
					registered = true
				}
			} else {
				c.logger.Warn(
					"server expects different agent version; Go agent auto-update is not implemented yet",
					"server_version", serverVersion,
					"agent_version", version.Version,
				)
			}
		case protocol.TypeError:
			return fmt.Errorf("server error: %s", string(msg.Data))
		case protocol.TypeRestart:
			c.logger.Info("restart requested by server")
			return fmt.Errorf("restart requested")
		case protocol.TypeStop:
			c.logger.Info("stop requested by server")
			return fmt.Errorf("stop requested")
		default:
			c.logger.Debug("ignored message", "type", msg.Type.String())
		}
	}
}

func serverWebSocketURL(serverURL string) (string, error) {
	switch {
	case strings.HasPrefix(serverURL, "https://"):
		return "wss://" + strings.TrimPrefix(serverURL, "https://") + protocol.ServerPath, nil
	case strings.HasPrefix(serverURL, "http://"):
		return "ws://" + strings.TrimPrefix(serverURL, "http://") + protocol.ServerPath, nil
	default:
		return "", fmt.Errorf("serverUrl must start with http:// or https://")
	}
}

func wait(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}