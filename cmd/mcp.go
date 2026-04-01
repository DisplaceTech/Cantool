package cmd

import (
	"fmt"

	"github.com/displacetech/cantool/internal/config"
	"github.com/displacetech/cantool/internal/ledger"
	"github.com/displacetech/cantool/internal/mcp"
	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "MCP server for AI assistants",
}

var mcpServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start MCP server",
	RunE:  runMCPServe,
}

func init() {
	mcpServeCmd.Flags().Bool("stdio", true, "Use stdin/stdout transport (default)")
	mcpCmd.AddCommand(mcpServeCmd)
	rootCmd.AddCommand(mcpCmd)
}

func runMCPServe(cmd *cobra.Command, _ []string) error {
	// Load config (optional — MCP can work without project context)
	cfg, _ := config.Load()
	if cfg == nil {
		cfg = &config.Config{Version: "1"}
	}

	// Resolve environment
	host, port, token := resolveActiveEnv()
	baseURL := fmt.Sprintf("http://%s:%d", host, port)

	var opts []ledger.ClientOption
	if token != "" {
		opts = append(opts, ledger.WithToken(token))
	}
	client := ledger.NewHTTPLedgerClient(baseURL, opts...)

	srv := mcp.NewServer(client, cfg)
	return srv.ServeStdio(cmd.Context())
}
