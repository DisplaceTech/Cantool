package plugin

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/displacetech/cantool/internal/exec"
)

// Loader discovers and loads plugins from the plugin directory.
type Loader struct {
	pluginDir string
	runner    exec.CommandRunner
}

// NewLoader creates a loader for the given plugin directory.
func NewLoader(pluginDir string, runner exec.CommandRunner) *Loader {
	return &Loader{pluginDir: pluginDir, runner: runner}
}

// Discover scans the plugin directory and returns metadata for all valid plugins.
func (l *Loader) Discover(ctx context.Context) ([]PluginMetadata, error) {
	entries, err := os.ReadDir(l.pluginDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var plugins []PluginMetadata
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		// Skip non-executable files
		if info.Mode()&0111 == 0 {
			continue
		}

		path := filepath.Join(l.pluginDir, entry.Name())
		meta, err := l.loadMetadata(ctx, path)
		if err != nil {
			continue // skip invalid plugins
		}
		plugins = append(plugins, *meta)
	}

	return plugins, nil
}

func (l *Loader) loadMetadata(ctx context.Context, path string) (*PluginMetadata, error) {
	out, err := l.runner.Run(ctx, path, "--metadata")
	if err != nil {
		return nil, err
	}
	if out.ExitCode != 0 {
		return nil, os.ErrInvalid
	}

	var meta PluginMetadata
	if err := json.Unmarshal([]byte(out.Stdout), &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}
