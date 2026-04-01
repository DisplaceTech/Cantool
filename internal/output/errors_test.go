package output

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCantoolError_Error(t *testing.T) {
	err := &CantoolError{Code: "CT1001", Message: "config not found"}
	assert.Equal(t, "[CT1001] config not found", err.Error())
}

func TestCantoolError_Unwrap(t *testing.T) {
	cause := errors.New("underlying")
	err := &CantoolError{Code: "CT1002", Message: "bad yaml", Cause: cause}
	assert.Equal(t, cause, errors.Unwrap(err))
}

func TestCantoolError_UnwrapNil(t *testing.T) {
	err := &CantoolError{Code: "CT1001", Message: "no cause"}
	assert.Nil(t, errors.Unwrap(err))
}

func TestCantoolError_ImplementsError(t *testing.T) {
	var _ error = &CantoolError{}
}

func TestCantoolError_JSONMarshal(t *testing.T) {
	err := &CantoolError{
		Code:       "CT6001",
		Message:    "connection refused",
		Suggestion: "run cantool dev",
		DocsURL:    "https://docs.example.com",
	}
	b, marshalErr := json.Marshal(err)
	require.NoError(t, marshalErr)

	var m map[string]string
	require.NoError(t, json.Unmarshal(b, &m))
	assert.Equal(t, "CT6001", m["code"])
	assert.Equal(t, "connection refused", m["message"])
	assert.Equal(t, "run cantool dev", m["suggestion"])
	assert.Equal(t, "https://docs.example.com", m["docs_url"])
}

func TestCantoolError_JSONOmitsEmptyDocsURL(t *testing.T) {
	err := &CantoolError{Code: "CT1001", Message: "test", Suggestion: "fix it"}
	b, marshalErr := json.Marshal(err)
	require.NoError(t, marshalErr)
	assert.NotContains(t, string(b), "docs_url")
}

func TestErrorf(t *testing.T) {
	err := Errorf("CT2001", "install SDK", "SDK %q not found", "daml")
	assert.Equal(t, "CT2001", err.Code)
	assert.Equal(t, `SDK "daml" not found`, err.Message)
	assert.Equal(t, "install SDK", err.Suggestion)
	assert.Nil(t, err.Cause)
}

func TestWrapf(t *testing.T) {
	cause := errors.New("dial refused")
	err := Wrapf(cause, "CT6001", "start node", "cannot connect to %s", "localhost")
	assert.Equal(t, "CT6001", err.Code)
	assert.Equal(t, "cannot connect to localhost", err.Message)
	assert.Equal(t, "start node", err.Suggestion)
	assert.Equal(t, cause, err.Cause)
}
