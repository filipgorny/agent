package agent

import (
	"context"
	"testing"
)

func TestShellRun(t *testing.T) {
	out, err := ShellRun{}.Run(context.Background(), map[string]any{
		"command": "echo hello-shell",
	})

	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if out != "hello-shell" {
		t.Errorf("output = %q, want hello-shell", out)
	}
}

func TestShellRunMissingCommand(t *testing.T) {
	_, err := ShellRun{}.Run(context.Background(), map[string]any{})

	if err == nil {
		t.Fatal("expected error for missing command")
	}
}

func TestShellRunFailingCommand(t *testing.T) {
	_, err := ShellRun{}.Run(context.Background(), map[string]any{
		"command": "exit 3",
	})

	if err == nil {
		t.Fatal("expected error for failing command")
	}
}
