package config

import (
	"os"
	"path/filepath"

	"github.com/displacetech/cantool/internal/output"
	"github.com/spf13/viper"
)

const configFileName = "cantool.yaml"

// Load finds and loads cantool.yaml by walking up from the current directory.
func Load() (*Config, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, output.Wrapf(err, "CT1001", "Ensure you are in a valid directory",
			"cannot determine working directory")
	}
	path, err := findConfig(dir)
	if err != nil {
		return nil, err
	}
	return LoadFrom(path)
}

// LoadFrom reads and validates a cantool.yaml at the given path.
func LoadFrom(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetEnvPrefix("CANTOOL")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, output.Wrapf(err, "CT1002",
			"Check that cantool.yaml is valid YAML",
			"failed to read config %s", path)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, output.Wrapf(err, "CT1002",
			"Check that cantool.yaml matches the expected schema",
			"failed to parse config %s", path)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// findConfig walks up from dir looking for cantool.yaml.
func findConfig(dir string) (string, error) {
	for {
		candidate := filepath.Join(dir, configFileName)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", output.Errorf("CT1001",
				"Run `cantool init` to create a new project, or cd into an existing project directory",
				"cantool.yaml not found")
		}
		dir = parent
	}
}
