package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var formatFlag string

var rootCmd = &cobra.Command{
	Use:   "cantool",
	Short: "Canton application development CLI",
	CompletionOptions: cobra.CompletionOptions{
		HiddenDefaultCmd: true,
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&formatFlag, "format", "human", "Output format: human, json, quiet")
}

// SetVersion sets the version string displayed by --version.
func SetVersion(v string) {
	rootCmd.Version = v
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Format returns the current output format flag value.
func Format() string {
	return formatFlag
}
