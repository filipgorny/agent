package interaction

import "github.com/filipgorny/agent/core"

// InteractionPlugin provides skills that ask the user (only effective when the
// agent is interactive, i.e. Deps.Ask is wired).
type InteractionPlugin struct{}

func (InteractionPlugin) Name() string {
	return "interaction"
}

func (InteractionPlugin) Description() string {
	return "Ask the user questions (free text or a choice between options)."
}

func (InteractionPlugin) Skills() map[string]func(core.Deps) core.Skill {
	return map[string]func(core.Deps) core.Skill{
		AskUserSkillName:   func(d core.Deps) core.Skill { return AskUser{ask: d.Ask} },
		AskChoiceSkillName: func(d core.Deps) core.Skill { return AskChoice{ask: d.Ask} },
	}
}

func (InteractionPlugin) Events() []core.EventSpec {
	var events []core.EventSpec

	events = append(events, AskUser{}.GetEvents()...)
	events = append(events, AskChoice{}.GetEvents()...)

	return events
}
