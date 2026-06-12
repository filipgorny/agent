package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	llm "github.com/filipgorny/llm-provider"
)

// Agent ties together an LLM, a set of configured skills and a hard-coded set
// of actions. The LLM decides which action to run; actions may, in turn, invoke
// a skill or prompt the LLM again.
type Agent struct {
	llm           *llm.LlmProvider
	skills        map[string]Skill
	actions       map[string]Action
	initialPrompt string
}

// NewAgent builds an agent from an LLM provider, the enabled skills and an
// initial prompt. The action set is hard-coded.
func NewAgent(provider *llm.LlmProvider, skills map[string]Skill, initialPrompt string) *Agent {
	return &Agent{
		llm:           provider,
		skills:        skills,
		actions:       builtinActions(),
		initialPrompt: initialPrompt,
	}
}

// InitialPrompt returns the agent's initial prompt.
func (a *Agent) InitialPrompt() string {
	return a.initialPrompt
}

// Run starts the agent from its initial prompt: it asks the LLM to choose an
// action and executes it.
func (a *Agent) Run(ctx context.Context) (string, error) {
	return a.Handle(ctx, a.initialPrompt)
}

// Handle asks the LLM which action to take for input, then executes it.
func (a *Agent) Handle(ctx context.Context, input string) (string, error) {
	call, err := a.Decide(ctx, input)

	if err != nil {
		return "", err
	}

	return a.Execute(ctx, call)
}

// Decide asks the LLM to pick the next action for the given input.
func (a *Agent) Decide(ctx context.Context, input string) (ActionCall, error) {
	out, err := a.llm.Prompt(ctx, a.decisionPrompt(input))

	if err != nil {
		return ActionCall{}, fmt.Errorf("agent: decide: %w", err)
	}

	var call ActionCall

	if err := json.Unmarshal([]byte(extractJSON(out)), &call); err != nil {
		return ActionCall{}, fmt.Errorf("agent: parse action from %q: %w", out, err)
	}

	return call, nil
}

// Execute runs the action named in the call.
func (a *Agent) Execute(ctx context.Context, call ActionCall) (string, error) {
	action, ok := a.actions[call.Action]

	if !ok {
		return "", fmt.Errorf("agent: unknown action %q", call.Action)
	}

	return action.Run(ctx, a, call.Params)
}

// decisionPrompt renders the prompt that asks the LLM to choose an action.
func (a *Agent) decisionPrompt(input string) string {
	var b strings.Builder

	b.WriteString("You are an autonomous agent. Choose the single next action.\n\n")
	b.WriteString("Available actions:\n")

	for _, name := range sortedKeys(a.actions) {
		fmt.Fprintf(&b, "- %s: %s\n", name, a.actions[name].Description)
	}

	b.WriteString("\nAvailable skills (for the \"skill\" action):\n")

	skillNames := sortedSkillKeys(a.skills)

	if len(skillNames) == 0 {
		b.WriteString("- (none)\n")
	}

	for _, name := range skillNames {
		fmt.Fprintf(&b, "- %s\n", name)
	}

	b.WriteString("\nRespond with ONLY a JSON object, no prose, of the form:\n")
	b.WriteString(`{"action": "<action>", "params": { ... }}`)
	b.WriteString("\n\nTask:\n")
	b.WriteString(input)

	return b.String()
}

func sortedKeys(m map[string]Action) []string {
	keys := make([]string, 0, len(m))

	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}

func sortedSkillKeys(m map[string]Skill) []string {
	keys := make([]string, 0, len(m))

	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}

// extractJSON pulls the first JSON object out of an LLM response, tolerating
// surrounding prose or ```json code fences.
func extractJSON(s string) string {
	start := strings.IndexByte(s, '{')
	end := strings.LastIndexByte(s, '}')

	if start == -1 || end == -1 || end < start {
		return s
	}

	return s[start : end+1]
}
