package doctor

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"

	"github.com/displacetech/cantool/internal/exec"
)

// CheckResult holds the outcome of a single prerequisite check.
type CheckResult struct {
	Name       string `json:"name"`
	Status     string `json:"status"` // "ok", "warn", "fail"
	Version    string `json:"version,omitempty"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
}

// Check is the interface for a single prerequisite check.
type Check interface {
	Run(ctx context.Context) CheckResult
}

// --- Java Check ---

// JavaCheck verifies that Java 21+ is installed.
type JavaCheck struct {
	Runner exec.CommandRunner
}

func (c *JavaCheck) Run(ctx context.Context) CheckResult {
	out, err := c.Runner.Run(ctx, "java", "-version")
	if err != nil {
		return CheckResult{
			Name:       "Java",
			Status:     "fail",
			Message:    "Java not found",
			Suggestion: "Install Java 21+: https://adoptium.net",
		}
	}

	// java -version outputs to stderr
	combined := out.Stdout + out.Stderr
	version := parseJavaVersion(combined)
	if version == "" {
		return CheckResult{
			Name:       "Java",
			Status:     "fail",
			Message:    "Could not determine Java version",
			Suggestion: "Install Java 21+: https://adoptium.net",
		}
	}

	major := parseMajorVersion(version)
	if major < 21 {
		return CheckResult{
			Name:       "Java",
			Status:     "fail",
			Version:    version,
			Message:    fmt.Sprintf("Java %s found, but 21+ is required", version),
			Suggestion: "Install Java 21+: https://adoptium.net",
		}
	}

	return CheckResult{
		Name:    "Java",
		Status:  "ok",
		Version: version,
		Message: fmt.Sprintf("Java %s", version),
	}
}

// parseJavaVersion extracts the version string from java -version output.
// Handles formats like:
//
//	openjdk version "21.0.1" ...
//	java version "1.8.0_362" ...
var javaVersionRe = regexp.MustCompile(`(?:java|openjdk) version "([^"]+)"`)

func parseJavaVersion(output string) string {
	m := javaVersionRe.FindStringSubmatch(output)
	if len(m) < 2 {
		return ""
	}
	return m[1]
}

// parseMajorVersion extracts the major version number.
// For "21.0.1" returns 21, for "1.8.0_362" returns 8.
func parseMajorVersion(version string) int {
	parts := strings.SplitN(version, ".", 3)
	if len(parts) == 0 {
		return 0
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0
	}
	// Old-style versioning: 1.8.x means Java 8
	if major == 1 && len(parts) > 1 {
		minor, err := strconv.Atoi(parts[1])
		if err != nil {
			return 0
		}
		return minor
	}
	return major
}

// --- DAML SDK Check ---

// DAMLCheck verifies that the DAML SDK is installed (daml or dpm).
type DAMLCheck struct {
	Runner exec.CommandRunner
}

func (c *DAMLCheck) Run(ctx context.Context) CheckResult {
	// Try `daml version` first
	out, err := c.Runner.Run(ctx, "daml", "version")
	if err == nil && out.ExitCode == 0 {
		version := strings.TrimSpace(out.Stdout + out.Stderr)
		return CheckResult{
			Name:    "DAML SDK",
			Status:  "ok",
			Version: firstLine(version),
			Message: "DAML SDK found",
		}
	}

	// Fall back to `dpm version`
	out, err = c.Runner.Run(ctx, "dpm", "version")
	if err == nil && out.ExitCode == 0 {
		version := strings.TrimSpace(out.Stdout + out.Stderr)
		return CheckResult{
			Name:    "DAML SDK",
			Status:  "ok",
			Version: firstLine(version),
			Message: "DAML SDK found (via dpm)",
		}
	}

	return CheckResult{
		Name:       "DAML SDK",
		Status:     "fail",
		Message:    "DAML SDK not found",
		Suggestion: "Install DAML SDK: https://docs.daml.com/getting-started/installation.html",
	}
}

// --- Docker Check ---

// DockerCheck verifies that Docker is available (optional).
type DockerCheck struct {
	Runner exec.CommandRunner
}

func (c *DockerCheck) Run(ctx context.Context) CheckResult {
	out, err := c.Runner.Run(ctx, "docker", "version")
	if err != nil || out.ExitCode != 0 {
		return CheckResult{
			Name:       "Docker",
			Status:     "warn",
			Message:    "Docker not found",
			Suggestion: "Install Docker for multi-node dev: https://docker.com",
		}
	}

	return CheckResult{
		Name:    "Docker",
		Status:  "ok",
		Message: "Docker available",
	}
}

// --- Port Check ---

// PortCheck verifies that a TCP port is available.
type PortCheck struct {
	Port int
}

func (c *PortCheck) Run(_ context.Context) CheckResult {
	name := fmt.Sprintf("Port %d", c.Port)
	addr := fmt.Sprintf(":%d", c.Port)

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return CheckResult{
			Name:       name,
			Status:     "warn",
			Message:    fmt.Sprintf("Port %d is in use", c.Port),
			Suggestion: fmt.Sprintf("Port %d is in use", c.Port),
		}
	}
	_ = ln.Close()

	return CheckResult{
		Name:    name,
		Status:  "ok",
		Message: fmt.Sprintf("Port %d is available", c.Port),
	}
}

// --- Go Check ---

// GoCheck verifies that Go is installed (optional).
type GoCheck struct {
	Runner exec.CommandRunner
}

func (c *GoCheck) Run(ctx context.Context) CheckResult {
	out, err := c.Runner.Run(ctx, "go", "version")
	if err != nil || out.ExitCode != 0 {
		return CheckResult{
			Name:       "Go",
			Status:     "warn",
			Message:    "Go not found",
			Suggestion: "Install Go for plugin development: https://go.dev",
		}
	}

	version := strings.TrimSpace(out.Stdout)
	return CheckResult{
		Name:    "Go",
		Status:  "ok",
		Version: firstLine(version),
		Message: "Go available",
	}
}

// firstLine returns the first non-empty line from the string.
func firstLine(s string) string {
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	return s
}
