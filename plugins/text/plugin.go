package text

import "github.com/filipgorny/agent/core"

// TextPlugin provides LLM-backed text skills.
type TextPlugin struct{}

func (TextPlugin) Name() string {
	return "text"
}

func (TextPlugin) Description() string {
	return "LLM-backed text operations: summarize and translate."
}

func (TextPlugin) Skills() map[string]func(core.Deps) core.Skill {
	return map[string]func(core.Deps) core.Skill{
		SummarizeTextSkillName: func(d core.Deps) core.Skill { return SummarizeText{llm: d.LLM} },
		TranslateSkillName:     func(d core.Deps) core.Skill { return Translate{llm: d.LLM} },
	}
}

func (TextPlugin) Events() []core.EventSpec {
	var events []core.EventSpec

	events = append(events, SummarizeText{}.GetEvents()...)
	events = append(events, Translate{}.GetEvents()...)

	return events
}
