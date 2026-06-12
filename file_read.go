package agent

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// FileReadSkillName is the registered name of the file_read skill.
const FileReadSkillName = "file_read"

func init() {
	RegisterSkill(FileReadSkillName, func(Deps) Skill {
		return FileRead{}
	})
}

// FileRead reads a file. Params: path, optional offset/limit (line-based).
type FileRead struct{}

func (FileRead) Name() string {
	return FileReadSkillName
}

func (FileRead) Run(ctx context.Context, params map[string]any) (string, error) {
	path, ok := paramString(params, "path")

	if !ok {
		return "", fmt.Errorf("file_read: missing string \"path\" parameter")
	}

	data, err := os.ReadFile(path)

	if err != nil {
		return "", fmt.Errorf("file_read: %w", err)
	}

	offset, hasOffset := paramInt(params, "offset")
	limit, hasLimit := paramInt(params, "limit")

	if !hasOffset && !hasLimit {
		return string(data), nil
	}

	lines := strings.Split(string(data), "\n")

	if hasOffset && offset > 0 {
		if offset >= len(lines) {
			return "", nil
		}

		lines = lines[offset:]
	}

	if hasLimit && limit >= 0 && limit < len(lines) {
		lines = lines[:limit]
	}

	return strings.Join(lines, "\n"), nil
}
