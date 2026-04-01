package doctor

import (
	"context"
	"fmt"
	"testing"

	"github.com/displacetech/cantool/internal/exec"
	"github.com/displacetech/cantool/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDoctor_HasAllChecks(t *testing.T) {
	runner := testutil.NewMockRunner()
	doc := NewDoctor(runner)

	// Should have 6 checks: Java, DAML, Docker, Port 5011, Port 7575, Go
	assert.Len(t, doc.checks, 6)
}

func TestDoctor_RunAll_AllPass(t *testing.T) {
	runner := testutil.NewMockRunner()
	runner.Responses["java"] = exec.Output{
		Stderr: `openjdk version "21.0.1" 2023-10-17`,
	}
	runner.Responses["daml"] = exec.Output{
		Stdout: "2.8.0",
	}
	runner.Responses["docker"] = exec.Output{
		Stdout: "Docker version 24.0.7",
	}
	runner.Responses["go"] = exec.Output{
		Stdout: "go version go1.22.0 linux/amd64",
	}

	doc := NewDoctor(runner)
	results, pass := doc.RunAll(context.Background())

	require.Len(t, results, 6)
	assert.True(t, pass)

	// Verify check names in order
	assert.Equal(t, "Java", results[0].Name)
	assert.Equal(t, "DAML SDK", results[1].Name)
	assert.Equal(t, "Docker", results[2].Name)
	assert.Equal(t, "Port 5011", results[3].Name)
	assert.Equal(t, "Port 7575", results[4].Name)
	assert.Equal(t, "Go", results[5].Name)

	// Required checks pass
	assert.Equal(t, "ok", results[0].Status)
	assert.Equal(t, "ok", results[1].Status)
}

func TestDoctor_RunAll_RequiredFail(t *testing.T) {
	runner := testutil.NewMockRunner()
	runner.Errors["java"] = fmt.Errorf("not found")
	runner.Errors["daml"] = fmt.Errorf("not found")
	runner.Errors["dpm"] = fmt.Errorf("not found")
	runner.Responses["docker"] = exec.Output{
		Stdout: "Docker version 24.0.7",
	}
	runner.Responses["go"] = exec.Output{
		Stdout: "go version go1.22.0 linux/amd64",
	}

	doc := NewDoctor(runner)
	results, pass := doc.RunAll(context.Background())

	require.Len(t, results, 6)
	assert.False(t, pass, "overall should fail when required checks fail")

	assert.Equal(t, "fail", results[0].Status) // Java
	assert.Equal(t, "fail", results[1].Status) // DAML
}

func TestDoctor_RunAll_OptionalWarnStillPasses(t *testing.T) {
	runner := testutil.NewMockRunner()
	runner.Responses["java"] = exec.Output{
		Stderr: `openjdk version "21.0.1" 2023-10-17`,
	}
	runner.Responses["daml"] = exec.Output{
		Stdout: "2.8.0",
	}
	// Docker not found (optional)
	runner.Errors["docker"] = fmt.Errorf("not found")
	// Go not found (optional)
	runner.Errors["go"] = fmt.Errorf("not found")

	doc := NewDoctor(runner)
	results, pass := doc.RunAll(context.Background())

	require.Len(t, results, 6)
	assert.True(t, pass, "overall should pass when only optional checks warn")

	assert.Equal(t, "ok", results[0].Status)   // Java
	assert.Equal(t, "ok", results[1].Status)   // DAML
	assert.Equal(t, "warn", results[2].Status) // Docker
	assert.Equal(t, "warn", results[5].Status) // Go
}

func TestDoctor_RunAll_MixedResults(t *testing.T) {
	runner := testutil.NewMockRunner()
	// Java too old - required fail
	runner.Responses["java"] = exec.Output{
		Stderr: `openjdk version "17.0.5" 2022-10-18`,
	}
	// DAML found - required pass
	runner.Responses["daml"] = exec.Output{
		Stdout: "2.8.0",
	}
	// Docker not found - optional warn
	runner.Errors["docker"] = fmt.Errorf("not found")
	// Go found
	runner.Responses["go"] = exec.Output{
		Stdout: "go version go1.22.0 linux/amd64",
	}

	doc := NewDoctor(runner)
	results, pass := doc.RunAll(context.Background())

	require.Len(t, results, 6)
	assert.False(t, pass, "overall should fail due to Java version")

	assert.Equal(t, "fail", results[0].Status) // Java too old
	assert.Equal(t, "ok", results[1].Status)   // DAML
	assert.Equal(t, "warn", results[2].Status) // Docker
}
