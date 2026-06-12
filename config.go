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

	// Llm configures which LLM backend to use (see llm-provider).
	Llm llm.Config `yaml:"llm"`

	// Skills lists the names of skills to enable.
	Skills []string `yaml:"skills"`
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

	skills, err := buildSkills(c.Skills)

	if err != nil {
		return nil, err
	}

	return NewAgent(provider, skills, c.InitialPrompt), nil
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
