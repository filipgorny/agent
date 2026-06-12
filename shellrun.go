package agent

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// ShellRunSkillName is the registered name of the shell skill.
const ShellRunSkillName = "shell_run"

func init() {
	RegisterSkill(ShellRunSkillName, func(Deps) Skill {
		return ShellRun{}
	})
}

// ShellRun is an example skill that runs a command in the shell. It expects a
// "command" string parameter.
type ShellRun struct{}

func (ShellRun) Name() string {
	return ShellRunSkillName
}

func (ShellRun) Run(ctx context.Context, params map[string]any) (string, error) {
	raw, ok := params["command"]

	if !ok {
		return "", fmt.Errorf("shell_run: missing \"command\" parameter")
	}

	command, ok := raw.(string)

	if !ok {
		return "", fmt.Errorf("shell_run: \"command\" must be a string, got %T", raw)
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
