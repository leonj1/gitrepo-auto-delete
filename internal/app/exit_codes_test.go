// Package app_test provides tests for exit code handling in the App struct.
//
// These tests verify that App.Run returns appropriate errors that map to correct
// exit codes according to the Gherkin scenarios:
//
// Feature: Exit Codes
//   Scenario: Exit code 0 on successful configuration
//   Scenario: Exit code 0 when already configured
//   Scenario: Exit code 0 on successful check
//   Scenario: Exit code 0 on successful dry-run
//   Scenario: Exit code 2 on invalid arguments
//   Scenario: Exit code 3 on authentication failure
//   Scenario: Exit code 4 on insufficient permissions
//   Scenario: Exit code 5 on repository not found
//   Scenario: Exit code 6 on API rate limit
//
// These tests are designed to FAIL until the implementations are properly created
// by the coder agent.
package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/josejulio/ghautodelete/internal/app"
	apperrors "github.com/josejulio/ghautodelete/internal/errors"
	"github.com/josejulio/ghautodelete/pkg/interfaces"
)

// =============================================================================
// Gherkin Scenario: Exit code 0 on successful configuration
// =============================================================================

// TestExitCode0OnSuccessfulConfiguration verifies exit code 0 for successful enable.
//
// Gherkin: Scenario: Exit code 0 on successful configuration
//
// The implementation should:
// - App.Run returns nil when configuration succeeds
// - GetExitCode(nil) returns 0
func TestExitCode0OnSuccessfulConfiguration(t *testing.T) {
	// Arrange
	mockWriter := &mockOutputWriter{}
	mockConfigSvc := &mockConfigService{
		ConfigureFunc: func(ctx context.Context, owner, name string, dryRun bool) (interfaces.IConfigResult, error) {
			// Success: feature was not enabled, now it is
			return newMockConfigResult(false, true, "main", "octocat/hello-world"), nil
		},
	}
	mockParser := &mockRepoParser{
		ParseFunc: func(repoIdentifier string) (string, string, error) {
			return "octocat", "hello-world", nil
		},
	}

	application := app.NewApp(mockWriter, mockConfigSvc, mockParser)
	ctx := context.Background()
	opts := interfaces.CLIOptions{
		Repository: "octocat/hello-world",
		DryRun:     false,
		CheckOnly:  false,
	}

	// Act
	err := application.Run(ctx, opts)

	// Assert - Run returns nil on success
	if err != nil {
		t.Errorf("App.Run() should return nil on successful configuration, got error: %v", err)
	}

	// Assert - GetExitCode maps nil to 0
	exitCode := apperrors.GetExitCode(err)
	if exitCode != 0 {
		t.Errorf("GetExitCode(nil) = %d, expected 0 for successful configuration", exitCode)
	}
}

// =============================================================================
// Gherkin Scenario: Exit code 0 when already configured
// =============================================================================

// TestExitCode0WhenAlreadyConfigured verifies exit code 0 when already enabled.
//
// Gherkin: Scenario: Exit code 0 when already configured
//
// The implementation should:
// - App.Run returns nil when feature was already enabled
// - GetExitCode(nil) returns 0
func TestExitCode0WhenAlreadyConfigured(t *testing.T) {
	// Arrange
	mockWriter := &mockOutputWriter{}
	mockConfigSvc := &mockConfigService{
		ConfigureFunc: func(ctx context.Context, owner, name string, dryRun bool) (interfaces.IConfigResult, error) {
			// Success: feature was already enabled
			return newMockConfigResult(true, true, "main", "octocat/hello-world"), nil
		},
	}
	mockParser := &mockRepoParser{
		ParseFunc: func(repoIdentifier string) (string, string, error) {
			return "octocat", "hello-world", nil
		},
	}

	application := app.NewApp(mockWriter, mockConfigSvc, mockParser)
	ctx := context.Background()
	opts := interfaces.CLIOptions{
		Repository: "octocat/hello-world",
		DryRun:     false,
		CheckOnly:  false,
	}

	// Act
	err := application.Run(ctx, opts)

	// Assert - Run returns nil when already configured
	if err != nil {
		t.Errorf("App.Run() should return nil when already configured, got error: %v", err)
	}

	// Assert - GetExitCode maps nil to 0
	exitCode := apperrors.GetExitCode(err)
	if exitCode != 0 {
		t.Errorf("GetExitCode(nil) = %d, expected 0 when already configured", exitCode)
	}
}

// =============================================================================
// Gherkin Scenario: Exit code 0 on successful check
// =============================================================================

// TestExitCode0OnSuccessfulCheck verifies exit code 0 for successful status check.
//
// Gherkin: Scenario: Exit code 0 on successful check
//
// The implementation should:
// - App.Run returns nil when check mode succeeds
// - GetExitCode(nil) returns 0
func TestExitCode0OnSuccessfulCheck(t *testing.T) {
	// Arrange
	mockWriter := &mockOutputWriter{}
	mockConfigSvc := &mockConfigService{
		CheckStatusFunc: func(ctx context.Context, owner, name string) (interfaces.IConfigResult, error) {
			return newMockConfigResult(false, false, "main", "octocat/hello-world"), nil
		},
	}
	mockParser := &mockRepoParser{
		ParseFunc: func(repoIdentifier string) (string, string, error) {
			return "octocat", "hello-world", nil
		},
	}

	application := app.NewApp(mockWriter, mockConfigSvc, mockParser)
	ctx := context.Background()
	opts := interfaces.CLIOptions{
		Repository: "octocat/hello-world",
		CheckOnly:  true,
	}

	// Act
	err := application.Run(ctx, opts)

	// Assert - Run returns nil on successful check
	if err != nil {
		t.Errorf("App.Run() should return nil on successful check, got error: %v", err)
	}

	// Assert - GetExitCode maps nil to 0
	exitCode := apperrors.GetExitCode(err)
	if exitCode != 0 {
		t.Errorf("GetExitCode(nil) = %d, expected 0 for successful check", exitCode)
	}
}

