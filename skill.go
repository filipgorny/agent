package agent

import (
	"context"
	"fmt"

	llm "github.com/filipgorny/llm-provider"
)

// Skill is a capability the agent can use. Concrete skills are registered by
// name and selected via configuration, not hard-coded into the agent.
type Skill interface {
	// Name is the identifier used in config and when an action invokes the skill.
	Name() string

	// Run executes the skill with the given parameters.
	Run(ctx context.Context, params map[string]any) (string, error)
}

// Deps are the agent services a skill may need at construction: the LLM
// provider (e.g. summarize_text, translate) and the event bus (e.g. file_watch).
type Deps struct {
	LLM *llm.LlmProvider
	Bus *EventBus
}

// skillRegistry maps skill names to factories. Concrete skills register
// themselves here (e.g. in an init function), so new skills can be added
// without touching the agent core.
var skillRegistry = map[string]func(Deps) Skill{}

// RegisterSkill registers a skill factory under name. It is meant to be called
// from init functions of concrete skills.
func RegisterSkill(name string, factory func(Deps) Skill) {
	skillRegistry[name] = factory
}

// buildSkills instantiates the named skills from the registry, injecting deps.
func buildSkills(names []string, deps Deps) (map[string]Skill, error) {
	skills := make(map[string]Skill, len(names))

	for _, name := range names {
		factory, ok := skillRegistry[name]

		if !ok {
			return nil, fmt.Errorf("agent: unknown skill %q (not registered)", name)
		}

		skills[name] = factory(deps)
	}

	return skills, nil
}
