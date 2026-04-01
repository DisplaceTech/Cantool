package sdk

import (
	"context"
	"strings"

	"github.com/displacetech/cantool/internal/exec"
	"github.com/displacetech/cantool/internal/output"
)

// SDKInfo describes a detected DAML SDK installation.
type SDKInfo struct {
	Path    string // full path or command name
	Command string // "dpm" or "daml"
	Version string // semver string
}

// SDKDetector finds the DAML SDK on the system.
type SDKDetector interface {
	Detect(ctx context.Context) (*SDKInfo, error)
}

// PathDetector looks for dpm/daml on PATH using a CommandRunner.
type PathDetector struct {
	Runner exec.CommandRunner
}

// Detect checks for dpm first, then daml.
func (d *PathDetector) Detect(ctx context.Context) (*SDKInfo, error) {
	// Try dpm first (newer SDK manager)
	if info, err := d.tryCommand(ctx, "dpm"); err == nil {
		return info, nil
	}
	// Fall back to daml
	if info, err := d.tryCommand(ctx, "daml"); err == nil {
		return info, nil
	}
	return nil, output.Errorf("CT2001",
		"Install the DAML SDK: https://docs.daml.com/getting-started/installation.html",
		"neither dpm nor daml found on PATH")
}

func (d *PathDetector) tryCommand(ctx context.Context, name string) (*SDKInfo, error) {
	out, err := d.Runner.Run(ctx, name, "version")
	if err != nil {
		return nil, err
	}
	if out.ExitCode != 0 {
		return nil, output.Errorf("CT2002", "", "%s version returned exit code %d", name, out.ExitCode)
	}
	version := parseVersion(out.Stdout)
	return &SDKInfo{
		Path:    name,
		Command: name,
		Version: version,
	}, nil
}

// parseVersion extracts the first semver-like string from SDK version output.
func parseVersion(s string) string {
	for _, line := range strings.Split(strings.TrimSpace(s), "\n") {
		line = strings.TrimSpace(line)
		// Look for a line containing digits and dots (semver-ish)
		for _, word := range strings.Fields(line) {
			if len(word) > 0 && word[0] >= '0' && word[0] <= '9' && strings.Contains(word, ".") {
				return word
			}
		}
	}
	return strings.TrimSpace(s)
}
