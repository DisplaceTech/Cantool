package plugin

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHookRunner_NoHooks(t *testing.T) {
	h := NewHookRunner()
	err := h.Run(context.Background(), HookPreBuild)
	require.NoError(t, err)
}

func TestHookRunner_HasHooks(t *testing.T) {
	h := NewHookRunner()
	assert.False(t, h.HasHooks(HookPreBuild))

	h.Register("test-plugin", HookRegistration{Point: HookPreBuild, Priority: 10})
	assert.True(t, h.HasHooks(HookPreBuild))
	assert.False(t, h.HasHooks(HookPostBuild))
}

func TestHookRunner_Register(t *testing.T) {
	h := NewHookRunner()
	h.Register("alpha", HookRegistration{Point: HookPreBuild, Priority: 20})
	h.Register("beta", HookRegistration{Point: HookPreBuild, Priority: 10})

	err := h.Run(context.Background(), HookPreBuild)
	require.NoError(t, err)
}

func TestHookRunner_MultiplePoints(t *testing.T) {
	h := NewHookRunner()
	h.Register("p1", HookRegistration{Point: HookPreBuild, Priority: 1})
	h.Register("p2", HookRegistration{Point: HookPostBuild, Priority: 1})

	assert.True(t, h.HasHooks(HookPreBuild))
	assert.True(t, h.HasHooks(HookPostBuild))
	assert.False(t, h.HasHooks(HookPreTest))
}
