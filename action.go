package agent

import (
	"context"
	"fmt"
	"strings"
)

// Action names hard-coded into the agent.
const (
	ActionPrompt   = "prompt"
	ActionSkill    = "skill"
	ActionRemember = "remember"
	ActionRead     = "read"
)

// defaultReadTopK is used when the read action omits "top_k".
const defaultReadTopK = 5

// Action is a capability the LLM can trigger. Unlike skills, actions are
// hard-coded into the agent. Each action takes a free-form parameter map so
// different actions can accept different parameters.
type Action struct {
	Name        string
	Description string
	Run         func(ctx context.Context, a *Agent, params map[string]any) (string, error)
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
		{
			Name:        ActionRemember,
			Description: `Store a memory. params: {"content": string, "meta": object?}`,
			Run:         runRememberAction,
		},
		{
			Name:        ActionRead,
			Description: `Search memory. params: {"query": string, "top_k": int?}`,
			Run:         runReadAction,
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

func runRememberAction(ctx context.Context, a *Agent, params map[string]any) (string, error) {
	if a.memory == nil {
		return "", fmt.Errorf("action remember: no memory configured")
	}

	content, ok := params["content"].(string)

	if !ok {
		return "", fmt.Errorf("action remember: missing string \"content\" parameter")
	}

	meta, _ := params["meta"].(map[string]any)

	if err := a.memory.Remember(ctx, content, meta); err != nil {
		return "", err
	}

	return "remembered", nil
}

func runReadAction(ctx context.Context, a *Agent, params map[string]any) (string, error) {
	if a.memory == nil {
		return "", fmt.Errorf("action read: no memory configured")
	}

	query, ok := params["query"].(string)

	if !ok {
		return "", fmt.Errorf("action read: missing string \"query\" parameter")
	}

	topK := defaultReadTopK

	if v, ok := params["top_k"].(float64); ok {
		topK = int(v)
	}

	records, err := a.memory.Read(ctx, query, topK)

	if err != nil {
		return "", err
	}

	if len(records) == 0 {
		return "(no matching memories)", nil
	}

	var b strings.Builder

	for i, r := range records {
		fmt.Fprintf(&b, "%d. %s\n", i+1, r.Content)
	}

	return strings.TrimRight(b.String(), "\n"), nil
}
