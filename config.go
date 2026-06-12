package agent

import (
	llm "github.com/filipgorny/llm-provider"
)

// Config describes an agent. The "llm" section is passed straight to the
// llm-provider package to build the LLM instance. Skills are selected by name
// (they must be registered via RegisterSkill).
type Config struct {
	// InitialPrompt is the agent's starting prompt.
	InitialPrompt string `yaml:"initial_prompt"`

	// Language is the response language; defaults to "English".
	Language string `yaml:"language"`

	// Llm configures which LLM backend to use (see llm-provider).
	Llm llm.Config `yaml:"llm"`

	// Skills lists the names of skills to enable.
	Skills []string `yaml:"skills"`

	// Memory configures the agent's long-term memory.
	Memory MemoryConfig `yaml:"memory"`
}

// MemoryConfig selects and configures the memory backend.
type MemoryConfig struct {
	Backend string `yaml:"backend"` // "sqlite" (default) | "inmemory"
	Path    string `yaml:"path"`    // SQLite database file (empty = in-memory)
}

// ConfigSource produces an agent Config. YAML is the default source.
type ConfigSource interface {
	Load() (Config, error)
}

// NewAgentFromConfig builds an agent from an in-memory Config. The LLM section
// is delegated to llm-provider; skills are resolved from the registry.
func NewAgentFromConfig(c Config) (*Agent, error) {
	provider, err := llm.NewLlmProviderFromConfig(c.Llm)

	if err != nil {
		return nil, err
	}

	memory, err := newMemory(c.Memory)

	if err != nil {
		return nil, err
	}

	bus := NewEventBus()

	skills, err := buildSkills(c.Skills, Deps{LLM: provider, Bus: bus})

	if err != nil {
		return nil, err
	}

	return NewAgent(provider, skills, bus, memory, c.InitialPrompt, c.Language), nil
}

// NewAgentFrom loads a Config from any source and builds an agent.
// For the default YAML source: NewAgentFrom(YamlFile(path)).
func NewAgentFrom(src ConfigSource) (*Agent, error) {
	c, err := src.Load()

	if err != nil {
		return nil, err
	}

	return NewAgentFromConfig(c)
}
