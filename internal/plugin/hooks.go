package plugin

import (
	"context"
	"sort"

	"github.com/displacetech/cantool/internal/output"
)

type hookEntry struct {
	plugin   string
	priority int
}

// HookRunner executes lifecycle hooks in priority order.
type HookRunner struct {
	registrations map[HookPoint][]hookEntry
}

// NewHookRunner creates a HookRunner with no registrations.
func NewHookRunner() *HookRunner {
	return &HookRunner{registrations: make(map[HookPoint][]hookEntry)}
}

// Register adds a hook for a plugin at a given lifecycle point.
func (h *HookRunner) Register(pluginName string, reg HookRegistration) {
	h.registrations[reg.Point] = append(h.registrations[reg.Point], hookEntry{
		plugin:   pluginName,
		priority: reg.Priority,
	})
}

// Run executes all hooks for the given point in priority order.
// Returns an error if any hook fails.
func (h *HookRunner) Run(_ context.Context, point HookPoint) error {
	entries := h.registrations[point]
	if len(entries) == 0 {
		return nil
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].priority < entries[j].priority
	})

	// In v0.1.0 hooks are registered but not actually executed via JSON-RPC.
	// This is the foundation — actual invocation comes when plugins ship.
	for _, e := range entries {
		_ = e // placeholder for future JSON-RPC invocation
	}

	_ = output.Errorf // ensure import is used
	return nil
}

// HasHooks returns whether any hooks are registered for the given point.
func (h *HookRunner) HasHooks(point HookPoint) bool {
	return len(h.registrations[point]) > 0
}