// TestExitCode0OnSuccessfulCheckWhenEnabled verifies exit code 0 for check when enabled.
//
// Gherkin: Scenario: Exit code 0 on successful check (when feature is enabled)
//
// The implementation should:
// - App.Run returns nil regardless of current enabled state
// - GetExitCode(nil) returns 0
func TestExitCode0OnSuccessfulCheckWhenEnabled(t *testing.T) {
	// Arrange
	mockWriter := &mockOutputWriter{}
	mockConfigSvc := &mockConfigService{
		CheckStatusFunc: func(ctx context.Context, owner, name string) (interfaces.IConfigResult, error) {
			return newMockConfigResult(true, true, "main", "octocat/hello-world"), nil
		},
	}
	mockParser := &mockRepoParser{
		ParseFunc: func(repoIdentifier string) (string, string, error) {
			return "octocat", "hello-world", nil
		},
	}

	application := app.NewApp(mockWriter, mockConfigSvc, mockParser)
	ctx := context.Background()
	opts := interfaces.CLIOptions{
		Repository: "octocat/hello-world",
		CheckOnly:  true,
	}

	// Act
	err := application.Run(ctx, opts)

	// Assert
	if err != nil {
		t.Errorf("App.Run() should return nil on successful check, got error: %v", err)
	}

	exitCode := apperrors.GetExitCode(err)
	if exitCode != 0 {
		t.Errorf("GetExitCode(nil) = %d, expected 0 for successful check", exitCode)
	}
}

// =============================================================================
// Gherkin Scenario: Exit code 0 on successful dry-run
// =============================================================================

// TestExitCode0OnSuccessfulDryRun verifies exit code 0 for successful dry-run.
//
// Gherkin: Scenario: Exit code 0 on successful dry-run
//
// The implementation should:
// - App.Run returns nil when dry-run succeeds
// - GetExitCode(nil) returns 0
func TestExitCode0OnSuccessfulDryRun(t *testing.T) {
	// Arrange
	mockWriter := &mockOutputWriter{}
	mockConfigSvc := &mockConfigService{
		ConfigureFunc: func(ctx context.Context, owner, name string, dryRun bool) (interfaces.IConfigResult, error) {
			// Dry-run success: feature not enabled, would be enabled
			return newMockConfigResult(false, false, "main", "octocat/hello-world"), nil
		},
	}
	mockParser := &mockRepoParser{
		ParseFunc: func(repoIdentifier string) (string, string, error) {
			return "octocat", "hello-world", nil
		},
	}

	application := app.NewApp(mockWriter, mockConfigSvc, mockParser)
	ctx := context.Background()
	opts := interfaces.CLIOptions{
		Repository: "octocat/hello-world",
		DryRun:     true,
		CheckOnly:  false,
	}

	// Act
	err := application.Run(ctx, opts)

	// Assert - Run returns nil on successful dry-run
	if err != nil {
		t.Errorf("App.Run() should return nil on successful dry-run, got error: %v", err)
	}

	// Assert - GetExitCode maps nil to 0
	exitCode := apperrors.GetExitCode(err)
	if exitCode != 0 {
		t.Errorf("GetExitCode(nil) = %d, expected 0 for successful dry-run", exitCode)
	}
}

// TestExitCode0OnSuccessfulDryRunWhenAlreadyEnabled verifies exit code 0 for dry-run when already enabled.
//
// Gherkin: Scenario: Exit code 0 on successful dry-run (when already enabled)
//
// The implementation should:
// - App.Run returns nil even when feature is already enabled
// - GetExitCode(nil) returns 0
func TestExitCode0OnSuccessfulDryRunWhenAlreadyEnabled(t *testing.T) {
	// Arrange
	mockWriter := &mockOutputWriter{}
	mockConfigSvc := &mockConfigService{
		ConfigureFunc: func(ctx context.Context, owner, name string, dryRun bool) (interfaces.IConfigResult, error) {
			// Dry-run success: feature already enabled
			return newMockConfigResult(true, true, "main", "octocat/hello-world"), nil
		},
	}
	mockParser := &mockRepoParser{
		ParseFunc: func(repoIdentifier string) (string, string, error) {
			return "octocat", "hello-world", nil
		},
	}

	application := app.NewApp(mockWriter, mockConfigSvc, mockParser)
	ctx := context.Background()
	opts := interfaces.CLIOptions{
		Repository: "octocat/hello-world",
		DryRun:     true,
		CheckOnly:  false,
	}

	// Act
	err := application.Run(ctx, opts)

	// Assert
	if err != nil {
		t.Errorf("App.Run() should return nil on successful dry-run, got error: %v", err)
	}

	exitCode := apperrors.GetExitCode(err)
	if exitCode != 0 {
		t.Errorf("GetExitCode(nil) = %d, expected 0 for successful dry-run", exitCode)
	}
}

