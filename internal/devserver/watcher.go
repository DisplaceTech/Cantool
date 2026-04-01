package devserver

import (
	"context"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watcher watches file paths and calls onChange on modifications.
type Watcher struct {
	paths    []string
	onChange func()
	debounce time.Duration
}

// NewWatcher creates a file watcher with the given debounce interval.
func NewWatcher(paths []string, debounce time.Duration, onChange func()) *Watcher {
	if debounce == 0 {
		debounce = 100 * time.Millisecond
	}
	return &Watcher{paths: paths, debounce: debounce, onChange: onChange}
}

// Watch blocks until the context is cancelled, triggering onChange on file changes.
func (w *Watcher) Watch(ctx context.Context) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	for _, p := range w.paths {
		if err := watcher.Add(p); err != nil {
			// Path may not exist yet — skip silently
			continue
		}
	}

	var timer *time.Timer
	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if event.Op&(fsnotify.Write|fsnotify.Create) == 0 {
				continue
			}
			if timer != nil {
				timer.Stop()
			}
			timer = time.AfterFunc(w.debounce, w.onChange)
		case _, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
		}
	}
}
