// Package errors_test provides tests for the AppError struct.
//
// These tests verify that AppError struct implements the error interface correctly
// and supports Go's error wrapping/unwrapping pattern.
//
// These tests are designed to FAIL until the implementations are properly created
// by the coder agent.
package errors_test

import (
	"errors"
	"testing"

	apperrors "github.com/josejulio/ghautodelete/internal/errors"
)

// =============================================================================
// AppError Struct Tests
// =============================================================================

// TestAppErrorImplementsError verifies AppError implements the error interface.
//
// The implementation should:
// - Define AppError struct in internal/errors/errors.go
// - Implement Error() method returning formatted message
func TestAppErrorImplementsError(t *testing.T) {
	// Arrange & Act - compile-time interface satisfaction check
	var err error = &apperrors.AppError{}

	// Assert
	if err == nil {
		t.Error("AppError should implement error interface")
	}
}

// TestAppErrorFieldsExist verifies AppError has required fields.
//
// The implementation should define AppError with:
// - Code field (ErrorCode type)
// - Message field (string)
// - Cause field (error, for wrapping)
func TestAppErrorFieldsExist(t *testing.T) {
	// Arrange - create AppError with all fields
	cause := errors.New("underlying error")
	appErr := apperrors.AppError{
		Code:    apperrors.ErrGeneral,
		Message: "test error message",
		Cause:   cause,
	}

	// Assert - verify all fields are accessible
	if appErr.Code != apperrors.ErrGeneral {
		t.Errorf("AppError.Code = %v, expected %v", appErr.Code, apperrors.ErrGeneral)
	}
	if appErr.Message != "test error message" {
		t.Errorf("AppError.Message = %q, expected %q", appErr.Message, "test error message")
	}
	if appErr.Cause != cause {
		t.Errorf("AppError.Cause = %v, expected %v", appErr.Cause, cause)
	}
}

// TestAppErrorErrorMethod verifies Error() returns formatted message.
//
// The implementation should:
// - Return formatted message including the error message
// - Optionally include cause information when present
func TestAppErrorErrorMethod(t *testing.T) {
	tests := []struct {
		name            string
		code            apperrors.ErrorCode
		message         string
		cause           error
		expectedContain string
	}{
		{
			name:            "returns message without cause",
			code:            apperrors.ErrGeneral,
			message:         "something went wrong",
			cause:           nil,
			expectedContain: "something went wrong",
		},
		{
			name:            "returns message with cause",
			code:            apperrors.ErrGeneral,
			message:         "operation failed",
			cause:           errors.New("connection refused"),
			expectedContain: "operation failed",
		},
		{
			name:            "authentication error message",
			code:            apperrors.ErrAuthenticationFailed,
			message:         "invalid token",
			cause:           nil,
			expectedContain: "invalid token",
		},
		{
			name:            "repository not found message",
			code:            apperrors.ErrRepositoryNotFound,
			message:         "repository owner/repo not found",
			cause:           nil,
			expectedContain: "repository owner/repo not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			appErr := &apperrors.AppError{
				Code:    tt.code,
				Message: tt.message,
				Cause:   tt.cause,
			}

			// Act
			result := appErr.Error()

			// Assert
			if result == "" {
				t.Error("Error() should not return empty string")
			}
			if !containsSubstring(result, tt.expectedContain) {
				t.Errorf("Error() = %q, expected to contain %q", result, tt.expectedContain)
			}
		})
	}
}

// TestAppErrorUnwrapMethod verifies Unwrap() returns the cause.
//
// The implementation should:
// - Return the Cause field from Unwrap()
// - Return nil when there is no cause
func TestAppErrorUnwrapMethod(t *testing.T) {
	tests := []struct {
		name          string
		cause         error
		expectedCause error
	}{
		{
			name:          "returns cause when present",
			cause:         errors.New("underlying error"),
			expectedCause: errors.New("underlying error"),
		},
		{
			name:          "returns nil when no cause",
			cause:         nil,
			expectedCause: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			appErr := &apperrors.AppError{
				Code:    apperrors.ErrGeneral,
				Message: "test error",
				Cause:   tt.cause,
			}

			// Act
			result := appErr.Unwrap()

			// Assert
			if tt.expectedCause == nil {
				if result != nil {
					t.Errorf("Unwrap() = %v, expected nil", result)
				}
			} else {
				if result == nil {
					t.Errorf("Unwrap() = nil, expected non-nil error")
				}
			}
		})
	}
}

// TestAppErrorUnwrapWorksWithErrorsIs verifies error chain works with errors.Is.
//
// The implementation should support standard Go error unwrapping.
func TestAppErrorUnwrapWorksWithErrorsIs(t *testing.T) {
	// Arrange
	sentinelErr := errors.New("sentinel error")
	appErr := &apperrors.AppError{
		Code:    apperrors.ErrGeneral,
		Message: "wrapper error",
		Cause:   sentinelErr,
	}

	// Act & Assert
	if !errors.Is(appErr, sentinelErr) {
		t.Error("errors.Is should find sentinel error in chain")
	}
}

// TestAppErrorCanBeAssertedFromError verifies type assertion works.
func TestAppErrorCanBeAssertedFromError(t *testing.T) {
	// Arrange
	var err error = apperrors.NewValidationError("test error")

	// Act
	appErr, ok := err.(*apperrors.AppError)

	// Assert
	if !ok {
		t.Fatal("Should be able to type assert error to *AppError")
	}
	if appErr == nil {
		t.Error("Type assertion result should not be nil")
	}
	if appErr.Code != apperrors.ErrInvalidArguments {
		t.Errorf("Code = %v, expected %v", appErr.Code, apperrors.ErrInvalidArguments)
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

// containsSubstring checks if s contains substr.
func containsSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
