package agent

import (
	"testing"

	llm "github.com/filipgorny/llm-provider"
)

func TestNewAgentFromConfig(t *testing.T) {
	a, err := NewAgentFromConfig(Config{
		InitialPrompt: "hello",
		Llm: llm.Config{
			Llm:    "ollama",
			Ollama: llm.OllamaConfig{URL: "http://localhost:11434", Model: "llama3"},
		},
		Skills: []string{"shell_run"},
	})

	if err != nil {
		t.Fatalf("build: %v", err)
	}

	if a.InitialPrompt() != "hello" {
		t.Errorf("initial prompt = %q", a.InitialPrompt())
	}

	if _, ok := a.skills["shell_run"]; !ok {
		t.Error("shell_run skill not enabled")
	}
}

func TestNewAgentFromConfigUnknownSkill(t *testing.T) {
	_, err := NewAgentFromConfig(Config{
		Llm: llm.Config{
			Llm:    "ollama",
			Ollama: llm.OllamaConfig{URL: "http://localhost:11434", Model: "llama3"},
		},
		Skills: []string{"does_not_exist"},
	})

	if err == nil {
		t.Fatal("expected error for unknown skill")
	}
}

func TestNewAgentFromConfigBadLlm(t *testing.T) {
	_, err := NewAgentFromConfig(Config{Llm: llm.Config{Llm: "unknown"}})

	if err == nil {
		t.Fatal("expected error for unknown llm")
	}
}

func TestYamlBytesSource(t *testing.T) {
	yaml := []byte(`
initial_prompt: do the thing
llm:
  llm: ollama
  ollama:
    url: http://localhost:11434
    model: qwen2.5-coder:14b
skills:
  - shell_run
`)

	a, err := NewAgentFrom(YamlBytes(yaml))

	if err != nil {
		t.Fatalf("build: %v", err)
	}

	if a.InitialPrompt() != "do the thing" {
		t.Errorf("initial prompt = %q", a.InitialPrompt())
	}

	if _, ok := a.skills["shell_run"]; !ok {
		t.Error("shell_run skill not enabled")
	}
}

func TestEnvSource(t *testing.T) {
	t.Setenv("AGENT_INITIAL_PROMPT", "from env")
	t.Setenv("AGENT_SKILLS", "shell_run")
	t.Setenv("AGENT_LLM", "ollama")
	t.Setenv("AGENT_LLM_OLLAMA_URL", "http://localhost:11434")
	t.Setenv("AGENT_LLM_OLLAMA_MODEL", "llama3")

	a, err := NewAgentFrom(Env("AGENT"))

	if err != nil {
		t.Fatalf("build: %v", err)
	}

	if a.InitialPrompt() != "from env" {
		t.Errorf("initial prompt = %q", a.InitialPrompt())
	}

	if _, ok := a.skills["shell_run"]; !ok {
		t.Error("shell_run skill not enabled")
	}
}
