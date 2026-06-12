package config

import (
	"os"
	"strings"

	llm "github.com/filipgorny/llm-provider"
)

type envSource struct {
	prefix string
}

// Env returns a Source backed by environment variables under prefix. For prefix
// "AGENT" it reads AGENT_INITIAL_MESSAGE, AGENT_LANGUAGE, AGENT_PLUGINS,
// AGENT_SKILLS (comma-separated) and AGENT_LLM* (delegated to llm.Env).
func Env(prefix string) Source {
	return envSource{prefix: prefix}
}

func (s envSource) Load() (Config, error) {
	llmCfg, err := llm.Env(s.prefix + "_LLM").Load()

	if err != nil {
		return Config{}, err
	}

	return Config{
		InitialMessage: os.Getenv(s.prefix + "_INITIAL_MESSAGE"),
		Language:       os.Getenv(s.prefix + "_LANGUAGE"),
		Llm:            llmCfg,
		Plugins:        splitList(os.Getenv(s.prefix + "_PLUGINS")),
		Skills:         splitList(os.Getenv(s.prefix + "_SKILLS")),
	}, nil
}

func splitList(raw string) []string {
	if raw == "" {
		return nil
	}

	var out []string

	for _, name := range strings.Split(raw, ",") {
		out = append(out, strings.TrimSpace(name))
	}

	return out
}
