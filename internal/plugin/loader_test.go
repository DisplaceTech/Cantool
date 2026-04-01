package plugin

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/displacetech/cantool/internal/exec"
	"github.com/displacetech/cantool/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoader_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	mock := testutil.NewMockRunner()
	loader := NewLoader(dir, mock)

	plugins, err := loader.Discover(context.Background())
	require.NoError(t, err)
	assert.Empty(t, plugins)
}

func TestLoader_NonexistentDir(t *testing.T) {
	mock := testutil.NewMockRunner()
	loader := NewLoader("/nonexistent/plugins", mock)

	plugins, err := loader.Discover(context.Background())
	require.NoError(t, err)
	assert.Nil(t, plugins)
}

func TestLoader_SkipsDirectories(t *testing.T) {
	dir := t.TempDir()
	os.Mkdir(filepath.Join(dir, "subdir"), 0755)
	mock := testutil.NewMockRunner()
	loader := NewLoader(dir, mock)

	plugins, err := loader.Discover(context.Background())
	require.NoError(t, err)
	assert.Empty(t, plugins)
}

func TestLoader_SkipsNonExecutable(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("not a plugin"), 0644)
	mock := testutil.NewMockRunner()
	loader := NewLoader(dir, mock)

	plugins, err := loader.Discover(context.Background())
	require.NoError(t, err)
	assert.Empty(t, plugins)
}

func TestLoader_LoadsValidPlugin(t *testing.T) {
	dir := t.TempDir()
	pluginPath := filepath.Join(dir, "my-plugin")
	os.WriteFile(pluginPath, []byte("#!/bin/sh\n"), 0755)

	mock := testutil.NewMockRunner()
	mock.Responses[pluginPath] = exec.Output{
		Stdout: `{"name":"my-plugin","version":"1.0.0","description":"test","author":"test","min_host_version":"0.1.0"}`,
	}
	loader := NewLoader(dir, mock)

	plugins, err := loader.Discover(context.Background())
	require.NoError(t, err)
	require.Len(t, plugins, 1)
	assert.Equal(t, "my-plugin", plugins[0].Name)
	assert.Equal(t, "1.0.0", plugins[0].Version)
}

func TestLoader_SkipsInvalidMetadata(t *testing.T) {
	dir := t.TempDir()
	pluginPath := filepath.Join(dir, "bad-plugin")
	os.WriteFile(pluginPath, []byte("#!/bin/sh\n"), 0755)

	mock := testutil.NewMockRunner()
	mock.Responses[pluginPath] = exec.Output{Stdout: "not json"}
	loader := NewLoader(dir, mock)

	plugins, err := loader.Discover(context.Background())
	require.NoError(t, err)
	assert.Empty(t, plugins)
}

func TestLoader_SkipsNonZeroExit(t *testing.T) {
	dir := t.TempDir()
	pluginPath := filepath.Join(dir, "failing-plugin")
	os.WriteFile(pluginPath, []byte("#!/bin/sh\n"), 0755)

	mock := testutil.NewMockRunner()
	mock.Responses[pluginPath] = exec.Output{ExitCode: 1}
	loader := NewLoader(dir, mock)

	plugins, err := loader.Discover(context.Background())
	require.NoError(t, err)
	assert.Empty(t, plugins)
}
