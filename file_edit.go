package agent

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// FileEditSkillName is the registered name of the file_edit skill.
const FileEditSkillName = "file_edit"

func init() {
	RegisterSkill(FileEditSkillName, func(Deps) Skill {
		return FileEdit{}
	})
}

// FileEdit replaces text in a file. Params: path, old, new, optional all.
// Without "all" the old string must occur exactly once.
type FileEdit struct{}

func (FileEdit) Name() string {
	return FileEditSkillName
}

func (FileEdit) Run(ctx context.Context, params map[string]any) (string, error) {
	path, ok := paramString(params, "path")

	if !ok {
		return "", fmt.Errorf("file_edit: missing string \"path\" parameter")
	}

	oldStr, ok := paramString(params, "old")

	if !ok {
		return "", fmt.Errorf("file_edit: missing string \"old\" parameter")
	}

	newStr, ok := paramString(params, "new")

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

	all := paramBool(params, "all")

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
