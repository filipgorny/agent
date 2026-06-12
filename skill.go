package agent

import (
	"context"
	"fmt"
)

// Skill is a capability the agent can use. Concrete skills are registered by
// name and selected via configuration, not hard-coded into the agent.
type Skill interface {
	// Name is the identifier used in config and when an action invokes the skill.
	Name() string

	// Run executes the skill with the given parameters.
	Run(ctx context.Context, params map[string]any) (string, error)
}

// skillRegistry maps skill names to factories. Concrete skills register
// themselves here (e.g. in an init function), so new skills can be added
// without touching the agent core.
var skillRegistry = map[string]func() Skill{}

// RegisterSkill registers a skill factory under name. It is meant to be called
// from init functions of concrete skills.
func RegisterSkill(name string, factory func() Skill) {
	skillRegistry[name] = factory
}

// buildSkills instantiates the named skills from the registry.
func buildSkills(names []string) (map[string]Skill, error) {
	skills := make(map[string]Skill, len(names))

	for _, name := range names {
		factory, ok := skillRegistry[name]

		if !ok {
			return nil, fmt.Errorf("agent: unknown skill %q (not registered)", name)
		}

		skills[name] = factory()
	}

	return skills, nil
}