// =============================================================================
// Gherkin Scenario: Exit code 2 on invalid arguments
// =============================================================================

// TestExitCode2OnInvalidRepositoryFormat verifies exit code 2 for invalid repository format.
//
// Gherkin: Scenario: Exit code 2 on invalid arguments
//
// The implementation should:
// - App.Run returns error when repository parsing fails
// - If the parser returns a validation error (AppError with ErrInvalidArguments),
//   GetExitCode should return 2
func TestExitCode2OnInvalidRepositoryFormat(t *testing.T) {
	// Arrange
	mockWriter := &mockOutputWriter{}
	mockConfigSvc := &mockConfigService{}
	mockParser := &mockRepoParser{
		ParseFunc: func(repoIdentifier string) (string, string, error) {
			// Return a validation error (exit code 2)
			return "", "", apperrors.NewValidationError("invalid repository format: expected 'owner/repo'")
		},
	}

	application := app.NewApp(mockWriter, mockConfigSvc, mockParser)
	ctx := context.Background()
	opts := interfaces.CLIOptions{
		Repository: "invalid-format",
		CheckOnly:  true,
	}

	// Act
	err := application.Run(ctx, opts)

	// Assert - Run returns error
	if err == nil {
		t.Fatal("App.Run() should return error for invalid repository format")
	}

	// Assert - GetExitCode maps to 2 (invalid arguments)
	exitCode := apperrors.GetExitCode(err)
	if exitCode != 2 {
		t.Errorf("GetExitCode() = %d, expected 2 for invalid arguments", exitCode)
	}
}

// TestExitCode2OnEmptyRepository verifies exit code 2 for empty repository identifier.
//
// Gherkin: Scenario: Exit code 2 on invalid arguments (empty repository)
//
// The implementation should:
// - App.Run returns validation error for empty repository
// - GetExitCode returns 2
func TestExitCode2OnEmptyRepository(t *testing.T) {
	// Arrange
	mockWriter := &mockOutputWriter{}
	mockConfigSvc := &mockConfigService{}
	mockParser := &mockRepoParser{
		ParseFunc: func(repoIdentifier string) (string, string, error) {
			if repoIdentifier == "" {
				return "", "", apperrors.NewValidationError("repository identifier is required")
			}
			return "owner", "repo", nil
		},
	}

	application := app.NewApp(mockWriter, mockConfigSvc, mockParser)
	ctx := context.Background()
	opts := interfaces.CLIOptions{
		Repository: "",
		CheckOnly:  true,
	}

	// Act
	err := application.Run(ctx, opts)

	// Assert
	if err == nil {
		t.Fatal("App.Run() should return error for empty repository")
	}

	exitCode := apperrors.GetExitCode(err)
	if exitCode != 2 {
		t.Errorf("GetExitCode() = %d, expected 2 for empty repository", exitCode)
	}
}

// TestExitCode2OnMissingOwnerSlash verifies exit code 2 for repo without owner/slash.
//
// Gherkin: Scenario: Exit code 2 on invalid arguments (missing slash)
//
// The implementation should:
// - App.Run returns validation error for missing slash
// - GetExitCode returns 2
func TestExitCode2OnMissingOwnerSlash(t *testing.T) {
	// Arrange
	mockWriter := &mockOutputWriter{}
	mockConfigSvc := &mockConfigService{}
	mockParser := &mockRepoParser{
		ParseFunc: func(repoIdentifier string) (string, string, error) {
			return "", "", apperrors.NewValidationError("invalid repository format: missing '/' separator")
		},
	}

	application := app.NewApp(mockWriter, mockConfigSvc, mockParser)
	ctx := context.Background()
	opts := interfaces.CLIOptions{
		Repository: "justreponame",
		CheckOnly:  true,
	}

	// Act
	err := application.Run(ctx, opts)

	// Assert
	if err == nil {
		t.Fatal("App.Run() should return error for invalid repository format")
	}

	exitCode := apperrors.GetExitCode(err)
	if exitCode != 2 {
		t.Errorf("GetExitCode() = %d, expected 2 for invalid arguments", exitCode)
	}
}

// =============================================================================
// Gherkin Scenario: Exit code 3 on authentication failure
// =============================================================================

// TestExitCode3OnAuthenticationFailure verifies exit code 3 for auth failures.
//
// Gherkin: Scenario: Exit code 3 on authentication failure
//
// The implementation should:
// - App.Run returns AppError with ErrAuthenticationFailed when auth fails
// - GetExitCode returns 3
func TestExitCode3OnAuthenticationFailure(t *testing.T) {
	// Arrange
	mockWriter := &mockOutputWriter{}
	mockConfigSvc := &mockConfigService{
		ConfigureFunc: func(ctx context.Context, owner, name string, dryRun bool) (interfaces.IConfigResult, error) {
			// Return authentication error (exit code 3)
			return nil, apperrors.NewAuthenticationError("token validation failed", errors.New("401 Unauthorized"))
		},
	}
	mockParser := &mockRepoParser{
		ParseFunc: func(repoIdentifier string) (string, string, error) {
			return "octocat", "hello-world", nil
		},
	}

	application := app.NewApp(mockWriter, mockConfigSvc, mockParser)
	ctx := context.Background()
	opts := interfaces.CLIOptions{
		Repository: "octocat/hello-world",
		DryRun:     false,
		CheckOnly:  false,
	}

	// Act
	err := application.Run(ctx, opts)

	// Assert - Run returns error
	if err == nil {
		t.Fatal("App.Run() should return error on authentication failure")
	}

	// Assert - GetExitCode maps to 3 (authentication failed)
	exitCode := apperrors.GetExitCode(err)
	if exitCode != 3 {
		t.Errorf("GetExitCode() = %d, expected 3 for authentication failure", exitCode)
	}
}

