package config

import "testing"

func TestYamlBytesParse(t *testing.T) {
	yaml := []byte(`
initial_message: do the thing
language: Polish
llm:
  llm: ollama
  ollama:
    url: http://localhost:11434
    model: llama3
plugins: [files, web]
skills: [file_read, web_get]
memory:
  backend: sqlite
  path: ./m.db
max_result_chars: 1500
`)

	c, err := YamlBytes(yaml).Load()

	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if c.InitialMessage != "do the thing" || c.Language != "Polish" {
		t.Errorf("scalars wrong: %+v", c)
	}

	if c.Llm.Llm != "ollama" || c.Llm.Ollama.Model != "llama3" {
		t.Errorf("llm wrong: %+v", c.Llm)
	}

	if len(c.Plugins) != 2 || c.Plugins[0] != "files" {
		t.Errorf("plugins = %v", c.Plugins)
	}

	if len(c.Skills) != 2 || c.Skills[1] != "web_get" {
		t.Errorf("skills = %v", c.Skills)
	}

	if c.Memory.Backend != "sqlite" || c.Memory.Path != "./m.db" {
		t.Errorf("memory = %+v", c.Memory)
	}

	if c.MaxResultChars != 1500 {
		t.Errorf("max_result_chars = %d", c.MaxResultChars)
	}
}

func TestEnvParse(t *testing.T) {
	t.Setenv("AGENT_INITIAL_MESSAGE", "from env")
	t.Setenv("AGENT_PLUGINS", "shell, files")
	t.Setenv("AGENT_SKILLS", "shell_run")
	t.Setenv("AGENT_LLM", "ollama")
	t.Setenv("AGENT_LLM_OLLAMA_URL", "http://localhost:11434")
	t.Setenv("AGENT_LLM_OLLAMA_MODEL", "llama3")

	c, err := Env("AGENT").Load()

	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if c.InitialMessage != "from env" {
		t.Errorf("initial = %q", c.InitialMessage)
	}

	if len(c.Plugins) != 2 || c.Plugins[1] != "files" {
		t.Errorf("plugins = %v", c.Plugins)
	}

	if c.Llm.Llm != "ollama" || c.Llm.Ollama.Model != "llama3" {
		t.Errorf("llm = %+v", c.Llm)
	}
}
