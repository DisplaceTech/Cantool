package cmd

import (
	"os"
	"path/filepath"

	"github.com/displacetech/cantool/internal/output"
	"github.com/spf13/cobra"
)

var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Manage plugins",
}

var pluginListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed plugins",
	RunE:  runPluginList,
}

func init() {
	pluginCmd.AddCommand(pluginListCmd)
	rootCmd.AddCommand(pluginCmd)
}

func runPluginList(_ *cobra.Command, _ []string) error {
	f := output.New(Format())

	var rows [][]string

	// Built-in plugins
	status := "disabled"
	if ConvenienceEnabled() {
		status = "enabled"
	}
	rows = append(rows, []string{"convenience", status, "built-in", "Wraps dpm/daml commands (build, test, clean, dev, doctor)"})

	// External plugins
	home, err := os.UserHomeDir()
	if err == nil {
		pluginDir := filepath.Join(home, ".cantool", "plugins")
		entries, err := os.ReadDir(pluginDir)
		if err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}
				info, err := entry.Info()
				if err != nil {
					continue
				}
				if info.Mode()&0111 == 0 {
					continue
				}
				rows = append(rows, []string{entry.Name(), "unknown", "external", filepath.Join(pluginDir, entry.Name())})
			}
		}
	}

	f.Info("Plugins\n")
	f.Table([]string{"NAME", "STATUS", "TYPE", "DESCRIPTION"}, rows)

	f.Info("\nSee https://github.com/DisplaceTech/Cantool#plugins for plugin management.")

	return nil
}
