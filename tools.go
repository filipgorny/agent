package agent

import (
	"github.com/filipgorny/agent/core"
	llm "github.com/filipgorny/llm-provider"
)

// buildToolSpecs maps the agent's actions and enabled skills to native tool
// specs. Skills are exposed as individual tools (the generic "skill" action is
// omitted); Execute's skill-name-as-action tolerance dispatches them.
func (a *Agent) buildToolSpecs() []llm.ToolSpec {
	var tools []llm.ToolSpec

	for _, name := range sortedActionKeys(a.actions) {
		if name == ActionSkill {
			continue
		}

		tools = append(tools, llm.ToolSpec{
			Name:        name,
			Description: a.actions[name].Description,
			Parameters:  objectSchema(),
		})
	}

	for _, name := range sortedSkillKeys(a.skills) {
		tools = append(tools, llm.ToolSpec{
			Name:        name,
			Description: skillToolDescription(a.skills[name]),
			Parameters:  objectSchema(),
		})
	}

	return tools
}

// objectSchema is a permissive JSON Schema for free-form tool arguments.
func objectSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"properties":           map[string]any{},
		"additionalProperties": true,
	}
}

func skillToolDescription(s core.Skill) string {
	d := s.Description()

	if s.IsAsync() {
		d += " (async: emits a result event; pair with wait_for or listen_for)"
	}

	return d
}
