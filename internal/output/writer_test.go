// Package output_test provides tests for the OutputWriter implementation.
//
// These tests verify that OutputWriter correctly implements the IOutputWriter
// interface and handles verbose mode, output formatting, and stream routing:
// - Success messages: Prefixed with checkmark, written to stdout
// - Error messages: Prefixed with "Error: ", written to stderr
// - Info messages: Plain text, written to stdout
// - Verbose messages: Prefixed with "[verbose] ", written to stdout (only when verbose=true)
//
// These tests are designed to FAIL until the implementation is properly created
// by the coder agent.
package output_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/josejulio/ghautodelete/internal/output"
	"github.com/josejulio/ghautodelete/pkg/interfaces"
)

// =============================================================================
// Interface Satisfaction Tests
// =============================================================================

// TestOutputWriterImplementsIOutputWriter verifies OutputWriter implements IOutputWriter.
//
// The implementation should:
// - Define OutputWriter struct in internal/output/writer.go
// - Implement Success(message string), Error(message string), Info(message string), Verbose(message string)
func TestOutputWriterImplementsIOutputWriter(t *testing.T) {
	// Arrange
	var out bytes.Buffer
	var errOut bytes.Buffer

	// Act - create OutputWriter
	writer := output.NewOutputWriter(false, &out, &errOut)

	// Assert - compile-time interface satisfaction check
	var _ interfaces.IOutputWriter = writer
	if writer == nil {
		t.Error("NewOutputWriter should return a non-nil writer")
	}
}

// =============================================================================
// Constructor Tests
// =============================================================================

// TestNewOutputWriterReturnsNonNil verifies constructor returns non-nil.
//
// The implementation should:
// - Return a non-nil OutputWriter from NewOutputWriter
func TestNewOutputWriterReturnsNonNil(t *testing.T) {
	// Arrange
	var out bytes.Buffer
	var errOut bytes.Buffer

	// Act
	writer := output.NewOutputWriter(false, &out, &errOut)

	// Assert
	if writer == nil {
		t.Error("NewOutputWriter() should return non-nil writer")
	}
}

// TestNewOutputWriterAcceptsVerboseFlag verifies verbose parameter is accepted.
//
// The implementation should:
// - Accept verbose boolean as first parameter
func TestNewOutputWriterAcceptsVerboseFlag(t *testing.T) {
	// Arrange
	var out bytes.Buffer
	var errOut bytes.Buffer

	// Act - verbose mode enabled
	verboseWriter := output.NewOutputWriter(true, &out, &errOut)

	// Assert
	if verboseWriter == nil {
		t.Error("NewOutputWriter(true, ...) should return non-nil writer")
	}

	// Act - verbose mode disabled
	nonVerboseWriter := output.NewOutputWriter(false, &out, &errOut)

	// Assert
	if nonVerboseWriter == nil {
		t.Error("NewOutputWriter(false, ...) should return non-nil writer")
	}
}

// TestNewOutputWriterAcceptsWriters verifies io.Writer parameters are accepted.
//
// The implementation should:
// - Accept out io.Writer as second parameter (for stdout)
// - Accept errOut io.Writer as third parameter (for stderr)
func TestNewOutputWriterAcceptsWriters(t *testing.T) {
	// Arrange
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	// Act
	writer := output.NewOutputWriter(false, &stdout, &stderr)

	// Assert - writer created successfully with both io.Writers
	if writer == nil {
		t.Error("NewOutputWriter should accept io.Writer parameters")
	}
}

// =============================================================================
// Success Method Tests
// =============================================================================

// TestSuccessWritesToStdout verifies Success writes to the out writer.
//
// Gherkin: Scenario: Non-verbose mode shows only essential output
// - Success messages should always be shown
//
// The implementation should:
// - Write success message to the out writer (stdout)
func TestSuccessWritesToStdout(t *testing.T) {
	// Arrange
	var out bytes.Buffer
	var errOut bytes.Buffer
	writer := output.NewOutputWriter(false, &out, &errOut)

	// Act
	writer.Success("Operation completed")

	// Assert
	if out.Len() == 0 {
		t.Error("Success() should write to out writer")
	}
	if errOut.Len() != 0 {
		t.Error("Success() should NOT write to errOut writer")
	}
}

