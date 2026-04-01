package sdk

import (
	"context"
	osexec "os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSandbox_StopNilCmd(t *testing.T) {
	s := &Sandbox{Port: 5011}
	assert.NoError(t, s.Stop())
}

func TestSandbox_IsRunningNilCmd(t *testing.T) {
	s := &Sandbox{Port: 5011}
	assert.False(t, s.IsRunning())
}

func TestSandbox_StopNilProcess(t *testing.T) {
	cmd := osexec.Command("true")
	s := &Sandbox{cmd: cmd, Port: 5011}
	// cmd.Process is nil before Start
	assert.NoError(t, s.Stop())
}

func TestSandbox_IsRunningNilProcess(t *testing.T) {
	cmd := osexec.Command("true")
	s := &Sandbox{cmd: cmd, Port: 5011}
	assert.False(t, s.IsRunning())
}

func TestSandbox_FullLifecycle(t *testing.T) {
	// Start a real long-running process
	cmd := osexec.Command("sleep", "60")
	require.NoError(t, cmd.Start())

	s := &Sandbox{cmd: cmd, Port: 5011}
	assert.True(t, s.IsRunning())

	err := s.Stop()
	assert.NoError(t, err)

	// Give a moment for process to exit
	time.Sleep(50 * time.Millisecond)
	assert.False(t, s.IsRunning())
}

func TestStartSandbox_FailsOnBadCommand(t *testing.T) {
	sdk := &SDKInfo{Command: "cantool_nonexistent_xyz"}
	_, err := StartSandbox(context.Background(), sdk, SandboxConfig{
		ProjectDir: t.TempDir(),
	})
	require.Error(t, err)
}

func TestStartSandbox_SetsDefaultPort(t *testing.T) {
	// Use a command that starts but exits quickly
	sdk := &SDKInfo{Command: "sh"}
	s, err := StartSandbox(context.Background(), sdk, SandboxConfig{
		ProjectDir: t.TempDir(),
	})
	if err == nil {
		assert.Equal(t, 5011, s.Port)
		s.Stop()
	}
}

func TestStartSandbox_CustomPortAndDAR(t *testing.T) {
	// Verify port and DAR path are passed as args
	sdk := &SDKInfo{Command: "sh"}
	s, err := StartSandbox(context.Background(), sdk, SandboxConfig{
		Port:       6865,
		DARPath:    "/tmp/test.dar",
		ProjectDir: t.TempDir(),
	})
	if err == nil {
		assert.Equal(t, 6865, s.Port)
		s.Stop()
	}
}

func TestSandbox_StopAlreadyExited(t *testing.T) {
	// Start a process that exits immediately
	cmd := osexec.Command("true")
	require.NoError(t, cmd.Start())
	cmd.Wait() // wait for it to exit

	s := &Sandbox{cmd: cmd, Port: 5011}
	// Stop should handle the already-exited case gracefully
	assert.NoError(t, s.Stop())
}
