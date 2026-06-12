package agent

import (
	"context"
	"fmt"
)

// Action names hard-coded into the agent.
const (
	ActionPrompt = "prompt"
	ActionSkill  = "skill"
)

// Action is a capability the LLM can trigger. Unlike skills, actions are
// hard-coded into the agent. Each action takes a free-form parameter map so
// different actions can accept different parameters.
type Action struct {
	Name        string
	Description string
	Run         func(ctx context.Context, a *Agent, params map[string]any) (string, error)
}

// ActionCall is the LLM's decision: which action to run and with what params.
type ActionCall struct {
	Action string         `json:"action"`
	Params map[string]any `json:"params"`
}

// builtinActions returns the hard-coded action set.
func builtinActions() map[string]Action {
	actions := []Action{
		{
			Name:        ActionPrompt,
			Description: `Ask the LLM directly. params: {"text": string}`,
			Run:         runPromptAction,
		},
		{
			Name:        ActionSkill,
			Description: `Run a configured skill. params: {"name": string, ...skill-specific params}`,
			Run:         runSkillAction,
		},
	}

	out := make(map[string]Action, len(actions))

	for _, act := range actions {
		out[act.Name] = act
	}

	return out
}

func runPromptAction(ctx context.Context, a *Agent, params map[string]any) (string, error) {
	text, ok := params["text"].(string)

	if !ok {
		return "", fmt.Errorf("action prompt: missing string \"text\" parameter")
	}

	return a.llm.Prompt(ctx, text)
}

func runSkillAction(ctx context.Context, a *Agent, params map[string]any) (string, error) {
	name, ok := params["name"].(string)

	if !ok {
		return "", fmt.Errorf("action skill: missing string \"name\" parameter")
	}

	skill, ok := a.skills[name]

	if !ok {
		return "", fmt.Errorf("action skill: skill %q is not enabled", name)
	}

	return skill.Run(ctx, params)
}
