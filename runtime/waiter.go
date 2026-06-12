package runtime

import "github.com/filipgorny/agent/core"

// waiter is a registered interest in events matching (event, threadID).
// threadID == "" matches any thread. One-shot waiters (wait_for) are removed
// after the first delivery; persistent waiters (listen_for) stay.
type waiter struct {
	event    string
	threadID string
	ch       chan core.Event
	oneShot  bool
}