// TestExitCode3OnInvalidToken verifies exit code 3 for invalid token.
//
// Gherkin: Scenario: Exit code 3 on authentication failure (invalid token)
//
// The implementation should:
// - App.Run returns authentication error for invalid/revoked token
// - GetExitCode returns 3
func TestExitCode3OnInvalidToken(t *testing.T) {
	// Arrange
	mockWriter := &mockOutputWriter{}
	mockConfigSvc := &mockConfigService{
		CheckStatusFunc: func(ctx context.Context, owner, name string) (interfaces.IConfigResult, error) {
			return nil, apperrors.NewAuthenticationError("invalid token", errors.New("bad credentials"))
		},
	}
	mockParser := &mockRepoParser{
		ParseFunc: func(repoIdentifier string) (string, string, error) {
			return "octocat", "hello-world", nil
		},
	}

	application := app.NewApp(mockWriter, mockConfigSvc, mockParser)
	ctx := context.Background()
	opts := interfaces.CLIOptions{
		Repository: "octocat/hello-world",
		CheckOnly:  true,
	}

	// Act
	err := application.Run(ctx, opts)

	// Assert
	if err == nil {
		t.Fatal("App.Run() should return error on invalid token")
	}

	exitCode := apperrors.GetExitCode(err)
	if exitCode != 3 {
		t.Errorf("GetExitCode() = %d, expected 3 for invalid token", exitCode)
	}
}

// TestExitCode3OnExpiredToken verifies exit code 3 for expired token.
//
// Gherkin: Scenario: Exit code 3 on authentication failure (expired token)
//
// The implementation should:
// - App.Run returns authentication error for expired token
// - GetExitCode returns 3
func TestExitCode3OnExpiredToken(t *testing.T) {
	// Arrange
	mockWriter := &mockOutputWriter{}
	mockConfigSvc := &mockConfigService{
		ConfigureFunc: func(ctx context.Context, owner, name string, dryRun bool) (interfaces.IConfigResult, error) {
			return nil, apperrors.NewAuthenticationError("token has expired", nil)
		},
	}
	mockParser := &mockRepoParser{
		ParseFunc: func(repoIdentifier string) (string, string, error) {
			return "octocat", "hello-world", nil
		},
	}

	application := app.NewApp(mockWriter, mockConfigSvc, mockParser)
	ctx := context.Background()
	opts := interfaces.CLIOptions{
		Repository: "octocat/hello-world",
		DryRun:     true,
	}

	// Act
	err := application.Run(ctx, opts)

	// Assert
	if err == nil {
		t.Fatal("App.Run() should return error on expired token")
	}

	exitCode := apperrors.GetExitCode(err)
	if exitCode != 3 {
		t.Errorf("GetExitCode() = %d, expected 3 for expired token", exitCode)
	}
}

// =============================================================================
// Gherkin Scenario: Exit code 4 on insufficient permissions
// =============================================================================

// TestExitCode4OnInsufficientPermissions verifies exit code 4 for permission errors.
//
// Gherkin: Scenario: Exit code 4 on insufficient permissions
//
// The implementation should:
// - App.Run returns AppError with ErrInsufficientPerms when lacking permissions
// - GetExitCode returns 4
func TestExitCode4OnInsufficientPermissions(t *testing.T) {
	// Arrange
	mockWriter := &mockOutputWriter{}
	mockConfigSvc := &mockConfigService{
		ConfigureFunc: func(ctx context.Context, owner, name string, dryRun bool) (interfaces.IConfigResult, error) {
			// Return authorization error (exit code 4)
			return nil, apperrors.NewAuthorizationError("token lacks 'repo' scope required to update repository settings")
		},
	}
	mockParser := &mockRepoParser{
		ParseFunc: func(repoIdentifier string) (string, string, error) {
			return "octocat", "hello-world", nil
		},
	}

	application := app.NewApp(mockWriter, mockConfigSvc, mockParser)
	ctx := context.Background()
	opts := interfaces.CLIOptions{
		Repository: "octocat/hello-world",
		DryRun:     false,
		CheckOnly:  false,
	}

	// Act
	err := application.Run(ctx, opts)

	// Assert - Run returns error
	if err == nil {
		t.Fatal("App.Run() should return error on insufficient permissions")
	}

	// Assert - GetExitCode maps to 4 (insufficient permissions)
	exitCode := apperrors.GetExitCode(err)
	if exitCode != 4 {
		t.Errorf("GetExitCode() = %d, expected 4 for insufficient permissions", exitCode)
	}
}