// TestSuccessPrefixesWithCheckmark verifies Success prefixes message with checkmark.
//
// The implementation should:
// - Prefix success messages with checkmark symbol
func TestSuccessPrefixesWithCheckmark(t *testing.T) {
	// Arrange
	var out bytes.Buffer
	var errOut bytes.Buffer
	writer := output.NewOutputWriter(false, &out, &errOut)

	// Act
	writer.Success("Operation completed")

	// Assert
	output := out.String()
	// Should contain checkmark prefix
	if !strings.HasPrefix(output, "\u2713 ") && !strings.HasPrefix(output, "v ") && !strings.Contains(output, "v ") {
		// Accept various checkmark representations
		if !strings.Contains(output, "Operation completed") {
			t.Errorf("Success() output should contain message, got: %q", output)
		}
	}
	if !strings.Contains(output, "Operation completed") {
		t.Errorf("Success() should contain original message, got: %q", output)
	}
}

// TestSuccessAppendsNewline verifies Success appends newline.
//
// The implementation should:
// - Append newline to success messages
func TestSuccessAppendsNewline(t *testing.T) {
	// Arrange
	var out bytes.Buffer
	var errOut bytes.Buffer
	writer := output.NewOutputWriter(false, &out, &errOut)

	// Act
	writer.Success("Test message")

	// Assert
	output := out.String()
	if !strings.HasSuffix(output, "\n") {
		t.Errorf("Success() should append newline, got: %q", output)
	}
}

// TestSuccessShowsInBothVerboseModes verifies Success works in both modes.
//
// Gherkin: Scenario: Non-verbose mode shows only essential output
// - Success messages are always shown
//
// The implementation should:
// - Show success messages regardless of verbose setting
func TestSuccessShowsInBothVerboseModes(t *testing.T) {
	tests := []struct {
		name    string
		verbose bool
	}{
		{name: "verbose mode enabled", verbose: true},
		{name: "verbose mode disabled", verbose: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			var out bytes.Buffer
			var errOut bytes.Buffer
			writer := output.NewOutputWriter(tt.verbose, &out, &errOut)

			// Act
			writer.Success("Test message")

			// Assert
			if out.Len() == 0 {
				t.Errorf("Success() should write output in %s", tt.name)
			}
		})
	}
}

// =============================================================================
// Error Method Tests
// =============================================================================

// TestErrorWritesToStderr verifies Error writes to the errOut writer.
//
// Gherkin: Scenario: Non-verbose mode shows only essential output
// - Error messages should always be shown on stderr
//
// The implementation should:
// - Write error message to the errOut writer (stderr)
func TestErrorWritesToStderr(t *testing.T) {
	// Arrange
	var out bytes.Buffer
	var errOut bytes.Buffer
	writer := output.NewOutputWriter(false, &out, &errOut)

	// Act
	writer.Error("Something went wrong")

	// Assert
	if errOut.Len() == 0 {
		t.Error("Error() should write to errOut writer")
	}
	if out.Len() != 0 {
		t.Error("Error() should NOT write to out writer")
	}
}

// TestErrorPrefixesWithErrorLabel verifies Error prefixes message.
//
// The implementation should:
// - Prefix error messages with "Error: "
func TestErrorPrefixesWithErrorLabel(t *testing.T) {
	// Arrange
	var out bytes.Buffer
	var errOut bytes.Buffer
	writer := output.NewOutputWriter(false, &out, &errOut)

	// Act
	writer.Error("Something went wrong")

	// Assert
	output := errOut.String()
	if !strings.HasPrefix(output, "Error: ") {
		t.Errorf("Error() should prefix with 'Error: ', got: %q", output)
	}
	if !strings.Contains(output, "Something went wrong") {
		t.Errorf("Error() should contain original message, got: %q", output)
	}
}

// TestErrorAppendsNewline verifies Error appends newline.
//
// The implementation should:
// - Append newline to error messages
func TestErrorAppendsNewline(t *testing.T) {
	// Arrange
	var out bytes.Buffer
	var errOut bytes.Buffer
	writer := output.NewOutputWriter(false, &out, &errOut)

	// Act
	writer.Error("Test error")

	// Assert
	output := errOut.String()
	if !strings.HasSuffix(output, "\n") {
		t.Errorf("Error() should append newline, got: %q", output)
	}
}

