// Package agent is a reactive, multi-threaded reasoning agent: plugins provide
// skills and events, the LLM communicates over a typed JSON protocol, and async
// skills run in side threads whose results return as events.
package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/filipgorny/agent/core"
	"github.com/filipgorny/agent/memory"
	"github.com/filipgorny/agent/message"
	llm "github.com/filipgorny/llm-provider"
)

const defaultMaxSteps = 12

// Agent ties together an LLM, plugins (skills + events), a hard-coded action
// set, an event bus with listeners, threads and memory.
type Agent struct {
	llm            *llm.LlmProvider
	plugins        map[string]core.Plugin
	skills         map[string]core.Skill
	actions        map[string]Action
	bus            *core.EventBus
	listeners      *listeners
	memory         memory.Memory
	threads        *threads
	reactions      chan message.InputMessage
	initialMessage string
	language       string
	maxSteps       int

	mu     sync.Mutex
	runCtx context.Context
}

// newAgent returns an initialized agent with no plugins/skills registered yet.
func newAgent(provider *llm.LlmProvider, mem memory.Memory, language, initialMessage string) *Agent {
	if language == "" {
		language = "English"
	}

	return &Agent{
		llm:            provider,
		plugins:        map[string]core.Plugin{},
		skills:         map[string]core.Skill{},
		actions:        builtinActions(),
		bus:            core.NewEventBus(),
		listeners:      newListeners(),
		memory:         mem,
		threads:        newThreads(),
		reactions:      make(chan message.InputMessage, 64),
		initialMessage: initialMessage,
		language:       language,
		maxSteps:       defaultMaxSteps,
		runCtx:         context.Background(),
	}
}

// RegisterPlugin makes a plugin's skills and events available to the agent.
func (a *Agent) RegisterPlugin(p core.Plugin) {
	a.plugins[p.Name()] = p
}

// Bus returns the agent's event bus.
func (a *Agent) Bus() *core.EventBus {
	return a.bus
}

// InitialMessage returns the agent's first message.
func (a *Agent) InitialMessage() string {
	return a.initialMessage
}

// buildSkills resolves the selected skills from the enabled plugins. An empty
// enabled list enables all registered plugins.
func (a *Agent) buildSkills(enabled, skillNames []string, deps core.Deps) (map[string]core.Skill, error) {
	var active []core.Plugin

	if len(enabled) == 0 {
		for _, p := range a.plugins {
			active = append(active, p)
		}
	} else {
		for _, name := range enabled {
			p, ok := a.plugins[name]

			if !ok {
				return nil, fmt.Errorf("agent: unknown plugin %q", name)
			}

			active = append(active, p)
		}
	}

	available := map[string]func(core.Deps) core.Skill{}

	for _, p := range active {
		for n, f := range p.Skills() {
			available[n] = f
		}
	}

	out := make(map[string]core.Skill, len(skillNames))

	for _, name := range skillNames {
		f, ok := available[name]

		if !ok {
			return nil, fmt.Errorf("agent: skill %q is not provided by any enabled plugin", name)
		}

		out[name] = f(deps)
	}

	return out, nil
}

// emit routes an event to listeners (wait_for/listen_for) and the bus.
func (a *Agent) emit(ev core.Event) {
	a.listeners.emit(ev)
	a.bus.Publish(ev)
}

// forward pipes a persistent listener channel into the reaction loop.
func (a *Agent) forward(ch <-chan core.Event) {
	for {
		select {

		case <-a.ctx().Done():
			return

		case ev, ok := <-ch:

			if !ok {
				return
			}

			select {

			case a.reactions <- message.NewEventMessage(ev):

			case <-a.ctx().Done():
				return
			}
		}
	}
}

func (a *Agent) setRunCtx(ctx context.Context) {
	a.mu.Lock()

	defer a.mu.Unlock()

	a.runCtx = ctx
}

func (a *Agent) ctx() context.Context {
	a.mu.Lock()

	defer a.mu.Unlock()

	return a.runCtx
}

// Run starts the agent from its initial prompt and reasons until a conclusion.
func (a *Agent) Run(ctx context.Context) (string, error) {
	a.setRunCtx(ctx)

	main := a.threads.ensureMain()

	return a.reason(ctx, main, message.NewUserInput(a.initialMessage))
}

// Ask sends free-form user text and reasons until a conclusion.
func (a *Agent) Ask(ctx context.Context, text string) (string, error) {
	a.setRunCtx(ctx)

	main := a.threads.ensureMain()

	return a.reason(ctx, main, message.NewUserInput(text))
}

// Listen runs the reactive loop: events registered via listen_for drive new
// reasoning, until the context is cancelled.
func (a *Agent) Listen(ctx context.Context) error {
	a.setRunCtx(ctx)

	main := a.threads.ensureMain()

	for {
		select {

		case <-ctx.Done():
			return ctx.Err()

		case msg := <-a.reactions:
			_, _ = a.reason(ctx, main, msg)
		}
	}
}

