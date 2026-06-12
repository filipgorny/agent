package config

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

// YamlBytes returns the default-format (YAML) Source from raw bytes.
func YamlBytes(data []byte) Source {
	return yamlBytesSource{data: data}
}

// YamlFile returns the default Source: a Config read from a YAML file.
func YamlFile(path string) Source {
	return yamlFileSource{path: path}
}

// Load is a convenience wrapper around the default (YAML file) source.
func Load(path string) (Config, error) {
	return YamlFile(path).Load()
}

func (s yamlBytesSource) Load() (Config, error) {
	var c Config

	if err := yaml.Unmarshal(s.data, &c); err != nil {
		return Config{}, fmt.Errorf("config: parse yaml: %w", err)
	}

	return c, nil
}

func (s yamlFileSource) Load() (Config, error) {
	data, err := os.ReadFile(s.path)

	if err != nil {
		return Config{}, fmt.Errorf("config: read %q: %w", s.path, err)
	}

	return YamlBytes(data).Load()
}
