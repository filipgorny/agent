package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/filipgorny/agent/core"
	"github.com/filipgorny/agent/stream"
)

// Action names hard-coded into the agent.
const (
	ActionPrompt     = "prompt"
	ActionSkill      = "skill"
	ActionRemember   = "remember"
	ActionRead       = "read"
	ActionListenFor  = "listen_for"
	ActionWaitFor    = "wait_for"
	ActionGetResult  = "get_result"
	ActionChangeRoot = "change_root"
)

const defaultWaitTimeout = 120 * time.Second

// Action is a capability the LLM can trigger. Unlike skills, actions are
// hard-coded into the agent. Each takes a free-form parameter map.
type Action struct {
	Name        string
	Description string
	Run         func(ctx context.Context, a *Agent, ec execContext, params map[string]any) (string, error)
}

// builtinActions returns the hard-coded action set.
func builtinActions() map[string]Action {
	actions := []Action{
		{ActionPrompt, `Ask the LLM directly. params: {"text": string}`, runPromptAction},
		{ActionSkill, `Run a skill (async skills run in a side thread and emit an event). params: {"name": string, ...skill params}`, runSkillAction},
		{ActionRemember, `Store a memory. params: {"content": string, "meta": object?}`, runRememberAction},
		{ActionRead, `Search memory. params: {"query": string, "top_k": int?}`, runReadAction},
		{ActionListenFor, `Asynchronously register interest in an event; matching events drive the agent later. params: {"event": string, "thread_id": string?}`, runListenForAction},
		{ActionWaitFor, `Synchronously block until an event arrives and return its payload. params: {"event": string, "thread_id": string?, "timeout_seconds": int?}`, runWaitForAction},
		{ActionGetResult, `Read a slice of a stored large result. params: {"result_id": string, "offset": int?, "limit": int?}`, runGetResultAction},
		{ActionChangeRoot, `Change the working project/repository root. params: {"path": string}`, runChangeRootAction},
	}

	out := make(map[string]Action, len(actions))

	for _, act := range actions {
		out[act.Name] = act
	}

	return out
}

func runPromptAction(ctx context.Context, a *Agent, ec execContext, params map[string]any) (string, error) {
	text, ok := core.ParamString(params, "text")

	if !ok {
		return "", fmt.Errorf("action prompt: missing string \"text\" parameter")
	}

	return a.llm.Prompt(ctx, text)
}

func runSkillAction(ctx context.Context, a *Agent, ec execContext, params map[string]any) (string, error) {
	name, ok := core.ParamString(params, "name")

	if !ok {
		return "", fmt.Errorf("action skill: missing string \"name\" parameter")
	}

	skill, ok := a.skills[name]

	if !ok {
		return "", fmt.Errorf("action skill: skill %q is not enabled", name)
	}

	if skill.IsAsync() {
		threadID := a.threads.Spawn(ec.threadID)
		actionUID := ec.actionUID
		runCtx := context.WithoutCancel(ctx)

		go func() {
			out, err := skill.Run(runCtx, params)

			data := map[string]any{}

			if err != nil {
				data["error"] = err.Error()
			} else {
				data["result"] = out
			}

			a.emitEvent(core.Event{
				Type:      core.ResultEvent(skill),
				Source:    name,
				ActionUID: actionUID,
				ThreadID:  threadID,
				Data:      data,
			})
		}()

		return fmt.Sprintf("spawned %s in thread %s", name, threadID), nil
	}

	out, err := skill.Run(ctx, params)

	if err != nil {
		return "", err
	}

	a.emitEvent(core.Event{
		Type:      core.ResultEvent(skill),
		Source:    name,
		ActionUID: ec.actionUID,
		ThreadID:  ec.threadID,
		Data:      map[string]any{"result": out},
	})

	return out, nil
}

func runRememberAction(ctx context.Context, a *Agent, ec execContext, params map[string]any) (string, error) {
	if a.memory == nil {
		return "", fmt.Errorf("action remember: no memory configured")
	}

	content, ok := core.ParamString(params, "content")

	if !ok {
		return "", fmt.Errorf("action remember: missing string \"content\" parameter")
	}

	meta, _ := params["meta"].(map[string]any)

	if err := a.memory.Remember(ctx, content, meta); err != nil {
		return "", err
	}

	return "remembered", nil
}

func runReadAction(ctx context.Context, a *Agent, ec execContext, params map[string]any) (string, error) {
	if a.memory == nil {
		return "", fmt.Errorf("action read: no memory configured")
	}

	query, ok := core.ParamString(params, "query")

	if !ok {
		return "", fmt.Errorf("action read: missing string \"query\" parameter")
	}

	topK, _ := core.ParamInt(params, "top_k")

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

func runListenForAction(ctx context.Context, a *Agent, ec execContext, params map[string]any) (string, error) {
	event, ok := core.ParamString(params, "event")

	if !ok {
		return "", fmt.Errorf("action listen_for: missing string \"event\" parameter")
	}

	threadID, _ := core.ParamString(params, "thread_id")

	ch, _ := a.listeners.Register(event, threadID, false)

	go a.forward(ch)

	return fmt.Sprintf("listening for %s", event), nil
}

func runWaitForAction(ctx context.Context, a *Agent, ec execContext, params map[string]any) (string, error) {
	event, ok := core.ParamString(params, "event")

	if !ok {
		return "", fmt.Errorf("action wait_for: missing string \"event\" parameter")
	}

	threadID, _ := core.ParamString(params, "thread_id")

	timeout := defaultWaitTimeout

	if secs, ok := core.ParamInt(params, "timeout_seconds"); ok && secs > 0 {
		timeout = time.Duration(secs) * time.Second
	}

	ch, cancel := a.listeners.Register(event, threadID, true)

	defer cancel()

	select {

	case ev := <-ch:
		return eventPayload(ev), nil

	case <-time.After(timeout):
		return "", fmt.Errorf("action wait_for: timeout waiting for %s", event)

	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func runGetResultAction(ctx context.Context, a *Agent, ec execContext, params map[string]any) (string, error) {
	id, ok := core.ParamString(params, "result_id")

	if !ok {
		return "", fmt.Errorf("action get_result: missing string \"result_id\" parameter")
	}

	full, ok := a.results.Get(id)

	if !ok {
		return "", fmt.Errorf("action get_result: unknown result_id %q", id)
	}

	offset, _ := core.ParamInt(params, "offset")

	if offset < 0 || offset > len(full) {
		return "", nil
	}

	slice := full[offset:]

	if limit, ok := core.ParamInt(params, "limit"); ok && limit > 0 && limit < len(slice) {
		slice = slice[:limit]
	}

	return slice, nil
}

func runChangeRootAction(ctx context.Context, a *Agent, ec execContext, params map[string]any) (string, error) {
	path, ok := core.ParamString(params, "path")

	if !ok {
		return "", fmt.Errorf("action change_root: missing string \"path\" parameter")
	}

	if err := os.Chdir(path); err != nil {
		return "", fmt.Errorf("action change_root: %w", err)
	}

	a.SetRoot(path)
	a.emitMsg(stream.TypeChangeRoot, "", path)

	return "root changed to " + path, nil
}

// eventPayload renders an event's data for feeding back to the LLM.
func eventPayload(ev core.Event) string {
	if r, ok := ev.Data["result"].(string); ok {
		return r
	}

	b, _ := json.Marshal(ev.Data)

	return string(b)
}
