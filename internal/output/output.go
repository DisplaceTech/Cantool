package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/fatih/color"
)

// Formatter is the interface for all output modes.
type Formatter interface {
	Success(msg string, data any)
	Error(err *CantoolError)
	Warning(msg string)
	Info(msg string)
	Table(headers []string, rows [][]string)
}

// New returns a Formatter for the given format name.
// Supported: "human" (default), "json", "quiet".
func New(format string) Formatter {
	return NewWithWriter(format, os.Stdout, os.Stderr)
}

// NewWithWriter returns a Formatter writing to the given writers.
func NewWithWriter(format string, stdout, stderr io.Writer) Formatter {
	switch format {
	case "json":
		return &JSONFormatter{out: stdout}
	case "quiet":
		return &QuietFormatter{errOut: stderr}
	default:
		return &HumanFormatter{out: stdout, errOut: stderr}
	}
}

// --- Human ---

// HumanFormatter renders colored terminal output.
type HumanFormatter struct {
	out    io.Writer
	errOut io.Writer
}

func (f *HumanFormatter) Success(msg string, _ any) {
	fmt.Fprintf(f.out, "%s %s\n", color.GreenString("✓"), msg)
}

func (f *HumanFormatter) Error(err *CantoolError) {
	fmt.Fprintf(f.errOut, "%s %s\n", color.RedString("✗"), err.Message)
	if err.Suggestion != "" {
		fmt.Fprintf(f.errOut, "  %s\n", err.Suggestion)
	}
}

func (f *HumanFormatter) Warning(msg string) {
	fmt.Fprintf(f.errOut, "%s %s\n", color.YellowString("⚠"), msg)
}

func (f *HumanFormatter) Info(msg string) {
	fmt.Fprintf(f.out, "%s\n", msg)
}

func (f *HumanFormatter) Table(headers []string, rows [][]string) {
	tw := tabwriter.NewWriter(f.out, 2, 4, 2, ' ', 0)
	for i, h := range headers {
		if i > 0 {
			fmt.Fprint(tw, "\t")
		}
		fmt.Fprint(tw, h)
	}
	fmt.Fprintln(tw)
	for _, row := range rows {
		for i, cell := range row {
			if i > 0 {
				fmt.Fprint(tw, "\t")
			}
			fmt.Fprint(tw, cell)
		}
		fmt.Fprintln(tw)
	}
	tw.Flush()
}

// --- JSON ---

// JSONFormatter outputs one JSON object per line.
type JSONFormatter struct {
	out io.Writer
}

type jsonMessage struct {
	Type       string      `json:"type"`
	Message    string      `json:"message,omitempty"`
	Data       any         `json:"data,omitempty"`
	Code       string      `json:"code,omitempty"`
	Suggestion string      `json:"suggestion,omitempty"`
	Headers    []string    `json:"headers,omitempty"`
	Rows       [][]string  `json:"rows,omitempty"`
}

func (f *JSONFormatter) emit(msg jsonMessage) {
	b, _ := json.Marshal(msg)
	fmt.Fprintf(f.out, "%s\n", b)
}

func (f *JSONFormatter) Success(msg string, data any) {
	f.emit(jsonMessage{Type: "success", Message: msg, Data: data})
}

func (f *JSONFormatter) Error(err *CantoolError) {
	f.emit(jsonMessage{Type: "error", Message: err.Message, Code: err.Code, Suggestion: err.Suggestion})
}

func (f *JSONFormatter) Warning(msg string) {
	f.emit(jsonMessage{Type: "warning", Message: msg})
}

func (f *JSONFormatter) Info(msg string) {
	f.emit(jsonMessage{Type: "info", Message: msg})
}

func (f *JSONFormatter) Table(headers []string, rows [][]string) {
	f.emit(jsonMessage{Type: "table", Headers: headers, Rows: rows})
}

// --- Quiet ---

// QuietFormatter only outputs errors.
type QuietFormatter struct {
	errOut io.Writer
}

func (f *QuietFormatter) Success(_ string, _ any) {}
func (f *QuietFormatter) Warning(_ string)        {}
func (f *QuietFormatter) Info(_ string)            {}

func (f *QuietFormatter) Error(err *CantoolError) {
	fmt.Fprintf(f.errOut, "[%s] %s\n", err.Code, err.Message)
}

func (f *QuietFormatter) Table(_ []string, _ [][]string) {}