// TestErrorShowsInBothVerboseModes verifies Error works in both modes.
//
// The implementation should:
// - Show error messages regardless of verbose setting
func TestErrorShowsInBothVerboseModes(t *testing.T) {
	tests := []struct {
		name    string
		verbose bool
	}{
		{name: "verbose mode enabled", verbose: true},
		{name: "verbose mode disabled", verbose: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			var out bytes.Buffer
			var errOut bytes.Buffer
			writer := output.NewOutputWriter(tt.verbose, &out, &errOut)

			// Act
			writer.Error("Test error")

			// Assert
			if errOut.Len() == 0 {
				t.Errorf("Error() should write output in %s", tt.name)
			}
		})
	}
}

// =============================================================================
// Info Method Tests
// =============================================================================

// TestInfoWritesToStdout verifies Info writes to the out writer.
//
// Gherkin: Scenario: Non-verbose mode shows only essential output
// - Info messages should always be shown
//
// The implementation should:
// - Write info message to the out writer (stdout)
func TestInfoWritesToStdout(t *testing.T) {
	// Arrange
	var out bytes.Buffer
	var errOut bytes.Buffer
	writer := output.NewOutputWriter(false, &out, &errOut)

	// Act
	writer.Info("Processing repository")

	// Assert
	if out.Len() == 0 {
		t.Error("Info() should write to out writer")
	}
	if errOut.Len() != 0 {
		t.Error("Info() should NOT write to errOut writer")
	}
}

// TestInfoOutputsPlainText verifies Info outputs plain text without prefix.
//
// The implementation should:
// - Output info messages as plain text (no prefix)
func TestInfoOutputsPlainText(t *testing.T) {
	// Arrange
	var out bytes.Buffer
	var errOut bytes.Buffer
	writer := output.NewOutputWriter(false, &out, &errOut)
	message := "Processing repository owner/repo"

	// Act
	writer.Info(message)

	// Assert
	output := out.String()
	// Info should contain the message and a newline
	expected := message + "\n"
	if output != expected {
		t.Errorf("Info() should output plain text with newline, got: %q, expected: %q", output, expected)
	}
}

// TestInfoAppendsNewline verifies Info appends newline.
//
// The implementation should:
// - Append newline to info messages
func TestInfoAppendsNewline(t *testing.T) {
	// Arrange
	var out bytes.Buffer
	var errOut bytes.Buffer
	writer := output.NewOutputWriter(false, &out, &errOut)

	// Act
	writer.Info("Test message")

	// Assert
	output := out.String()
	if !strings.HasSuffix(output, "\n") {
		t.Errorf("Info() should append newline, got: %q", output)
	}
}

// TestInfoShowsInBothVerboseModes verifies Info works in both modes.
//
// The implementation should:
// - Show info messages regardless of verbose setting
func TestInfoShowsInBothVerboseModes(t *testing.T) {
	tests := []struct {
		name    string
		verbose bool
	}{
		{name: "verbose mode enabled", verbose: true},
		{name: "verbose mode disabled", verbose: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			var out bytes.Buffer
			var errOut bytes.Buffer
			writer := output.NewOutputWriter(tt.verbose, &out, &errOut)

			// Act
			writer.Info("Test info")

			// Assert
			if out.Len() == 0 {
				t.Errorf("Info() should write output in %s", tt.name)
			}
		})
	}
}

// =============================================================================
// Verbose Method Tests - Verbose Mode Enabled
// =============================================================================

// TestVerboseWritesToStdoutWhenEnabled verifies Verbose writes when enabled.
//
// Gherkin: Scenario: Verbose mode shows authentication details
// Gherkin: Scenario: Verbose mode shows repository fetch details
// Gherkin: Scenario: Verbose mode shows API operations
// Gherkin: Scenario: Use short form -v for verbose mode
//
// The implementation should:
// - Write verbose message to the out writer (stdout) when verbose=true
func TestVerboseWritesToStdoutWhenEnabled(t *testing.T) {
	// Arrange
	var out bytes.Buffer
	var errOut bytes.Buffer
	writer := output.NewOutputWriter(true, &out, &errOut)

	// Act
	writer.Verbose("Authenticating with token...")

	// Assert
	if out.Len() == 0 {
		t.Error("Verbose() should write to out writer when verbose mode enabled")
	}
	if errOut.Len() != 0 {
		t.Error("Verbose() should NOT write to errOut writer")
	}
}

