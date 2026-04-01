package env

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTempStore creates an EnvironmentStore backed by a temporary directory.
func newTempStore(t *testing.T) *EnvironmentStore {
	t.Helper()
	dir := t.TempDir()
	s, err := NewStore(dir)
	require.NoError(t, err)
	return s
}

// --- NewStore ---

func TestNewStore_CreatesDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "sub", "dir")
	s, err := NewStore(dir)
	require.NoError(t, err)
	require.NotNil(t, s)

	info, err := os.Stat(dir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestNewStore_DefaultActiveIsLocal(t *testing.T) {
	s := newTempStore(t)
	assert.Equal(t, "local", s.Active())
}

func TestNewStore_MalformedYAML_ResetsWithWarning(t *testing.T) {
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, storeFile), []byte("{{{{not yaml"), 0o644)
	require.NoError(t, err)

	s, err := NewStore(dir)
	require.NoError(t, err)
	assert.True(t, s.MalformedReset)
	assert.Equal(t, "local", s.Active())
	assert.Len(t, s.List(), 1) // only local
}

func TestNewStore_LoadsExistingData(t *testing.T) {
	dir := t.TempDir()
	yaml := `active: staging
environments:
  - name: staging
    host: staging.example.com
    port: 6865
`
	err := os.WriteFile(filepath.Join(dir, storeFile), []byte(yaml), 0o644)
	require.NoError(t, err)

	s, err := NewStore(dir)
	require.NoError(t, err)
	assert.Equal(t, "staging", s.Active())
	envs := s.List()
	assert.Len(t, envs, 2) // local + staging
}

func TestNewStore_EmptyActive_DefaultsToLocal(t *testing.T) {
	dir := t.TempDir()
	yaml := `active: ""
environments: []
`
	err := os.WriteFile(filepath.Join(dir, storeFile), []byte(yaml), 0o644)
	require.NoError(t, err)

	s, err := NewStore(dir)
	require.NoError(t, err)
	assert.Equal(t, "local", s.Active())
}

// --- List ---

func TestList_AlwaysIncludesLocal(t *testing.T) {
	s := newTempStore(t)
	envs := s.List()
	require.Len(t, envs, 1)
	assert.Equal(t, "local", envs[0].Name)
	assert.Equal(t, "localhost", envs[0].Host)
	assert.Equal(t, 5011, envs[0].Port)
}

func TestList_IncludesAddedEnvironments(t *testing.T) {
	s := newTempStore(t)
	require.Nil(t, s.Add(Environment{Name: "dev", Host: "dev.example.com", Port: 6865}))
	require.Nil(t, s.Add(Environment{Name: "staging", Host: "staging.example.com", Port: 6866}))

	envs := s.List()
	require.Len(t, envs, 3)
	assert.Equal(t, "local", envs[0].Name)
	assert.Equal(t, "dev", envs[1].Name)
	assert.Equal(t, "staging", envs[2].Name)
}

// --- Get ---

func TestGet_Local(t *testing.T) {
	s := newTempStore(t)
	e := s.Get("local")
	require.NotNil(t, e)
	assert.Equal(t, "localhost", e.Host)
	assert.Equal(t, 5011, e.Port)
}

func TestGet_Added(t *testing.T) {
	s := newTempStore(t)
	require.Nil(t, s.Add(Environment{Name: "prod", Host: "prod.example.com", Port: 6865, Token: "jwt123"}))

	e := s.Get("prod")
	require.NotNil(t, e)
	assert.Equal(t, "prod.example.com", e.Host)
	assert.Equal(t, 6865, e.Port)
	assert.Equal(t, "jwt123", e.Token)
}

func TestGet_NotFound(t *testing.T) {
	s := newTempStore(t)
	assert.Nil(t, s.Get("nonexistent"))
}

// --- Add ---

func TestAdd_Success(t *testing.T) {
	s := newTempStore(t)
	err := s.Add(Environment{Name: "staging", Host: "staging.example.com", Port: 6865})
	assert.Nil(t, err)
	assert.NotNil(t, s.Get("staging"))
}

func TestAdd_PersistsToDisk(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(dir)
	require.NoError(t, err)
	require.Nil(t, s.Add(Environment{Name: "prod", Host: "prod.example.com", Port: 6865}))

	// Reload from disk
	s2, err := NewStore(dir)
	require.NoError(t, err)
	e := s2.Get("prod")
	require.NotNil(t, e)
	assert.Equal(t, "prod.example.com", e.Host)
}

func TestAdd_DuplicateName_CT9003(t *testing.T) {
	s := newTempStore(t)
	require.Nil(t, s.Add(Environment{Name: "dup", Host: "a.com", Port: 1}))

	err := s.Add(Environment{Name: "dup", Host: "b.com", Port: 2})
	require.NotNil(t, err)
	assert.Equal(t, "CT9003", err.Code)
}

func TestAdd_DuplicateLocal_CT9003(t *testing.T) {
	s := newTempStore(t)
	err := s.Add(Environment{Name: "local", Host: "anywhere", Port: 9999})
	require.NotNil(t, err)
	assert.Equal(t, "CT9003", err.Code)
}

