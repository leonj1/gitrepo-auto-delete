// Package errors_test provides tests for error constructor functions.
//
// These tests verify that error constructor functions create AppError instances
// with the correct error codes for each Gherkin scenario.
//
// Gherkin Scenarios:
// - Repository not found -> code 5
// - Insufficient permissions -> code 4
// - API rate limit exceeded -> code 6
// - Network connection failure -> code 1
// - GitHub API server error -> code 1
// - Invalid command line arguments -> code 2
//
// These tests are designed to FAIL until the implementations are properly created
// by the coder agent.
package errors_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	apperrors "github.com/josejulio/ghautodelete/internal/errors"
)

// =============================================================================
// NewValidationError Tests
// =============================================================================

// TestNewValidationError verifies NewValidationError creates correct error.
//
// Gherkin: Scenario: Invalid command line arguments -> code 2
//
// The implementation should:
// - Create AppError with ErrInvalidArguments code
// - Set the provided message
// - Have nil cause (validation errors don't wrap other errors)
func TestNewValidationError(t *testing.T) {
	tests := []struct {
		name            string
		message         string
		expectedCode    apperrors.ErrorCode
		expectedMessage string
	}{
		{
			name:            "creates validation error with message",
			message:         "invalid repository format",
			expectedCode:    apperrors.ErrInvalidArguments,
			expectedMessage: "invalid repository format",
		},
		{
			name:            "creates validation error for missing argument",
			message:         "repository is required",
			expectedCode:    apperrors.ErrInvalidArguments,
			expectedMessage: "repository is required",
		},
		{
			name:            "creates validation error with empty message",
			message:         "",
			expectedCode:    apperrors.ErrInvalidArguments,
			expectedMessage: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := apperrors.NewValidationError(tt.message)

			// Assert
			if err == nil {
				t.Fatal("NewValidationError() returned nil")
			}
			if err.Code != tt.expectedCode {
				t.Errorf("Code = %v, expected %v", err.Code, tt.expectedCode)
			}
			if err.Message != tt.expectedMessage {
				t.Errorf("Message = %q, expected %q", err.Message, tt.expectedMessage)
			}
			if err.Cause != nil {
				t.Errorf("Cause = %v, expected nil", err.Cause)
			}
		})
	}
}

// =============================================================================
// NewAuthenticationError Tests
// =============================================================================

// TestNewAuthenticationError verifies NewAuthenticationError creates correct error.
//
// The implementation should:
// - Create AppError with ErrAuthenticationFailed code
// - Set the provided message
// - Wrap the cause error
func TestNewAuthenticationError(t *testing.T) {
	tests := []struct {
		name            string
		message         string
		cause           error
		expectedCode    apperrors.ErrorCode
		expectedMessage string
	}{
		{
			name:            "creates authentication error with cause",
			message:         "token validation failed",
			cause:           errors.New("401 unauthorized"),
			expectedCode:    apperrors.ErrAuthenticationFailed,
			expectedMessage: "token validation failed",
		},
		{
			name:            "creates authentication error without cause",
			message:         "token expired",
			cause:           nil,
			expectedCode:    apperrors.ErrAuthenticationFailed,
			expectedMessage: "token expired",
		},
		{
			name:            "creates authentication error for invalid token",
			message:         "invalid token format",
			cause:           errors.New("malformed JWT"),
			expectedCode:    apperrors.ErrAuthenticationFailed,
			expectedMessage: "invalid token format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := apperrors.NewAuthenticationError(tt.message, tt.cause)

			// Assert
			if err == nil {
				t.Fatal("NewAuthenticationError() returned nil")
			}
			if err.Code != tt.expectedCode {
				t.Errorf("Code = %v, expected %v", err.Code, tt.expectedCode)
			}
			if err.Message != tt.expectedMessage {
				t.Errorf("Message = %q, expected %q", err.Message, tt.expectedMessage)
			}
			if tt.cause != nil && err.Cause == nil {
				t.Error("Cause should not be nil when cause is provided")
			}
			if tt.cause == nil && err.Cause != nil {
				t.Errorf("Cause = %v, expected nil", err.Cause)
			}
		})
	}
}

// =============================================================================
// NewAuthorizationError Tests
// =============================================================================

