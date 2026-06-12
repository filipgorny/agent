package core

import "context"

// Skill is a capability the agent can use. Skills belong to plugins. The
// metadata methods (Description/IsAsync/GetEvents) are read from the struct and
// rendered into the agent's preamble so the LLM knows how to use the skill.
type Skill interface {
	// Name is the identifier used in config and when an action invokes the skill.
	Name() string

	// Description explains, in English, what the skill does.
	Description() string

	// IsAsync reports whether running the skill spawns a side thread. The agent
	// returns immediately and the result arrives later as an event.
	IsAsync() bool

	// GetEvents lists the events the skill emits. The first entry is treated as
	// the skill's completion/result event.
	GetEvents() []EventSpec

	// Run executes the skill with the given parameters.
	Run(ctx context.Context, params map[string]any) (string, error)
}

// ResultEvent returns the skill's completion event name, or "<name>.result".
func ResultEvent(s Skill) string {
	events := s.GetEvents()

	if len(events) > 0 {
		return events[0].Name
	}

	return s.Name() + ".result"
}
