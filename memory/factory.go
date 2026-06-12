package memory

import (
	"fmt"

	"github.com/filipgorny/agent/config"
)

// New builds a memory backend from config. Empty/"sqlite" → SQLite FTS5
// (in-memory when no path); "inmemory" → dependency-free lexical store.
func New(c config.MemoryConfig) (Memory, error) {
	switch c.Backend {

	case "", "sqlite":
		return NewSQLite(c.Path)

	case "inmemory":
		return NewInMemory(), nil

	default:
		return nil, fmt.Errorf("memory: unknown backend %q", c.Backend)
	}
}
