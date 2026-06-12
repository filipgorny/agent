package agent

import (
	"context"
	"testing"

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

func newAgentWithLlm(t *testing.T, strat llm.Llm, skills []string, initial string) *Agent {
	t.Helper()

	built, err := buildSkills(skills)

	if err != nil {
		t.Fatalf("buildSkills: %v", err)
	}

	return NewAgent(llm.NewLlmProvider(strat), built, initial)
}

// TestAgentRunsShellSkill is the closing test: the LLM decides to invoke the
// shell_run skill, and the agent actually runs the command.
func TestAgentRunsShellSkill(t *testing.T) {
	strat := &scriptedLlm{
		reply: func(n int, prompt string) string {
			return `{"action":"skill","params":{"name":"shell_run","command":"echo hello-from-agent"}}`
		},
	}

	a := newAgentWithLlm(t, strat, []string{"shell_run"}, "Run a shell command that prints a greeting.")

	out, err := a.Run(context.Background())

	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if out != "hello-from-agent" {
		t.Errorf("output = %q, want hello-from-agent", out)
	}

	// The decision prompt must include the initial prompt and the skill name.
	if len(strat.calls) != 1 {
		t.Fatalf("llm calls = %d, want 1", len(strat.calls))
	}
}

func TestAgentPromptAction(t *testing.T) {
	strat := &scriptedLlm{
		reply: func(n int, prompt string) string {
			if n == 0 {
				return "```json\n{\"action\":\"prompt\",\"params\":{\"text\":\"What is 2+2?\"}}\n```"
			}

			return "4"
		},
	}

	a := newAgentWithLlm(t, strat, nil, "Answer a question.")

	out, err := a.Run(context.Background())

	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if out != "4" {
		t.Errorf("output = %q, want 4", out)
	}

	// Two calls: one to decide, one for the prompt action itself.
	if len(strat.calls) != 2 {
		t.Fatalf("llm calls = %d, want 2", len(strat.calls))
	}

	if strat.calls[1] != "What is 2+2?" {
		t.Errorf("prompt action got %q, want What is 2+2?", strat.calls[1])
	}
}

func TestAgentUnknownAction(t *testing.T) {
	strat := &scriptedLlm{
		reply: func(n int, prompt string) string {
			return `{"action":"nope","params":{}}`
		},
	}

	a := newAgentWithLlm(t, strat, nil, "x")

	_, err := a.Run(context.Background())

	if err == nil {
		t.Fatal("expected error for unknown action")
	}
}
