package agent

import (
	"sync"

	"github.com/filipgorny/agent/core"
)

// listeners routes emitted events to wait_for/listen_for registrations. Events
// emitted before a waiter registers are buffered in pending, so a wait_for that
// runs after an async skill already finished still receives the result.
type listeners struct {
	mu      sync.Mutex
	waiters []*waiter
	pending []core.Event
}

func newListeners() *listeners {
	return &listeners{}
}

func eventMatches(event, threadID string, ev core.Event) bool {
	if event != ev.Type {
		return false
	}

	return threadID == "" || threadID == ev.ThreadID
}

// register adds a waiter. If a buffered pending event already matches, it is
// delivered immediately (and, for one-shot waiters, no waiter is added).
// Returns the delivery channel and a cancel function.
func (l *listeners) register(event, threadID string, oneShot bool) (<-chan core.Event, func()) {
	l.mu.Lock()

	defer l.mu.Unlock()

	ch := make(chan core.Event, 8)

	if oneShot {
		for i, ev := range l.pending {
			if eventMatches(event, threadID, ev) {
				ch <- ev
				l.pending = append(l.pending[:i], l.pending[i+1:]...)

				return ch, func() {}
			}
		}
	} else {
		for _, ev := range l.pending {
			if eventMatches(event, threadID, ev) {
				ch <- ev
			}
		}
	}

	w := &waiter{event: event, threadID: threadID, ch: ch, oneShot: oneShot}
	l.waiters = append(l.waiters, w)

	cancel := func() {
		l.mu.Lock()

		defer l.mu.Unlock()

		for i, x := range l.waiters {
			if x == w {
				l.waiters = append(l.waiters[:i], l.waiters[i+1:]...)

				break
			}
		}
	}

	return ch, cancel
}

// emit delivers ev to matching waiters. If none match, it is buffered.
func (l *listeners) emit(ev core.Event) {
	l.mu.Lock()

	defer l.mu.Unlock()

	matched := false
	kept := l.waiters[:0]

	for _, w := range l.waiters {
		if eventMatches(w.event, w.threadID, ev) {
			select {

			case w.ch <- ev:

			default:
			}

			matched = true

			if w.oneShot {
				continue
			}
		}

		kept = append(kept, w)
	}

	l.waiters = kept

	if !matched {
		l.pending = append(l.pending, ev)
	}
}
