package shell

import (
	"context"
	"testing"
)

func TestShellRun(t *testing.T) {
	out, err := (ShellRun{}).Run(context.Background(), map[string]any{"command": "echo hello-shell"})

	if err != nil {
		t.Fatalf("run: %v", err)
	}

	if out != "hello-shell" {
		t.Errorf("out = %q", out)
	}

	if _, err := (ShellRun{}).Run(context.Background(), map[string]any{}); err == nil {
		t.Error("expected error for missing command")
	}
}
