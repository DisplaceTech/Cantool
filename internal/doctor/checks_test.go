package doctor

import (
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/displacetech/cantool/internal/exec"
	"github.com/displacetech/cantool/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJavaCheck_OK(t *testing.T) {
	runner := testutil.NewMockRunner()
	runner.Responses["java"] = exec.Output{
		Stderr: `openjdk version "21.0.1" 2023-10-17`,
	}

	check := &JavaCheck{Runner: runner}
	result := check.Run(context.Background())

	assert.Equal(t, "Java", result.Name)
	assert.Equal(t, "ok", result.Status)
	assert.Equal(t, "21.0.1", result.Version)
	assert.Contains(t, result.Message, "Java 21.0.1")
	assert.Empty(t, result.Suggestion)
}

func TestJavaCheck_OldVersion(t *testing.T) {
	runner := testutil.NewMockRunner()
	runner.Responses["java"] = exec.Output{
		Stderr: `java version "1.8.0_362"`,
	}

	check := &JavaCheck{Runner: runner}
	result := check.Run(context.Background())

	assert.Equal(t, "Java", result.Name)
	assert.Equal(t, "fail", result.Status)
	assert.Equal(t, "1.8.0_362", result.Version)
	assert.Contains(t, result.Message, "21+ is required")
	assert.NotEmpty(t, result.Suggestion)
}

func TestJavaCheck_Version17(t *testing.T) {
	runner := testutil.NewMockRunner()
	runner.Responses["java"] = exec.Output{
		Stderr: `openjdk version "17.0.5" 2022-10-18`,
	}

	check := &JavaCheck{Runner: runner}
	result := check.Run(context.Background())

	assert.Equal(t, "fail", result.Status)
	assert.Equal(t, "17.0.5", result.Version)
}

func TestJavaCheck_NotFound(t *testing.T) {
	runner := testutil.NewMockRunner()
	runner.Errors["java"] = fmt.Errorf("exec: \"java\": executable file not found in $PATH")

	check := &JavaCheck{Runner: runner}
	result := check.Run(context.Background())

	assert.Equal(t, "Java", result.Name)
	assert.Equal(t, "fail", result.Status)
	assert.Equal(t, "Java not found", result.Message)
	assert.Contains(t, result.Suggestion, "adoptium.net")
}

func TestJavaCheck_UnparseableVersion(t *testing.T) {
	runner := testutil.NewMockRunner()
	runner.Responses["java"] = exec.Output{
		Stderr: "some unexpected output",
	}

	check := &JavaCheck{Runner: runner}
	result := check.Run(context.Background())

	assert.Equal(t, "fail", result.Status)
	assert.Equal(t, "Could not determine Java version", result.Message)
}

func TestDAMLCheck_DamlFound(t *testing.T) {
	runner := testutil.NewMockRunner()
	runner.Responses["daml"] = exec.Output{
		Stdout: "2.8.0",
	}

	check := &DAMLCheck{Runner: runner}
	result := check.Run(context.Background())

	assert.Equal(t, "DAML SDK", result.Name)
	assert.Equal(t, "ok", result.Status)
	assert.Equal(t, "2.8.0", result.Version)
	assert.Equal(t, "DAML SDK found", result.Message)
}

func TestDAMLCheck_DpmFallback(t *testing.T) {
	runner := testutil.NewMockRunner()
	runner.Errors["daml"] = fmt.Errorf("not found")
	runner.Responses["dpm"] = exec.Output{
		Stdout: "1.0.0",
	}

	check := &DAMLCheck{Runner: runner}
	result := check.Run(context.Background())

	assert.Equal(t, "DAML SDK", result.Name)
	assert.Equal(t, "ok", result.Status)
	assert.Equal(t, "1.0.0", result.Version)
	assert.Contains(t, result.Message, "via dpm")
}

func TestDAMLCheck_NotFound(t *testing.T) {
	runner := testutil.NewMockRunner()
	runner.Errors["daml"] = fmt.Errorf("not found")
	runner.Errors["dpm"] = fmt.Errorf("not found")

	check := &DAMLCheck{Runner: runner}
	result := check.Run(context.Background())

	assert.Equal(t, "DAML SDK", result.Name)
	assert.Equal(t, "fail", result.Status)
	assert.Contains(t, result.Suggestion, "docs.daml.com")
}

