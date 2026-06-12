package agent

import (
	"context"
	"path/filepath"
	"testing"
)

func TestMemoryBackends(t *testing.T) {
	ctx := context.Background()

	backends := map[string]func(t *testing.T) Memory{
		"inmemory": func(t *testing.T) Memory {
			return NewInMemoryMemory()
		},
		"sqlite": func(t *testing.T) Memory {
			m, err := NewSQLiteMemory(filepath.Join(t.TempDir(), "mem.db"))

			if err != nil {
				t.Fatalf("sqlite: %v", err)
			}

			return m
		},
	}

	for name, build := range backends {

		t.Run(name, func(t *testing.T) {
			m := build(t)

			facts := []string{
				"The capital of France is Paris",
				"Bananas are yellow fruit",
				"The capital of Japan is Tokyo",
			}

			for _, f := range facts {
				if err := m.Remember(ctx, f, map[string]any{"source": "test"}); err != nil {
					t.Fatalf("remember: %v", err)
				}
			}

			records, err := m.Read(ctx, "capital France", 5)

			if err != nil {
				t.Fatalf("read: %v", err)
			}

			if len(records) == 0 {
				t.Fatal("expected at least one record")
			}

			if records[0].Content != "The capital of France is Paris" {
				t.Errorf("top record = %q, want the France fact", records[0].Content)
			}
		})
	}
}
