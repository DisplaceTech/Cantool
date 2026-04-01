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

	home, err := os.UserHomeDir()
	if err != nil {
		f.Info("No plugins installed.")
		f.Info("\nInstall plugins with: cantool plugin install <name>")
		return nil
	}

	pluginDir := filepath.Join(home, ".cantool", "plugins")
	entries, err := os.ReadDir(pluginDir)
	if err != nil || len(entries) == 0 {
		f.Info("No plugins installed.")
		f.Info("\nInstall plugins with: cantool plugin install <name>")
		return nil
	}

	var rows [][]string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		// Only list executable files
		if info.Mode()&0111 == 0 {
			continue
		}
		rows = append(rows, []string{entry.Name(), "unknown", "local", filepath.Join(pluginDir, entry.Name())})
	}

	if len(rows) == 0 {
		f.Info("No plugins installed.")
		f.Info("\nInstall plugins with: cantool plugin install <name>")
		return nil
	}

	f.Info("Installed Plugins\n")
	f.Table([]string{"NAME", "VERSION", "AUTHOR", "SOURCE"}, rows)
	return nil
}
