package files

import (
	"context"
	"fmt"
	"os"

	"github.com/filipgorny/agent/core"
)

// FileWriteSkillName is the registered name of the file_write skill.
const FileWriteSkillName = "file_write"

// FileWrite writes content to a file. Params: path, content, optional append.
type FileWrite struct{}

func (FileWrite) Name() string {
	return FileWriteSkillName
}

func (FileWrite) Description() string {
	return "Write (or append) content to a file. params: {\"path\": string, \"content\": string, \"append\": bool?}"
}

func (FileWrite) IsAsync() bool {
	return false
}

func (FileWrite) GetEvents() []core.EventSpec {
	return []core.EventSpec{{Name: "file_write.result", Description: "Emitted when file_write finishes."}}
}

func (FileWrite) Run(ctx context.Context, params map[string]any) (string, error) {
	path, ok := core.ParamString(params, "path")

	if !ok {
		return "", fmt.Errorf("file_write: missing string \"path\" parameter")
	}

	content, ok := core.ParamString(params, "content")

	if !ok {
		return "", fmt.Errorf("file_write: missing string \"content\" parameter")
	}

	if core.ParamBool(params, "append") {
		f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)

		if err != nil {
			return "", fmt.Errorf("file_write: %w", err)
		}

		defer f.Close()

		if _, err := f.WriteString(content); err != nil {
			return "", fmt.Errorf("file_write: %w", err)
		}

		return fmt.Sprintf("appended %d bytes to %s", len(content), path), nil
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("file_write: %w", err)
	}

	return fmt.Sprintf("wrote %d bytes to %s", len(content), path), nil
}
