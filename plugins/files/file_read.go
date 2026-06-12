package files

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/filipgorny/agent/core"
)

// FileReadSkillName is the registered name of the file_read skill.
const FileReadSkillName = "file_read"

// FileRead reads a file. Params: path, optional offset/limit (line-based).
type FileRead struct{}

func (FileRead) Name() string {
	return FileReadSkillName
}

func (FileRead) Description() string {
	return "Read a file's contents. params: {\"path\": string, \"offset\": int?, \"limit\": int?}"
}

func (FileRead) IsAsync() bool {
	return false
}

func (FileRead) GetEvents() []core.EventSpec {
	return []core.EventSpec{{Name: "file_read.result", Description: "Emitted with the file contents when file_read finishes."}}
}

func (FileRead) Run(ctx context.Context, params map[string]any) (string, error) {
	path, ok := core.ParamString(params, "path")

	if !ok {
		return "", fmt.Errorf("file_read: missing string \"path\" parameter")
	}

	data, err := os.ReadFile(path)

	if err != nil {
		return "", fmt.Errorf("file_read: %w", err)
	}

	offset, hasOffset := core.ParamInt(params, "offset")
	limit, hasLimit := core.ParamInt(params, "limit")

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
