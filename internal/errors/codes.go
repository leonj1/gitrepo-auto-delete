// Package errors provides application-specific error types and exit code handling.
//
// This package defines ErrorCode constants that map to specific exit codes
// according to the Gherkin scenarios:
// - Repository not found -> code 5
// - Insufficient permissions -> code 4
// - API rate limit exceeded -> code 6
// - Network connection failure -> code 1
// - GitHub API server error -> code 1
// - Invalid command line arguments -> code 2
package errors

import "errors"

// ErrorCode represents the type of error that occurred.
type ErrorCode int

const (
	// ErrGeneral represents general errors including network and API server errors (exit code 1).
	ErrGeneral ErrorCode = 1

	// ErrInvalidArguments represents invalid command line arguments (exit code 2).
	ErrInvalidArguments ErrorCode = 2

	// ErrAuthenticationFailed represents authentication failures (exit code 3).
	ErrAuthenticationFailed ErrorCode = 3

	// ErrInsufficientPerms represents insufficient permissions (exit code 4).
	ErrInsufficientPerms ErrorCode = 4

	// ErrRepositoryNotFound represents repository not found errors (exit code 5).
	ErrRepositoryNotFound ErrorCode = 5

	// ErrAPIRateLimited represents API rate limit exceeded errors (exit code 6).
	ErrAPIRateLimited ErrorCode = 6
)

// GetExitCode maps an error to its corresponding exit code.
//
// If the error is an AppError, it returns the error code as an int.
// If the error is any other error type, it returns 1 (general error).
// If the error is nil, it returns 0.
//
// The function can unwrap error chains to find AppError instances.
func GetExitCode(err error) int {
	if err == nil {
		return 0
	}

	// Try to type assert to *AppError
	var appErr *AppError
	if errors.As(err, &appErr) {
		return int(appErr.Code)
	}

	// For any other error type, return 1 (general error)
	return 1
}
