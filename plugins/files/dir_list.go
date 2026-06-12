package files

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/filipgorny/agent/core"
)

// DirListSkillName is the registered name of the dir_list skill.
const DirListSkillName = "dir_list"

// DirList lists a directory tree. Params: path, optional depth (default 1).
type DirList struct{}

func (DirList) Name() string {
	return DirListSkillName
}

func (DirList) Description() string {
	return "List a directory tree. params: {\"path\": string, \"depth\": int?}"
}

func (DirList) IsAsync() bool {
	return false
}

func (DirList) GetEvents() []core.EventSpec {
	return []core.EventSpec{{Name: "dir_list.result", Description: "Emitted with the listing when dir_list finishes."}}
}

func (DirList) Run(ctx context.Context, params map[string]any) (string, error) {
	root, ok := core.ParamString(params, "path")

	if !ok {
		return "", fmt.Errorf("dir_list: missing string \"path\" parameter")
	}

	depth := 1

	if d, ok := core.ParamInt(params, "depth"); ok {
		depth = d
	}

	var lines []string

	walk := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		rel, relErr := filepath.Rel(root, path)

		if relErr != nil || rel == "." {
			return nil
		}

		level := strings.Count(rel, string(filepath.Separator))

		if depth > 0 && level >= depth {
			if d.IsDir() {
				return fs.SkipDir
			}

			return nil
		}

		if d.IsDir() {
			lines = append(lines, rel+"/")
		} else {
			lines = append(lines, rel)
		}

		return nil
	}

	if err := filepath.WalkDir(root, walk); err != nil {
		return "", fmt.Errorf("dir_list: %w", err)
	}

	if len(lines) == 0 {
		return "(empty)", nil
	}

	return strings.Join(lines, "\n"), nil
}