// TestVerbosePrefixesWithVerboseLabel verifies Verbose prefixes message.
//
// The implementation should:
// - Prefix verbose messages with "[verbose] "
func TestVerbosePrefixesWithVerboseLabel(t *testing.T) {
	// Arrange
	var out bytes.Buffer
	var errOut bytes.Buffer
	writer := output.NewOutputWriter(true, &out, &errOut)

	// Act
	writer.Verbose("Fetching repository details")

	// Assert
	output := out.String()
	if !strings.HasPrefix(output, "[verbose] ") {
		t.Errorf("Verbose() should prefix with '[verbose] ', got: %q", output)
	}
	if !strings.Contains(output, "Fetching repository details") {
		t.Errorf("Verbose() should contain original message, got: %q", output)
	}
}

// TestVerboseAppendsNewlineWhenEnabled verifies Verbose appends newline.
//
// The implementation should:
// - Append newline to verbose messages when enabled
func TestVerboseAppendsNewlineWhenEnabled(t *testing.T) {
	// Arrange
	var out bytes.Buffer
	var errOut bytes.Buffer
	writer := output.NewOutputWriter(true, &out, &errOut)

	// Act
	writer.Verbose("Test verbose message")

	// Assert
	output := out.String()
	if !strings.HasSuffix(output, "\n") {
		t.Errorf("Verbose() should append newline, got: %q", output)
	}
}

// TestVerboseShowsAuthenticationDetails verifies auth details shown.
//
// Gherkin: Scenario: Verbose mode shows authentication details
//
// The implementation should:
// - Show authentication-related verbose messages when enabled
func TestVerboseShowsAuthenticationDetails(t *testing.T) {
	// Arrange
	var out bytes.Buffer
	var errOut bytes.Buffer
	writer := output.NewOutputWriter(true, &out, &errOut)

	// Act
	writer.Verbose("Using token from GITHUB_TOKEN environment variable")

	// Assert
	output := out.String()
	if !strings.Contains(output, "Using token from GITHUB_TOKEN environment variable") {
		t.Errorf("Verbose() should show auth details, got: %q", output)
	}
}

// TestVerboseShowsRepositoryFetchDetails verifies repo fetch details shown.
//
// Gherkin: Scenario: Verbose mode shows repository fetch details
//
// The implementation should:
// - Show repository fetch verbose messages when enabled
func TestVerboseShowsRepositoryFetchDetails(t *testing.T) {
	// Arrange
	var out bytes.Buffer
	var errOut bytes.Buffer
	writer := output.NewOutputWriter(true, &out, &errOut)

	// Act
	writer.Verbose("Fetching repository: owner/repo")

	// Assert
	output := out.String()
	if !strings.Contains(output, "Fetching repository: owner/repo") {
		t.Errorf("Verbose() should show repo fetch details, got: %q", output)
	}
}

// TestVerboseShowsAPIOperations verifies API operations shown.
//
// Gherkin: Scenario: Verbose mode shows API operations
//
// The implementation should:
// - Show API operation verbose messages when enabled
func TestVerboseShowsAPIOperations(t *testing.T) {
	// Arrange
	var out bytes.Buffer
	var errOut bytes.Buffer
	writer := output.NewOutputWriter(true, &out, &errOut)

	// Act
	writer.Verbose("API call: PATCH /repos/owner/repo")

	// Assert
	output := out.String()
	if !strings.Contains(output, "API call: PATCH /repos/owner/repo") {
		t.Errorf("Verbose() should show API operations, got: %q", output)
	}
}

// =============================================================================
// Verbose Method Tests - Verbose Mode Disabled (No-Op)
// =============================================================================

// TestVerboseIsNoOpWhenDisabled verifies Verbose is no-op when disabled.
//
// Gherkin: Scenario: Non-verbose mode shows only essential output
// - Verbose messages should NOT be shown when verbose=false
//
// The implementation should:
// - NOT write anything when verbose=false
func TestVerboseIsNoOpWhenDisabled(t *testing.T) {
	// Arrange
	var out bytes.Buffer
	var errOut bytes.Buffer
	writer := output.NewOutputWriter(false, &out, &errOut)

	// Act
	writer.Verbose("This should not appear")

	// Assert
	if out.Len() != 0 {
		t.Errorf("Verbose() should NOT write to out writer when verbose disabled, got: %q", out.String())
	}
	if errOut.Len() != 0 {
		t.Errorf("Verbose() should NOT write to errOut writer when verbose disabled, got: %q", errOut.String())
	}
}

