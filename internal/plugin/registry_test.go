package plugin

import (
	"errors"
	"testing"

	"github.com/displacetech/cantool/internal/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistry_RegisterAndList(t *testing.T) {
	r := NewRegistry()
	require.NoError(t, r.Register(PluginMetadata{Name: "bravo", Version: "1.0"}))
	require.NoError(t, r.Register(PluginMetadata{Name: "alpha", Version: "2.0"}))

	list := r.List()
	require.Len(t, list, 2)
	assert.Equal(t, "alpha", list[0].Name) // sorted
	assert.Equal(t, "bravo", list[1].Name)
}

func TestRegistry_DuplicateName(t *testing.T) {
	r := NewRegistry()
	require.NoError(t, r.Register(PluginMetadata{Name: "foo"}))
	err := r.Register(PluginMetadata{Name: "foo"})
	var ce *output.CantoolError
	require.True(t, errors.As(err, &ce))
	assert.Equal(t, "CT8001", ce.Code)
}

func TestRegistry_Get(t *testing.T) {
	r := NewRegistry()
	r.Register(PluginMetadata{Name: "foo", Version: "1.0"})

	m, ok := r.Get("foo")
	require.True(t, ok)
	assert.Equal(t, "1.0", m.Version)

	_, ok = r.Get("bar")
	assert.False(t, ok)
}

func TestRegistry_EmptyList(t *testing.T) {
	r := NewRegistry()
	assert.Empty(t, r.List())
}
