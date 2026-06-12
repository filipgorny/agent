package agent

import (
	"fmt"

	"github.com/filipgorny/agent/config"
	"github.com/filipgorny/agent/memory"
)

// newMemory builds a memory backend from config. Empty/"sqlite" → SQLite FTS5
// (in-memory when no path); "inmemory" → dependency-free lexical store.
func newMemory(c config.MemoryConfig) (memory.Memory, error) {
	switch c.Backend {

	case "", "sqlite":
		return memory.NewSQLite(c.Path)

	case "inmemory":
		return memory.NewInMemory(), nil

	default:
		return nil, fmt.Errorf("agent: unknown memory backend %q", c.Backend)
	}
}
