package agent

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileReadWriteEdit(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f.txt")
	ctx := context.Background()

	// write
	if _, err := (FileWrite{}).Run(ctx, map[string]any{"path": path, "content": "hello world"}); err != nil {
		t.Fatalf("write: %v", err)
	}

	// read
	out, err := (FileRead{}).Run(ctx, map[string]any{"path": path})

	if err != nil {
		t.Fatalf("read: %v", err)
	}

	if out != "hello world" {
		t.Errorf("read = %q", out)
	}

	// append
	if _, err := (FileWrite{}).Run(ctx, map[string]any{"path": path, "content": "!", "append": true}); err != nil {
		t.Fatalf("append: %v", err)
	}

	// edit (unique)
	if _, err := (FileEdit{}).Run(ctx, map[string]any{"path": path, "old": "world", "new": "go"}); err != nil {
		t.Fatalf("edit: %v", err)
	}

	out, _ = (FileRead{}).Run(ctx, map[string]any{"path": path})

	if out != "hello go!" {
		t.Errorf("after edit = %q, want %q", out, "hello go!")
	}
}

func TestFileEditNotFoundAndAmbiguous(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f.txt")
	ctx := context.Background()

	_ = os.WriteFile(path, []byte("a a a"), 0o644)

	if _, err := (FileEdit{}).Run(ctx, map[string]any{"path": path, "old": "zzz", "new": "x"}); err == nil {
		t.Error("expected not-found error")
	}

	if _, err := (FileEdit{}).Run(ctx, map[string]any{"path": path, "old": "a", "new": "b"}); err == nil {
		t.Error("expected ambiguous error (multiple occurrences without all)")
	}

	if _, err := (FileEdit{}).Run(ctx, map[string]any{"path": path, "old": "a", "new": "b", "all": true}); err != nil {
		t.Errorf("all replace: %v", err)
	}
}

func TestGrep(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "a.txt"), []byte("foo\nbar\nFOObar"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "b.log"), []byte("nothing here"), 0o644)
	ctx := context.Background()

	out, err := (Grep{}).Run(ctx, map[string]any{"pattern": "foo", "path": dir})

	if err != nil {
		t.Fatalf("grep: %v", err)
	}

	if !strings.Contains(out, "a.txt:1:foo") {
		t.Errorf("grep out missing match: %q", out)
	}

	// ignore_case picks up FOObar too
	out, _ = (Grep{}).Run(ctx, map[string]any{"pattern": "foo", "path": dir, "ignore_case": true})

	if !strings.Contains(out, "FOObar") {
		t.Errorf("ignore_case missed FOObar: %q", out)
	}

	// glob limits to .log
	out, _ = (Grep{}).Run(ctx, map[string]any{"pattern": ".", "path": dir, "glob": "*.log"})

	if strings.Contains(out, "a.txt") {
		t.Errorf("glob should exclude a.txt: %q", out)
	}
}

func TestDirList(t *testing.T) {
	dir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(dir, "sub"), 0o755)
	_ = os.WriteFile(filepath.Join(dir, "top.txt"), []byte("x"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "sub", "inner.txt"), []byte("y"), 0o644)
	ctx := context.Background()

	out, err := (DirList{}).Run(ctx, map[string]any{"path": dir, "depth": 1})

	if err != nil {
		t.Fatalf("dir_list: %v", err)
	}

	if !strings.Contains(out, "top.txt") || !strings.Contains(out, "sub/") {
		t.Errorf("depth-1 listing wrong: %q", out)
	}

	if strings.Contains(out, "inner.txt") {
		t.Errorf("depth-1 should not include nested file: %q", out)
	}

	out, _ = (DirList{}).Run(ctx, map[string]any{"path": dir, "depth": 2})

	if !strings.Contains(out, filepath.Join("sub", "inner.txt")) {
		t.Errorf("depth-2 should include nested file: %q", out)
	}
}
