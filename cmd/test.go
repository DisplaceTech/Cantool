package cmd

import (
	"fmt"
	"time"

	"github.com/displacetech/cantool/internal/config"
	"github.com/displacetech/cantool/internal/convenience"
	"github.com/displacetech/cantool/internal/exec"
	"github.com/displacetech/cantool/internal/output"
	"github.com/displacetech/cantool/internal/sdk"
	"github.com/spf13/cobra"
)

var TestVerbose bool

var TestCmd = &cobra.Command{
	Use:   "test",
	Short: "Run DAML Script tests",
	RunE:  runTest,
}

func init() {
	TestCmd.Flags().BoolVarP(&TestVerbose, "verbose", "v", false, "Show full SDK output")
}

func runTest(cmd *cobra.Command, _ []string) error {
	f := output.New(Format())
	runner := &exec.DefaultRunner{}
	ctx := cmd.Context()

	_, err := config.Load()
	if err != nil {
		if ce, ok := err.(*output.CantoolError); ok {
			f.Error(ce)
		}
		return err
	}

	detector := &sdk.PathDetector{Runner: runner}
	sdkInfo, err := detector.Detect(ctx)
	if err != nil {
		if ce, ok := err.(*output.CantoolError); ok {
			f.Error(ce)
		}
		return err
	}

	convenience.PrintDelegation(sdkInfo.Command, "test")

	result, err := sdk.Test(ctx, runner, sdkInfo, ".", TestVerbose)
	if err != nil {
		if ce, ok := err.(*output.CantoolError); ok {
			f.Error(ce)
		}
		return err
	}

	if TestVerbose {
		f.Info(result.RawOutput)
	}

	if result.Total == 0 {
		f.Info("No tests found.")
		return nil
	}

	// Show individual results
	for _, tc := range result.Tests {
		if tc.Passed {
			f.Info(fmt.Sprintf("  ✓ %-25s (%s)", tc.Name, tc.Duration.Round(time.Millisecond)))
		} else {
			f.Info(fmt.Sprintf("  ✗ %-25s (%s)", tc.Name, tc.Duration.Round(time.Millisecond)))
			if tc.Error != "" {
				f.Info(fmt.Sprintf("    Error: %s", tc.Error))
			}
		}
	}

	if result.Failed > 0 {
		f.Error(output.Errorf("CT5002", "Fix failing tests",
			"%d passed, %d failed (%s)", result.Passed, result.Failed, result.Duration.Round(time.Millisecond)))
		return fmt.Errorf("%d tests failed", result.Failed)
	}

	f.Success(fmt.Sprintf("%d tests passed (%s)", result.Passed, result.Duration.Round(time.Millisecond)),
		map[string]any{
			"passed":      result.Passed,
			"failed":      result.Failed,
			"total":       result.Total,
			"duration_ms": result.Duration.Milliseconds(),
		})
	return nil
}
