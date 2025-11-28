// Package errors_test provides tests for the errors package error codes and exit code mapping.
//
// These tests verify that ErrorCode constants are defined correctly and that
// GetExitCode() maps errors to correct exit codes per Gherkin scenarios.
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
	"testing"
	"time"

	apperrors "github.com/josejulio/ghautodelete/internal/errors"
)

// =============================================================================
// ErrorCode Constants Tests
// =============================================================================

// TestErrorCodeConstantsExist verifies all ErrorCode constants are defined.
//
// The implementation should define these constants in internal/errors/codes.go:
// - ErrGeneral              ErrorCode = 1
// - ErrInvalidArguments     ErrorCode = 2
// - ErrAuthenticationFailed ErrorCode = 3
// - ErrInsufficientPerms    ErrorCode = 4
// - ErrRepositoryNotFound   ErrorCode = 5
// - ErrAPIRateLimited       ErrorCode = 6
func TestErrorCodeConstantsExist(t *testing.T) {
	tests := []struct {
		name         string
		code         apperrors.ErrorCode
		expectedInt  int
		description  string
	}{
		{
			name:        "ErrGeneral is 1",
			code:        apperrors.ErrGeneral,
			expectedInt: 1,
			description: "General errors including network and API server errors",
		},
		{
			name:        "ErrInvalidArguments is 2",
			code:        apperrors.ErrInvalidArguments,
			expectedInt: 2,
			description: "Invalid command line arguments",
		},
		{
			name:        "ErrAuthenticationFailed is 3",
			code:        apperrors.ErrAuthenticationFailed,
			expectedInt: 3,
			description: "Authentication failures (invalid token)",
		},
		{
			name:        "ErrInsufficientPerms is 4",
			code:        apperrors.ErrInsufficientPerms,
			expectedInt: 4,
			description: "Insufficient permissions",
		},
		{
			name:        "ErrRepositoryNotFound is 5",
			code:        apperrors.ErrRepositoryNotFound,
			expectedInt: 5,
			description: "Repository not found",
		},
		{
			name:        "ErrAPIRateLimited is 6",
			code:        apperrors.ErrAPIRateLimited,
			expectedInt: 6,
			description: "API rate limit exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			actualInt := int(tt.code)

			// Assert
			if actualInt != tt.expectedInt {
				t.Errorf("ErrorCode %s = %d, expected %d (%s)",
					tt.name, actualInt, tt.expectedInt, tt.description)
			}
		})
	}
}

// TestErrorCodeTypeIsInt verifies ErrorCode is based on int type.
//
// The implementation should define: type ErrorCode int
func TestErrorCodeTypeIsInt(t *testing.T) {
	// Arrange
	var code apperrors.ErrorCode = 1

	// Act - verify it can be converted to int
	intValue := int(code)

	// Assert
	if intValue != 1 {
		t.Errorf("ErrorCode should be convertible to int, got %d", intValue)
	}
}

// TestErrorCodeValuesAreUnique verifies all ErrorCode values are unique.
func TestErrorCodeValuesAreUnique(t *testing.T) {
	// Arrange
	codes := []apperrors.ErrorCode{
		apperrors.ErrGeneral,
		apperrors.ErrInvalidArguments,
		apperrors.ErrAuthenticationFailed,
		apperrors.ErrInsufficientPerms,
		apperrors.ErrRepositoryNotFound,
		apperrors.ErrAPIRateLimited,
	}

	// Act - build map to check for duplicates
	seen := make(map[apperrors.ErrorCode]bool)
	for _, code := range codes {
		if seen[code] {
			t.Errorf("Duplicate ErrorCode value: %d", code)
		}
		seen[code] = true
	}

	// Assert - verify we have 6 unique codes
	if len(seen) != 6 {
		t.Errorf("Expected 6 unique error codes, got %d", len(seen))
	}
}

// =============================================================================
// GetExitCode Function Tests
// =============================================================================

