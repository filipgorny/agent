// Package config holds the agent configuration and its loaders (YAML/env).
package config

import llm "github.com/filipgorny/llm-provider"

// Config describes an agent. The "llm" section is passed to the llm-provider
// package; plugins are enabled and skills selected from them.
type Config struct {
	// InitialMessage is the first message sent to the agent. The system prompt
	// (protocol preamble) is generated from the enabled plugins/skills, not set here.
	InitialMessage string `yaml:"initial_message"`

	// Language is the response language; defaults to "English".
	Language string `yaml:"language"`

	// Llm configures which LLM backend to use (see llm-provider).
	Llm llm.Config `yaml:"llm"`

	// Plugins lists the plugins to enable.
	Plugins []string `yaml:"plugins"`

	// Skills lists the skills to enable (must come from an enabled plugin).
	Skills []string `yaml:"skills"`

	// Memory configures the agent's long-term memory.
	Memory MemoryConfig `yaml:"memory"`

	// MaxResultChars caps how many chars of a result are inlined into the LLM
	// context before offloading to the result store (0 = default).
	MaxResultChars int `yaml:"max_result_chars"`

	// MaxSteps caps reasoning steps per turn (0 = default).
	MaxSteps int `yaml:"max_steps"`

	// Verbose renders full descriptions in the generated system prompt (larger).
	Verbose bool `yaml:"verbose"`

	// Interactive allows the agent to ask the user (ask_user/ask_choice skills).
	Interactive bool `yaml:"interactive"`
}

// MemoryConfig selects and configures the memory backend.
type MemoryConfig struct {
	Backend string `yaml:"backend"` // "sqlite" (default) | "inmemory"
	Path    string `yaml:"path"`    // SQLite database file (empty = in-memory)
}

// Source produces a Config. YAML is the default source.
type Source interface {
	Load() (Config, error)
}
