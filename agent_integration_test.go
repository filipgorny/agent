//go:build integration

package agent

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/filipgorny/agent/config"
	"github.com/filipgorny/agent/message"
	"github.com/filipgorny/agent/plugins/shell"
	llm "github.com/filipgorny/llm-provider"
)

// TestAgentShellViaInitialPromptIntegration is the full end-to-end test: a real
// LLM reads the agent's initial prompt, decides on its own to invoke the
// shell_run skill via an action, and the agent runs the command.
//
// Uses Ollama (a raw instruction-following model follows the JSON protocol
// reliably; the claude CLI is itself an agent and ignores the protocol).
// Configure with OLLAMA_URL / OLLAMA_MODEL. Run: go test -tags integration ./...
func TestAgentShellViaInitialPromptIntegration(t *testing.T) {
	url := os.Getenv("OLLAMA_URL")

	if url == "" {
		url = "http://localhost:11434"
	}

	model := os.Getenv("OLLAMA_MODEL")

	if model == "" {
		model = "qwen2.5-coder:14b"
	}

	const marker = "agent-shell-ok"

	a, err := NewAgentFromConfig(config.Config{
		InitialMessage: "Use the shell_run skill to run exactly this shell command: echo " + marker,
		Llm:            llm.Config{Llm: "ollama", Ollama: llm.OllamaConfig{URL: url, Model: model}},
		Plugins:        []string{"shell"},
		Skills:         []string{"shell_run"},
	})

	if err != nil {
		t.Fatalf("build agent: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)

	defer cancel()

	// 1) The LLM must DECIDE, from the initial prompt alone, to run the shell skill.
	call, err := a.Decide(ctx, message.NewUserInput(a.InitialMessage()))

	if err != nil {
		t.Fatalf("Decide: %v", err)
	}

	t.Logf("llm decided: action=%q params=%v", call.Action, call.Params)

	if call.Action != ActionSkill {
		t.Fatalf("llm chose action %q, want %q", call.Action, ActionSkill)
	}

	if name, _ := call.Params["name"].(string); name != shell.SkillName {
		t.Fatalf("llm chose skill %q, want %q", name, shell.SkillName)
	}

	// 2) Executing that decision (through the skill action) actually runs the command.
	out, err := a.Execute(ctx, execContext{threadID: "main", actionUID: "itest"}, call)

	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	t.Logf("shell output: %q", out)

	if !strings.Contains(out, marker) {
		t.Errorf("expected shell output to contain %q, got %q", marker, out)
	}
}
