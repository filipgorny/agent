package agent

import (
	"context"
	"fmt"

	"github.com/fsnotify/fsnotify"
)

// FileWatchSkillName is the registered name of the file_watch skill.
const FileWatchSkillName = "file_watch"

// FileChangedEvent is the event type published when a watched path changes.
const FileChangedEvent = "file.changed"

func init() {
	RegisterSkill(FileWatchSkillName, func(d Deps) Skill {
		return &FileWatch{bus: d.Bus}
	})
}

// FileWatch asynchronously watches a path and publishes file.changed events on
// the agent's event bus. Params: path. The watch lives until ctx is cancelled.
type FileWatch struct {
	bus *EventBus
}

func (*FileWatch) Name() string {
	return FileWatchSkillName
}

func (w *FileWatch) Run(ctx context.Context, params map[string]any) (string, error) {
	if w.bus == nil {
		return "", fmt.Errorf("file_watch: no event bus available")
	}

	path, ok := paramString(params, "path")

	if !ok {
		return "", fmt.Errorf("file_watch: missing string \"path\" parameter")
	}

	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		return "", fmt.Errorf("file_watch: %w", err)
	}

	if err := watcher.Add(path); err != nil {
		watcher.Close()

		return "", fmt.Errorf("file_watch: %w", err)
	}

	go func() {
		defer watcher.Close()

		for {
			select {

			case <-ctx.Done():
				return

			case ev, ok := <-watcher.Events:

				if !ok {
					return
				}

				w.bus.Publish(Event{
					Type:   FileChangedEvent,
					Source: FileWatchSkillName,
					Data: map[string]any{
						"path": ev.Name,
						"op":   ev.Op.String(),
					},
				})

			case _, ok := <-watcher.Errors:

				if !ok {
					return
				}
			}
		}
	}()

	return fmt.Sprintf("watching %s", path), nil
}