// TestNewAuthorizationError verifies NewAuthorizationError creates correct error.
//
// Gherkin: Scenario: Insufficient permissions -> code 4
//
// The implementation should:
// - Create AppError with ErrInsufficientPerms code
// - Set the provided message
func TestNewAuthorizationError(t *testing.T) {
	tests := []struct {
		name            string
		message         string
		expectedCode    apperrors.ErrorCode
		expectedMessage string
	}{
		{
			name:            "creates authorization error for insufficient permissions",
			message:         "token lacks 'repo' scope",
			expectedCode:    apperrors.ErrInsufficientPerms,
			expectedMessage: "token lacks 'repo' scope",
		},
		{
			name:            "creates authorization error for forbidden action",
			message:         "user does not have write access",
			expectedCode:    apperrors.ErrInsufficientPerms,
			expectedMessage: "user does not have write access",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := apperrors.NewAuthorizationError(tt.message)

			// Assert
			if err == nil {
				t.Fatal("NewAuthorizationError() returned nil")
			}
			if err.Code != tt.expectedCode {
				t.Errorf("Code = %v, expected %v", err.Code, tt.expectedCode)
			}
			if err.Message != tt.expectedMessage {
				t.Errorf("Message = %q, expected %q", err.Message, tt.expectedMessage)
			}
		})
	}
}

// =============================================================================
// NewRepositoryNotFoundError Tests
// =============================================================================

// TestNewRepositoryNotFoundError verifies NewRepositoryNotFoundError creates correct error.
//
// Gherkin: Scenario: Repository not found -> code 5
//
// The implementation should:
// - Create AppError with ErrRepositoryNotFound code
// - Set message that includes owner and repo
func TestNewRepositoryNotFoundError(t *testing.T) {
	tests := []struct {
		name                   string
		owner                  string
		repo                   string
		expectedCode           apperrors.ErrorCode
		expectedMessageContain string
	}{
		{
			name:                   "creates error with owner and repo",
			owner:                  "octocat",
			repo:                   "hello-world",
			expectedCode:           apperrors.ErrRepositoryNotFound,
			expectedMessageContain: "octocat/hello-world",
		},
		{
			name:                   "creates error for organization repo",
			owner:                  "github",
			repo:                   "docs",
			expectedCode:           apperrors.ErrRepositoryNotFound,
			expectedMessageContain: "github/docs",
		},
		{
			name:                   "creates error with empty owner",
			owner:                  "",
			repo:                   "repo",
			expectedCode:           apperrors.ErrRepositoryNotFound,
			expectedMessageContain: "repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := apperrors.NewRepositoryNotFoundError(tt.owner, tt.repo)

			// Assert
			if err == nil {
				t.Fatal("NewRepositoryNotFoundError() returned nil")
			}
			if err.Code != tt.expectedCode {
				t.Errorf("Code = %v, expected %v", err.Code, tt.expectedCode)
			}
			if !strings.Contains(err.Message, tt.expectedMessageContain) {
				t.Errorf("Message = %q, expected to contain %q", err.Message, tt.expectedMessageContain)
			}
		})
	}
}

// =============================================================================
// NewRateLimitError Tests
// =============================================================================

// TestNewRateLimitError verifies NewRateLimitError creates correct error.
//
// Gherkin: Scenario: API rate limit exceeded -> code 6
//
// The implementation should:
// - Create AppError with ErrAPIRateLimited code
// - Set message that includes reset time information
func TestNewRateLimitError(t *testing.T) {
	tests := []struct {
		name                   string
		resetTime              time.Time
		expectedCode           apperrors.ErrorCode
		expectedMessageContain string
	}{
		{
			name:                   "creates rate limit error with reset time",
			resetTime:              time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			expectedCode:           apperrors.ErrAPIRateLimited,
			expectedMessageContain: "rate limit",
		},
		{
			name:                   "creates rate limit error with future reset time",
			resetTime:              time.Now().Add(1 * time.Hour),
			expectedCode:           apperrors.ErrAPIRateLimited,
			expectedMessageContain: "rate limit",
		},
		{
			name:                   "creates rate limit error with zero time",
			resetTime:              time.Time{},
			expectedCode:           apperrors.ErrAPIRateLimited,
			expectedMessageContain: "rate limit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := apperrors.NewRateLimitError(tt.resetTime)

			// Assert
			if err == nil {
				t.Fatal("NewRateLimitError() returned nil")
			}
			if err.Code != tt.expectedCode {
				t.Errorf("Code = %v, expected %v", err.Code, tt.expectedCode)
			}
			errStr := strings.ToLower(err.Error())
			expectedLower := strings.ToLower(tt.expectedMessageContain)
			if !strings.Contains(errStr, expectedLower) {
				t.Errorf("Error() = %q, expected to contain %q", err.Error(), tt.expectedMessageContain)
			}
		})
	}
}

// =============================================================================
// NewNetworkError Tests
// =============================================================================

