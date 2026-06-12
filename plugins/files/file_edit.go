package files

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/filipgorny/agent/core"
)

// FileEditSkillName is the registered name of the file_edit skill.
const FileEditSkillName = "file_edit"

// FileEdit replaces text in a file. Params: path, old, new, optional all.
// Without "all" the old string must occur exactly once.
type FileEdit struct{}

func (FileEdit) Name() string {
	return FileEditSkillName
}

func (FileEdit) Description() string {
	return "Replace text in a file. params: {\"path\": string, \"old\": string, \"new\": string, \"all\": bool?}"
}

func (FileEdit) IsAsync() bool {
	return false
}

func (FileEdit) GetEvents() []core.EventSpec {
	return []core.EventSpec{{Name: "file_edit.result", Description: "Emitted when file_edit finishes."}}
}

func (FileEdit) Run(ctx context.Context, params map[string]any) (string, error) {
	path, ok := core.ParamString(params, "path")

	if !ok {
		return "", fmt.Errorf("file_edit: missing string \"path\" parameter")
	}

	oldStr, ok := core.ParamString(params, "old")

	if !ok {
		return "", fmt.Errorf("file_edit: missing string \"old\" parameter")
	}

	newStr, ok := core.ParamString(params, "new")

	if !ok {
		return "", fmt.Errorf("file_edit: missing string \"new\" parameter")
	}

	data, err := os.ReadFile(path)

	if err != nil {
		return "", fmt.Errorf("file_edit: %w", err)
	}

	content := string(data)
	count := strings.Count(content, oldStr)

	if count == 0 {
		return "", fmt.Errorf("file_edit: %q not found in %s", oldStr, path)
	}

	all := core.ParamBool(params, "all")

	if count > 1 && !all {
		return "", fmt.Errorf("file_edit: %q occurs %d times in %s (set \"all\" to replace all)", oldStr, count, path)
	}

	var updated string

	if all {
		updated = strings.ReplaceAll(content, oldStr, newStr)
	} else {
		updated = strings.Replace(content, oldStr, newStr, 1)
	}

	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		return "", fmt.Errorf("file_edit: %w", err)
	}

	return fmt.Sprintf("edited %s (%d replacement(s))", path, count), nil
}
