package sdk

import (
	"context"
	"fmt"
	"os"
	osexec "os/exec"
	"syscall"
	"time"

	"github.com/displacetech/cantool/internal/output"
)

// SandboxConfig configures a local Canton sandbox.
type SandboxConfig struct {
	Port       int
	DARPath    string
	ProjectDir string
}

// Sandbox is a running Canton sandbox process.
type Sandbox struct {
	cmd  *osexec.Cmd
	Port int
}

// StartSandbox starts a Canton sandbox as a subprocess.
func StartSandbox(ctx context.Context, sdk *SDKInfo, cfg SandboxConfig) (*Sandbox, error) {
	args := []string{"sandbox"}
	if cfg.Port > 0 {
		args = append(args, "--port", fmt.Sprintf("%d", cfg.Port))
	}
	if cfg.DARPath != "" {
		args = append(args, cfg.DARPath)
	}

	cmd := osexec.CommandContext(ctx, sdk.Command, args...)
	cmd.Dir = cfg.ProjectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, output.Wrapf(err, "CT3001",
			"Check that the DAML SDK is installed and no other sandbox is running",
			"failed to start sandbox")
	}

	port := cfg.Port
	if port == 0 {
		port = 5011
	}

	return &Sandbox{cmd: cmd, Port: port}, nil
}

// Stop gracefully stops the sandbox.
func (s *Sandbox) Stop() error {
	if s.cmd == nil || s.cmd.Process == nil {
		return nil
	}
	// Send SIGTERM
	if err := s.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		// Process may have already exited
		return nil
	}
	// Wait briefly for clean exit
	done := make(chan error, 1)
	go func() { done <- s.cmd.Wait() }()
	select {
	case <-done:
		return nil
	case <-time.After(5 * time.Second):
		s.cmd.Process.Kill()
		return nil
	}
}

// IsRunning reports whether the sandbox process is still alive.
func (s *Sandbox) IsRunning() bool {
	if s.cmd == nil || s.cmd.Process == nil {
		return false
	}
	err := s.cmd.Process.Signal(syscall.Signal(0))
	return err == nil
}