// reason runs the decide→execute loop, feeding each action result back, until
// the LLM responds with no action (its final answer) or maxSteps is reached.
func (a *Agent) reason(ctx context.Context, threadID string, msg message.InputMessage) (string, error) {
	for step := 0; step < a.maxSteps; step++ {
		out, err := a.llm.Prompt(ctx, a.buildPrompt(msg))

		if err != nil {
			return "", fmt.Errorf("agent: reason: %w", err)
		}

		call, ok := parseActionCall(out)

		if !ok {
			return strings.TrimSpace(extractText(out)), nil
		}

		result, err := a.Execute(ctx, execContext{threadID: threadID, actionUID: newUID()}, call)

		if err != nil {
			result = "error: " + err.Error()
		}

		msg = message.NewActionResult(call.Action, result)
	}

	return "", fmt.Errorf("agent: reasoning did not conclude in %d steps", a.maxSteps)
}

// Decide asks the LLM for a single action for the given message.
func (a *Agent) Decide(ctx context.Context, msg message.InputMessage) (message.ActionCall, error) {
	out, err := a.llm.Prompt(ctx, a.buildPrompt(msg))

	if err != nil {
		return message.ActionCall{}, fmt.Errorf("agent: decide: %w", err)
	}

	call, ok := parseActionCall(out)

	if !ok {
		return message.ActionCall{}, fmt.Errorf("agent: llm returned no action: %q", out)
	}

	return call, nil
}

// Execute runs the action named in the call. As a convenience, if the action
// name is actually a skill name, it is dispatched through the skill action (so
// both {"action":"skill","params":{"name":"x"}} and {"action":"x"} work).
func (a *Agent) Execute(ctx context.Context, ec execContext, call message.ActionCall) (string, error) {
	action, ok := a.actions[call.Action]

	if !ok {
		if _, isSkill := a.skills[call.Action]; isSkill {
			params := map[string]any{"name": call.Action}

			for k, v := range call.Params {
				params[k] = v
			}

			return a.actions[ActionSkill].Run(ctx, a, ec, params)
		}

		return "", fmt.Errorf("agent: unknown action %q", call.Action)
	}

	return action.Run(ctx, a, ec, call.Params)
}

// buildPrompt renders the protocol preamble plus the incoming message JSON.
func (a *Agent) buildPrompt(msg message.InputMessage) string {
	payload, _ := json.Marshal(msg)

	return a.protocolPreamble() + "\n\nIncoming message:\n" + string(payload)
}

// protocolPreamble describes the protocol, actions, plugins, skills (with async
// flag and events) and the language. It is prepended to every LLM message.
func (a *Agent) protocolPreamble() string {
	var b strings.Builder

	b.WriteString("You are an autonomous agent communicating over a JSON protocol.\n")
	b.WriteString("Each incoming message is JSON with a \"msg_type\" field ")
	b.WriteString("(\"user_input\", \"event\", \"action_result\").\n\n")

	b.WriteString("Available actions:\n")

	for _, name := range sortedActionKeys(a.actions) {
		fmt.Fprintf(&b, "- %s: %s\n", name, a.actions[name].Description)
	}

	b.WriteString("\nEnabled plugins:\n")

	for _, p := range a.sortedPlugins() {
		fmt.Fprintf(&b, "- %s: %s\n", p.Name(), p.Description())

		for _, ev := range p.Events() {
			fmt.Fprintf(&b, "    event %s: %s\n", ev.Name, ev.Description)
		}
	}

	b.WriteString("\nEnabled skills:\n")

	for _, name := range sortedSkillKeys(a.skills) {
		s := a.skills[name]

		fmt.Fprintf(&b, "- %s (async=%t): %s\n", name, s.IsAsync(), s.Description())

		for _, ev := range s.GetEvents() {
			fmt.Fprintf(&b, "    emits %s: %s\n", ev.Name, ev.Description)
		}
	}

	b.WriteString("\nRespond with ONLY a JSON object, no prose, of the form:\n")
	b.WriteString(`{"msg_type": "action", "action": "<action>", "params": { ... }}`)
	b.WriteString("\nWhen you are done, respond with your final answer as plain text (no JSON).")
	fmt.Fprintf(&b, "\n\nAlways respond in %s.", a.language)

	return b.String()
}

func (a *Agent) sortedPlugins() []core.Plugin {
	out := make([]core.Plugin, 0, len(a.plugins))

	for _, p := range a.plugins {
		out = append(out, p)
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].Name() < out[j].Name()
	})

	return out
}

func sortedActionKeys(m map[string]Action) []string {
	keys := make([]string, 0, len(m))

	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}

func sortedSkillKeys(m map[string]core.Skill) []string {
	keys := make([]string, 0, len(m))

	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}

// parseActionCall extracts an action call from an LLM response. Returns ok=false
// when the response carries no action (a plain-text final answer).
func parseActionCall(out string) (message.ActionCall, bool) {
	var call message.ActionCall

	if err := json.Unmarshal([]byte(extractJSON(out)), &call); err != nil {
		return message.ActionCall{}, false
	}

	if call.Action == "" {
		return message.ActionCall{}, false
	}

	return call, true
}

// extractJSON returns the outermost {...} span of s, or s unchanged.
func extractJSON(s string) string {
	start := strings.IndexByte(s, '{')
	end := strings.LastIndexByte(s, '}')

	if start == -1 || end == -1 || end < start {
		return s
	}

	return s[start : end+1]
}

// extractText strips ``` fences from a plain-text final answer.
func extractText(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```")

	return strings.TrimSuffix(s, "```")
}
