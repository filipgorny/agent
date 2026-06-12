package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	llm "github.com/filipgorny/llm-provider"
)

// Agent ties together an LLM, configured skills, a hard-coded action set, an
// event bus and a memory. The LLM communicates via a typed JSON message
// protocol: it receives InputMessages and returns an action (OutputMessage).
type Agent struct {
	llm           *llm.LlmProvider
	skills        map[string]Skill
	actions       map[string]Action
	bus           *EventBus
	memory        Memory
	initialPrompt string
	language      string
}

// NewAgent builds an agent. The action set is hard-coded; skills, bus, memory
// and language come from configuration.
func NewAgent(provider *llm.LlmProvider, skills map[string]Skill, bus *EventBus, memory Memory, initialPrompt, language string) *Agent {
	if language == "" {
		language = "English"
	}

	return &Agent{
		llm:           provider,
		skills:        skills,
		actions:       builtinActions(),
		bus:           bus,
		memory:        memory,
		initialPrompt: initialPrompt,
		language:      language,
	}
}

// InitialPrompt returns the agent's initial prompt.
func (a *Agent) InitialPrompt() string {
	return a.initialPrompt
}

// Bus returns the agent's event bus (for publishing/subscribing).
func (a *Agent) Bus() *EventBus {
	return a.bus
}

// Run starts the agent from its initial prompt.
func (a *Agent) Run(ctx context.Context) (string, error) {
	return a.Handle(ctx, NewUserInput(a.initialPrompt))
}

// Ask sends free-form user text to the agent.
func (a *Agent) Ask(ctx context.Context, text string) (string, error) {
	return a.Handle(ctx, NewUserInput(text))
}

// Handle decides on an action for the message, then executes it.
func (a *Agent) Handle(ctx context.Context, msg InputMessage) (string, error) {
	call, err := a.Decide(ctx, msg)

	if err != nil {
		return "", err
	}

	return a.Execute(ctx, call)
}

// Listen consumes events from the bus and reacts to each as an EventMessage,
// until the context is cancelled. Errors handling individual events are ignored
// so one bad event does not stop the loop.
func (a *Agent) Listen(ctx context.Context) error {
	if a.bus == nil {
		return fmt.Errorf("agent: no event bus configured")
	}

	events, unsubscribe := a.bus.Subscribe()

	defer unsubscribe()

	for {
		select {

		case <-ctx.Done():
			return ctx.Err()

		case ev, ok := <-events:

			if !ok {
				return nil
			}

			_, _ = a.Handle(ctx, NewEventMessage(ev.Type, ev.Source, ev.Data))
		}
	}
}

// Decide sends an input message to the LLM and parses the action it returns.
func (a *Agent) Decide(ctx context.Context, msg InputMessage) (ActionCall, error) {
	payload, err := json.Marshal(msg)

	if err != nil {
		return ActionCall{}, fmt.Errorf("agent: marshal message: %w", err)
	}

	prompt := a.protocolPreamble() + "\n\nIncoming message:\n" + string(payload)

	out, err := a.llm.Prompt(ctx, prompt)

	if err != nil {
		return ActionCall{}, fmt.Errorf("agent: decide: %w", err)
	}

	var call ActionCall

	if err := json.Unmarshal([]byte(extractJSON(out)), &call); err != nil {
		return ActionCall{}, fmt.Errorf("agent: parse action from %q: %w", out, err)
	}

	if call.Action == "" {
		return ActionCall{}, fmt.Errorf("agent: llm returned no action: %q", out)
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

// protocolPreamble describes the JSON message protocol, available actions and
// skills, the required response shape and the configured language. It is
// prepended to every message sent to the LLM.
func (a *Agent) protocolPreamble() string {
	var b strings.Builder

	b.WriteString("You are an autonomous agent communicating over a JSON protocol.\n")
	b.WriteString("Each incoming message is a JSON object with a \"msg_type\" field ")
	b.WriteString("(e.g. \"user_input\", \"event\").\n\n")

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
	b.WriteString(`{"msg_type": "action", "action": "<action>", "params": { ... }}`)
	fmt.Fprintf(&b, "\n\nAlways respond in %s.", a.language)

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
