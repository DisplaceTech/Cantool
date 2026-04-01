package sdk

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	"github.com/displacetech/cantool/internal/exec"
	"github.com/displacetech/cantool/internal/output"
)

// BuildResult holds the outcome of a DAML build.
type BuildResult struct {
	Success  bool
	DARPath  string
	Errors   []string
	Warnings []string
	Duration time.Duration
}

// Build compiles DAML sources using the detected SDK.
func Build(ctx context.Context, runner exec.CommandRunner, sdk *SDKInfo, projectDir string) (*BuildResult, error) {
	start := time.Now()

	out, err := runner.Run(ctx, sdk.Command, "build")
	if err != nil {
		return nil, output.Wrapf(err, "CT4001",
			"Ensure the DAML SDK is installed and working",
			"build command failed to execute")
	}

	result := &BuildResult{
		Duration: time.Since(start),
	}

	if out.ExitCode != 0 {
		result.Success = false
		result.Errors = parseErrors(out.Stderr + out.Stdout)
		return result, nil
	}

	result.Success = true
	result.Warnings = parseWarnings(out.Stderr + out.Stdout)
	result.DARPath = findDAR(projectDir)
	return result, nil
}

func parseErrors(s string) []string {
	var errs []string
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line != "" && (strings.Contains(line, "error") || strings.Contains(line, "Error")) {
			errs = append(errs, line)
		}
	}
	if len(errs) == 0 && strings.TrimSpace(s) != "" {
		errs = append(errs, strings.TrimSpace(s))
	}
	return errs
}

func parseWarnings(s string) []string {
	var warns []string
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "warning") || strings.Contains(line, "Warning") {
			warns = append(warns, line)
		}
	}
	return warns
}

func findDAR(projectDir string) string {
	matches, _ := filepath.Glob(filepath.Join(projectDir, ".daml", "dist", "*.dar"))
	if len(matches) > 0 {
		return matches[0]
	}
	return ""
}