// TestGetExitCodeForAppErrors verifies GetExitCode returns correct codes for AppError.
//
// The implementation should:
// - Extract the Code field from AppError
// - Return the code as int
func TestGetExitCodeForAppErrors(t *testing.T) {
	tests := []struct {
		name             string
		appError         *apperrors.AppError
		expectedExitCode int
		gherkinScenario  string
	}{
		{
			name: "repository not found returns code 5",
			appError: &apperrors.AppError{
				Code:    apperrors.ErrRepositoryNotFound,
				Message: "repository not found",
			},
			expectedExitCode: 5,
			gherkinScenario:  "Repository not found -> code 5",
		},
		{
			name: "insufficient permissions returns code 4",
			appError: &apperrors.AppError{
				Code:    apperrors.ErrInsufficientPerms,
				Message: "insufficient permissions",
			},
			expectedExitCode: 4,
			gherkinScenario:  "Insufficient permissions -> code 4",
		},
		{
			name: "API rate limit exceeded returns code 6",
			appError: &apperrors.AppError{
				Code:    apperrors.ErrAPIRateLimited,
				Message: "rate limit exceeded",
			},
			expectedExitCode: 6,
			gherkinScenario:  "API rate limit exceeded -> code 6",
		},
		{
			name: "network connection failure returns code 1",
			appError: &apperrors.AppError{
				Code:    apperrors.ErrGeneral,
				Message: "network error",
				Cause:   errors.New("connection refused"),
			},
			expectedExitCode: 1,
			gherkinScenario:  "Network connection failure -> code 1",
		},
		{
			name: "GitHub API server error returns code 1",
			appError: &apperrors.AppError{
				Code:    apperrors.ErrGeneral,
				Message: "API server error",
				Cause:   errors.New("HTTP 500"),
			},
			expectedExitCode: 1,
			gherkinScenario:  "GitHub API server error -> code 1",
		},
		{
			name: "invalid command line arguments returns code 2",
			appError: &apperrors.AppError{
				Code:    apperrors.ErrInvalidArguments,
				Message: "invalid arguments",
			},
			expectedExitCode: 2,
			gherkinScenario:  "Invalid command line arguments -> code 2",
		},
		{
			name: "authentication failed returns code 3",
			appError: &apperrors.AppError{
				Code:    apperrors.ErrAuthenticationFailed,
				Message: "authentication failed",
			},
			expectedExitCode: 3,
			gherkinScenario:  "Authentication failures -> code 3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			exitCode := apperrors.GetExitCode(tt.appError)

			// Assert
			if exitCode != tt.expectedExitCode {
				t.Errorf("GetExitCode() = %d, expected %d (Gherkin: %s)",
					exitCode, tt.expectedExitCode, tt.gherkinScenario)
			}
		})
	}
}

// TestGetExitCodeForNonAppErrors verifies non-AppError errors map to code 1.
//
// The implementation should:
// - Return 1 for any error that is not an AppError
// - Handle nil gracefully (return 0 or appropriate default)
func TestGetExitCodeForNonAppErrors(t *testing.T) {
	tests := []struct {
		name             string
		err              error
		expectedExitCode int
	}{
		{
			name:             "standard error returns code 1",
			err:              errors.New("some error"),
			expectedExitCode: 1,
		},
		{
			name:             "wrapped standard error returns code 1",
			err:              errors.New("wrapped: inner error"),
			expectedExitCode: 1,
		},
		{
			name:             "nil error returns code 0",
			err:              nil,
			expectedExitCode: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			exitCode := apperrors.GetExitCode(tt.err)

			// Assert
			if exitCode != tt.expectedExitCode {
				t.Errorf("GetExitCode() = %d, expected %d", exitCode, tt.expectedExitCode)
			}
		})
	}
}

// TestGetExitCodeWithWrappedAppError verifies wrapped AppErrors are handled.
//
// The implementation should be able to find AppError in wrapped error chains.
func TestGetExitCodeWithWrappedAppError(t *testing.T) {
	// Arrange
	appErr := &apperrors.AppError{
		Code:    apperrors.ErrRepositoryNotFound,
		Message: "repository not found",
	}
	// Wrap the AppError
	wrappedErr := wrapError(appErr, "context")

	// Act
	exitCode := apperrors.GetExitCode(wrappedErr)

	// Assert - should find AppError in chain and return its code
	// Note: The expected behavior depends on implementation.
	// If GetExitCode unwraps, it should return 5.
	// If it doesn't unwrap, it should return 1.
	// For best practice, it should unwrap and return 5.
	if exitCode != 5 && exitCode != 1 {
		t.Errorf("GetExitCode() = %d, expected 5 (if unwraps) or 1 (if doesn't unwrap)", exitCode)
	}
}

// =============================================================================
// Gherkin Scenario Integration Tests
// =============================================================================

// TestGherkinScenarioRepositoryNotFound tests: Repository not found -> code 5
func TestGherkinScenarioRepositoryNotFound(t *testing.T) {
	// Arrange - create the error using constructor
	err := apperrors.NewRepositoryNotFoundError("octocat", "nonexistent")

	// Act
	exitCode := apperrors.GetExitCode(err)

	// Assert
	if exitCode != 5 {
		t.Errorf("Gherkin: 'Repository not found -> code 5' FAILED: got %d", exitCode)
	}
}

// TestGherkinScenarioInsufficientPermissions tests: Insufficient permissions -> code 4
func TestGherkinScenarioInsufficientPermissions(t *testing.T) {
	// Arrange - create the error using constructor
	err := apperrors.NewAuthorizationError("token lacks required scope")

	// Act
	exitCode := apperrors.GetExitCode(err)

	// Assert
	if exitCode != 4 {
		t.Errorf("Gherkin: 'Insufficient permissions -> code 4' FAILED: got %d", exitCode)
	}
}