func TestDAMLCheck_DamlNonZeroExit(t *testing.T) {
	runner := testutil.NewMockRunner()
	runner.Responses["daml"] = exec.Output{ExitCode: 1}
	runner.Responses["dpm"] = exec.Output{
		Stdout: "1.2.3",
	}

	check := &DAMLCheck{Runner: runner}
	result := check.Run(context.Background())

	assert.Equal(t, "ok", result.Status)
	assert.Contains(t, result.Message, "via dpm")
}

func TestDockerCheck_OK(t *testing.T) {
	runner := testutil.NewMockRunner()
	runner.Responses["docker"] = exec.Output{
		Stdout: "Docker version 24.0.7",
	}

	check := &DockerCheck{Runner: runner}
	result := check.Run(context.Background())

	assert.Equal(t, "Docker", result.Name)
	assert.Equal(t, "ok", result.Status)
	assert.Equal(t, "Docker available", result.Message)
}

func TestDockerCheck_NotFound(t *testing.T) {
	runner := testutil.NewMockRunner()
	runner.Errors["docker"] = fmt.Errorf("not found")

	check := &DockerCheck{Runner: runner}
	result := check.Run(context.Background())

	assert.Equal(t, "Docker", result.Name)
	assert.Equal(t, "warn", result.Status)
	assert.Contains(t, result.Suggestion, "docker.com")
}

func TestDockerCheck_NonZeroExit(t *testing.T) {
	runner := testutil.NewMockRunner()
	runner.Responses["docker"] = exec.Output{ExitCode: 1}

	check := &DockerCheck{Runner: runner}
	result := check.Run(context.Background())

	assert.Equal(t, "warn", result.Status)
}

func TestPortCheck_Available(t *testing.T) {
	// Find a port that is available
	ln, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()

	check := &PortCheck{Port: port}
	result := check.Run(context.Background())

	assert.Equal(t, fmt.Sprintf("Port %d", port), result.Name)
	assert.Equal(t, "ok", result.Status)
	assert.Contains(t, result.Message, "available")
}

func TestPortCheck_InUse(t *testing.T) {
	// Occupy a port
	ln, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	defer ln.Close()

	port := ln.Addr().(*net.TCPAddr).Port

	check := &PortCheck{Port: port}
	result := check.Run(context.Background())

	assert.Equal(t, fmt.Sprintf("Port %d", port), result.Name)
	assert.Equal(t, "warn", result.Status)
	assert.Contains(t, result.Message, "in use")
	assert.Contains(t, result.Suggestion, "in use")
}

func TestGoCheck_OK(t *testing.T) {
	runner := testutil.NewMockRunner()
	runner.Responses["go"] = exec.Output{
		Stdout: "go version go1.22.0 linux/amd64",
	}

	check := &GoCheck{Runner: runner}
	result := check.Run(context.Background())

	assert.Equal(t, "Go", result.Name)
	assert.Equal(t, "ok", result.Status)
	assert.Contains(t, result.Version, "go1.22.0")
	assert.Equal(t, "Go available", result.Message)
}

func TestGoCheck_NotFound(t *testing.T) {
	runner := testutil.NewMockRunner()
	runner.Errors["go"] = fmt.Errorf("not found")

	check := &GoCheck{Runner: runner}
	result := check.Run(context.Background())

	assert.Equal(t, "Go", result.Name)
	assert.Equal(t, "warn", result.Status)
	assert.Contains(t, result.Suggestion, "go.dev")
}

func TestGoCheck_NonZeroExit(t *testing.T) {
	runner := testutil.NewMockRunner()
	runner.Responses["go"] = exec.Output{ExitCode: 1}

	check := &GoCheck{Runner: runner}
	result := check.Run(context.Background())

	assert.Equal(t, "warn", result.Status)
}

func TestParseJavaVersion(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`openjdk version "21.0.1" 2023-10-17`, "21.0.1"},
		{`java version "1.8.0_362"`, "1.8.0_362"},
		{`openjdk version "17.0.5" 2022-10-18`, "17.0.5"},
		{"some garbage", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseJavaVersion(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseMajorVersion(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"21.0.1", 21},
		{"17.0.5", 17},
		{"1.8.0_362", 8},
		{"1.7.0", 7},
		{"22", 22},
		{"", 0},
		{"abc", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseMajorVersion(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFirstLine(t *testing.T) {
	assert.Equal(t, "hello", firstLine("hello\nworld"))
	assert.Equal(t, "hello", firstLine("hello"))
	assert.Equal(t, "hello", firstLine("\nhello\n"))
	assert.Equal(t, "", firstLine(""))
}
