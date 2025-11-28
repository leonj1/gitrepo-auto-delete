// Package errors provides application-specific error types with error codes.
package errors

import "fmt"

// AppError represents an application-specific error with a code, message, and optional cause.
//
// AppError implements the error interface and supports Go's error wrapping pattern
// through the Unwrap method.
type AppError struct {
	// Code is the specific error code for this error
	Code ErrorCode

	// Message is a human-readable error message
	Message string

	// Cause is the underlying error that caused this error (can be nil)
	Cause error
}

// Error implements the error interface and returns a formatted error message.
//
// If there is a cause, it includes the cause in the error message.
// Otherwise, it returns just the message.
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap returns the underlying cause error, enabling Go's error unwrapping.
//
// This allows errors.Is and errors.As to work correctly with AppError.
func (e *AppError) Unwrap() error {
	return e.Cause
}
