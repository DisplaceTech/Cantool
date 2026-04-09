package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeTestConfig(t *testing.T, dir, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "cantool.yaml"), []byte(content), 0644))
}

func removeConvenienceCommands(t *testing.T) {
	t.Helper()
	for _, name := range []string{"build", "test", "clean", "dev", "doctor"} {
		if c := findSubCommand(name); c != nil {
			rootCmd.RemoveCommand(c)
		}
	}
}

// isolateHome sets HOME to an empty temp dir so tests don't pick up
// the real user's global config at ~/.config/cantool/config.yaml.
func isolateHome(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
}

func TestConvenienceCommands_RegisterWhenEnabled(t *testing.T) {
	isolateHome(t)
	dir := t.TempDir()
	writeTestConfig(t, dir, `version: "1"
project:
  name: test-app
plugins:
  convenience:
    enabled: true
`)

	orig, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() {
		_ = os.Chdir(orig)
		removeConvenienceCommands(t)
	})

	registerConvenienceCommands()

	names := commandNames(rootCmd)
	assert.Contains(t, names, "build")
	assert.Contains(t, names, "test")
	assert.Contains(t, names, "clean")
	assert.Contains(t, names, "dev")
	assert.Contains(t, names, "doctor")
}

func TestConvenienceCommands_NotRegisteredWhenDisabled(t *testing.T) {
	isolateHome(t)
	dir := t.TempDir()
	writeTestConfig(t, dir, `version: "1"
project:
  name: test-app
plugins:
  convenience:
    enabled: false
`)

	orig, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(orig) })

	removeConvenienceCommands(t)
	registerConvenienceCommands()

	names := commandNames(rootCmd)
	assert.NotContains(t, names, "build")
	assert.NotContains(t, names, "test")
	assert.NotContains(t, names, "clean")
	assert.NotContains(t, names, "dev")
	assert.NotContains(t, names, "doctor")
}

func TestConvenienceCommands_NotRegisteredWithNoConfig(t *testing.T) {
	isolateHome(t)
	dir := t.TempDir()

	orig, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(orig) })

	removeConvenienceCommands(t)
	registerConvenienceCommands()

	names := commandNames(rootCmd)
	assert.NotContains(t, names, "build")
	assert.NotContains(t, names, "doctor")
}

func TestConvenienceEnabled_True(t *testing.T) {
	isolateHome(t)
	dir := t.TempDir()
	writeTestConfig(t, dir, `version: "1"
project:
  name: test-app
plugins:
  convenience:
    enabled: true
`)

	orig, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(orig) })

	assert.True(t, ConvenienceEnabled())
}

func TestConvenienceEnabled_False(t *testing.T) {
	isolateHome(t)
	dir := t.TempDir()
	writeTestConfig(t, dir, `version: "1"
project:
  name: test-app
`)

	orig, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(orig) })

	assert.False(t, ConvenienceEnabled())
}

func TestPluginList_ShowsConveniencePlugin(t *testing.T) {
	isolateHome(t)
	dir := t.TempDir()
	writeTestConfig(t, dir, `version: "1"
project:
  name: test-app
plugins:
  convenience:
    enabled: true
`)

	orig, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(orig) })

	out := captureStdout(t, func() {
		rootCmd.SetArgs([]string{"plugin", "list"})
		t.Cleanup(func() { rootCmd.SetArgs(nil) })
		err := rootCmd.Execute()
		require.NoError(t, err)
	})

	assert.Contains(t, out, "convenience")
	assert.Contains(t, out, "enabled")
	assert.Contains(t, out, "built-in")
}

func TestPluginList_ShowsDisabledWhenNotEnabled(t *testing.T) {
	isolateHome(t)
	dir := t.TempDir()
	writeTestConfig(t, dir, `version: "1"
project:
  name: test-app
`)

	orig, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(orig) })

	out := captureStdout(t, func() {
		rootCmd.SetArgs([]string{"plugin", "list"})
		t.Cleanup(func() { rootCmd.SetArgs(nil) })
		err := rootCmd.Execute()
		require.NoError(t, err)
	})

	assert.Contains(t, out, "convenience")
	assert.Contains(t, out, "disabled")
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	require.NoError(t, err)

	origStdout := os.Stdout
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = origStdout })

	fn()

	_ = w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	return buf.String()
}

func commandNames(cmd *cobra.Command) []string {
	var names []string
	for _, c := range cmd.Commands() {
		names = append(names, c.Name())
	}
	return names
}

func findSubCommand(name string) *cobra.Command {
	for _, c := range rootCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}
