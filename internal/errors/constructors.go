// Package errors provides constructor functions for creating AppError instances.
package errors

import (
	"fmt"
	"time"
)

// NewValidationError creates an AppError for validation failures.
//
// This error type is used for invalid command line arguments and other
// input validation failures. Maps to exit code 2 (ErrInvalidArguments).
//
// Example: NewValidationError("invalid repository format: expected 'owner/repo'")
func NewValidationError(message string) *AppError {
	return &AppError{
		Code:    ErrInvalidArguments,
		Message: message,
		Cause:   nil,
	}
}

// NewAuthenticationError creates an AppError for authentication failures.
//
// This error type is used when authentication fails (e.g., invalid token).
// Maps to exit code 3 (ErrAuthenticationFailed).
//
// Example: NewAuthenticationError("token validation failed", err)
func NewAuthenticationError(message string, cause error) *AppError {
	return &AppError{
		Code:    ErrAuthenticationFailed,
		Message: message,
		Cause:   cause,
	}
}

// NewAuthorizationError creates an AppError for authorization/permission failures.
//
// This error type is used when the user lacks required permissions.
// Maps to exit code 4 (ErrInsufficientPerms).
//
// Example: NewAuthorizationError("token lacks 'repo' scope")
func NewAuthorizationError(message string) *AppError {
	return &AppError{
		Code:    ErrInsufficientPerms,
		Message: message,
		Cause:   nil,
	}
}

// NewRepositoryNotFoundError creates an AppError for repository not found.
//
// This error type is used when a repository doesn't exist or cannot be accessed.
// Maps to exit code 5 (ErrRepositoryNotFound).
//
// The error message includes the owner/repo format and actionable advice.
//
// Example: NewRepositoryNotFoundError("octocat", "hello-world")
func NewRepositoryNotFoundError(owner, repo string) *AppError {
	var repoPath string
	if owner != "" {
		repoPath = fmt.Sprintf("%s/%s", owner, repo)
	} else {
		repoPath = repo
	}

	message := fmt.Sprintf("Repository not found: %s. Ensure the repository exists and you have access to it", repoPath)

	return &AppError{
		Code:    ErrRepositoryNotFound,
		Message: message,
		Cause:   nil,
	}
}

// NewRateLimitError creates an AppError for API rate limit exceeded.
//
// This error type is used when the GitHub API rate limit is exceeded.
// Maps to exit code 6 (ErrAPIRateLimited).
//
// The error message includes the reset time and actionable advice.
//
// Example: NewRateLimitError(resetTime)
func NewRateLimitError(resetTime time.Time) *AppError {
	message := fmt.Sprintf("API rate limit exceeded. Rate limit resets at: %s", resetTime.Format(time.RFC3339))

	return &AppError{
		Code:    ErrAPIRateLimited,
		Message: message,
		Cause:   nil,
	}
}

// NewNetworkError creates an AppError for network connection failures.
//
// This error type is used for network-related errors (connection refused, timeout, etc.).
// Maps to exit code 1 (ErrGeneral).
//
// The error message includes actionable advice about checking internet connection.
//
// Example: NewNetworkError(errors.New("connection refused"))
func NewNetworkError(cause error) *AppError {
	message := fmt.Sprintf("Network error: check your internet connection. %v", cause)

	return &AppError{
		Code:    ErrGeneral,
		Message: message,
		Cause:   cause,
	}
}

// NewAPIError creates an AppError for GitHub API server errors.
//
// This error type is used for GitHub API server errors (5xx responses, etc.).
// Maps to exit code 1 (ErrGeneral).
//
// Example: NewAPIError("GitHub API returned 500 Internal Server Error", err)
func NewAPIError(message string, cause error) *AppError {
	return &AppError{
		Code:    ErrGeneral,
		Message: message,
		Cause:   cause,
	}
}
