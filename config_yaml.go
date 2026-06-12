package agent

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type yamlBytesSource struct {
	data []byte
}

type yamlFileSource struct {
	path string
}

// YamlBytes returns the default-format (YAML) ConfigSource from raw bytes.
func YamlBytes(data []byte) ConfigSource {
	return yamlBytesSource{data: data}
}

// YamlFile returns the default ConfigSource: an agent Config read from YAML.
func YamlFile(path string) ConfigSource {
	return yamlFileSource{path: path}
}

// LoadConfig is a convenience wrapper around the default (YAML file) source.
func LoadConfig(path string) (Config, error) {
	return YamlFile(path).Load()
}

func (s yamlBytesSource) Load() (Config, error) {
	var c Config

	if err := yaml.Unmarshal(s.data, &c); err != nil {
		return Config{}, fmt.Errorf("agent: parse yaml config: %w", err)
	}

	return c, nil
}

func (s yamlFileSource) Load() (Config, error) {
	data, err := os.ReadFile(s.path)

	if err != nil {
		return Config{}, fmt.Errorf("agent: read config %q: %w", s.path, err)
	}

	return YamlBytes(data).Load()
}
