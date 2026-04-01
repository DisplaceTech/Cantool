package config

import (
	"errors"
	"testing"

	"github.com/displacetech/cantool/internal/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validConfig() Config {
	return Config{
		Version: "1",
		Project: ProjectConfig{Name: "my-app", SDKVersion: "3.4.11"},
		Parties: []PartyConfig{
			{Name: "Alice", Display: "Alice Corp"},
			{Name: "Bob", Display: "Bob Trading"},
		},
		Environments: map[string]EnvConfig{
			"local": {Host: "localhost", LedgerPort: 5011, JSONAPIPort: 7575},
		},
		Dev: DevConfig{WatchPaths: []string{"daml/"}, HotReload: true, SandboxPort: 5011},
	}
}

func requireCode(t *testing.T, err error, code string) {
	t.Helper()
	var ce *output.CantoolError
	require.True(t, errors.As(err, &ce), "expected CantoolError, got %T: %v", err, err)
	assert.Equal(t, code, ce.Code)
}

func TestValidConfig(t *testing.T) {
	c := validConfig()
	assert.NoError(t, c.Validate())
}

func TestValidate_MissingProjectName(t *testing.T) {
	c := validConfig()
	c.Project.Name = ""
	requireCode(t, c.Validate(), "CT1003")
}

func TestValidate_InvalidProjectName(t *testing.T) {
	tests := []string{"has space", "special!", "-leading-hyphen", ""}
	for _, name := range tests {
		c := validConfig()
		c.Project.Name = name
		err := c.Validate()
		require.Error(t, err, "expected error for name %q", name)
	}
}

func TestValidate_ValidProjectNames(t *testing.T) {
	tests := []string{"my-app", "app123", "A", "my-cool-app-2"}
	for _, name := range tests {
		c := validConfig()
		c.Project.Name = name
		assert.NoError(t, c.Validate(), "expected valid for name %q", name)
	}
}

func TestValidate_InvalidSDKVersion(t *testing.T) {
	c := validConfig()
	c.Project.SDKVersion = "not-semver"
	requireCode(t, c.Validate(), "CT1005")
}

func TestValidate_EmptySDKVersion(t *testing.T) {
	c := validConfig()
	c.Project.SDKVersion = ""
	assert.NoError(t, c.Validate())
}

func TestValidate_DuplicatePartyNames(t *testing.T) {
	c := validConfig()
	c.Parties = append(c.Parties, PartyConfig{Name: "Alice", Display: "Dup"})
	requireCode(t, c.Validate(), "CT1006")
}

func TestValidate_InvalidPort(t *testing.T) {
	c := validConfig()
	c.Environments["local"] = EnvConfig{Host: "localhost", LedgerPort: 99999}
	requireCode(t, c.Validate(), "CT1007")
}

func TestValidate_ZeroPort(t *testing.T) {
	c := validConfig()
	c.Environments["local"] = EnvConfig{Host: "localhost", LedgerPort: 0}
	assert.NoError(t, c.Validate())
}

func TestValidate_UnsupportedVersion(t *testing.T) {
	c := validConfig()
	c.Version = "2"
	requireCode(t, c.Validate(), "CT1004")
}

func TestValidate_InvalidDevSandboxPort(t *testing.T) {
	c := validConfig()
	c.Dev.SandboxPort = -1
	requireCode(t, c.Validate(), "CT1007")
}
