package agent

import (
	"testing"

	"github.com/filipgorny/agent/config"
	llm "github.com/filipgorny/llm-provider"
)

func ollamaCfg() llm.Config {
	return llm.Config{Llm: "ollama", Ollama: llm.OllamaConfig{URL: "http://localhost:11434", Model: "llama3"}}
}

func TestNewAgentFromConfig(t *testing.T) {
	a, err := NewAgentFromConfig(config.Config{
		InitialMessage: "hello",
		Llm:            ollamaCfg(),
		Plugins:        []string{"files"},
		Skills:         []string{"file_read"},
	})

	if err != nil {
		t.Fatalf("build: %v", err)
	}

	if a.InitialMessage() != "hello" {
		t.Errorf("initial = %q", a.InitialMessage())
	}

	if _, ok := a.skills["file_read"]; !ok {
		t.Error("file_read not enabled")
	}
}

func TestUnknownPlugin(t *testing.T) {
	_, err := NewAgentFromConfig(config.Config{Llm: ollamaCfg(), Plugins: []string{"nope"}, Skills: []string{"x"}})

	if err == nil {
		t.Fatal("expected error for unknown plugin")
	}
}

func TestSkillNotInEnabledPlugin(t *testing.T) {
	_, err := NewAgentFromConfig(config.Config{Llm: ollamaCfg(), Plugins: []string{"files"}, Skills: []string{"shell_run"}})

	if err == nil {
		t.Fatal("expected error: shell_run not in files plugin")
	}
}

func TestBadLlm(t *testing.T) {
	_, err := NewAgentFromConfig(config.Config{Llm: llm.Config{Llm: "unknown"}})

	if err == nil {
		t.Fatal("expected error for unknown llm")
	}
}

func TestYamlBytesSource(t *testing.T) {
	yaml := []byte(`
initial_message: do the thing
llm:
  llm: ollama
  ollama:
    url: http://localhost:11434
    model: llama3
plugins: [files]
skills: [file_read]
`)

	a, err := NewAgentFrom(config.YamlBytes(yaml))

	if err != nil {
		t.Fatalf("build: %v", err)
	}

	if a.InitialMessage() != "do the thing" {
		t.Errorf("initial = %q", a.InitialMessage())
	}

	if _, ok := a.skills["file_read"]; !ok {
		t.Error("file_read not enabled")
	}
}

func TestEnvSource(t *testing.T) {
	t.Setenv("AGENT_INITIAL_MESSAGE", "from env")
	t.Setenv("AGENT_PLUGINS", "shell")
	t.Setenv("AGENT_SKILLS", "shell_run")
	t.Setenv("AGENT_LLM", "ollama")
	t.Setenv("AGENT_LLM_OLLAMA_URL", "http://localhost:11434")
	t.Setenv("AGENT_LLM_OLLAMA_MODEL", "llama3")

	a, err := NewAgentFrom(config.Env("AGENT"))

	if err != nil {
		t.Fatalf("build: %v", err)
	}

	if a.InitialMessage() != "from env" {
		t.Errorf("initial = %q", a.InitialMessage())
	}

	if _, ok := a.skills["shell_run"]; !ok {
		t.Error("shell_run not enabled")
	}
}
