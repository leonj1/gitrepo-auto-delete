// Package output provides implementations for writing formatted output messages.
//
// This package implements the IOutputWriter interface for handling various types
// of output messages with different verbosity levels and formatting.
package output

import (
	"fmt"
	"io"

	"github.com/josejulio/ghautodelete/pkg/interfaces"
)

// OutputWriter implements the IOutputWriter interface for writing formatted output.
//
// It routes messages to appropriate output streams (stdout/stderr) and respects
// the verbose flag for debug output.
type OutputWriter struct {
	verbose bool
	out     io.Writer
	errOut  io.Writer
}

// NewOutputWriter creates a new OutputWriter instance.
//
// Parameters:
//   - verbose: if true, Verbose() messages will be output; if false, they are suppressed
//   - out: the writer for standard output (Success, Info, Verbose messages)
//   - errOut: the writer for error output (Error messages)
//
// Returns:
//   - *OutputWriter: a new OutputWriter instance
func NewOutputWriter(verbose bool, out io.Writer, errOut io.Writer) *OutputWriter {
	return &OutputWriter{
		verbose: verbose,
		out:     out,
		errOut:  errOut,
	}
}

// Success writes a success message to stdout with a checkmark prefix.
//
// Format: "✓ message\n"
//
// Parameters:
//   - message: the success message to display
func (w *OutputWriter) Success(message string) {
	fmt.Fprintf(w.out, "\u2713 %s\n", message)
}

// Error writes an error message to stderr with an "Error: " prefix.
//
// Format: "Error: message\n"
//
// Parameters:
//   - message: the error message to display
func (w *OutputWriter) Error(message string) {
	fmt.Fprintf(w.errOut, "Error: %s\n", message)
}

// Info writes an informational message to stdout as plain text.
//
// Format: "message\n"
//
// Parameters:
//   - message: the informational message to display
func (w *OutputWriter) Info(message string) {
	fmt.Fprintf(w.out, "%s\n", message)
}

// Verbose writes a verbose/debug message to stdout with a "[verbose] " prefix.
//
// This method is a no-op when verbose mode is disabled (verbose=false).
//
// Format: "[verbose] message\n" (only when verbose=true)
//
// Parameters:
//   - message: the verbose message to display
func (w *OutputWriter) Verbose(message string) {
	if w.verbose {
		fmt.Fprintf(w.out, "[verbose] %s\n", message)
	}
}

// Compile-time interface satisfaction check
var _ interfaces.IOutputWriter = (*OutputWriter)(nil)
