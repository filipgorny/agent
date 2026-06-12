package agent

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/filipgorny/agent/core"
	"github.com/filipgorny/agent/message"
	"github.com/filipgorny/agent/plugins/files"
	"github.com/filipgorny/agent/plugins/reasoning"
	"github.com/filipgorny/agent/plugins/shell"
)

func TestMemoryActions(t *testing.T) {
	a := newAgentWithLlm(t, &scriptedLlm{reply: func(int, string) string { return "" }}, nil, nil, "")
	ctx := context.Background()

	if _, err := a.Execute(ctx, execContext{}, message.ActionCall{Action: ActionRemember, Params: map[string]any{"content": "Paris is the capital of France"}}); err != nil {
		t.Fatalf("remember: %v", err)
	}

	out, err := a.Execute(ctx, execContext{}, message.ActionCall{Action: ActionRead, Params: map[string]any{"query": "capital France"}})

	if err != nil {
		t.Fatalf("read: %v", err)
	}

	if !strings.Contains(out, "Paris") {
		t.Errorf("read result missing fact: %q", out)
	}
}

// TestAsyncSkillEmitsEvent: running an async skill returns immediately and emits
// a completion event carrying action_uid + thread_id + result.
func TestAsyncSkillEmitsEvent(t *testing.T) {
	strat := &scriptedLlm{reply: func(int, string) string { return "CONCLUSION" }}
	a := newAgentWithLlm(t, strat, []core.Plugin{reasoning.ReasoningPlugin{}}, []string{"think"}, "")

	ch, _ := a.Bus().Subscribe()

	out, err := a.Execute(context.Background(), execContext{threadID: "main", actionUID: "act1"},
		message.ActionCall{Action: ActionSkill, Params: map[string]any{"name": "think", "topic": "x"}})

	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	if !strings.Contains(out, "spawned think") {
		t.Errorf("async skill should return spawn confirmation, got %q", out)
	}

	select {

	case ev := <-ch:

		if ev.Type != "think.result" || ev.ActionUID != "act1" || ev.ThreadID == "" {
			t.Errorf("bad event: %+v", ev)
		}

		if ev.Data["result"] != "CONCLUSION" {
			t.Errorf("result = %v", ev.Data["result"])
		}

	case <-time.After(3 * time.Second):
		t.Fatal("no completion event")
	}
}

func TestWaitFor(t *testing.T) {
	strat := &scriptedLlm{reply: func(int, string) string { return "RESULT" }}
	a := newAgentWithLlm(t, strat, []core.Plugin{reasoning.ReasoningPlugin{}}, []string{"think"}, "")
	ctx := context.Background()

	if _, err := a.Execute(ctx, execContext{threadID: "main", actionUID: "a"},
		message.ActionCall{Action: ActionSkill, Params: map[string]any{"name": "think", "topic": "x"}}); err != nil {
		t.Fatalf("spawn: %v", err)
	}

	out, err := a.Execute(ctx, execContext{},
		message.ActionCall{Action: ActionWaitFor, Params: map[string]any{"event": "think.result", "timeout_seconds": float64(5)}})

	if err != nil {
		t.Fatalf("wait_for: %v", err)
	}

	if out != "RESULT" {
		t.Errorf("wait_for returned %q, want RESULT", out)
	}
}

// signalLlm records each prompt and always returns a shell action.
type signalLlm struct {
	got chan string
}

func (s *signalLlm) Prompt(ctx context.Context, prompt string) (string, error) {
	select {

	case s.got <- prompt:

	default:
	}

	return `{"action":"skill","params":{"name":"shell_run","command":"echo hi"}}`, nil
}

func TestListenForReactsToEvent(t *testing.T) {
	sl := &signalLlm{got: make(chan string, 16)}
	a := newAgentWithLlm(t, sl, []core.Plugin{shell.ShellPlugin{}}, []string{"shell_run"}, "")

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	if _, err := a.Execute(ctx, execContext{threadID: "main"},
		message.ActionCall{Action: ActionListenFor, Params: map[string]any{"event": "file.changed"}}); err != nil {
		t.Fatalf("listen_for: %v", err)
	}

	go func() {
		_ = a.Listen(ctx)
	}()

	deadline := time.After(3 * time.Second)
	tick := time.NewTicker(50 * time.Millisecond)

	defer tick.Stop()

	for {
		select {

		case p := <-sl.got:

			if strings.Contains(p, `"msg_type":"event"`) {
				return
			}

		case <-tick.C:
			a.emitEvent(core.Event{Type: "file.changed", Source: "file_watch", Data: map[string]any{"path": "/x"}})

		case <-deadline:
			t.Fatal("agent did not react to event")
		}
	}
}

func TestPreamble(t *testing.T) {
	a := newAgentWithLlm(t, &scriptedLlm{reply: func(int, string) string { return "" }},
		[]core.Plugin{files.FilesPlugin{}, reasoning.ReasoningPlugin{}},
		[]string{"file_read", "think"}, "")

	pre := a.protocolPreamble()

	for _, want := range []string{
		ActionListenFor, ActionWaitFor, ActionRemember, ActionRead, ActionSkill, ActionGetResult,
		"file_read (async=false)", "think (async=true)", "think.result", "English",
	} {
		if !strings.Contains(pre, want) {
			t.Errorf("preamble missing %q", want)
		}
	}

	// Compact mode must not include long descriptions.
	if strings.Contains(pre, "Reason step-by-step") {
		t.Error("compact preamble should omit long descriptions")
	}
}

func TestResultOffloading(t *testing.T) {
	a := newAgentWithLlm(t, &scriptedLlm{reply: func(int, string) string { return "" }}, nil, nil, "")
	a.maxResultChars = 10

	small := "tiny"

	if a.condense(small) != small {
		t.Error("small result must pass through unchanged")
	}

	big := strings.Repeat("x", 100)
	condensed := a.condense(big)

	if !strings.Contains(condensed, "get_result") || !strings.Contains(condensed, "id=") {
		t.Errorf("condensed should reference get_result + id: %q", condensed)
	}

	if strings.Contains(condensed, big) {
		t.Error("full big result must not be inlined")
	}

	// The full result is retrievable losslessly via get_result.
	id := a.results.Put(big)

	out, err := a.Execute(context.Background(), execContext{}, message.ActionCall{
		Action: ActionGetResult,
		Params: map[string]any{"result_id": id, "offset": float64(0), "limit": float64(100)},
	})

	if err != nil {
		t.Fatalf("get_result: %v", err)
	}

	if out != big {
		t.Errorf("get_result did not return full content (%d chars)", len(out))
	}
}
