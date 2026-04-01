package testutil

import (
	"context"
	"fmt"
	"sync"

	"github.com/displacetech/cantool/internal/exec"
)

// RunCall records a single invocation of CommandRunner.Run.
type RunCall struct {
	Name string
	Args []string
}

// MockRunner is a test double for exec.CommandRunner.
// Configure responses by command name; all calls are recorded.
type MockRunner struct {
	mu        sync.Mutex
	Responses map[string]exec.Output
	Errors    map[string]error
	Calls     []RunCall
}

// NewMockRunner creates a MockRunner with empty response maps.
func NewMockRunner() *MockRunner {
	return &MockRunner{
		Responses: make(map[string]exec.Output),
		Errors:    make(map[string]error),
	}
}

func (m *MockRunner) Run(_ context.Context, name string, args ...string) (exec.Output, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Calls = append(m.Calls, RunCall{Name: name, Args: args})

	if err, ok := m.Errors[name]; ok {
		return exec.Output{}, err
	}
	if out, ok := m.Responses[name]; ok {
		return out, nil
	}
	return exec.Output{}, fmt.Errorf("mock: no response configured for %q", name)
}
