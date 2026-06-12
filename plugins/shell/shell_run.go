package shell

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/filipgorny/agent/core"
)

// SkillName is the registered name of the shell_run skill.
const SkillName = "shell_run"

// ShellRun runs a command in the shell. It expects a "command" string parameter.
type ShellRun struct{}

func (ShellRun) Name() string {
	return SkillName
}

func (ShellRun) Description() string {
	return "Run a shell command and return its stdout. params: {\"command\": string}"
}

func (ShellRun) IsAsync() bool {
	return false
}

func (ShellRun) GetEvents() []core.EventSpec {
	return []core.EventSpec{{Name: "shell_run.result", Description: "Emitted with the command output when shell_run finishes."}}
}

func (ShellRun) Run(ctx context.Context, params map[string]any) (string, error) {
	command, ok := core.ParamString(params, "command")

	if !ok {
		return "", fmt.Errorf("shell_run: missing \"command\" parameter")
	}

	cmd := exec.CommandContext(ctx, "sh", "-c", command)

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("shell_run: %w: %s", err, strings.TrimSpace(stderr.String()))
	}

	return strings.TrimSpace(stdout.String()), nil
}
