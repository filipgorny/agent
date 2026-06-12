//go:build integration

package agent

import (
	"context"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/filipgorny/agent/config"
	"github.com/filipgorny/agent/message"
	"github.com/filipgorny/agent/plugins/shell"
	llm "github.com/filipgorny/llm-provider"
)

func ollamaTarget() (url, model string) {
	url = os.Getenv("OLLAMA_URL")

	if url == "" {
		url = "http://localhost:11434"
	}

	model = os.Getenv("OLLAMA_MODEL")

	if model == "" {
		model = "qwen2.5-coder:14b"
	}

	return url, model
}

// TestAgentMultiSkillReasoningIntegration: the agent must investigate its own
// codebase and decide on its own to run at least 3 skills (e.g. dir_list, grep,
// file_read/think) before answering which plugins/ subfolder defines the think
// skill. Verifies the reasoning loop, multi-skill use, and result correctness.
// Run from the agent module dir so relative paths resolve.
func TestAgentMultiSkillReasoningIntegration(t *testing.T) {
	url, model := ollamaTarget()

	a, err := NewAgentFromConfig(config.Config{
		InitialMessage: "Investigate this Go codebase and answer precisely: which subfolder under the " +
			"\"plugins\" directory defines the \"think\" skill, and how many Go source files (.go) does that " +
			"subfolder contain?",
		Llm: llm.Config{Llm: "ollama", Ollama: llm.OllamaConfig{
			URL: url, Model: model, Options: map[string]any{"num_ctx": 8192},
		}},
		Plugins:        []string{"files", "reasoning"},
		Skills:         []string{"dir_list", "grep", "file_read", "think"},
		MaxResultChars: 1500,
	})

	if err != nil {
		t.Fatalf("build agent: %v", err)
	}

	// Count skill executions via their completion events.
	ch, unsub := a.Bus().Subscribe()

	var (
		mu  sync.Mutex
		ran []string
	)

	go func() {
		for ev := range ch {
			if strings.HasSuffix(ev.Type, ".result") {
				r, _ := ev.Data["result"].(string)

				if len(r) > 120 {
					r = r[:120]
				}

				mu.Lock()
				ran = append(ran, ev.Source+" -> "+r)
				mu.Unlock()
			}
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 240*time.Second)

	defer cancel()

	out, err := a.Run(ctx)

	time.Sleep(200 * time.Millisecond) // let any async (think) result events arrive

	mu.Lock()
	count := len(ran)
	skills := append([]string(nil), ran...)
	mu.Unlock()

	unsub()

	t.Logf("run err: %v", err)
	t.Logf("skills run (%d): %v", count, skills)
	t.Logf("final answer: %q", out)

	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if count < 3 {
		t.Errorf("agent ran %d skills, want >= 3", count)
	}

	if !strings.Contains(strings.ToLower(out), "reasoning") {
		t.Errorf("final answer should name the 'reasoning' subfolder, got %q", out)
	}
}

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
