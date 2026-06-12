package files

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/filipgorny/agent/core"
)

func TestFileReadWriteEdit(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f.txt")
	ctx := context.Background()

	if _, err := (FileWrite{}).Run(ctx, map[string]any{"path": path, "content": "hello world"}); err != nil {
		t.Fatalf("write: %v", err)
	}

	out, err := (FileRead{}).Run(ctx, map[string]any{"path": path})

	if err != nil {
		t.Fatalf("read: %v", err)
	}

	if out != "hello world" {
		t.Errorf("read = %q", out)
	}

	if _, err := (FileWrite{}).Run(ctx, map[string]any{"path": path, "content": "!", "append": true}); err != nil {
		t.Fatalf("append: %v", err)
	}

	if _, err := (FileEdit{}).Run(ctx, map[string]any{"path": path, "old": "world", "new": "go"}); err != nil {
		t.Fatalf("edit: %v", err)
	}

	out, _ = (FileRead{}).Run(ctx, map[string]any{"path": path})

	if out != "hello go!" {
		t.Errorf("after edit = %q, want %q", out, "hello go!")
	}
}

func TestFileEditErrors(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f.txt")
	_ = os.WriteFile(path, []byte("a a a"), 0o644)
	ctx := context.Background()

	if _, err := (FileEdit{}).Run(ctx, map[string]any{"path": path, "old": "zzz", "new": "x"}); err == nil {
		t.Error("expected not-found error")
	}

	if _, err := (FileEdit{}).Run(ctx, map[string]any{"path": path, "old": "a", "new": "b"}); err == nil {
		t.Error("expected ambiguous error")
	}

	if _, err := (FileEdit{}).Run(ctx, map[string]any{"path": path, "old": "a", "new": "b", "all": true}); err != nil {
		t.Errorf("all replace: %v", err)
	}
}

func TestGrepAndDirList(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "a.txt"), []byte("foo\nbar"), 0o644)
	_ = os.MkdirAll(filepath.Join(dir, "sub"), 0o755)
	_ = os.WriteFile(filepath.Join(dir, "sub", "inner.txt"), []byte("y"), 0o644)
	ctx := context.Background()

	out, err := (Grep{}).Run(ctx, map[string]any{"pattern": "foo", "path": dir})

	if err != nil || !strings.Contains(out, "a.txt:1:foo") {
		t.Errorf("grep = %q, err=%v", out, err)
	}

	out, _ = (DirList{}).Run(ctx, map[string]any{"path": dir, "depth": 1})

	if !strings.Contains(out, "a.txt") || !strings.Contains(out, "sub/") || strings.Contains(out, "inner.txt") {
		t.Errorf("dir_list depth1 = %q", out)
	}
}

func TestFileWatch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "w.txt")
	_ = os.WriteFile(path, []byte("init"), 0o644)

	events := make(chan core.Event, 8)
	skill := &FileWatch{emit: func(ev core.Event) { events <- ev }}

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	if _, err := skill.Run(ctx, map[string]any{"path": path}); err != nil {
		t.Fatalf("run: %v", err)
	}

	_ = os.WriteFile(path, []byte("changed"), 0o644)

	select {

	case ev := <-events:

		if ev.Type != FileChangedEvent {
			t.Errorf("event type = %q", ev.Type)
		}

	case <-time.After(3 * time.Second):
		t.Fatal("no file.changed event")
	}
}
