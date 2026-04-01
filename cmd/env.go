package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/displacetech/cantool/internal/env"
	"github.com/displacetech/cantool/internal/output"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(envCmd)
	envCmd.AddCommand(envAddCmd)
	envCmd.AddCommand(envUseCmd)
	envCmd.AddCommand(envListCmd)
	envCmd.AddCommand(envRemoveCmd)

	envAddCmd.Flags().StringVar(&envAddHost, "host", "", "Ledger API host (required)")
	envAddCmd.Flags().IntVar(&envAddPort, "port", 0, "Ledger API port (required)")
	envAddCmd.Flags().StringVar(&envAddToken, "token", "", "JWT access token")
	_ = envAddCmd.MarkFlagRequired("host")
	_ = envAddCmd.MarkFlagRequired("port")
}

var (
	envAddHost  string
	envAddPort  int
	envAddToken string
)

// defaultStoreDir returns ~/.cantool for production use.
func defaultStoreDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".cantool")
	}
	return filepath.Join(home, ".cantool")
}

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Manage environment profiles",
	Long:  "Add, list, switch, and remove Canton ledger environment profiles.",
}

var envAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add a new environment profile",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		f := output.New(Format())
		store, err := env.NewStore(defaultStoreDir())
		if err != nil {
			f.Error(output.Wrapf(err, "CT9000", "check filesystem permissions", "failed to load environments"))
			return
		}
		if store.MalformedReset {
			f.Warning("environments file was malformed and has been reset to defaults")
		}

		e := env.Environment{
			Name:  args[0],
			Host:  envAddHost,
			Port:  envAddPort,
			Token: envAddToken,
		}
		if cerr := store.Add(e); cerr != nil {
			f.Error(cerr)
			return
		}
		f.Success(fmt.Sprintf("Added environment %q", args[0]), e)
	},
}

var envUseCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "Set the active environment",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		f := output.New(Format())
		store, err := env.NewStore(defaultStoreDir())
		if err != nil {
			f.Error(output.Wrapf(err, "CT9000", "check filesystem permissions", "failed to load environments"))
			return
		}
		if store.MalformedReset {
			f.Warning("environments file was malformed and has been reset to defaults")
		}

		if cerr := store.Use(args[0]); cerr != nil {
			f.Error(cerr)
			return
		}
		f.Success(fmt.Sprintf("Switched to environment %q", args[0]), nil)
	},
}

var envListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all environment profiles",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		f := output.New(Format())
		store, err := env.NewStore(defaultStoreDir())
		if err != nil {
			f.Error(output.Wrapf(err, "CT9000", "check filesystem permissions", "failed to load environments"))
			return
		}
		if store.MalformedReset {
			f.Warning("environments file was malformed and has been reset to defaults")
		}

		envs := store.List()
		active := store.Active()

		headers := []string{"", "NAME", "HOST", "PORT", "TOKEN"}
		var rows [][]string
		for _, e := range envs {
			indicator := ""
			if e.Name == active {
				indicator = "*"
			}
			token := ""
			if e.Token != "" {
				token = "***"
			}
			rows = append(rows, []string{
				indicator,
				e.Name,
				e.Host,
				fmt.Sprintf("%d", e.Port),
				token,
			})
		}
		f.Table(headers, rows)
	},
}

var envRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove an environment profile",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		f := output.New(Format())
		store, err := env.NewStore(defaultStoreDir())
		if err != nil {
			f.Error(output.Wrapf(err, "CT9000", "check filesystem permissions", "failed to load environments"))
			return
		}
		if store.MalformedReset {
			f.Warning("environments file was malformed and has been reset to defaults")
		}

		if cerr := store.Remove(args[0]); cerr != nil {
			f.Error(cerr)
			return
		}
		f.Success(fmt.Sprintf("Removed environment %q", args[0]), nil)
	},
}
