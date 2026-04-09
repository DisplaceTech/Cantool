package convenience

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var dimPrint = color.New(color.Faint).FprintfFunc()

// PrintDelegation prints a single dimmed attribution line to stderr,
// making the delegation to an external tool visible to the user.
func PrintDelegation(tool, subcommand string) {
	dimPrint(os.Stderr, "→ delegating to %s %s\n", tool, subcommand)
}

// Register adds convenience commands to root when the convenience plugin
// is enabled. Pass the commands to register; only non-nil commands are added.
func Register(root *cobra.Command, commands []*cobra.Command) {
	for _, cmd := range commands {
		if cmd != nil {
			root.AddCommand(cmd)
		}
	}
}

// CommandNames returns the names of all convenience commands.
func CommandNames() []string {
	return []string{"build", "test", "clean", "dev", "doctor"}
}

// DelegationMessage returns the formatted delegation string without printing it.
// Useful for testing.
func DelegationMessage(tool, subcommand string) string {
	return fmt.Sprintf("→ delegating to %s %s\n", tool, subcommand)
}
