package output

import "fmt"

// CantoolError is a structured error with code, message, and fix suggestion.
//
// Error code scheme by subsystem:
//
//	CT1xxx — Configuration
//	CT2xxx — SDK/Tooling
//	CT3xxx — Sandbox/Runtime
//	CT4xxx — Build
//	CT5xxx — Test
//	CT6xxx — Ledger API
//	CT7xxx — MCP Server
//	CT8xxx — Plugin System
//	CT9xxx — Environment Management
type CantoolError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion"`
	DocsURL    string `json:"docs_url,omitempty"`
	Cause      error  `json:"-"`
}

func (e *CantoolError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *CantoolError) Unwrap() error {
	return e.Cause
}

// Errorf creates a CantoolError with the given code, suggestion, and formatted message.
func Errorf(code, suggestion, format string, args ...any) *CantoolError {
	return &CantoolError{
		Code:       code,
		Message:    fmt.Sprintf(format, args...),
		Suggestion: suggestion,
	}
}

// Wrapf creates a CantoolError wrapping an existing error.
func Wrapf(cause error, code, suggestion, format string, args ...any) *CantoolError {
	return &CantoolError{
		Code:       code,
		Message:    fmt.Sprintf(format, args...),
		Suggestion: suggestion,
		Cause:      cause,
	}
}
