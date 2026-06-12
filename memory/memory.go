// Package memory provides the agent's pluggable long-term memory and backends.
package memory

import "context"

// DefaultTopK is the number of results returned when a read omits top_k.
const DefaultTopK = 5

// Memory is the agent's long-term store. Backends are pluggable (lexical now,
// vector later), mirroring the LLM strategy.
type Memory interface {
	Remember(ctx context.Context, content string, meta map[string]any) error
	Read(ctx context.Context, query string, topK int) ([]Record, error)
}
