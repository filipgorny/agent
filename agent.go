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
	"time"

	"github.com/filipgorny/agent/core"
	"github.com/filipgorny/agent/memory"
	"github.com/filipgorny/agent/message"
	"github.com/filipgorny/agent/runtime"
	"github.com/filipgorny/agent/stream"
	llm "github.com/filipgorny/llm-provider"
)

const (
	defaultMaxSteps       = 12
	defaultMaxResultChars = 2000
	maxStepsInContext     = 6 // bounded window of recent steps kept in the prompt
)

// Agent ties together an LLM, plugins (skills + events), a hard-coded action
// set, an event bus with listeners, threads and memory.
type Agent struct {
	llm            *llm.LlmProvider
	plugins        map[string]core.Plugin
	skills         map[string]core.Skill
	actions        map[string]Action
	bus            *core.EventBus
	listeners      *runtime.Listeners
	memory         memory.Memory
	threads        *runtime.Threads
	results        *runtime.ResultStore
	reactions      chan message.InputMessage
	initialMessage string
	language       string
	maxSteps       int
	maxResultChars int
	verbose        bool

	// session stream + interactive ask
	msgs        chan stream.Message
	interactive bool
	root        string
	answers     chan string

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
		listeners:      runtime.NewListeners(),
		memory:         mem,
		threads:        runtime.NewThreads(),
		results:        runtime.NewResultStore(),
		reactions:      make(chan message.InputMessage, 64),
		initialMessage: initialMessage,
		language:       language,
		maxSteps:       defaultMaxSteps,
		maxResultChars: defaultMaxResultChars,
		msgs:           make(chan stream.Message, 256),
		answers:        make(chan string, 1),
		runCtx:         context.Background(),
	}
}

// Messages returns the agent's outbound message stream (LOG, ANSWER_USER,
// ASK_USER, CHANGE_ROOT_FOLDER, …). A session/UI consumes it.
func (a *Agent) Messages() <-chan stream.Message {
	return a.msgs
}

// SetInteractive controls whether the agent may ask the user (ask_user/ask_choice).
func (a *Agent) SetInteractive(v bool) {
	a.interactive = v
}

// Root returns the agent's current project/working root.
func (a *Agent) Root() string {
	return a.root
}

// SetRoot sets the agent's working root.
func (a *Agent) SetRoot(path string) {
	a.root = path
}

// emitMsg pushes a message on the outbound stream (non-blocking; dropped if no
// consumer or buffer full, so the agent never stalls on a missing UI).
func (a *Agent) emitMsg(msgType, subtype string, payload any) {
	select {

	case a.msgs <- stream.Message{Type: msgType, Subtype: subtype, Payload: payload, CreatedAt: time.Now()}:

	default:
	}
}

// askUser emits an ASK_USER message and blocks until Answer delivers a reply.
func (a *Agent) askUser(ctx context.Context, req stream.AskRequest) (string, error) {
	if !a.interactive {
		return "", fmt.Errorf("agent: cannot ask the user (non-interactive)")
	}

	subtype := ""

	if len(req.Choices) > 0 {
		subtype = stream.SubtypeChoice
	}

	a.emitMsg(stream.TypeAskUser, subtype, map[string]any{"question": req.Question, "choices": req.Choices})

	select {

	case ans := <-a.answers:
		return ans, nil

	case <-ctx.Done():
		return "", ctx.Err()
	}
}

// Answer delivers a user reply to a pending ask_user/ask_choice.
func (a *Agent) Answer(text string) {
	select {

	case a.answers <- text:

	default:
	}
}

// execute emits TOOL_CALL/TOOL_RESULT around an action and returns its result.
func (a *Agent) execute(ctx context.Context, ec execContext, call message.ActionCall) (string, error) {
	a.emitMsg(stream.TypeLog, stream.LogToolCall, map[string]any{"action": call.Action, "params": call.Params})

	result, err := a.Execute(ctx, ec, call)

	if err != nil {
		a.emitMsg(stream.TypeLog, stream.LogError, err.Error())

		return "", err
	}

	a.emitMsg(stream.TypeLog, stream.LogToolResult, map[string]any{"action": call.Action, "result": a.condense(result)})

	return result, nil
}

// condense keeps large results OUT of the LLM context but retrievable: results
// over maxResultChars are stored and replaced with a short preview + an id the
// LLM can fetch from with the get_result action (context offloading). Small
// results pass through unchanged — nothing is ever lost.
func (a *Agent) condense(result string) string {
	if a.maxResultChars <= 0 || len(result) <= a.maxResultChars {
		return result
	}

	id := a.results.Put(result)

	return fmt.Sprintf("[large result stored: id=%s, %d bytes — read more with get_result {\"result_id\":%q, \"offset\":N, \"limit\":N}]\npreview:\n%s",
		id, len(result), id, result[:a.maxResultChars])
}

