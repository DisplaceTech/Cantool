package convenience

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDelegationMessage_Format(t *testing.T) {
	tests := []struct {
		tool       string
		subcommand string
		expected   string
	}{
		{"dpm", "build", "→ delegating to dpm build\n"},
		{"daml", "build", "→ delegating to daml build\n"},
		{"dpm", "test", "→ delegating to dpm test\n"},
		{"dpm", "sandbox", "→ delegating to dpm sandbox\n"},
	}
	for _, tt := range tests {
		t.Run(tt.tool+"_"+tt.subcommand, func(t *testing.T) {
			assert.Equal(t, tt.expected, DelegationMessage(tt.tool, tt.subcommand))
		})
	}
}

func TestRegister_AddsCommands(t *testing.T) {
	root := &cobra.Command{Use: "test-root"}
	cmd1 := &cobra.Command{Use: "build", Short: "Build things"}
	cmd2 := &cobra.Command{Use: "test", Short: "Test things"}

	Register(root, []*cobra.Command{cmd1, cmd2})

	found := root.Commands()
	require.Len(t, found, 2)
	assert.Equal(t, "build", found[0].Name())
	assert.Equal(t, "test", found[1].Name())
}

func TestRegister_SkipsNilCommands(t *testing.T) {
	root := &cobra.Command{Use: "test-root"}
	cmd1 := &cobra.Command{Use: "build", Short: "Build things"}

	Register(root, []*cobra.Command{nil, cmd1, nil})

	found := root.Commands()
	require.Len(t, found, 1)
	assert.Equal(t, "build", found[0].Name())
}

func TestRegister_EmptySlice(t *testing.T) {
	root := &cobra.Command{Use: "test-root"}
	Register(root, []*cobra.Command{})
	assert.Empty(t, root.Commands())
}

func TestCommandNames(t *testing.T) {
	names := CommandNames()
	assert.Equal(t, []string{"build", "test", "clean", "dev", "doctor"}, names)
}
