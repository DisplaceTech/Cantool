package doctor

import (
	"context"

	"github.com/displacetech/cantool/internal/exec"
)

// Doctor runs a suite of prerequisite health checks.
type Doctor struct {
	checks []Check
}

// NewDoctor creates a Doctor with the standard set of checks.
func NewDoctor(runner exec.CommandRunner) *Doctor {
	return &Doctor{
		checks: []Check{
			&JavaCheck{Runner: runner},
			&DAMLCheck{Runner: runner},
			&DockerCheck{Runner: runner},
			&PortCheck{Port: 5011},
			&PortCheck{Port: 7575},
			&GoCheck{Runner: runner},
		},
	}
}

// RunAll executes every check and returns the results plus an overall
// pass/fail. Overall pass is true when no required check has failed.
// Required checks produce "fail" status; optional checks produce "warn".
func (d *Doctor) RunAll(ctx context.Context) ([]CheckResult, bool) {
	results := make([]CheckResult, 0, len(d.checks))
	pass := true

	for _, c := range d.checks {
		r := c.Run(ctx)
		results = append(results, r)
		if r.Status == "fail" {
			pass = false
		}
	}

	return results, pass
}
