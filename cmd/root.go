package cmd

import (
	"fmt"
	"os"

	"github.com/displacetech/cantool/internal/config"
	"github.com/displacetech/cantool/internal/convenience"
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
	cobra.OnInitialize(registerConvenienceCommands)
}

var convenienceRegistered bool

func registerConvenienceCommands() {
	if convenienceRegistered {
		return
	}
	convenienceRegistered = true

	cfg, err := config.LoadWithGlobal()
	if err != nil || cfg == nil {
		return
	}

	if !cfg.Plugins.Convenience.Enabled {
		return
	}

	convenience.Register(rootCmd, []*cobra.Command{
		BuildCmd,
		TestCmd,
		CleanCmd,
		DevCmd,
		DoctorCmd,
	})
}

// ConvenienceEnabled reports whether the convenience plugin is currently
// enabled according to the loaded config. This is used by plugin list.
func ConvenienceEnabled() bool {
	cfg, err := config.LoadWithGlobal()
	if err != nil || cfg == nil {
		return false
	}
	return cfg.Plugins.Convenience.Enabled
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
