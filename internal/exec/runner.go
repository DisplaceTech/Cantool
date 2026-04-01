package exec

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
)

// Output holds the result of a command execution.
type Output struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// CommandRunner is the interface for executing external commands.
type CommandRunner interface {
	Run(ctx context.Context, name string, args ...string) (Output, error)
}

// DefaultRunner executes commands via os/exec.
type DefaultRunner struct{}

func (r *DefaultRunner) Run(ctx context.Context, name string, args ...string) (Output, error) {
	cmd := exec.CommandContext(ctx, name, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	out := Output{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}

	if err != nil {
		// If the context was cancelled or timed out, surface that.
		if ctx.Err() != nil {
			return out, ctx.Err()
		}
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			out.ExitCode = exitErr.ExitCode()
			return out, nil
		}
		return out, err
	}

	return out, nil
}