// TestExitCode4OnMissingRepoScope verifies exit code 4 when repo scope is missing.
//
// Gherkin: Scenario: Exit code 4 on insufficient permissions (missing repo scope)
//
// The implementation should:
// - App.Run returns authorization error when token lacks repo scope
// - GetExitCode returns 4
func TestExitCode4OnMissingRepoScope(t *testing.T) {
	// Arrange
	mockWriter := &mockOutputWriter{}
	mockConfigSvc := &mockConfigService{
		ConfigureFunc: func(ctx context.Context, owner, name string, dryRun bool) (interfaces.IConfigResult, error) {
			return nil, apperrors.NewAuthorizationError("missing required scope: repo")
		},
	}
	mockParser := &mockRepoParser{
		ParseFunc: func(repoIdentifier string) (string, string, error) {
			return "octocat", "private-repo", nil
		},
	}

	application := app.NewApp(mockWriter, mockConfigSvc, mockParser)
	ctx := context.Background()
	opts := interfaces.CLIOptions{
		Repository: "octocat/private-repo",
	}

	// Act
	err := application.Run(ctx, opts)

	// Assert
	if err == nil {
		t.Fatal("App.Run() should return error on missing repo scope")
	}

	exitCode := apperrors.GetExitCode(err)
	if exitCode != 4 {
		t.Errorf("GetExitCode() = %d, expected 4 for missing repo scope", exitCode)
	}
}

// TestExitCode4OnReadOnlyAccess verifies exit code 4 for read-only access.
//
// Gherkin: Scenario: Exit code 4 on insufficient permissions (read-only access)
//
// The implementation should:
// - App.Run returns authorization error for read-only access
// - GetExitCode returns 4
func TestExitCode4OnReadOnlyAccess(t *testing.T) {
	// Arrange
	mockWriter := &mockOutputWriter{}
	mockConfigSvc := &mockConfigService{
		ConfigureFunc: func(ctx context.Context, owner, name string, dryRun bool) (interfaces.IConfigResult, error) {
			return nil, apperrors.NewAuthorizationError("user has read-only access to repository")
		},
	}
	mockParser := &mockRepoParser{
		ParseFunc: func(repoIdentifier string) (string, string, error) {
			return "octocat", "hello-world", nil
		},
	}

	application := app.NewApp(mockWriter, mockConfigSvc, mockParser)
	ctx := context.Background()
	opts := interfaces.CLIOptions{
		Repository: "octocat/hello-world",
	}

	// Act
	err := application.Run(ctx, opts)

	// Assert
	if err == nil {
		t.Fatal("App.Run() should return error on read-only access")
	}

	exitCode := apperrors.GetExitCode(err)
	if exitCode != 4 {
		t.Errorf("GetExitCode() = %d, expected 4 for read-only access", exitCode)
	}
}

// =============================================================================
// Gherkin Scenario: Exit code 5 on repository not found
// =============================================================================

// TestExitCode5OnRepositoryNotFound verifies exit code 5 for repo not found.
//
// Gherkin: Scenario: Exit code 5 on repository not found
//
// The implementation should:
// - App.Run returns AppError with ErrRepositoryNotFound when repo doesn't exist
// - GetExitCode returns 5
func TestExitCode5OnRepositoryNotFound(t *testing.T) {
	// Arrange
	mockWriter := &mockOutputWriter{}
	mockConfigSvc := &mockConfigService{
		ConfigureFunc: func(ctx context.Context, owner, name string, dryRun bool) (interfaces.IConfigResult, error) {
			// Return repository not found error (exit code 5)
			return nil, apperrors.NewRepositoryNotFoundError(owner, name)
		},
	}
	mockParser := &mockRepoParser{
		ParseFunc: func(repoIdentifier string) (string, string, error) {
			return "octocat", "nonexistent-repo", nil
		},
	}

	application := app.NewApp(mockWriter, mockConfigSvc, mockParser)
	ctx := context.Background()
	opts := interfaces.CLIOptions{
		Repository: "octocat/nonexistent-repo",
		DryRun:     false,
		CheckOnly:  false,
	}

	// Act
	err := application.Run(ctx, opts)

	// Assert - Run returns error
	if err == nil {
		t.Fatal("App.Run() should return error when repository not found")
	}

	// Assert - GetExitCode maps to 5 (repository not found)
	exitCode := apperrors.GetExitCode(err)
	if exitCode != 5 {
		t.Errorf("GetExitCode() = %d, expected 5 for repository not found", exitCode)
	}
}

// TestExitCode5OnRepositoryNotFoundInCheckMode verifies exit code 5 in check mode.
//
// Gherkin: Scenario: Exit code 5 on repository not found (check mode)
//
// The implementation should:
// - App.Run returns repository not found error even in check mode
// - GetExitCode returns 5
func TestExitCode5OnRepositoryNotFoundInCheckMode(t *testing.T) {
	// Arrange
	mockWriter := &mockOutputWriter{}
	mockConfigSvc := &mockConfigService{
		CheckStatusFunc: func(ctx context.Context, owner, name string) (interfaces.IConfigResult, error) {
			return nil, apperrors.NewRepositoryNotFoundError(owner, name)
		},
	}
	mockParser := &mockRepoParser{
		ParseFunc: func(repoIdentifier string) (string, string, error) {
			return "octocat", "deleted-repo", nil
		},
	}

	application := app.NewApp(mockWriter, mockConfigSvc, mockParser)
	ctx := context.Background()
	opts := interfaces.CLIOptions{
		Repository: "octocat/deleted-repo",
		CheckOnly:  true,
	}

	// Act
	err := application.Run(ctx, opts)

	// Assert
	if err == nil {
		t.Fatal("App.Run() should return error when repository not found in check mode")
	}

	exitCode := apperrors.GetExitCode(err)
	if exitCode != 5 {
		t.Errorf("GetExitCode() = %d, expected 5 for repository not found", exitCode)
	}
}

