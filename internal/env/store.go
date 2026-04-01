package env

import (
	"os"
	"path/filepath"

	"github.com/displacetech/cantool/internal/output"
	"gopkg.in/yaml.v3"
)

const storeFile = "environments.yaml"

// Environment represents a Canton ledger endpoint.
type Environment struct {
	Name  string `yaml:"name"`
	Host  string `yaml:"host"`
	Port  int    `yaml:"port"`
	Token string `yaml:"token,omitempty"`
}

// storeData is the on-disk YAML schema.
type storeData struct {
	Active       string        `yaml:"active"`
	Environments []Environment `yaml:"environments"`
}

// EnvironmentStore manages environment profiles persisted to YAML.
type EnvironmentStore struct {
	dir  string
	data storeData

	// MalformedReset is set to true when load detected malformed YAML and
	// reset the store to defaults. Callers may use this to emit a warning.
	MalformedReset bool
}

// localEnv is the implicit default environment.
var localEnv = Environment{
	Name: "local",
	Host: "localhost",
	Port: 5011,
}

// NewStore creates an EnvironmentStore rooted at the given directory.
// It loads existing data from disk (creating the directory if needed)
// and resets to defaults if the YAML is malformed.
func NewStore(dir string) (*EnvironmentStore, error) {
	s := &EnvironmentStore{dir: dir}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

// filePath returns the full path to the environments YAML file.
func (s *EnvironmentStore) filePath() string {
	return filepath.Join(s.dir, storeFile)
}

// load reads the YAML file from disk. If the directory or file does not
// exist it initialises defaults. Malformed YAML triggers a reset with a
// warning message returned via the WarningOnLoad field.
func (s *EnvironmentStore) load() error {
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return err
	}

	raw, err := os.ReadFile(s.filePath())
	if err != nil {
		if os.IsNotExist(err) {
			s.data = storeData{Active: "local"}
			return nil
		}
		return err
	}

	var data storeData
	if err := yaml.Unmarshal(raw, &data); err != nil {
		// Malformed YAML — reset to defaults.
		s.data = storeData{Active: "local"}
		s.MalformedReset = true
		return s.save()
	}

	s.data = data
	if s.data.Active == "" {
		s.data.Active = "local"
	}
	return nil
}

// save persists the current state to disk.
func (s *EnvironmentStore) save() error {
	raw, err := yaml.Marshal(&s.data)
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath(), raw, 0o644)
}

// List returns all environments including the implicit local one.
// The local environment is always first.
func (s *EnvironmentStore) List() []Environment {
	envs := []Environment{localEnv}
	for _, e := range s.data.Environments {
		if e.Name != "local" {
			envs = append(envs, e)
		}
	}
	return envs
}

// Active returns the name of the currently active environment.
func (s *EnvironmentStore) Active() string {
	return s.data.Active
}

// Get returns the environment with the given name, or nil if not found.
func (s *EnvironmentStore) Get(name string) *Environment {
	if name == "local" {
		e := localEnv
		return &e
	}
	for i := range s.data.Environments {
		if s.data.Environments[i].Name == name {
			return &s.data.Environments[i]
		}
	}
	return nil
}

// Add creates a new environment profile and persists it.
// Returns CT9003 if the name already exists.
func (s *EnvironmentStore) Add(e Environment) *output.CantoolError {
	if s.Get(e.Name) != nil {
		return output.Errorf("CT9003",
			"choose a different name or remove the existing environment first",
			"environment %q already exists", e.Name)
	}
	s.data.Environments = append(s.data.Environments, e)
	if err := s.save(); err != nil {
		return output.Wrapf(err, "CT9000", "check filesystem permissions", "failed to save environments")
	}
	return nil
}

// Use sets the active environment. Returns an error if the name is not found.
func (s *EnvironmentStore) Use(name string) *output.CantoolError {
	if s.Get(name) == nil {
		return output.Errorf("CT9004",
			"run 'cantool env list' to see available environments",
			"environment %q not found", name)
	}
	s.data.Active = name
	if err := s.save(); err != nil {
		return output.Wrapf(err, "CT9000", "check filesystem permissions", "failed to save environments")
	}
	return nil
}

// Remove deletes an environment profile.
// Returns CT9001 if attempting to remove the active environment.
// Returns CT9002 if attempting to remove the implicit local environment.
func (s *EnvironmentStore) Remove(name string) *output.CantoolError {
	if name == "local" {
		return output.Errorf("CT9002",
			"the local environment is built-in and cannot be removed",
			"cannot remove the local environment")
	}
	if name == s.data.Active {
		return output.Errorf("CT9001",
			"switch to another environment with 'cantool env use <name>' first",
			"cannot remove the active environment %q", name)
	}

	idx := -1
	for i, e := range s.data.Environments {
		if e.Name == name {
			idx = i
			break
		}
	}
	if idx == -1 {
		return output.Errorf("CT9004",
			"run 'cantool env list' to see available environments",
			"environment %q not found", name)
	}

	s.data.Environments = append(s.data.Environments[:idx], s.data.Environments[idx+1:]...)
	if err := s.save(); err != nil {
		return output.Wrapf(err, "CT9000", "check filesystem permissions", "failed to save environments")
	}
	return nil
}
