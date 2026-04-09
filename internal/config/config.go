package config

import (
	"os"
	"path/filepath"

	"github.com/displacetech/cantool/internal/output"
	"github.com/spf13/viper"
)

const (
	configFileName       = "cantool.yaml"
	globalConfigFileName = "config.yaml"
)

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

// LoadWithGlobal loads config with global fallback. It first loads the global
// config file (if any), then loads the project config. Project values override
// global values. If no project config exists but a global config does, the
// global config is returned. If neither exists, returns nil and no error.
func LoadWithGlobal() (*Config, error) {
	global, err := LoadGlobal()
	if err != nil {
		return nil, err
	}

	dir, cwdErr := os.Getwd()
	if cwdErr != nil {
		if global != nil {
			return global, nil
		}
		return nil, nil
	}

	projectPath, findErr := findConfig(dir)
	if findErr != nil {
		if global != nil {
			return global, nil
		}
		return nil, nil
	}

	project, projectErr := loadFromWithKeys(projectPath)
	if projectErr != nil {
		return nil, projectErr
	}

	if global == nil {
		return project.config, nil
	}

	return mergeConfig(global, project), nil
}

// LoadGlobal reads the user-level config file. Search order:
//  1. $XDG_CONFIG_HOME/cantool/config.yaml (default: ~/.config/cantool/)
//  2. os.UserConfigDir()/cantool/config.yaml (macOS: ~/Library/Application Support/cantool/)
//  3. ~/.cantool/config.yaml
//
// Returns nil with no error if no config file is found.
func LoadGlobal() (*Config, error) {
	return loadGlobal(os.UserHomeDir)
}

func loadGlobal(homeFn func() (string, error)) (*Config, error) {
	home, err := homeFn()
	if err != nil {
		return nil, nil
	}

	candidates := globalConfigPaths(home, os.UserConfigDir)

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return loadGlobalFrom(path)
		}
	}

	return nil, nil
}

// globalConfigPaths returns the ordered candidate paths for global config:
//  1. $XDG_CONFIG_HOME/cantool/config.yaml (falls back to $HOME/.config/cantool/)
//  2. os.UserConfigDir()/cantool/config.yaml (macOS: ~/Library/Application Support/)
//  3. $HOME/.cantool/config.yaml (legacy dotfile fallback)
//
// Duplicate paths are suppressed so each location is checked at most once.
func globalConfigPaths(home string, configDirFn func() (string, error)) []string {
	var xdgBase string
	if v := os.Getenv("XDG_CONFIG_HOME"); v != "" && filepath.IsAbs(v) {
		xdgBase = v
	} else {
		xdgBase = filepath.Join(home, ".config")
	}
	xdgPath := filepath.Join(xdgBase, "cantool", globalConfigFileName)

	seen := map[string]bool{xdgPath: true}
	paths := []string{xdgPath}

	if configDirFn != nil {
		if configDir, err := configDirFn(); err == nil {
			nativePath := filepath.Join(configDir, "cantool", globalConfigFileName)
			if !seen[nativePath] {
				seen[nativePath] = true
				paths = append(paths, nativePath)
			}
		}
	}

	dotfilePath := filepath.Join(home, ".cantool", globalConfigFileName)
	if !seen[dotfilePath] {
		paths = append(paths, dotfilePath)
	}

	return paths
}

func loadGlobalFrom(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetEnvPrefix("CANTOOL")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, output.Wrapf(err, "CT1002",
			"Check that the global config is valid YAML",
			"failed to read global config %s", path)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, output.Wrapf(err, "CT1002",
			"Check that the global config matches the expected schema",
			"failed to parse global config %s", path)
	}

	if err := cfg.ValidateGlobal(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// loadedConfig pairs a parsed Config with the set of keys that were
// explicitly present in the YAML file, so merge logic can distinguish
// "not set" from "explicitly set to zero value."
type loadedConfig struct {
	config *Config
	setKeys map[string]bool
}

func loadFromWithKeys(path string) (*loadedConfig, error) {
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

	keys := make(map[string]bool)
	for _, k := range v.AllKeys() {
		if v.IsSet(k) {
			keys[k] = true
		}
	}

	return &loadedConfig{config: &cfg, setKeys: keys}, nil
}

// mergeConfig returns a new Config with project values taking precedence.
// Global config supplies defaults for plugin settings only when the project
// config does not explicitly set them.
func mergeConfig(global *Config, project *loadedConfig) *Config {
	merged := *project.config

	if !project.setKeys["plugins.convenience.enabled"] {
		merged.Plugins.Convenience.Enabled = global.Plugins.Convenience.Enabled
	}

	return &merged
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