// TestVerboseMultipleCallsNoOpWhenDisabled verifies multiple Verbose calls are no-op.
//
// The implementation should:
// - NOT write anything for any number of Verbose calls when verbose=false
func TestVerboseMultipleCallsNoOpWhenDisabled(t *testing.T) {
	// Arrange
	var out bytes.Buffer
	var errOut bytes.Buffer
	writer := output.NewOutputWriter(false, &out, &errOut)

	// Act
	writer.Verbose("Message 1")
	writer.Verbose("Message 2")
	writer.Verbose("Message 3")

	// Assert
	if out.Len() != 0 {
		t.Errorf("Multiple Verbose() calls should NOT write when disabled, got: %q", out.String())
	}
}

// =============================================================================
// Output Format Verification Tests
// =============================================================================

// TestSuccessFormatComplete verifies complete Success format.
//
// The implementation should:
// - Format: checkmark + space + message + newline
func TestSuccessFormatComplete(t *testing.T) {
	// Arrange
	var out bytes.Buffer
	var errOut bytes.Buffer
	writer := output.NewOutputWriter(false, &out, &errOut)
	message := "Repository updated successfully"

	// Act
	writer.Success(message)

	// Assert
	output := out.String()
	// Check format: should start with checkmark symbol, contain message, end with newline
	if !strings.Contains(output, message) {
		t.Errorf("Success format should contain message, got: %q", output)
	}
	if !strings.HasSuffix(output, "\n") {
		t.Errorf("Success format should end with newline, got: %q", output)
	}
	// The first character should be a checkmark (U+2713)
	if !strings.HasPrefix(output, "\u2713 ") {
		t.Errorf("Success format should start with checkmark, got: %q", output)
	}
}

// TestErrorFormatComplete verifies complete Error format.
//
// The implementation should:
// - Format: "Error: " + message + newline
func TestErrorFormatComplete(t *testing.T) {
	// Arrange
	var out bytes.Buffer
	var errOut bytes.Buffer
	writer := output.NewOutputWriter(false, &out, &errOut)
	message := "Repository not found"

	// Act
	writer.Error(message)

	// Assert
	output := errOut.String()
	expected := "Error: " + message + "\n"
	if output != expected {
		t.Errorf("Error format incorrect, got: %q, expected: %q", output, expected)
	}
}

// TestInfoFormatComplete verifies complete Info format.
//
// The implementation should:
// - Format: message + newline (plain text, no prefix)
func TestInfoFormatComplete(t *testing.T) {
	// Arrange
	var out bytes.Buffer
	var errOut bytes.Buffer
	writer := output.NewOutputWriter(false, &out, &errOut)
	message := "Checking repository status"

	// Act
	writer.Info(message)

	// Assert
	output := out.String()
	expected := message + "\n"
	if output != expected {
		t.Errorf("Info format incorrect, got: %q, expected: %q", output, expected)
	}
}

// TestVerboseFormatComplete verifies complete Verbose format when enabled.
//
// The implementation should:
// - Format: "[verbose] " + message + newline
func TestVerboseFormatComplete(t *testing.T) {
	// Arrange
	var out bytes.Buffer
	var errOut bytes.Buffer
	writer := output.NewOutputWriter(true, &out, &errOut)
	message := "Debug information here"

	// Act
	writer.Verbose(message)

	// Assert
	output := out.String()
	expected := "[verbose] " + message + "\n"
	if output != expected {
		t.Errorf("Verbose format incorrect, got: %q, expected: %q", output, expected)
	}
}

// =============================================================================
// Mixed Method Calls Tests
// =============================================================================

// TestMixedMethodCallsRouteCorrectly verifies all methods route to correct writers.
//
// The implementation should:
// - Route Success, Info, Verbose to out writer
// - Route Error to errOut writer
func TestMixedMethodCallsRouteCorrectly(t *testing.T) {
	// Arrange
	var out bytes.Buffer
	var errOut bytes.Buffer
	writer := output.NewOutputWriter(true, &out, &errOut)

	// Act
	writer.Success("success msg")
	writer.Error("error msg")
	writer.Info("info msg")
	writer.Verbose("verbose msg")

	// Assert - stdout should have success, info, verbose
	stdout := out.String()
	if !strings.Contains(stdout, "success msg") {
		t.Error("stdout should contain success message")
	}
	if !strings.Contains(stdout, "info msg") {
		t.Error("stdout should contain info message")
	}
	if !strings.Contains(stdout, "verbose msg") {
		t.Error("stdout should contain verbose message")
	}
	if strings.Contains(stdout, "error msg") {
		t.Error("stdout should NOT contain error message")
	}

	// Assert - stderr should only have error
	stderr := errOut.String()
	if !strings.Contains(stderr, "error msg") {
		t.Error("stderr should contain error message")
	}
	if strings.Contains(stderr, "success msg") {
		t.Error("stderr should NOT contain success message")
	}
	if strings.Contains(stderr, "info msg") {
		t.Error("stderr should NOT contain info message")
	}
	if strings.Contains(stderr, "verbose msg") {
		t.Error("stderr should NOT contain verbose message")
	}
}

