package files

import (
	"bufio"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/filipgorny/agent/core"
)

// GrepSkillName is the registered name of the grep skill.
const GrepSkillName = "grep"

// Grep searches files for a regexp. Params: pattern, path, optional glob,
// ignore_case. path may be a file or a directory (searched recursively).
type Grep struct{}

func (Grep) Name() string {
	return GrepSkillName
}

func (Grep) Description() string {
	return "Search files for a regexp. params: {\"pattern\": string, \"path\": string, \"glob\": string?, \"ignore_case\": bool?}"
}

func (Grep) IsAsync() bool {
	return false
}

func (Grep) GetEvents() []core.EventSpec {
	return []core.EventSpec{{Name: "grep.result", Description: "Emitted with matching lines when grep finishes."}}
}

func (Grep) Run(ctx context.Context, params map[string]any) (string, error) {
	pattern, ok := core.ParamString(params, "pattern")

	if !ok {
		return "", fmt.Errorf("grep: missing string \"pattern\" parameter")
	}

	root, ok := core.ParamString(params, "path")

	if !ok {
		return "", fmt.Errorf("grep: missing string \"path\" parameter")
	}

	if core.ParamBool(params, "ignore_case") {
		pattern = "(?i)" + pattern
	}

	re, err := regexp.Compile(pattern)

	if err != nil {
		return "", fmt.Errorf("grep: invalid pattern: %w", err)
	}

	glob, _ := core.ParamString(params, "glob")

	var matches []string

	walk := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if d.IsDir() {
			return nil
		}

		if glob != "" {
			ok, _ := filepath.Match(glob, filepath.Base(path))

			if !ok {
				return nil
			}
		}

		found, err := grepFile(path, re)

		if err != nil {
			return nil
		}

		matches = append(matches, found...)

		return nil
	}

	info, err := os.Stat(root)

	if err != nil {
		return "", fmt.Errorf("grep: %w", err)
	}

	if info.IsDir() {
		if err := filepath.WalkDir(root, walk); err != nil {
			return "", fmt.Errorf("grep: %w", err)
		}
	} else {
		found, err := grepFile(root, re)

		if err != nil {
			return "", fmt.Errorf("grep: %w", err)
		}

		matches = found
	}

	if len(matches) == 0 {
		return "(no matches)", nil
	}

	return strings.Join(matches, "\n"), nil
}

func grepFile(path string, re *regexp.Regexp) ([]string, error) {
	f, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	defer f.Close()

	var out []string

	scanner := bufio.NewScanner(f)
	line := 0

	for scanner.Scan() {
		line++

		text := scanner.Text()

		if re.MatchString(text) {
			out = append(out, fmt.Sprintf("%s:%d:%s", path, line, text))
		}
	}

	return out, scanner.Err()
}
