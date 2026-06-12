package memory

import (
	"context"
	"sort"
	"strings"
	"sync"
)

// InMemory is a dependency-free lexical memory: it scores records by how often
// the query terms appear in their content. Useful as a fallback and in tests.
type InMemory struct {
	mu      sync.RWMutex
	records []Record
}

// NewInMemory returns an empty in-memory store.
func NewInMemory() *InMemory {
	return &InMemory{}
}

func (m *InMemory) Remember(ctx context.Context, content string, meta map[string]any) error {
	m.mu.Lock()

	defer m.mu.Unlock()

	m.records = append(m.records, Record{Content: content, Meta: meta})

	return nil
}

func (m *InMemory) Read(ctx context.Context, query string, topK int) ([]Record, error) {
	terms := strings.Fields(strings.ToLower(query))

	m.mu.RLock()

	defer m.mu.RUnlock()

	var scored []Record

	for _, r := range m.records {
		content := strings.ToLower(r.Content)
		score := 0.0

		for _, t := range terms {
			score += float64(strings.Count(content, t))
		}

		if score > 0 {
			hit := r
			hit.Score = score
			scored = append(scored, hit)
		}
	}

	sort.SliceStable(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	if topK > 0 && len(scored) > topK {
		scored = scored[:topK]
	}

	return scored, nil
}