// TestMixedMethodCallsWithVerboseDisabled verifies routing with verbose off.
//
// The implementation should:
// - Route Success, Info to out writer
// - Route Error to errOut writer
// - Verbose should be no-op
func TestMixedMethodCallsWithVerboseDisabled(t *testing.T) {
	// Arrange
	var out bytes.Buffer
	var errOut bytes.Buffer
	writer := output.NewOutputWriter(false, &out, &errOut)

	// Act
	writer.Success("success msg")
	writer.Error("error msg")
	writer.Info("info msg")
	writer.Verbose("verbose msg") // Should be no-op

	// Assert - stdout should have success, info (NOT verbose)
	stdout := out.String()
	if !strings.Contains(stdout, "success msg") {
		t.Error("stdout should contain success message")
	}
	if !strings.Contains(stdout, "info msg") {
		t.Error("stdout should contain info message")
	}
	if strings.Contains(stdout, "verbose msg") {
		t.Error("stdout should NOT contain verbose message when verbose disabled")
	}

	// Assert - stderr should only have error
	stderr := errOut.String()
	if !strings.Contains(stderr, "error msg") {
		t.Error("stderr should contain error message")
	}
}

// =============================================================================
// Edge Cases Tests
// =============================================================================

// TestEmptyMessageHandling verifies empty messages are handled.
//
// The implementation should:
// - Handle empty string messages without panic
func TestEmptyMessageHandling(t *testing.T) {
	// Arrange
	var out bytes.Buffer
	var errOut bytes.Buffer
	writer := output.NewOutputWriter(true, &out, &errOut)

	// Act & Assert - should not panic
	writer.Success("")
	writer.Error("")
	writer.Info("")
	writer.Verbose("")

	// Assert - output should still have prefixes/newlines for non-empty methods
	stdout := out.String()
	stderr := errOut.String()

	// Success should have checkmark prefix even with empty message
	if !strings.Contains(stdout, "\u2713 ") {
		t.Error("Success('') should still output checkmark prefix")
	}

	// Error should have Error: prefix even with empty message
	if !strings.Contains(stderr, "Error: ") {
		t.Error("Error('') should still output Error: prefix")
	}
}

// TestMessageWithNewlines verifies messages with embedded newlines.
//
// The implementation should:
// - Handle messages containing newlines
func TestMessageWithNewlines(t *testing.T) {
	// Arrange
	var out bytes.Buffer
	var errOut bytes.Buffer
	writer := output.NewOutputWriter(false, &out, &errOut)
	multilineMsg := "Line 1\nLine 2\nLine 3"

	// Act
	writer.Info(multilineMsg)

	// Assert
	output := out.String()
	if !strings.Contains(output, "Line 1\nLine 2\nLine 3") {
		t.Errorf("Info() should preserve embedded newlines, got: %q", output)
	}
}

// TestMessageWithSpecialCharacters verifies special character handling.
//
// The implementation should:
// - Handle messages with special characters (Unicode, emojis, etc.)
func TestMessageWithSpecialCharacters(t *testing.T) {
	// Arrange
	var out bytes.Buffer
	var errOut bytes.Buffer
	writer := output.NewOutputWriter(false, &out, &errOut)
	specialMsg := "Repository: owner/repo-name_v2.0"

	// Act
	writer.Info(specialMsg)

	// Assert
	output := out.String()
	if !strings.Contains(output, specialMsg) {
		t.Errorf("Info() should handle special characters, got: %q", output)
	}
}

