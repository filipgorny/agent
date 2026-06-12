package agent

import (
	"context"
	"strings"
	"testing"

	"github.com/filipgorny/agent/core"
	"github.com/filipgorny/agent/memory"
	"github.com/filipgorny/agent/message"
	"github.com/filipgorny/agent/plugins/shell"
	llm "github.com/filipgorny/llm-provider"
)

// scriptedLlm is a fake Llm strategy that returns a scripted reply per call.
type scriptedLlm struct {
	calls []string
	reply func(n int, prompt string) string
}

func (s *scriptedLlm) Prompt(ctx context.Context, prompt string) (string, error) {
	n := len(s.calls)

	s.calls = append(s.calls, prompt)

	return s.reply(n, prompt), nil
}

func newAgentWithLlm(t *testing.T, strat llm.Llm, plugins []core.Plugin, skills []string, initial string) *Agent {
	t.Helper()

	provider := llm.NewLlmProvider(strat)
	a := newAgent(provider, memory.NewInMemory(), "English", initial)

	for _, p := range plugins {
		a.RegisterPlugin(p)
	}

	built, err := a.buildSkills(nil, skills, core.Deps{LLM: provider, Emit: a.emitEvent})

	if err != nil {
		t.Fatalf("buildSkills: %v", err)
	}

	a.skills = built

	return a
}

// TestAgentReasonRunsShellSkill: the LLM decides to run shell_run, the agent
// executes it, feeds the result back, and the LLM concludes with plain text.
func TestAgentReasonRunsShellSkill(t *testing.T) {
	strat := &scriptedLlm{
		reply: func(n int, prompt string) string {
			if n == 0 {
				return `{"action":"skill","params":{"name":"shell_run","command":"echo hello-from-agent"}}`
			}

			return "done"
		},
	}

	a := newAgentWithLlm(t, strat, []core.Plugin{shell.ShellPlugin{}}, []string{"shell_run"}, "run echo")

	out, err := a.Run(context.Background())

	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if out != "done" {
		t.Errorf("final = %q, want done", out)
	}

	if len(strat.calls) < 2 || !strings.Contains(strat.calls[1], "hello-from-agent") {
		t.Errorf("shell result not fed back: %v", strat.calls)
	}
}

func TestAgentPromptAction(t *testing.T) {
	strat := &scriptedLlm{
		reply: func(n int, prompt string) string {
			if n == 0 {
				return `{"action":"prompt","params":{"text":"What is 2+2?"}}`
			}

			return "4"
		},
	}

	a := newAgentWithLlm(t, strat, nil, nil, "answer")

	out, err := a.Run(context.Background())

	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if out != "4" {
		t.Errorf("final = %q, want 4", out)
	}

	if strat.calls[1] != "What is 2+2?" {
		t.Errorf("prompt action got %q", strat.calls[1])
	}
}

func TestExecuteUnknownAction(t *testing.T) {
	a := newAgentWithLlm(t, &scriptedLlm{reply: func(int, string) string { return "" }}, nil, nil, "")

	_, err := a.Execute(context.Background(), execContext{}, message.ActionCall{Action: "nope"})

	if err == nil {
		t.Fatal("expected error for unknown action")
	}
}
