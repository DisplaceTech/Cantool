package cmd

import (
	"github.com/displacetech/cantool/internal/config"
	"github.com/displacetech/cantool/internal/convenience"
	"github.com/displacetech/cantool/internal/devserver"
	"github.com/displacetech/cantool/internal/exec"
	"github.com/displacetech/cantool/internal/output"
	"github.com/displacetech/cantool/internal/sdk"
	"github.com/spf13/cobra"
)

var DevPort int

var DevCmd = &cobra.Command{
	Use:   "dev",
	Short: "Start local Canton sandbox with hot-reload",
	RunE:  runDev,
}

func init() {
	DevCmd.Flags().IntVar(&DevPort, "port", 0, "Sandbox port (default: from cantool.yaml or 5011)")
}

func runDev(cmd *cobra.Command, _ []string) error {
	f := output.New(Format())
	runner := &exec.DefaultRunner{}

	cfg, err := config.Load()
	if err != nil {
		if ce, ok := err.(*output.CantoolError); ok {
			f.Error(ce)
		}
		return err
	}

	detector := &sdk.PathDetector{Runner: runner}
	sdkInfo, err := detector.Detect(cmd.Context())
	if err != nil {
		if ce, ok := err.(*output.CantoolError); ok {
			f.Error(ce)
		}
		return err
	}

	convenience.PrintDelegation(sdkInfo.Command, "sandbox")

	srv := &devserver.DevServer{
		Config:    cfg,
		Runner:    runner,
		Formatter: f,
		Port:      DevPort,
	}

	return srv.Start(cmd.Context())
}