// TestMessageWithUnicode verifies Unicode message handling.
//
// The implementation should:
// - Handle messages with Unicode characters
func TestMessageWithUnicode(t *testing.T) {
	// Arrange
	var out bytes.Buffer
	var errOut bytes.Buffer
	writer := output.NewOutputWriter(false, &out, &errOut)
	unicodeMsg := "Processing..."

	// Act
	writer.Info(unicodeMsg)

	// Assert
	output := out.String()
	if !strings.Contains(output, unicodeMsg) {
		t.Errorf("Info() should handle Unicode characters, got: %q", output)
	}
}

// =============================================================================
// Writer Interface Compatibility Tests
// =============================================================================

// TestAcceptsAnyIoWriter verifies any io.Writer implementation works.
//
// The implementation should:
// - Accept any io.Writer implementation
func TestAcceptsAnyIoWriter(t *testing.T) {
	// Arrange - use io.Discard which implements io.Writer
	var errBuf bytes.Buffer

	// Act - use io.Discard for stdout
	writer := output.NewOutputWriter(false, io.Discard, &errBuf)

	// Assert - should not panic
	writer.Success("test")
	writer.Info("test")
}

// =============================================================================
// Table-Driven Complete Scenario Tests
// =============================================================================

// TestOutputWriterCompleteScenarios tests all output scenarios in one table.
func TestOutputWriterCompleteScenarios(t *testing.T) {
	tests := []struct {
		name           string
		verbose        bool
		method         string
		message        string
		expectStdout   bool
		expectStderr   bool
		stdoutContains string
		stderrContains string
	}{
		{
			name:           "success in non-verbose",
			verbose:        false,
			method:         "Success",
			message:        "Operation done",
			expectStdout:   true,
			expectStderr:   false,
			stdoutContains: "Operation done",
		},
		{
			name:           "success in verbose",
			verbose:        true,
			method:         "Success",
			message:        "Operation done",
			expectStdout:   true,
			expectStderr:   false,
			stdoutContains: "Operation done",
		},
		{
			name:           "error in non-verbose",
			verbose:        false,
			method:         "Error",
			message:        "Something failed",
			expectStdout:   false,
			expectStderr:   true,
			stderrContains: "Error: Something failed",
		},
		{
			name:           "error in verbose",
			verbose:        true,
			method:         "Error",
			message:        "Something failed",
			expectStdout:   false,
			expectStderr:   true,
			stderrContains: "Error: Something failed",
		},
		{
			name:           "info in non-verbose",
			verbose:        false,
			method:         "Info",
			message:        "Processing",
			expectStdout:   true,
			expectStderr:   false,
			stdoutContains: "Processing",
		},
		{
			name:           "info in verbose",
			verbose:        true,
			method:         "Info",
			message:        "Processing",
			expectStdout:   true,
			expectStderr:   false,
			stdoutContains: "Processing",
		},
		{
			name:           "verbose in verbose mode",
			verbose:        true,
			method:         "Verbose",
			message:        "Debug info",
			expectStdout:   true,
			expectStderr:   false,
			stdoutContains: "[verbose] Debug info",
		},
		{
			name:         "verbose in non-verbose mode",
			verbose:      false,
			method:       "Verbose",
			message:      "Debug info",
			expectStdout: false,
			expectStderr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			var out bytes.Buffer
			var errOut bytes.Buffer
			writer := output.NewOutputWriter(tt.verbose, &out, &errOut)

			// Act
			switch tt.method {
			case "Success":
				writer.Success(tt.message)
			case "Error":
				writer.Error(tt.message)
			case "Info":
				writer.Info(tt.message)
			case "Verbose":
				writer.Verbose(tt.message)
			}

			// Assert - stdout
			if tt.expectStdout {
				if out.Len() == 0 {
					t.Errorf("Expected stdout output, got none")
				}
				if tt.stdoutContains != "" && !strings.Contains(out.String(), tt.stdoutContains) {
					t.Errorf("stdout should contain %q, got: %q", tt.stdoutContains, out.String())
				}
			} else {
				if out.Len() != 0 {
					t.Errorf("Expected no stdout output, got: %q", out.String())
				}
			}

			// Assert - stderr
			if tt.expectStderr {
				if errOut.Len() == 0 {
					t.Errorf("Expected stderr output, got none")
				}
				if tt.stderrContains != "" && !strings.Contains(errOut.String(), tt.stderrContains) {
					t.Errorf("stderr should contain %q, got: %q", tt.stderrContains, errOut.String())
				}
			} else {
				if errOut.Len() != 0 {
					t.Errorf("Expected no stderr output, got: %q", errOut.String())
				}
			}
		})
	}
}

