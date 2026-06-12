//go:build integration

package agent

import (
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"

	llm "github.com/filipgorny/llm-provider"
)

// TestAgentShellViaInitialPromptIntegration is the full end-to-end test: a real
// LLM (Claude headless) reads the agent's initial prompt, decides on its own to
// invoke the shell_run skill, and the agent actually runs the command.
//
// It is skipped when the claude binary is not on PATH.
// Run with: go test -tags integration ./...
func TestAgentShellViaInitialPromptIntegration(t *testing.T) {
	if _, err := exec.LookPath("claude"); err != nil {
		t.Skip("claude binary not found on PATH")
	}

	const marker = "agent-shell-ok"

	a, err := NewAgentFromConfig(Config{
		InitialPrompt: "Use the shell_run skill to run exactly this shell command: echo " + marker,
		Llm:           llm.Config{Llm: "claude"},
		Skills:        []string{ShellRunSkillName},
	})

	if err != nil {
		t.Fatalf("build agent: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)

	defer cancel()

	// 1) The LLM must DECIDE, from the initial prompt alone, to run the shell skill.
	call, err := a.Decide(ctx, NewUserInput(a.InitialPrompt()))

	if err != nil {
		t.Fatalf("Decide: %v", err)
	}

	t.Logf("llm decided: action=%q params=%v", call.Action, call.Params)

	if call.Action != ActionSkill {
		t.Fatalf("llm chose action %q, want %q", call.Action, ActionSkill)
	}

	if name, _ := call.Params["name"].(string); name != ShellRunSkillName {
		t.Fatalf("llm chose skill %q, want %q", name, ShellRunSkillName)
	}

	// 2) Executing that decision actually runs the command.
	out, err := a.Execute(ctx, call)

	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	t.Logf("shell output: %q", out)

	if !strings.Contains(out, marker) {
		t.Errorf("expected shell output to contain %q, got %q", marker, out)
	}
}
