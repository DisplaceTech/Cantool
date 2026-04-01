package exec

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultRunner(t *testing.T) {
	runner := &DefaultRunner{}
	ctx := context.Background()

	tests := []struct {
		name       string
		cmd        string
		args       []string
		wantStdout string
		wantExit   int
		wantErr    bool
	}{
		{
			name:       "captures stdout",
			cmd:        "echo",
			args:       []string{"hello"},
			wantStdout: "hello\n",
			wantExit:   0,
		},
		{
			name:     "non-zero exit code",
			cmd:      "sh",
			args:     []string{"-c", "exit 42"},
			wantExit: 42,
		},
		{
			name:    "command not found",
			cmd:     "cantool_nonexistent_binary_xyz",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := runner.Run(ctx, tt.cmd, tt.args...)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantExit, out.ExitCode)
			if tt.wantStdout != "" {
				assert.Equal(t, tt.wantStdout, out.Stdout)
			}
		})
	}
}

func TestDefaultRunner_Stderr(t *testing.T) {
	runner := &DefaultRunner{}
	ctx := context.Background()

	out, err := runner.Run(ctx, "sh", "-c", "echo oops >&2")
	require.NoError(t, err)
	assert.Equal(t, "oops\n", out.Stderr)
	assert.Equal(t, "", out.Stdout)
}

func TestDefaultRunner_ContextCancellation(t *testing.T) {
	runner := &DefaultRunner{}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := runner.Run(ctx, "sleep", "10")
	require.Error(t, err)
}