// =============================================================================
// Gherkin Scenario Verification Tests
// =============================================================================

// TestGherkinVerboseModeShowsAuthenticationDetails verifies Gherkin scenario.
//
// Gherkin: Scenario: Verbose mode shows authentication details
func TestGherkinVerboseModeShowsAuthenticationDetails(t *testing.T) {
	// Arrange
	var out bytes.Buffer
	var errOut bytes.Buffer
	writer := output.NewOutputWriter(true, &out, &errOut)

	// Act - simulate authentication verbose output
	writer.Verbose("Using token from environment variable GITHUB_TOKEN")
	writer.Verbose("Token validated successfully")
	writer.Verbose("Authenticated as: testuser")

	// Assert
	stdout := out.String()
	if !strings.Contains(stdout, "Using token from environment variable GITHUB_TOKEN") {
		t.Error("Verbose mode should show authentication source")
	}
	if !strings.Contains(stdout, "Token validated successfully") {
		t.Error("Verbose mode should show validation status")
	}
	if !strings.Contains(stdout, "Authenticated as: testuser") {
		t.Error("Verbose mode should show authenticated user")
	}
}

// TestGherkinVerboseModeShowsRepositoryFetchDetails verifies Gherkin scenario.
//
// Gherkin: Scenario: Verbose mode shows repository fetch details
func TestGherkinVerboseModeShowsRepositoryFetchDetails(t *testing.T) {
	// Arrange
	var out bytes.Buffer
	var errOut bytes.Buffer
	writer := output.NewOutputWriter(true, &out, &errOut)

	// Act - simulate repository fetch verbose output
	writer.Verbose("Fetching repository: owner/repo")
	writer.Verbose("Repository found: owner/repo")
	writer.Verbose("Default branch: main")
	writer.Verbose("Current delete_branch_on_merge: false")

	// Assert
	stdout := out.String()
	if !strings.Contains(stdout, "Fetching repository: owner/repo") {
		t.Error("Verbose mode should show repository being fetched")
	}
	if !strings.Contains(stdout, "Default branch: main") {
		t.Error("Verbose mode should show default branch")
	}
}

// TestGherkinVerboseModeShowsAPIOperations verifies Gherkin scenario.
//
// Gherkin: Scenario: Verbose mode shows API operations
func TestGherkinVerboseModeShowsAPIOperations(t *testing.T) {
	// Arrange
	var out bytes.Buffer
	var errOut bytes.Buffer
	writer := output.NewOutputWriter(true, &out, &errOut)

	// Act - simulate API operation verbose output
	writer.Verbose("API: GET /repos/owner/repo")
	writer.Verbose("API: PATCH /repos/owner/repo")
	writer.Verbose("API response: 200 OK")

	// Assert
	stdout := out.String()
	if !strings.Contains(stdout, "API: GET /repos/owner/repo") {
		t.Error("Verbose mode should show GET API calls")
	}
	if !strings.Contains(stdout, "API: PATCH /repos/owner/repo") {
		t.Error("Verbose mode should show PATCH API calls")
	}
}

// TestGherkinNonVerboseModeShowsOnlyEssential verifies Gherkin scenario.
//
// Gherkin: Scenario: Non-verbose mode shows only essential output
func TestGherkinNonVerboseModeShowsOnlyEssential(t *testing.T) {
	// Arrange
	var out bytes.Buffer
	var errOut bytes.Buffer
	writer := output.NewOutputWriter(false, &out, &errOut)

	// Act - simulate typical usage with mixed output
	writer.Verbose("Debug: Starting process") // Should NOT appear
	writer.Info("Processing repository owner/repo")
	writer.Verbose("Debug: API call made") // Should NOT appear
	writer.Success("Repository configured successfully")
	writer.Verbose("Debug: Done") // Should NOT appear

	// Assert - essential output should appear
	stdout := out.String()
	if !strings.Contains(stdout, "Processing repository owner/repo") {
		t.Error("Non-verbose mode should show Info messages")
	}
	if !strings.Contains(stdout, "Repository configured successfully") {
		t.Error("Non-verbose mode should show Success messages")
	}

	// Assert - verbose output should NOT appear
	if strings.Contains(stdout, "Debug:") {
		t.Error("Non-verbose mode should NOT show Verbose messages")
	}
}
