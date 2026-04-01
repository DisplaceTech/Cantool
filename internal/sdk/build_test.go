package sdk

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/displacetech/cantool/internal/exec"
	"github.com/displacetech/cantool/internal/output"
	"github.com/displacetech/cantool/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuild_Success(t *testing.T) {
	dir := t.TempDir()
	darDir := filepath.Join(dir, ".daml", "dist")
	require.NoError(t, os.MkdirAll(darDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(darDir, "app-0.0.1.dar"), []byte("fake"), 0644))

	mock := testutil.NewMockRunner()
	mock.Responses["daml"] = exec.Output{Stdout: "Compiling...\nDone.\n", ExitCode: 0}
	sdk := &SDKInfo{Command: "daml"}

	result, err := Build(context.Background(), mock, sdk, dir)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Contains(t, result.DARPath, "app-0.0.1.dar")
	assert.Empty(t, result.Errors)
	assert.True(t, result.Duration > 0)
}

func TestBuild_Failure(t *testing.T) {
	mock := testutil.NewMockRunner()
	mock.Responses["daml"] = exec.Output{
		Stderr:   "daml/Main.daml:12:5: error: type mismatch\n",
		ExitCode: 1,
	}
	sdk := &SDKInfo{Command: "daml"}

	result, err := Build(context.Background(), mock, sdk, t.TempDir())
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.NotEmpty(t, result.Errors)
}

func TestBuild_WithWarnings(t *testing.T) {
	mock := testutil.NewMockRunner()
	mock.Responses["daml"] = exec.Output{
		Stdout:   "Warning: unused import\nDone.\n",
		ExitCode: 0,
	}
	sdk := &SDKInfo{Command: "daml"}

	result, err := Build(context.Background(), mock, sdk, t.TempDir())
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Len(t, result.Warnings, 1)
}

func TestBuild_RunnerError(t *testing.T) {
	mock := testutil.NewMockRunner()
	mock.Errors["daml"] = errors.New("binary crashed")
	sdk := &SDKInfo{Command: "daml"}

	_, err := Build(context.Background(), mock, sdk, t.TempDir())
	var ce *output.CantoolError
	require.True(t, errors.As(err, &ce))
	assert.Equal(t, "CT4001", ce.Code)
}

func TestBuild_NoDAR(t *testing.T) {
	mock := testutil.NewMockRunner()
	mock.Responses["daml"] = exec.Output{Stdout: "Done.\n", ExitCode: 0}
	sdk := &SDKInfo{Command: "daml"}

	result, err := Build(context.Background(), mock, sdk, t.TempDir())
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Empty(t, result.DARPath)
}

func TestBuild_FailureWithGenericOutput(t *testing.T) {
	mock := testutil.NewMockRunner()
	mock.Responses["daml"] = exec.Output{
		Stderr:   "something went wrong\n",
		ExitCode: 1,
	}
	sdk := &SDKInfo{Command: "daml"}

	result, err := Build(context.Background(), mock, sdk, t.TempDir())
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.NotEmpty(t, result.Errors)
}

func TestParseErrors_Empty(t *testing.T) {
	assert.Empty(t, parseErrors(""))
}

func TestParseWarnings_Empty(t *testing.T) {
	assert.Empty(t, parseWarnings(""))
}

func TestParseWarnings_NoWarnings(t *testing.T) {
	assert.Empty(t, parseWarnings("all good\nno issues\n"))
}
