package agent

// execContext carries per-invocation data available while an action runs: the
// thread it runs in and the uid of the triggering action (propagated to events).
type execContext struct {
	threadID  string
	actionUID string
}
