package files

import (
	"context"
	"fmt"

	"github.com/filipgorny/agent/core"
	"github.com/fsnotify/fsnotify"
)

// FileWatchSkillName is the registered name of the file_watch skill.
const FileWatchSkillName = "file_watch"

// FileChangedEvent is the event type published when a watched path changes.
const FileChangedEvent = "file.changed"

// FileWatch asynchronously watches a path and emits file.changed events. Params:
// path. The watch lives until ctx is cancelled.
type FileWatch struct {
	emit func(core.Event)
}

func (*FileWatch) Name() string {
	return FileWatchSkillName
}

func (*FileWatch) Description() string {
	return "Watch a path and emit file.changed events when it changes. params: {\"path\": string}"
}

func (*FileWatch) IsAsync() bool {
	return false
}

func (*FileWatch) GetEvents() []core.EventSpec {
	return []core.EventSpec{
		{Name: "file_watch.result", Description: "Emitted when the watch has started."},
		{Name: FileChangedEvent, Description: "Emitted whenever the watched path changes."},
	}
}

func (w *FileWatch) Run(ctx context.Context, params map[string]any) (string, error) {
	if w.emit == nil {
		return "", fmt.Errorf("file_watch: no event sink available")
	}

	path, ok := core.ParamString(params, "path")

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

				w.emit(core.Event{
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
