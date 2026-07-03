package event

import "sync"

// Bus is a simple in-process pub/sub bus replacing ListenerRegistry.
type Bus struct {
	mu          sync.RWMutex
	subscribers map[string][]func(any)
}

// NewBus creates an event bus.
func NewBus() *Bus {
	return &Bus{subscribers: make(map[string][]func(any))}
}

// Publish dispatches payload to subscribers of eventType.
func (b *Bus) Publish(eventType string, payload any) {
	b.mu.RLock()
	subs := append([]func(any){}, b.subscribers[eventType]...)
	b.mu.RUnlock()
	for _, sub := range subs {
		sub(payload)
	}
}

// Subscribe registers a handler for eventType.
func (b *Bus) Subscribe(eventType string, handler func(any)) {
	b.mu.Lock()
	b.subscribers[eventType] = append(b.subscribers[eventType], handler)
	b.mu.Unlock()
}