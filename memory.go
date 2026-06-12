package agent

import (
	"context"
	"fmt"
)

// Record is one stored memory returned from a search.
type Record struct {
	Content string
	Meta    map[string]any
	Score   float64
}

// Memory is the agent's long-term store. Backends are pluggable (lexical now,
// vector later) and selected via configuration, mirroring the LLM strategy.
type Memory interface {
	Remember(ctx context.Context, content string, meta map[string]any) error
	Read(ctx context.Context, query string, topK int) ([]Record, error)
}

// newMemory builds a Memory from config. Empty/"sqlite" → SQLite FTS5 (default,
// in-memory when no path); "inmemory" → dependency-free lexical store.
func newMemory(c MemoryConfig) (Memory, error) {
	switch c.Backend {

	case "", "sqlite":
		return NewSQLiteMemory(c.Path)

	case "inmemory":
		return NewInMemoryMemory(), nil

	default:
		return nil, fmt.Errorf("agent: unknown memory backend %q", c.Backend)
	}
}