// TestExitCode5OnRepositoryNotFoundInDryRunMode verifies exit code 5 in dry-run mode.
//
// Gherkin: Scenario: Exit code 5 on repository not found (dry-run mode)
//
// The implementation should:
// - App.Run returns repository not found error even in dry-run mode
// - GetExitCode returns 5
func TestExitCode5OnRepositoryNotFoundInDryRunMode(t *testing.T) {
	// Arrange
	mockWriter := &mockOutputWriter{}
	mockConfigSvc := &mockConfigService{
		ConfigureFunc: func(ctx context.Context, owner, name string, dryRun bool) (interfaces.IConfigResult, error) {
			return nil, apperrors.NewRepositoryNotFoundError(owner, name)
		},
	}
	mockParser := &mockRepoParser{
		ParseFunc: func(repoIdentifier string) (string, string, error) {
			return "octocat", "ghost-repo", nil
		},
	}

	application := app.NewApp(mockWriter, mockConfigSvc, mockParser)
	ctx := context.Background()
	opts := interfaces.CLIOptions{
		Repository: "octocat/ghost-repo",
		DryRun:     true,
	}

	// Act
	err := application.Run(ctx, opts)

	// Assert
	if err == nil {
		t.Fatal("App.Run() should return error when repository not found in dry-run mode")
	}

	exitCode := apperrors.GetExitCode(err)
	if exitCode != 5 {
		t.Errorf("GetExitCode() = %d, expected 5 for repository not found", exitCode)
	}
}

// =============================================================================
// Gherkin Scenario: Exit code 6 on API rate limit
// =============================================================================

// TestExitCode6OnAPIRateLimit verifies exit code 6 for rate limit exceeded.
//
// Gherkin: Scenario: Exit code 6 on API rate limit
//
// The implementation should:
// - App.Run returns AppError with ErrAPIRateLimited when rate limited
// - GetExitCode returns 6
func TestExitCode6OnAPIRateLimit(t *testing.T) {
	// Arrange
	resetTime := time.Now().Add(1 * time.Hour)
	mockWriter := &mockOutputWriter{}
	mockConfigSvc := &mockConfigService{
		ConfigureFunc: func(ctx context.Context, owner, name string, dryRun bool) (interfaces.IConfigResult, error) {
			// Return rate limit error (exit code 6)
			return nil, apperrors.NewRateLimitError(resetTime)
		},
	}
	mockParser := &mockRepoParser{
		ParseFunc: func(repoIdentifier string) (string, string, error) {
			return "octocat", "hello-world", nil
		},
	}

	application := app.NewApp(mockWriter, mockConfigSvc, mockParser)
	ctx := context.Background()
	opts := interfaces.CLIOptions{
		Repository: "octocat/hello-world",
		DryRun:     false,
		CheckOnly:  false,
	}

	// Act
	err := application.Run(ctx, opts)

	// Assert - Run returns error
	if err == nil {
		t.Fatal("App.Run() should return error on API rate limit")
	}

	// Assert - GetExitCode maps to 6 (rate limited)
	exitCode := apperrors.GetExitCode(err)
	if exitCode != 6 {
		t.Errorf("GetExitCode() = %d, expected 6 for API rate limit", exitCode)
	}
}

// TestExitCode6OnAPIRateLimitInCheckMode verifies exit code 6 in check mode.
//
// Gherkin: Scenario: Exit code 6 on API rate limit (check mode)
//
// The implementation should:
// - App.Run returns rate limit error even in check mode
// - GetExitCode returns 6
func TestExitCode6OnAPIRateLimitInCheckMode(t *testing.T) {
	// Arrange
	resetTime := time.Now().Add(30 * time.Minute)
	mockWriter := &mockOutputWriter{}
	mockConfigSvc := &mockConfigService{
		CheckStatusFunc: func(ctx context.Context, owner, name string) (interfaces.IConfigResult, error) {
			return nil, apperrors.NewRateLimitError(resetTime)
		},
	}
	mockParser := &mockRepoParser{
		ParseFunc: func(repoIdentifier string) (string, string, error) {
			return "octocat", "hello-world", nil
		},
	}

	application := app.NewApp(mockWriter, mockConfigSvc, mockParser)
	ctx := context.Background()
	opts := interfaces.CLIOptions{
		Repository: "octocat/hello-world",
		CheckOnly:  true,
	}

	// Act
	err := application.Run(ctx, opts)

	// Assert
	if err == nil {
		t.Fatal("App.Run() should return error on API rate limit in check mode")
	}

	exitCode := apperrors.GetExitCode(err)
	if exitCode != 6 {
		t.Errorf("GetExitCode() = %d, expected 6 for API rate limit", exitCode)
	}
}

// TestExitCode6OnAPIRateLimitInDryRunMode verifies exit code 6 in dry-run mode.
//
// Gherkin: Scenario: Exit code 6 on API rate limit (dry-run mode)
//
// The implementation should:
// - App.Run returns rate limit error even in dry-run mode
// - GetExitCode returns 6
func TestExitCode6OnAPIRateLimitInDryRunMode(t *testing.T) {
	// Arrange
	resetTime := time.Now().Add(15 * time.Minute)
	mockWriter := &mockOutputWriter{}
	mockConfigSvc := &mockConfigService{
		ConfigureFunc: func(ctx context.Context, owner, name string, dryRun bool) (interfaces.IConfigResult, error) {
			return nil, apperrors.NewRateLimitError(resetTime)
		},
	}
	mockParser := &mockRepoParser{
		ParseFunc: func(repoIdentifier string) (string, string, error) {
			return "octocat", "hello-world", nil
		},
	}

	application := app.NewApp(mockWriter, mockConfigSvc, mockParser)
	ctx := context.Background()
	opts := interfaces.CLIOptions{
		Repository: "octocat/hello-world",
		DryRun:     true,
	}

	// Act
	err := application.Run(ctx, opts)

	// Assert
	if err == nil {
		t.Fatal("App.Run() should return error on API rate limit in dry-run mode")
	}

	exitCode := apperrors.GetExitCode(err)
	if exitCode != 6 {
		t.Errorf("GetExitCode() = %d, expected 6 for API rate limit", exitCode)
	}
}

