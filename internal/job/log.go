package job

import (
	"fmt"
	"io"
	"sync"
)

// Logger streams job output to subscribers and optional sinks.
type Logger struct {
	token string
	mu    sync.Mutex
	subs  []chan string
	sink  io.Writer
}

// NewLogger creates a job logger identified by token.
func NewLogger(token string, sink io.Writer) *Logger {
	return &Logger{token: token, sink: sink}
}

// Token returns the logger identifier used in agent LOG messages.
func (l *Logger) Token() string {
	return l.token
}

// Log writes a line to subscribers and sink.
func (l *Logger) Log(line string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.sink != nil {
		_, _ = fmt.Fprintln(l.sink, line)
	}
	for _, sub := range l.subs {
		select {
		case sub <- line:
		default:
		}
	}
}

// Subscribe returns a channel receiving log lines.
func (l *Logger) Subscribe() <-chan string {
	ch := make(chan string, 128)
	l.mu.Lock()
	l.subs = append(l.subs, ch)
	l.mu.Unlock()
	return ch
}