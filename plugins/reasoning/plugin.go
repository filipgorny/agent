package reasoning

import "github.com/filipgorny/agent/core"

// ReasoningPlugin provides the asynchronous think skill.
type ReasoningPlugin struct{}

func (ReasoningPlugin) Name() string {
	return "reasoning"
}

func (ReasoningPlugin) Description() string {
	return "Asynchronous reasoning: run think steps in side threads while doing research."
}

func (ReasoningPlugin) Skills() map[string]func(core.Deps) core.Skill {
	return map[string]func(core.Deps) core.Skill{
		ThinkSkillName: func(d core.Deps) core.Skill { return Think{llm: d.LLM} },
	}
}

func (ReasoningPlugin) Events() []core.EventSpec {
	return Think{}.GetEvents()
}