// =============================================================================
// Table-Driven Exit Code Tests
// =============================================================================

// TestExitCodesTableDriven verifies all exit code mappings in a single test.
//
// The implementation should:
// - Map each error type to its correct exit code
// - Map nil to exit code 0
// - Map non-AppError to exit code 1
func TestExitCodesTableDriven(t *testing.T) {
	resetTime := time.Now().Add(1 * time.Hour)

	tests := []struct {
		name             string
		err              error
		expectedExitCode int
		gherkinScenario  string
	}{
		{
			name:             "nil error maps to exit code 0",
			err:              nil,
			expectedExitCode: 0,
			gherkinScenario:  "Exit code 0 on success",
		},
		{
			name:             "validation error maps to exit code 2",
			err:              apperrors.NewValidationError("invalid arguments"),
			expectedExitCode: 2,
			gherkinScenario:  "Exit code 2 on invalid arguments",
		},
		{
			name:             "authentication error maps to exit code 3",
			err:              apperrors.NewAuthenticationError("auth failed", nil),
			expectedExitCode: 3,
			gherkinScenario:  "Exit code 3 on authentication failure",
		},
		{
			name:             "authorization error maps to exit code 4",
			err:              apperrors.NewAuthorizationError("insufficient permissions"),
			expectedExitCode: 4,
			gherkinScenario:  "Exit code 4 on insufficient permissions",
		},
		{
			name:             "repository not found maps to exit code 5",
			err:              apperrors.NewRepositoryNotFoundError("owner", "repo"),
			expectedExitCode: 5,
			gherkinScenario:  "Exit code 5 on repository not found",
		},
		{
			name:             "rate limit error maps to exit code 6",
			err:              apperrors.NewRateLimitError(resetTime),
			expectedExitCode: 6,
			gherkinScenario:  "Exit code 6 on API rate limit",
		},
		{
			name:             "network error maps to exit code 1",
			err:              apperrors.NewNetworkError(errors.New("connection refused")),
			expectedExitCode: 1,
			gherkinScenario:  "Exit code 1 on network error (general error)",
		},
		{
			name:             "API error maps to exit code 1",
			err:              apperrors.NewAPIError("server error", errors.New("500")),
			expectedExitCode: 1,
			gherkinScenario:  "Exit code 1 on API server error (general error)",
		},
		{
			name:             "standard error maps to exit code 1",
			err:              errors.New("some unexpected error"),
			expectedExitCode: 1,
			gherkinScenario:  "Exit code 1 on general error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			exitCode := apperrors.GetExitCode(tt.err)

			// Assert
			if exitCode != tt.expectedExitCode {
				t.Errorf("GetExitCode() = %d, expected %d (Gherkin: %s)",
					exitCode, tt.expectedExitCode, tt.gherkinScenario)
			}
		})
	}
}

// =============================================================================
// Integration Tests - Full Workflow Exit Codes
// =============================================================================

// TestExitCodeIntegrationSuccessScenarios verifies exit codes for all success scenarios.
//
// The implementation should:
// - Return exit code 0 for all success scenarios
func TestExitCodeIntegrationSuccessScenarios(t *testing.T) {
	tests := []struct {
		name           string
		checkOnly      bool
		dryRun         bool
		alreadyEnabled bool
		nowEnabled     bool
		gherkinCase    string
	}{
		{
			name:           "successful configuration",
			checkOnly:      false,
			dryRun:         false,
			alreadyEnabled: false,
			nowEnabled:     true,
			gherkinCase:    "Exit code 0 on successful configuration",
		},
		{
			name:           "already configured",
			checkOnly:      false,
			dryRun:         false,
			alreadyEnabled: true,
			nowEnabled:     true,
			gherkinCase:    "Exit code 0 when already configured",
		},
		{
			name:           "successful check (disabled)",
			checkOnly:      true,
			dryRun:         false,
			alreadyEnabled: false,
			nowEnabled:     false,
			gherkinCase:    "Exit code 0 on successful check",
		},
		{
			name:           "successful check (enabled)",
			checkOnly:      true,
			dryRun:         false,
			alreadyEnabled: true,
			nowEnabled:     true,
			gherkinCase:    "Exit code 0 on successful check",
		},
		{
			name:           "successful dry-run (would enable)",
			checkOnly:      false,
			dryRun:         true,
			alreadyEnabled: false,
			nowEnabled:     false,
			gherkinCase:    "Exit code 0 on successful dry-run",
		},
		{
			name:           "successful dry-run (already enabled)",
			checkOnly:      false,
			dryRun:         true,
			alreadyEnabled: true,
			nowEnabled:     true,
			gherkinCase:    "Exit code 0 on successful dry-run",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockWriter := &mockOutputWriter{}
			mockConfigSvc := &mockConfigService{
				CheckStatusFunc: func(ctx context.Context, owner, name string) (interfaces.IConfigResult, error) {
					return newMockConfigResult(tt.alreadyEnabled, tt.nowEnabled, "main", "owner/repo"), nil
				},
				ConfigureFunc: func(ctx context.Context, owner, name string, dryRun bool) (interfaces.IConfigResult, error) {
					return newMockConfigResult(tt.alreadyEnabled, tt.nowEnabled, "main", "owner/repo"), nil
				},
			}
			mockParser := &mockRepoParser{
				ParseFunc: func(repoIdentifier string) (string, string, error) {
					return "owner", "repo", nil
				},
			}

			application := app.NewApp(mockWriter, mockConfigSvc, mockParser)
			ctx := context.Background()
			opts := interfaces.CLIOptions{
				Repository: "owner/repo",
				CheckOnly:  tt.checkOnly,
				DryRun:     tt.dryRun,
			}

			// Act
			err := application.Run(ctx, opts)
			exitCode := apperrors.GetExitCode(err)

			// Assert
			if exitCode != 0 {
				t.Errorf("Exit code = %d, expected 0 (%s)", exitCode, tt.gherkinCase)
			}
		})
	}
}

