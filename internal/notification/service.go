package notification

import (
	"context"
	"log/slog"
)

// Event describes a notification-worthy event.
type Event struct {
	Type    string
	Project string
	Subject string
	Body    string
}

// Sender delivers notifications.
type Sender interface {
	Send(ctx context.Context, event Event) error
}

// LogSender logs notifications until real channels are ported.
type LogSender struct {
	logger *slog.Logger
}

// NewLogSender creates a logging notification sender.
func NewLogSender(logger *slog.Logger) *LogSender {
	if logger == nil {
		logger = slog.Default()
	}
	return &LogSender{logger: logger}
}

// Send logs the notification event.
func (s *LogSender) Send(ctx context.Context, event Event) error {
	_ = ctx
	s.logger.Info("notification", "type", event.Type, "project", event.Project, "subject", event.Subject)
	return nil
}