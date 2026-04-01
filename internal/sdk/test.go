package sdk

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/displacetech/cantool/internal/exec"
	"github.com/displacetech/cantool/internal/output"
)

// TestResult holds structured test run results.
type TestResult struct {
	Passed   int
	Failed   int
	Total    int
	Tests    []TestCase
	Duration time.Duration
	RawOutput string
}

// TestCase is a single test's result.
type TestCase struct {
	Name     string
	Passed   bool
	Duration time.Duration
	Error    string
}

// Test runs DAML script tests using the detected SDK.
func Test(ctx context.Context, runner exec.CommandRunner, sdk *SDKInfo, projectDir string, verbose bool) (*TestResult, error) {
	start := time.Now()

	out, err := runner.Run(ctx, sdk.Command, "test")
	if err != nil {
		return nil, output.Wrapf(err, "CT5001",
			"Ensure the DAML SDK is installed and working",
			"test command failed to execute")
	}

	combined := out.Stdout + out.Stderr
	result := &TestResult{
		Duration:  time.Since(start),
		RawOutput: combined,
	}

	result.Tests = parseTestCases(combined)
	for _, tc := range result.Tests {
		if tc.Passed {
			result.Passed++
		} else {
			result.Failed++
		}
	}
	result.Total = len(result.Tests)

	return result, nil
}

var testLineRe = regexp.MustCompile(`(?i)(ok|pass|fail|error)\s+.*?(\w+Test\w*|\w+test\w*)`)

func parseTestCases(s string) []TestCase {
	var cases []TestCase
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Look for lines with test indicators
		lower := strings.ToLower(line)
		if strings.Contains(lower, "pass") || strings.Contains(lower, "ok") {
			name := extractTestName(line)
			if name != "" {
				cases = append(cases, TestCase{Name: name, Passed: true})
			}
		} else if strings.Contains(lower, "fail") || strings.Contains(lower, "error") {
			name := extractTestName(line)
			errMsg := ""
			// Look ahead for error detail
			if i+1 < len(lines) && strings.TrimSpace(lines[i+1]) != "" {
				errMsg = strings.TrimSpace(lines[i+1])
			}
			if name != "" {
				cases = append(cases, TestCase{Name: name, Passed: false, Error: errMsg})
			}
		}
	}
	return cases
}

var testNameRe = regexp.MustCompile(`\b(test\w+)\b`)

func extractTestName(line string) string {
	matches := testNameRe.FindStringSubmatch(line)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}