func TestAdd_WithToken(t *testing.T) {
	s := newTempStore(t)
	err := s.Add(Environment{Name: "secure", Host: "secure.example.com", Port: 6865, Token: "my-jwt"})
	assert.Nil(t, err)
	e := s.Get("secure")
	require.NotNil(t, e)
	assert.Equal(t, "my-jwt", e.Token)
}

// --- Use ---

func TestUse_SwitchActive(t *testing.T) {
	s := newTempStore(t)
	require.Nil(t, s.Add(Environment{Name: "staging", Host: "staging.example.com", Port: 6865}))

	err := s.Use("staging")
	assert.Nil(t, err)
	assert.Equal(t, "staging", s.Active())
}

func TestUse_Local(t *testing.T) {
	s := newTempStore(t)
	require.Nil(t, s.Add(Environment{Name: "staging", Host: "staging.example.com", Port: 6865}))
	require.Nil(t, s.Use("staging"))

	err := s.Use("local")
	assert.Nil(t, err)
	assert.Equal(t, "local", s.Active())
}

func TestUse_PersistsToDisk(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(dir)
	require.NoError(t, err)
	require.Nil(t, s.Add(Environment{Name: "staging", Host: "staging.example.com", Port: 6865}))
	require.Nil(t, s.Use("staging"))

	s2, err := NewStore(dir)
	require.NoError(t, err)
	assert.Equal(t, "staging", s2.Active())
}

func TestUse_NotFound(t *testing.T) {
	s := newTempStore(t)
	err := s.Use("nonexistent")
	require.NotNil(t, err)
	assert.Equal(t, "CT9004", err.Code)
}

// --- Remove ---

func TestRemove_Success(t *testing.T) {
	s := newTempStore(t)
	require.Nil(t, s.Add(Environment{Name: "staging", Host: "staging.example.com", Port: 6865}))

	err := s.Remove("staging")
	assert.Nil(t, err)
	assert.Nil(t, s.Get("staging"))
	assert.Len(t, s.List(), 1) // only local
}

func TestRemove_PersistsToDisk(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(dir)
	require.NoError(t, err)
	require.Nil(t, s.Add(Environment{Name: "staging", Host: "staging.example.com", Port: 6865}))
	require.Nil(t, s.Remove("staging"))

	s2, err := NewStore(dir)
	require.NoError(t, err)
	assert.Nil(t, s2.Get("staging"))
}

func TestRemove_ActiveEnv_CT9001(t *testing.T) {
	s := newTempStore(t)
	require.Nil(t, s.Add(Environment{Name: "staging", Host: "staging.example.com", Port: 6865}))
	require.Nil(t, s.Use("staging"))

	err := s.Remove("staging")
	require.NotNil(t, err)
	assert.Equal(t, "CT9001", err.Code)
	// Environment should still exist
	assert.NotNil(t, s.Get("staging"))
}

func TestRemove_Local_CT9002(t *testing.T) {
	s := newTempStore(t)
	err := s.Remove("local")
	require.NotNil(t, err)
	assert.Equal(t, "CT9002", err.Code)
}

func TestRemove_NotFound(t *testing.T) {
	s := newTempStore(t)
	err := s.Remove("nonexistent")
	require.NotNil(t, err)
	assert.Equal(t, "CT9004", err.Code)
}

// --- Persistence round-trip ---

func TestRoundTrip_FullWorkflow(t *testing.T) {
	dir := t.TempDir()

	// Create and populate
	s, err := NewStore(dir)
	require.NoError(t, err)
	require.Nil(t, s.Add(Environment{Name: "dev", Host: "dev.example.com", Port: 6865}))
	require.Nil(t, s.Add(Environment{Name: "staging", Host: "staging.example.com", Port: 6866, Token: "tok"}))
	require.Nil(t, s.Use("dev"))

	// Reload and verify
	s2, err := NewStore(dir)
	require.NoError(t, err)
	assert.Equal(t, "dev", s2.Active())
	envs := s2.List()
	require.Len(t, envs, 3)
	assert.Equal(t, "local", envs[0].Name)
	assert.Equal(t, "dev", envs[1].Name)
	assert.Equal(t, "staging", envs[2].Name)
	assert.Equal(t, "tok", s2.Get("staging").Token)

	// Remove and reload
	require.Nil(t, s2.Use("local"))
	require.Nil(t, s2.Remove("dev"))
	s3, err := NewStore(dir)
	require.NoError(t, err)
	assert.Equal(t, "local", s3.Active())
	assert.Len(t, s3.List(), 2) // local + staging
	assert.Nil(t, s3.Get("dev"))
}

// --- Token omitted from YAML when empty ---

func TestAdd_TokenOmittedWhenEmpty(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(dir)
	require.NoError(t, err)
	require.Nil(t, s.Add(Environment{Name: "plain", Host: "host.com", Port: 1234}))

	raw, err := os.ReadFile(filepath.Join(dir, storeFile))
	require.NoError(t, err)
	assert.NotContains(t, string(raw), "token")
}
