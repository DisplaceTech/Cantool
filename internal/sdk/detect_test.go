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

func TestPathDetector_PrefersDpm(t *testing.T) {
	mock := testutil.NewMockRunner()
	mock.Responses["dpm"] = exec.Output{Stdout: "dpm version 3.4.11\n"}
	mock.Responses["daml"] = exec.Output{Stdout: "daml version 3.4.10\n"}
	d := &PathDetector{Runner: mock}

	info, err := d.Detect(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "dpm", info.Command)
	assert.Equal(t, "3.4.11", info.Version)
}

func TestPathDetector_FallbackToDaml(t *testing.T) {
	mock := testutil.NewMockRunner()
	mock.Errors["dpm"] = errors.New("not found")
	mock.Responses["daml"] = exec.Output{Stdout: "SDK 3.4.10\n"}
	d := &PathDetector{Runner: mock}

	info, err := d.Detect(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "daml", info.Command)
	assert.Equal(t, "3.4.10", info.Version)
}

func TestPathDetector_NeitherFound(t *testing.T) {
	mock := testutil.NewMockRunner()
	mock.Errors["dpm"] = errors.New("not found")
	mock.Errors["daml"] = errors.New("not found")
	d := &PathDetector{Runner: mock}

	_, err := d.Detect(context.Background())
	var ce *output.CantoolError
	require.True(t, errors.As(err, &ce))
	assert.Equal(t, "CT2001", ce.Code)
}

func TestPathDetector_NonZeroExitCode(t *testing.T) {
	mock := testutil.NewMockRunner()
	mock.Errors["dpm"] = errors.New("not found")
	mock.Responses["daml"] = exec.Output{Stdout: "", ExitCode: 1}
	d := &PathDetector{Runner: mock}

	_, err := d.Detect(context.Background())
	require.Error(t, err)
}

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple", "3.4.11", "3.4.11"},
		{"prefixed", "SDK version 3.4.11", "3.4.11"},
		{"multiline", "DAML SDK\nVersion: 3.4.11\n", "3.4.11"},
		{"no version", "no version here", "no version here"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, parseVersion(tt.input))
		})
	}
}