// TestGherkinScenarioAPIRateLimitExceeded tests: API rate limit exceeded -> code 6
func TestGherkinScenarioAPIRateLimitExceeded(t *testing.T) {
	// Arrange - create the error using constructor
	err := apperrors.NewRateLimitError(fixedResetTime())

	// Act
	exitCode := apperrors.GetExitCode(err)

	// Assert
	if exitCode != 6 {
		t.Errorf("Gherkin: 'API rate limit exceeded -> code 6' FAILED: got %d", exitCode)
	}
}

// TestGherkinScenarioNetworkConnectionFailure tests: Network connection failure -> code 1
func TestGherkinScenarioNetworkConnectionFailure(t *testing.T) {
	// Arrange - create the error using constructor
	err := apperrors.NewNetworkError(errors.New("connection refused"))

	// Act
	exitCode := apperrors.GetExitCode(err)

	// Assert
	if exitCode != 1 {
		t.Errorf("Gherkin: 'Network connection failure -> code 1' FAILED: got %d", exitCode)
	}
}

// TestGherkinScenarioGitHubAPIServerError tests: GitHub API server error -> code 1
func TestGherkinScenarioGitHubAPIServerError(t *testing.T) {
	// Arrange - create the error using constructor
	err := apperrors.NewAPIError("GitHub API returned 500", errors.New("HTTP 500"))

	// Act
	exitCode := apperrors.GetExitCode(err)

	// Assert
	if exitCode != 1 {
		t.Errorf("Gherkin: 'GitHub API server error -> code 1' FAILED: got %d", exitCode)
	}
}

// TestGherkinScenarioInvalidCommandLineArguments tests: Invalid command line arguments -> code 2
func TestGherkinScenarioInvalidCommandLineArguments(t *testing.T) {
	// Arrange - create the error using constructor
	err := apperrors.NewValidationError("invalid repository format")

	// Act
	exitCode := apperrors.GetExitCode(err)

	// Assert
	if exitCode != 2 {
		t.Errorf("Gherkin: 'Invalid command line arguments -> code 2' FAILED: got %d", exitCode)
	}
}

// =============================================================================
// ErrorCode to Exit Code Mapping Table Tests
// =============================================================================

// TestAllErrorCodesMappedToExitCodes verifies complete coverage of error codes.
func TestAllErrorCodesMappedToExitCodes(t *testing.T) {
	// Arrange - all defined error codes and their expected exit codes
	testCases := []struct {
		code         apperrors.ErrorCode
		expectedExit int
	}{
		{apperrors.ErrGeneral, 1},
		{apperrors.ErrInvalidArguments, 2},
		{apperrors.ErrAuthenticationFailed, 3},
		{apperrors.ErrInsufficientPerms, 4},
		{apperrors.ErrRepositoryNotFound, 5},
		{apperrors.ErrAPIRateLimited, 6},
	}

	for _, tc := range testCases {
		t.Run("ErrorCode_"+string(rune('0'+tc.expectedExit)), func(t *testing.T) {
			// Arrange
			appErr := &apperrors.AppError{
				Code:    tc.code,
				Message: "test error",
			}

			// Act
			exitCode := apperrors.GetExitCode(appErr)

			// Assert
			if exitCode != tc.expectedExit {
				t.Errorf("ErrorCode %d should map to exit code %d, got %d",
					tc.code, tc.expectedExit, exitCode)
			}
		})
	}
}

// =============================================================================
// Edge Cases Tests
// =============================================================================

// TestGetExitCodeWithZeroCode verifies behavior with zero ErrorCode.
func TestGetExitCodeWithZeroCode(t *testing.T) {
	// Arrange
	appErr := &apperrors.AppError{
		Code:    0, // Undefined/zero code
		Message: "test error",
	}

	// Act
	exitCode := apperrors.GetExitCode(appErr)

	// Assert - zero code should map to something sensible (likely 1 or 0)
	if exitCode < 0 {
		t.Errorf("GetExitCode() returned negative value: %d", exitCode)
	}
}

// TestGetExitCodeWithHighCode verifies behavior with undefined high ErrorCode.
func TestGetExitCodeWithHighCode(t *testing.T) {
	// Arrange
	appErr := &apperrors.AppError{
		Code:    apperrors.ErrorCode(999), // Undefined high code
		Message: "test error",
	}

	// Act
	exitCode := apperrors.GetExitCode(appErr)

	// Assert - undefined code should return the code value or default to 1
	if exitCode != 999 && exitCode != 1 {
		t.Errorf("GetExitCode() = %d, expected 999 or 1 for undefined code", exitCode)
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

// wrapError wraps an error with additional context (simulates fmt.Errorf wrapping).
type wrappedError struct {
	msg   string
	cause error
}

func (e *wrappedError) Error() string {
	return e.msg + ": " + e.cause.Error()
}

func (e *wrappedError) Unwrap() error {
	return e.cause
}

func wrapError(err error, msg string) error {
	return &wrappedError{msg: msg, cause: err}
}

// fixedResetTime returns a fixed time for deterministic testing.
func fixedResetTime() time.Time {
	return time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
}
