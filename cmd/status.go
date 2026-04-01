package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/displacetech/cantool/internal/ledger"
	"github.com/displacetech/cantool/internal/output"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show Canton node health and metadata",
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, _ []string) error {
	f := output.New(Format())
	ctx := cmd.Context()

	// Resolve active environment
	host, port, token := resolveActiveEnv()
	baseURL := fmt.Sprintf("http://%s:%d", host, port)

	var opts []ledger.ClientOption
	if token != "" {
		opts = append(opts, ledger.WithToken(token))
	}
	client := ledger.NewHTTPLedgerClient(baseURL, opts...)

	health, err := client.Health(ctx)
	if err != nil {
		if ce, ok := err.(*output.CantoolError); ok {
			f.Error(ce)
		} else {
			f.Error(output.Errorf("CT6001",
				"Run `cantool dev` to start a local sandbox, or check `cantool env list`",
				"Cannot reach Canton node at %s:%d", host, port))
		}
		return err
	}

	parties, _ := client.ListParties(ctx)
	var partyNames []string
	for _, p := range parties {
		partyNames = append(partyNames, p.DisplayName)
	}

	envName := "local"
	f.Info("Canton Node Status")
	f.Info(fmt.Sprintf("  Environment:  %s (%s:%d)", envName, host, port))
	f.Info(fmt.Sprintf("  Status:       ✓ %s", health.Status))
	f.Info(fmt.Sprintf("  Version:      %s", health.Version))
	f.Info(fmt.Sprintf("  Uptime:       %s", health.Uptime))
	if len(partyNames) > 0 {
		f.Info(fmt.Sprintf("  Parties:      %s", strings.Join(partyNames, ", ")))
	}

	return nil
}

// resolveActiveEnv reads the active environment from ~/.cantool/environments.yaml.
// Falls back to localhost:5011 if not configured.
func resolveActiveEnv() (host string, port int, token string) {
	host = "localhost"
	port = 5011

	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	data, err := os.ReadFile(filepath.Join(home, ".cantool", "environments.yaml"))
	if err != nil {
		return
	}

	var store struct {
		Active       string `yaml:"active"`
		Environments []struct {
			Name  string `yaml:"name"`
			Host  string `yaml:"host"`
			Port  int    `yaml:"port"`
			Token string `yaml:"token"`
		} `yaml:"environments"`
	}
	if err := yaml.Unmarshal(data, &store); err != nil {
		return
	}

	active := store.Active
	if active == "" {
		active = "local"
	}
	for _, env := range store.Environments {
		if env.Name == active {
			if env.Host != "" {
				host = env.Host
			}
			if env.Port > 0 {
				port = env.Port
			}
			token = env.Token
			return
		}
	}
	return
}
