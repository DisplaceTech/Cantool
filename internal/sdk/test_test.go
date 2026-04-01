package sdk

import (
	"context"
	"errors"
	"testing"

	"github.com/displacetech/cantool/internal/exec"
	"github.com/displacetech/cantool/internal/output"
	"github.com/displacetech/cantool/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTest_AllPass(t *testing.T) {
	mock := testutil.NewMockRunner()
	mock.Responses["daml"] = exec.Output{
		Stdout:   "ok testCreate\nok testTransfer\n",
		ExitCode: 0,
	}
	sdk := &SDKInfo{Command: "daml"}

	result, err := Test(context.Background(), mock, sdk, t.TempDir(), false)
	require.NoError(t, err)
	assert.Equal(t, 2, result.Passed)
	assert.Equal(t, 0, result.Failed)
	assert.Equal(t, 2, result.Total)
}

func TestTest_SomeFail(t *testing.T) {
	mock := testutil.NewMockRunner()
	mock.Responses["daml"] = exec.Output{
		Stdout:   "ok testCreate\nfail testTransfer\n  assertion failed\n",
		ExitCode: 1,
	}
	sdk := &SDKInfo{Command: "daml"}

	result, err := Test(context.Background(), mock, sdk, t.TempDir(), false)
	require.NoError(t, err)
	assert.Equal(t, 1, result.Passed)
	assert.Equal(t, 1, result.Failed)
	assert.Equal(t, 2, result.Total)
}

func TestTest_NoTests(t *testing.T) {
	mock := testutil.NewMockRunner()
	mock.Responses["daml"] = exec.Output{Stdout: "", ExitCode: 0}
	sdk := &SDKInfo{Command: "daml"}

	result, err := Test(context.Background(), mock, sdk, t.TempDir(), false)
	require.NoError(t, err)
	assert.Equal(t, 0, result.Total)
}

func TestTest_RunnerError(t *testing.T) {
	mock := testutil.NewMockRunner()
	mock.Errors["daml"] = errors.New("crashed")
	sdk := &SDKInfo{Command: "daml"}

	_, err := Test(context.Background(), mock, sdk, t.TempDir(), false)
	var ce *output.CantoolError
	require.True(t, errors.As(err, &ce))
	assert.Equal(t, "CT5001", ce.Code)
}

func TestTest_Verbose(t *testing.T) {
	mock := testutil.NewMockRunner()
	mock.Responses["daml"] = exec.Output{Stdout: "ok testOne\nverbose detail\n", ExitCode: 0}
	sdk := &SDKInfo{Command: "daml"}

	result, err := Test(context.Background(), mock, sdk, t.TempDir(), true)
	require.NoError(t, err)
	assert.Contains(t, result.RawOutput, "verbose detail")
}

func TestExtractTestName(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{"ok testCreate", "testCreate"},
		{"fail testTransfer", "testTransfer"},
		{"no test here", ""},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, extractTestName(tt.line))
	}
}

func TestParseTestCases_Empty(t *testing.T) {
	assert.Empty(t, parseTestCases(""))
}