// TestNewNetworkError verifies NewNetworkError creates correct error.
//
// Gherkin: Scenario: Network connection failure -> code 1
//
// The implementation should:
// - Create AppError with ErrGeneral code
// - Wrap the cause error (network errors always have underlying cause)
func TestNewNetworkError(t *testing.T) {
	tests := []struct {
		name         string
		cause        error
		expectedCode apperrors.ErrorCode
	}{
		{
			name:         "creates network error with connection refused",
			cause:        errors.New("connection refused"),
			expectedCode: apperrors.ErrGeneral,
		},
		{
			name:         "creates network error with timeout",
			cause:        errors.New("connection timed out"),
			expectedCode: apperrors.ErrGeneral,
		},
		{
			name:         "creates network error with DNS failure",
			cause:        errors.New("no such host"),
			expectedCode: apperrors.ErrGeneral,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := apperrors.NewNetworkError(tt.cause)

			// Assert
			if err == nil {
				t.Fatal("NewNetworkError() returned nil")
			}
			if err.Code != tt.expectedCode {
				t.Errorf("Code = %v, expected %v", err.Code, tt.expectedCode)
			}
			if err.Cause == nil {
				t.Error("Cause should not be nil for network errors")
			}
			// Verify cause can be unwrapped
			if !errors.Is(err, tt.cause) {
				t.Error("errors.Is should find cause in chain")
			}
		})
	}
}

// =============================================================================
// NewAPIError Tests
// =============================================================================

// TestNewAPIError verifies NewAPIError creates correct error.
//
// Gherkin: Scenario: GitHub API server error -> code 1
//
// The implementation should:
// - Create AppError with ErrGeneral code
// - Set the provided message
// - Wrap the cause error
func TestNewAPIError(t *testing.T) {
	tests := []struct {
		name            string
		message         string
		cause           error
		expectedCode    apperrors.ErrorCode
		expectedMessage string
	}{
		{
			name:            "creates API error with 500 response",
			message:         "GitHub API returned 500 Internal Server Error",
			cause:           errors.New("HTTP 500"),
			expectedCode:    apperrors.ErrGeneral,
			expectedMessage: "GitHub API returned 500 Internal Server Error",
		},
		{
			name:            "creates API error with 502 response",
			message:         "GitHub API returned 502 Bad Gateway",
			cause:           errors.New("HTTP 502"),
			expectedCode:    apperrors.ErrGeneral,
			expectedMessage: "GitHub API returned 502 Bad Gateway",
		},
		{
			name:            "creates API error without cause",
			message:         "unexpected API response",
			cause:           nil,
			expectedCode:    apperrors.ErrGeneral,
			expectedMessage: "unexpected API response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := apperrors.NewAPIError(tt.message, tt.cause)

			// Assert
			if err == nil {
				t.Fatal("NewAPIError() returned nil")
			}
			if err.Code != tt.expectedCode {
				t.Errorf("Code = %v, expected %v", err.Code, tt.expectedCode)
			}
			if err.Message != tt.expectedMessage {
				t.Errorf("Message = %q, expected %q", err.Message, tt.expectedMessage)
			}
			if tt.cause != nil && err.Cause == nil {
				t.Error("Cause should not be nil when cause is provided")
			}
		})
	}
}

// =============================================================================
// Error Message Quality Tests
// =============================================================================

// TestErrorMessagesContainActionableAdvice verifies error messages are helpful.
//
// The implementation should provide actionable advice in error messages.
func TestErrorMessagesContainActionableAdvice(t *testing.T) {
	tests := []struct {
		name              string
		createError       func() *apperrors.AppError
		expectedSubstring string
	}{
		{
			name: "validation error suggests correct format",
			createError: func() *apperrors.AppError {
				return apperrors.NewValidationError("invalid repository format: expected 'owner/repo'")
			},
			expectedSubstring: "owner/repo",
		},
		{
			name: "authorization error mentions required scope",
			createError: func() *apperrors.AppError {
				return apperrors.NewAuthorizationError("token lacks 'repo' scope")
			},
			expectedSubstring: "repo",
		},
		{
			name: "repository not found includes repo name",
			createError: func() *apperrors.AppError {
				return apperrors.NewRepositoryNotFoundError("octocat", "nonexistent")
			},
			expectedSubstring: "octocat",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := tt.createError()

			// Assert
			errorMsg := err.Error()
			if !strings.Contains(errorMsg, tt.expectedSubstring) {
				t.Errorf("Error message %q should contain %q", errorMsg, tt.expectedSubstring)
			}
		})
	}
}
