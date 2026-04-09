package cmd

import (
	"os"

	"github.com/displacetech/cantool/internal/doctor"
	"github.com/displacetech/cantool/internal/exec"
	"github.com/displacetech/cantool/internal/output"
	"github.com/spf13/cobra"
)

var DoctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check development environment prerequisites",
	Long:  "Run health checks to verify that required tools and ports are available.",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt := output.New(Format())
		runner := &exec.DefaultRunner{}
		doc := doctor.NewDoctor(runner)

		results, pass := doc.RunAll(cmd.Context())

		for _, r := range results {
			switch r.Status {
			case "ok":
				fmt.Success(r.Message, r)
			case "warn":
				fmt.Warning(r.Message)
				if r.Suggestion != "" {
					fmt.Info("  " + r.Suggestion)
				}
			case "fail":
				fmt.Error(&output.CantoolError{
					Code:       "CT2001",
					Message:    r.Message,
					Suggestion: r.Suggestion,
				})
			}
		}

		if !pass {
			os.Exit(1)
		}
		return nil
	},
}
