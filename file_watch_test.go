package agent

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileWatchPublishesEvent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "watched.txt")

	if err := os.WriteFile(path, []byte("initial"), 0o644); err != nil {
		t.Fatal(err)
	}

	bus := NewEventBus()
	ch, unsubscribe := bus.Subscribe()

	defer unsubscribe()

	skill := &FileWatch{bus: bus}

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	if _, err := skill.Run(ctx, map[string]any{"path": path}); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Trigger a change.
	if err := os.WriteFile(path, []byte("changed"), 0o644); err != nil {
		t.Fatal(err)
	}

	select {

	case ev := <-ch:

		if ev.Type != FileChangedEvent {
			t.Errorf("event type = %q, want %q", ev.Type, FileChangedEvent)
		}

		if ev.Source != FileWatchSkillName {
			t.Errorf("event source = %q", ev.Source)
		}

	case <-time.After(3 * time.Second):
		t.Fatal("no file.changed event received")
	}
}