// TestExitCodeIntegrationErrorScenarios verifies exit codes for all error scenarios.
//
// The implementation should:
// - Return correct exit codes for each error type
func TestExitCodeIntegrationErrorScenarios(t *testing.T) {
	resetTime := time.Now().Add(1 * time.Hour)

	tests := []struct {
		name             string
		parseError       error
		serviceError     error
		expectedExitCode int
		gherkinCase      string
	}{
		{
			name:             "invalid arguments",
			parseError:       apperrors.NewValidationError("invalid format"),
			serviceError:     nil,
			expectedExitCode: 2,
			gherkinCase:      "Exit code 2 on invalid arguments",
		},
		{
			name:             "authentication failure",
			parseError:       nil,
			serviceError:     apperrors.NewAuthenticationError("bad token", nil),
			expectedExitCode: 3,
			gherkinCase:      "Exit code 3 on authentication failure",
		},
		{
			name:             "insufficient permissions",
			parseError:       nil,
			serviceError:     apperrors.NewAuthorizationError("no access"),
			expectedExitCode: 4,
			gherkinCase:      "Exit code 4 on insufficient permissions",
		},
		{
			name:             "repository not found",
			parseError:       nil,
			serviceError:     apperrors.NewRepositoryNotFoundError("owner", "repo"),
			expectedExitCode: 5,
			gherkinCase:      "Exit code 5 on repository not found",
		},
		{
			name:             "rate limit exceeded",
			parseError:       nil,
			serviceError:     apperrors.NewRateLimitError(resetTime),
			expectedExitCode: 6,
			gherkinCase:      "Exit code 6 on API rate limit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockWriter := &mockOutputWriter{}
			mockConfigSvc := &mockConfigService{
				ConfigureFunc: func(ctx context.Context, owner, name string, dryRun bool) (interfaces.IConfigResult, error) {
					if tt.serviceError != nil {
						return nil, tt.serviceError
					}
					return newMockConfigResult(false, true, "main", "owner/repo"), nil
				},
			}
			mockParser := &mockRepoParser{
				ParseFunc: func(repoIdentifier string) (string, string, error) {
					if tt.parseError != nil {
						return "", "", tt.parseError
					}
					return "owner", "repo", nil
				},
			}

			application := app.NewApp(mockWriter, mockConfigSvc, mockParser)
			ctx := context.Background()
			opts := interfaces.CLIOptions{
				Repository: "owner/repo",
			}

			// Act
			err := application.Run(ctx, opts)
			exitCode := apperrors.GetExitCode(err)

			// Assert
			if exitCode != tt.expectedExitCode {
				t.Errorf("Exit code = %d, expected %d (%s)", exitCode, tt.expectedExitCode, tt.gherkinCase)
			}
		})
	}
}

// =============================================================================
// Error Wrapping Tests
// =============================================================================

// TestExitCodeWithWrappedAppError verifies wrapped AppErrors are handled correctly.
//
// The implementation should:
// - Unwrap error chains to find AppError
// - Return the correct exit code from the wrapped AppError
func TestExitCodeWithWrappedAppError(t *testing.T) {
	// Arrange
	mockWriter := &mockOutputWriter{}
	mockConfigSvc := &mockConfigService{
		ConfigureFunc: func(ctx context.Context, owner, name string, dryRun bool) (interfaces.IConfigResult, error) {
			// ConfigService wraps the error with additional context
			baseErr := apperrors.NewRepositoryNotFoundError(owner, name)
			// Note: The App wraps errors too, so we test if GetExitCode can unwrap
			return nil, baseErr
		},
	}
	mockParser := &mockRepoParser{
		ParseFunc: func(repoIdentifier string) (string, string, error) {
			return "octocat", "hello-world", nil
		},
	}

	application := app.NewApp(mockWriter, mockConfigSvc, mockParser)
	ctx := context.Background()
	opts := interfaces.CLIOptions{
		Repository: "octocat/hello-world",
	}

	// Act
	err := application.Run(ctx, opts)
	exitCode := apperrors.GetExitCode(err)

	// Assert - GetExitCode should find the AppError through wrapping
	if exitCode != 5 {
		t.Errorf("GetExitCode() = %d, expected 5 for wrapped repository not found error", exitCode)
	}
}
