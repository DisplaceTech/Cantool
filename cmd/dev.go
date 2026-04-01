package cmd

import (
	"github.com/displacetech/cantool/internal/config"
	"github.com/displacetech/cantool/internal/devserver"
	"github.com/displacetech/cantool/internal/exec"
	"github.com/displacetech/cantool/internal/output"
	"github.com/spf13/cobra"
)

var devPort int

var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "Start local Canton sandbox with hot-reload",
	RunE:  runDev,
}

func init() {
	devCmd.Flags().IntVar(&devPort, "port", 0, "Sandbox port (default: from cantool.yaml or 5011)")
	rootCmd.AddCommand(devCmd)
}

func runDev(cmd *cobra.Command, _ []string) error {
	f := output.New(Format())

	cfg, err := config.Load()
	if err != nil {
		if ce, ok := err.(*output.CantoolError); ok {
			f.Error(ce)
		}
		return err
	}

	srv := &devserver.DevServer{
		Config:    cfg,
		Runner:    &exec.DefaultRunner{},
		Formatter: f,
		Port:      devPort,
	}

	return srv.Start(cmd.Context())
}
