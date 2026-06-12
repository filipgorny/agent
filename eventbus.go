package agent

import "sync"

// Event is a message published on the EventBus by skills (e.g. file_watch) and
// consumed by the agent's reactive Listen loop.
type Event struct {
	Type   string
	Source string
	Data   map[string]any
}

// EventBus is a simple in-process pub/sub. Publish is non-blocking: subscribers
// with a full buffer miss the event rather than stalling the publisher.
type EventBus struct {
	mu     sync.RWMutex
	subs   []chan Event
	closed bool
}

// NewEventBus returns an empty bus.
func NewEventBus() *EventBus {
	return &EventBus{}
}

// Subscribe returns a channel of events and an unsubscribe function.
func (b *EventBus) Subscribe() (<-chan Event, func()) {
	b.mu.Lock()

	defer b.mu.Unlock()

	ch := make(chan Event, 16)
	b.subs = append(b.subs, ch)

	unsubscribe := func() {
		b.mu.Lock()

		defer b.mu.Unlock()

		for i, c := range b.subs {
			if c == ch {
				b.subs = append(b.subs[:i], b.subs[i+1:]...)
				close(ch)

				break
			}
		}
	}

	return ch, unsubscribe
}

// Publish delivers e to all current subscribers, best-effort and non-blocking.
func (b *EventBus) Publish(e Event) {
	b.mu.RLock()

	defer b.mu.RUnlock()

	if b.closed {
		return
	}

	for _, ch := range b.subs {
		select {

		case ch <- e:

		default:
		}
	}
}

// Close closes the bus and all subscriber channels.
func (b *EventBus) Close() {
	b.mu.Lock()

	defer b.mu.Unlock()

	if b.closed {
		return
	}

	b.closed = true

	for _, ch := range b.subs {
		close(ch)
	}

	b.subs = nil
}
