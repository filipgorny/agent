package agent

import (
	"os"
	"path/filepath"
)

// detectRoot returns the project/repository root: the nearest ancestor of the
// working directory containing a .git directory, or the working directory.
func detectRoot() string {
	cwd, err := os.Getwd()

	if err != nil {
		return "."
	}

	dir := cwd

	for {
		if fi, err := os.Stat(filepath.Join(dir, ".git")); err == nil && fi.IsDir() {
			return dir
		}

		parent := filepath.Dir(dir)

		if parent == dir {
			return cwd
		}

		dir = parent
	}
}
