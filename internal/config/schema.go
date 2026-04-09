package config

import (
	"fmt"
	"regexp"

	"github.com/displacetech/cantool/internal/output"
)

// Config is the top-level cantool.yaml schema.
type Config struct {
	Version      string               `yaml:"version" mapstructure:"version"`
	Project      ProjectConfig        `yaml:"project" mapstructure:"project"`
	Parties      []PartyConfig        `yaml:"parties" mapstructure:"parties"`
	Environments map[string]EnvConfig `yaml:"environments" mapstructure:"environments"`
	Dev          DevConfig            `yaml:"dev" mapstructure:"dev"`
	Plugins      PluginsConfig        `yaml:"plugins" mapstructure:"plugins"`
}

// PluginsConfig holds configuration for built-in and external plugins.
type PluginsConfig struct {
	Convenience ConvenienceConfig `yaml:"convenience" mapstructure:"convenience"`
}

// ConvenienceConfig controls the built-in convenience commands plugin.
// Disabled by default (zero value of bool is false).
type ConvenienceConfig struct {
	Enabled bool `yaml:"enabled" mapstructure:"enabled"`
}

// ProjectConfig identifies the project.
type ProjectConfig struct {
	Name       string `yaml:"name" mapstructure:"name"`
	SDKVersion string `yaml:"sdk-version" mapstructure:"sdk-version"`
}

// PartyConfig defines a party to allocate.
type PartyConfig struct {
	Name    string `yaml:"name" mapstructure:"name"`
	Display string `yaml:"display" mapstructure:"display"`
}

// EnvConfig defines connection parameters for an environment.
type EnvConfig struct {
	Host        string `yaml:"host" mapstructure:"host"`
	LedgerPort  int    `yaml:"ledger-port" mapstructure:"ledger-port"`
	JSONAPIPort int    `yaml:"json-api-port" mapstructure:"json-api-port"`
}

// DevConfig controls local development behaviour.
type DevConfig struct {
	WatchPaths  []string `yaml:"watch-paths" mapstructure:"watch-paths"`
	HotReload   bool     `yaml:"hot-reload" mapstructure:"hot-reload"`
	SandboxPort int      `yaml:"sandbox-port" mapstructure:"sandbox-port"`
}

var (
	nameRe    = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]*$`)
	semverRe  = regexp.MustCompile(`^\d+\.\d+\.\d+`)
)

// Validate checks the config for structural correctness.
func (c *Config) Validate() error {
	if c.Version != "1" {
		return output.Errorf("CT1004", `Only version "1" is supported`,
			"unsupported config version %q", c.Version)
	}
	if c.Project.Name == "" {
		return output.Errorf("CT1003", "Add a project.name field to cantool.yaml",
			"project name is required")
	}
	if !nameRe.MatchString(c.Project.Name) {
		return output.Errorf("CT1003",
			"Use only letters, numbers, and hyphens (must start with letter or number)",
			"invalid project name %q", c.Project.Name)
	}
	if c.Project.SDKVersion != "" && !semverRe.MatchString(c.Project.SDKVersion) {
		return output.Errorf("CT1005", "Use a semver string like 3.4.11",
			"invalid sdk-version %q", c.Project.SDKVersion)
	}

	seen := make(map[string]bool)
	for _, p := range c.Parties {
		if seen[p.Name] {
			return output.Errorf("CT1006", fmt.Sprintf("Remove duplicate party %q", p.Name),
				"duplicate party name %q", p.Name)
		}
		seen[p.Name] = true
	}

	for name, env := range c.Environments {
		if err := validatePort(env.LedgerPort, name, "ledger-port"); err != nil {
			return err
		}
		if err := validatePort(env.JSONAPIPort, name, "json-api-port"); err != nil {
			return err
		}
	}

	if err := validatePort(c.Dev.SandboxPort, "dev", "sandbox-port"); err != nil {
		return err
	}

	return nil
}

func validatePort(port int, section, field string) error {
	if port == 0 {
		return nil // unset is ok
	}
	if port < 1 || port > 65535 {
		return output.Errorf("CT1007", "Use a port between 1 and 65535",
			"invalid %s in %s: %d", field, section, port)
	}
	return nil
}
