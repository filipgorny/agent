package agent

import (
	"github.com/filipgorny/agent/config"
	"github.com/filipgorny/agent/core"
	llm "github.com/filipgorny/llm-provider"
)

// NewAgentFromConfig builds an agent from config. The given plugins are made
// available (default: DefaultPlugins()); c.Plugins enables a subset (empty =
// all) and c.Skills selects skills from the enabled plugins.
func NewAgentFromConfig(c config.Config, plugins ...core.Plugin) (*Agent, error) {
	if len(plugins) == 0 {
		plugins = DefaultPlugins()
	}

	provider, err := llm.NewLlmProviderFromConfig(c.Llm)

	if err != nil {
		return nil, err
	}

	mem, err := newMemory(c.Memory)

	if err != nil {
		return nil, err
	}

	a := newAgent(provider, mem, c.Language, c.InitialMessage)

	if c.MaxResultChars > 0 {
		a.maxResultChars = c.MaxResultChars
	}

	if c.MaxSteps > 0 {
		a.maxSteps = c.MaxSteps
	}

	a.verbose = c.Verbose

	for _, p := range plugins {
		a.RegisterPlugin(p)
	}

	skills, err := a.buildSkills(c.Plugins, c.Skills, core.Deps{LLM: provider, Emit: a.emit})

	if err != nil {
		return nil, err
	}

	a.skills = skills

	return a, nil
}

// NewAgentFrom loads a Config from any source and builds an agent.
// For the default YAML source: NewAgentFrom(config.YamlFile(path)).
func NewAgentFrom(src config.Source, plugins ...core.Plugin) (*Agent, error) {
	c, err := src.Load()

	if err != nil {
		return nil, err
	}

	return NewAgentFromConfig(c, plugins...)
}
