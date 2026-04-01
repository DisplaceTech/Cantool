package plugin

import (
	"sort"

	"github.com/displacetech/cantool/internal/output"
)

// Registry manages discovered plugin metadata.
type Registry struct {
	plugins map[string]PluginMetadata
}

// NewRegistry creates an empty plugin registry.
func NewRegistry() *Registry {
	return &Registry{plugins: make(map[string]PluginMetadata)}
}

// Register adds a plugin to the registry.
func (r *Registry) Register(meta PluginMetadata) error {
	if _, exists := r.plugins[meta.Name]; exists {
		return output.Errorf("CT8001",
			"Remove the existing plugin first",
			"plugin %q is already registered", meta.Name)
	}
	r.plugins[meta.Name] = meta
	return nil
}

// List returns all registered plugins sorted by name.
func (r *Registry) List() []PluginMetadata {
	result := make([]PluginMetadata, 0, len(r.plugins))
	for _, m := range r.plugins {
		result = append(result, m)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// Get returns a plugin by name.
func (r *Registry) Get(name string) (*PluginMetadata, bool) {
	m, ok := r.plugins[name]
	if !ok {
		return nil, false
	}
	return &m, true
}
