package agent

import (
	"context"
	"testing"

	"github.com/filipgorny/agent/core"
	"github.com/filipgorny/agent/plugins/shell"
	llm "github.com/filipgorny/llm-provider"
)

// toolLlm is a fake strategy implementing both Llm and ToolCaller (native path).
type toolLlm struct {
	n     int
	reply func(n int) llm.ToolResponse
}

func (t *toolLlm) Prompt(ctx context.Context, prompt string) (string, error) {
	return "", nil
}

func (t *toolLlm) CallTools(ctx context.Context, msgs []llm.Message, tools []llm.ToolSpec) (llm.ToolResponse, error) {
	n := t.n
	t.n++

	return t.reply(n), nil
}

func TestReasonWithToolsNative(t *testing.T) {
	tl := &toolLlm{
		reply: func(n int) llm.ToolResponse {
			if n == 0 {
				return llm.ToolResponse{Calls: []llm.ToolCall{
					{Name: "shell_run", Arguments: map[string]any{"command": "echo native-ok"}},
				}}
			}

			return llm.ToolResponse{Text: "done"}
		},
	}

	a := newAgentWithLlm(t, tl, []core.Plugin{shell.ShellPlugin{}}, []string{"shell_run"}, "do it")

	out, err := a.Run(context.Background())

	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if out != "done" {
		t.Errorf("final = %q, want done", out)
	}

	if tl.n < 2 {
		t.Errorf("expected native tool loop (>=2 CallTools), got %d", tl.n)
	}
}

func TestBuildToolSpecs(t *testing.T) {
	a := newAgentWithLlm(t, &scriptedLlm{reply: func(int, string) string { return "" }},
		[]core.Plugin{shell.ShellPlugin{}}, []string{"shell_run"}, "")

	names := map[string]bool{}

	for _, tool := range a.buildToolSpecs() {
		names[tool.Name] = true
	}

	if names[ActionSkill] {
		t.Error("generic skill action should be excluded (skills are individual tools)")
	}

	if !names["shell_run"] {
		t.Error("shell_run skill tool missing")
	}

	for _, want := range []string{ActionGetResult, ActionWaitFor, ActionListenFor, ActionRemember, ActionRead} {
		if !names[want] {
			t.Errorf("core action tool %q missing", want)
		}
	}
}

func TestFallbackWhenNoToolCaller(t *testing.T) {
	// scriptedLlm implements Llm but NOT ToolCaller -> drive uses prompt-based reason.
	strat := &scriptedLlm{
		reply: func(n int, prompt string) string {
			if n == 0 {
				return `{"action":"skill","params":{"name":"shell_run","command":"echo fb"}}`
			}

			return "done"
		},
	}

	a := newAgentWithLlm(t, strat, []core.Plugin{shell.ShellPlugin{}}, []string{"shell_run"}, "x")

	out, err := a.Run(context.Background())

	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if out != "done" {
		t.Errorf("final = %q, want done", out)
	}

	if len(strat.calls) < 2 {
		t.Errorf("expected prompt-based fallback path, calls=%d", len(strat.calls))
	}
}
