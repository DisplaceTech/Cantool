package output

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestFormatters() (human *HumanFormatter, jsonf *JSONFormatter, quiet *QuietFormatter, hOut, hErr, jOut, qErr *bytes.Buffer) {
	hOut = &bytes.Buffer{}
	hErr = &bytes.Buffer{}
	jOut = &bytes.Buffer{}
	qErr = &bytes.Buffer{}
	human = &HumanFormatter{out: hOut, errOut: hErr}
	jsonf = &JSONFormatter{out: jOut}
	quiet = &QuietFormatter{errOut: qErr}
	return
}

// --- Success ---

func TestHumanFormatter_Success(t *testing.T) {
	h, _, _, hOut, _, _, _ := newTestFormatters()
	h.Success("build succeeded", nil)
	assert.Contains(t, hOut.String(), "build succeeded")
}

func TestJSONFormatter_Success(t *testing.T) {
	_, j, _, _, _, jOut, _ := newTestFormatters()
	j.Success("ok", map[string]int{"count": 5})
	var m map[string]any
	require.NoError(t, json.Unmarshal(jOut.Bytes(), &m))
	assert.Equal(t, "success", m["type"])
	assert.Equal(t, "ok", m["message"])
	data := m["data"].(map[string]any)
	assert.Equal(t, float64(5), data["count"])
}

func TestQuietFormatter_Success(t *testing.T) {
	_, _, q, _, _, _, qErr := newTestFormatters()
	q.Success("should not appear", nil)
	assert.Empty(t, qErr.String())
}

// --- Error ---

func TestHumanFormatter_Error(t *testing.T) {
	h, _, _, _, hErr, _, _ := newTestFormatters()
	h.Error(&CantoolError{Code: "CT1001", Message: "not found", Suggestion: "create file"})
	out := hErr.String()
	assert.Contains(t, out, "not found")
	assert.Contains(t, out, "create file")
}

func TestJSONFormatter_Error(t *testing.T) {
	_, j, _, _, _, jOut, _ := newTestFormatters()
	j.Error(&CantoolError{Code: "CT1001", Message: "bad", Suggestion: "fix"})
	var m map[string]any
	require.NoError(t, json.Unmarshal(jOut.Bytes(), &m))
	assert.Equal(t, "error", m["type"])
	assert.Equal(t, "CT1001", m["code"])
	assert.Equal(t, "bad", m["message"])
	assert.Equal(t, "fix", m["suggestion"])
}

func TestQuietFormatter_Error(t *testing.T) {
	_, _, q, _, _, _, qErr := newTestFormatters()
	q.Error(&CantoolError{Code: "CT2001", Message: "sdk missing"})
	assert.Equal(t, "[CT2001] sdk missing\n", qErr.String())
}

// --- Warning ---

func TestHumanFormatter_Warning(t *testing.T) {
	h, _, _, _, hErr, _, _ := newTestFormatters()
	h.Warning("docker not found")
	assert.Contains(t, hErr.String(), "docker not found")
}

func TestJSONFormatter_Warning(t *testing.T) {
	_, j, _, _, _, jOut, _ := newTestFormatters()
	j.Warning("heads up")
	var m map[string]any
	require.NoError(t, json.Unmarshal(jOut.Bytes(), &m))
	assert.Equal(t, "warning", m["type"])
	assert.Equal(t, "heads up", m["message"])
}

func TestQuietFormatter_Warning(t *testing.T) {
	_, _, q, _, _, _, qErr := newTestFormatters()
	q.Warning("should not appear")
	assert.Empty(t, qErr.String())
}

// --- Info ---

func TestHumanFormatter_Info(t *testing.T) {
	h, _, _, hOut, _, _, _ := newTestFormatters()
	h.Info("some info")
	assert.Equal(t, "some info\n", hOut.String())
}

func TestJSONFormatter_Info(t *testing.T) {
	_, j, _, _, _, jOut, _ := newTestFormatters()
	j.Info("detail")
	var m map[string]any
	require.NoError(t, json.Unmarshal(jOut.Bytes(), &m))
	assert.Equal(t, "info", m["type"])
}

func TestQuietFormatter_Info(t *testing.T) {
	_, _, q, _, _, _, qErr := newTestFormatters()
	q.Info("should not appear")
	assert.Empty(t, qErr.String())
}

// --- Table ---

func TestHumanFormatter_Table(t *testing.T) {
	h, _, _, hOut, _, _, _ := newTestFormatters()
	h.Table([]string{"NAME", "VERSION"}, [][]string{{"foo", "1.0"}, {"bar", "2.0"}})
	out := hOut.String()
	assert.Contains(t, out, "NAME")
	assert.Contains(t, out, "foo")
	assert.Contains(t, out, "bar")
}

func TestJSONFormatter_Table(t *testing.T) {
	_, j, _, _, _, jOut, _ := newTestFormatters()
	j.Table([]string{"A", "B"}, [][]string{{"1", "2"}})
	var m map[string]any
	require.NoError(t, json.Unmarshal(jOut.Bytes(), &m))
	assert.Equal(t, "table", m["type"])
	headers := m["headers"].([]any)
	assert.Equal(t, "A", headers[0])
}

func TestQuietFormatter_Table(t *testing.T) {
	_, _, q, _, _, _, qErr := newTestFormatters()
	q.Table([]string{"X"}, [][]string{{"y"}})
	assert.Empty(t, qErr.String())
}

// --- New/NewWithWriter ---

func TestNew_ReturnsCorrectType(t *testing.T) {
	tests := []struct {
		format   string
		wantType any
	}{
		{"human", &HumanFormatter{}},
		{"json", &JSONFormatter{}},
		{"quiet", &QuietFormatter{}},
		{"unknown", &HumanFormatter{}},
	}
	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			f := New(tt.format)
			assert.IsType(t, tt.wantType, f)
		})
	}
}

func TestNewWithWriter(t *testing.T) {
	var out, errOut bytes.Buffer
	f := NewWithWriter("human", &out, &errOut)
	f.Success("hello", nil)
	assert.Contains(t, out.String(), "hello")
}