// condenseEvent applies result offloading to an event's "result" payload.
func (a *Agent) condenseEvent(ev core.Event) core.Event {
	r, ok := ev.Data["result"].(string)

	if !ok || a.maxResultChars <= 0 || len(r) <= a.maxResultChars {
		return ev
	}

	id := a.results.Put(r)

	data := make(map[string]any, len(ev.Data)+3)

	for k, v := range ev.Data {
		data[k] = v
	}

	data["result"] = r[:a.maxResultChars]
	data["result_id"] = id
	data["bytes"] = len(r)
	ev.Data = data

	return ev
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
func (a *Agent) emitEvent(ev core.Event) {
	a.listeners.Emit(ev)
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

			case a.reactions <- message.NewEventMessage(a.condenseEvent(ev)):

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

	main := a.threads.EnsureMain()

	return a.drive(ctx, main, message.NewUserInput(a.initialMessage))
}

// Ask sends free-form user text and reasons until a conclusion.
func (a *Agent) Ask(ctx context.Context, text string) (string, error) {
	a.setRunCtx(ctx)

	main := a.threads.EnsureMain()

	return a.drive(ctx, main, message.NewUserInput(text))
}

// Listen runs the reactive loop: events registered via listen_for drive new
// reasoning, until the context is cancelled.
func (a *Agent) Listen(ctx context.Context) error {
	a.setRunCtx(ctx)

	main := a.threads.EnsureMain()

	for {
		select {

		case <-ctx.Done():
			return ctx.Err()

		case msg := <-a.reactions:
			_, _ = a.drive(ctx, main, msg)
		}
	}
}

// reason runs the decide→execute loop until the LLM gives a plain-text answer or
// maxSteps is reached. The goal is PINNED in every prompt (so the model never
// loses the task), followed by a bounded window of recent steps (so context
// stays small for limited-context models).
func (a *Agent) reason(ctx context.Context, threadID string, goal message.InputMessage) (string, error) {
	goalJSON, _ := json.Marshal(goal)

	var steps []string

	for step := 0; step < a.maxSteps; step++ {
		out, err := a.llm.Prompt(ctx, a.reasonPrompt(goalJSON, steps))

		if err != nil {
			return "", fmt.Errorf("agent: reason: %w", err)
		}

		call, ok := parseActionCall(out)

		if !ok {
			return strings.TrimSpace(extractText(out)), nil
		}

		result, err := a.execute(ctx, execContext{threadID: threadID, actionUID: runtime.NewUID()}, call)

		if err != nil {
			result = "error: " + err.Error()
		}

		steps = append(steps, fmt.Sprintf("- %s -> %s", call.Action, a.condense(result)))

		if len(steps) > maxStepsInContext {
			steps = steps[len(steps)-maxStepsInContext:]
		}
	}

	return "", fmt.Errorf("agent: reasoning did not conclude in %d steps", a.maxSteps)
}

// reasonPrompt renders the preamble, the pinned goal and the recent steps.
func (a *Agent) reasonPrompt(goalJSON []byte, steps []string) string {
	var b strings.Builder

	b.WriteString(a.protocolPreamble())
	b.WriteString("\n\nGoal (keep working with actions until you can answer it; do not ask the user):\n")
	b.Write(goalJSON)

	if len(steps) > 0 {
		b.WriteString("\n\nSteps so far (their results are already known — do NOT repeat them):\n")
		b.WriteString(strings.Join(steps, "\n"))
	}

	b.WriteString("\n\nReply with the next action JSON to gather missing information, ")
	b.WriteString("or — as soon as the steps already contain enough to answer the Goal — ")
	b.WriteString("reply with the final answer as plain text.")

	return b.String()
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

// protocolPreamble is the generated system prompt. It is kept compact (one line
// per action/skill, event names only) so it fits limited-context models; set
// verbose for full descriptions when debugging.
func (a *Agent) protocolPreamble() string {
	var b strings.Builder

	b.WriteString("You are an agent using a JSON protocol. Incoming messages are JSON ")
	b.WriteString("with msg_type (user_input|event|action_result).\n")
	b.WriteString("Reply with ONLY {\"action\":\"<name>\",\"params\":{...}}, or plain text when done.\n")

	b.WriteString("\nActions:\n")

	for _, name := range sortedActionKeys(a.actions) {
		if a.verbose {
			fmt.Fprintf(&b, "- %s: %s\n", name, a.actions[name].Description)
		} else {
			fmt.Fprintf(&b, "- %s\n", name)
		}
	}

	b.WriteString("\nSkills (run via the skill action; async ones emit a result event to wait_for/listen_for):\n")

	for _, name := range sortedSkillKeys(a.skills) {
		s := a.skills[name]

		fmt.Fprintf(&b, "- %s (async=%t)", name, s.IsAsync())

		if events := eventNames(s.GetEvents()); events != "" {
			fmt.Fprintf(&b, " events:[%s]", events)
		}

		if a.verbose {
			fmt.Fprintf(&b, " — %s", s.Description())
		}

		b.WriteString("\n")
	}

	b.WriteString("\nExample of the loop (act, then answer from the results — never repeat an action):\n")
	b.WriteString("Goal: count .go files in dir \"x\"\n")
	b.WriteString("you: {\"action\":\"dir_list\",\"params\":{\"path\":\"x\"}}\n")
	b.WriteString("steps so far: - dir_list -> a.go\\nb.go\n")
	b.WriteString("you: There are 2 .go files: a.go and b.go.\n")

	fmt.Fprintf(&b, "\nRespond in %s.", a.language)

	return b.String()
}

// eventNames joins event spec names with commas.
func eventNames(specs []core.EventSpec) string {
	names := make([]string, 0, len(specs))

	for _, e := range specs {
		names = append(names, e.Name)
	}

	return strings.Join(names, ",")
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
