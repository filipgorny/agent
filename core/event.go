package core

// Event is published on the EventBus by skills (e.g. file_watch) and by the
// agent after a skill completes. It is consumed by wait_for/listen_for and by
// external subscribers.
type Event struct {
	Type      string
	Source    string
	ActionUID string // uid of the action that triggered the emitting skill
	ThreadID  string // thread the skill ran in
	Data      map[string]any
}
