package agent

import (
	"os"
	"strings"

	llm "github.com/filipgorny/llm-provider"
)

// envSource loads an agent Config from environment variables under a prefix.
// For a prefix of "AGENT" it reads:
//
//	AGENT_INITIAL_PROMPT  -> Config.InitialPrompt
//	AGENT_SKILLS          -> Config.Skills (comma-separated)
//	AGENT_LLM*            -> Config.Llm (delegated to llm.Env("AGENT_LLM"))
type envSource struct {
	prefix string
}

// Env returns a ConfigSource backed by environment variables under prefix.
func Env(prefix string) ConfigSource {
	return envSource{prefix: prefix}
}

func (s envSource) Load() (Config, error) {
	llmCfg, err := llm.Env(s.prefix + "_LLM").Load()

	if err != nil {
		return Config{}, err
	}

	var skills []string

	if raw := os.Getenv(s.prefix + "_SKILLS"); raw != "" {
		for _, name := range strings.Split(raw, ",") {
			skills = append(skills, strings.TrimSpace(name))
		}
	}

	return Config{
		InitialPrompt: os.Getenv(s.prefix + "_INITIAL_PROMPT"),
		Llm:           llmCfg,
		Skills:        skills,
	}, nil
}
