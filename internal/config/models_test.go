// Package config_test provides tests for the config package data models.
//
// These tests verify that ConfigResult struct correctly implements the IConfigResult interface.
//
// These tests are designed to FAIL until the implementations are properly created
// by the coder agent.
package config_test

import (
	"testing"

	"github.com/josejulio/ghautodelete/internal/config"
	"github.com/josejulio/ghautodelete/pkg/interfaces"
)

// =============================================================================
// ConfigResult Struct Tests
// =============================================================================

// TestConfigResultImplementsIConfigResult verifies ConfigResult implements IConfigResult.
//
// The implementation should:
// - Define ConfigResult struct in internal/config/models.go
// - Implement all IConfigResult methods
func TestConfigResultImplementsIConfigResult(t *testing.T) {
	// Arrange & Act - compile-time interface satisfaction check
	var result interfaces.IConfigResult = &config.ConfigResult{}

	// Assert
	if result == nil {
		t.Error("ConfigResult should implement IConfigResult interface")
	}
}

// TestConfigResultWasAlreadyEnabled verifies ConfigResult.WasAlreadyEnabled() returns correct value.
//
// The implementation should:
// - Store was_already_enabled flag in ConfigResult struct
// - Return the boolean value from WasAlreadyEnabled()
func TestConfigResultWasAlreadyEnabled(t *testing.T) {
	tests := []struct {
		name              string
		wasAlreadyEnabled bool
		expected          bool
	}{
		{
			name:              "returns true when was already enabled",
			wasAlreadyEnabled: true,
			expected:          true,
		},
		{
			name:              "returns false when was not already enabled",
			wasAlreadyEnabled: false,
			expected:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			result := config.ConfigResult{
				AlreadyEnabled: tt.wasAlreadyEnabled,
			}

			// Act
			actual := result.WasAlreadyEnabled()

			// Assert
			if actual != tt.expected {
				t.Errorf("WasAlreadyEnabled() = %v, expected %v", actual, tt.expected)
			}
		})
	}
}

// TestConfigResultIsNowEnabled verifies ConfigResult.IsNowEnabled() returns correct value.
//
// The implementation should:
// - Store is_now_enabled flag in ConfigResult struct
// - Return the boolean value from IsNowEnabled()
func TestConfigResultIsNowEnabled(t *testing.T) {
	tests := []struct {
		name         string
		isNowEnabled bool
		expected     bool
	}{
		{
			name:         "returns true when is now enabled",
			isNowEnabled: true,
			expected:     true,
		},
		{
			name:         "returns false when is not now enabled",
			isNowEnabled: false,
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			result := config.ConfigResult{
				NowEnabled: tt.isNowEnabled,
			}

			// Act
			actual := result.IsNowEnabled()

			// Assert
			if actual != tt.expected {
				t.Errorf("IsNowEnabled() = %v, expected %v", actual, tt.expected)
			}
		})
	}
}

// TestConfigResultGetDefaultBranch verifies ConfigResult.GetDefaultBranch() returns the branch.
//
// The implementation should:
// - Store default_branch in ConfigResult struct
// - Return the default branch string from GetDefaultBranch()
func TestConfigResultGetDefaultBranch(t *testing.T) {
	tests := []struct {
		name          string
		defaultBranch string
		expected      string
	}{
		{
			name:          "returns main as default branch",
			defaultBranch: "main",
			expected:      "main",
		},
		{
			name:          "returns master as default branch",
			defaultBranch: "master",
			expected:      "master",
		},
		{
			name:          "returns custom default branch",
			defaultBranch: "develop",
			expected:      "develop",
		},
		{
			name:          "returns empty string for empty default branch",
			defaultBranch: "",
			expected:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			result := config.ConfigResult{
				DefaultBranch: tt.defaultBranch,
			}

			// Act
			actual := result.GetDefaultBranch()

			// Assert
			if actual != tt.expected {
				t.Errorf("GetDefaultBranch() = %q, expected %q", actual, tt.expected)
			}
		})
	}
}

// TestConfigResultGetRepositoryFullName verifies ConfigResult.GetRepositoryFullName() returns the full name.
//
// The implementation should:
// - Store repository_full_name in ConfigResult struct
// - Return the repository full name string from GetRepositoryFullName()
func TestConfigResultGetRepositoryFullName(t *testing.T) {
	tests := []struct {
		name               string
		repositoryFullName string
		expected           string
	}{
		{
			name:               "returns full name for standard repository",
			repositoryFullName: "octocat/hello-world",
			expected:           "octocat/hello-world",
		},
		{
			name:               "returns full name for organization repository",
			repositoryFullName: "github/docs",
			expected:           "github/docs",
		},
		{
			name:               "returns empty string for empty full name",
			repositoryFullName: "",
			expected:           "",
		},
		{
			name:               "handles repository with dots",
			repositoryFullName: "owner/my.repo.name",
			expected:           "owner/my.repo.name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			result := config.ConfigResult{
				RepositoryFullName: tt.repositoryFullName,
			}

			// Act
			actual := result.GetRepositoryFullName()

			// Assert
			if actual != tt.expected {
				t.Errorf("GetRepositoryFullName() = %q, expected %q", actual, tt.expected)
			}
		})
	}
}

// TestConfigResultAllFieldsSet verifies ConfigResult works with all fields set.
//
// The implementation should correctly return all values when all fields are set.
func TestConfigResultAllFieldsSet(t *testing.T) {
	// Arrange
	result := config.ConfigResult{
		AlreadyEnabled:     true,
		NowEnabled:         true,
		DefaultBranch:      "main",
		RepositoryFullName: "octocat/hello-world",
	}

	// Act & Assert
	if wasEnabled := result.WasAlreadyEnabled(); wasEnabled != true {
		t.Errorf("WasAlreadyEnabled() = %v, expected %v", wasEnabled, true)
	}
	if isEnabled := result.IsNowEnabled(); isEnabled != true {
		t.Errorf("IsNowEnabled() = %v, expected %v", isEnabled, true)
	}
	if branch := result.GetDefaultBranch(); branch != "main" {
		t.Errorf("GetDefaultBranch() = %q, expected %q", branch, "main")
	}
	if fullName := result.GetRepositoryFullName(); fullName != "octocat/hello-world" {
		t.Errorf("GetRepositoryFullName() = %q, expected %q", fullName, "octocat/hello-world")
	}
}

// TestConfigResultAsIConfigResult verifies ConfigResult can be used as IConfigResult.
//
// The implementation should allow ConfigResult to be passed where IConfigResult is expected.
func TestConfigResultAsIConfigResult(t *testing.T) {
	// Arrange
	result := &config.ConfigResult{
		AlreadyEnabled:     false,
		NowEnabled:         true,
		DefaultBranch:      "main",
		RepositoryFullName: "octocat/hello-world",
	}

	// Act - function that accepts IConfigResult
	fullName := acceptIConfigResult(result)

	// Assert
	if fullName != "octocat/hello-world" {
		t.Errorf("acceptIConfigResult() = %q, expected %q", fullName, "octocat/hello-world")
	}
}

// Helper function that accepts IConfigResult interface
func acceptIConfigResult(result interfaces.IConfigResult) string {
	return result.GetRepositoryFullName()
}

// =============================================================================
// Constructor Tests
// =============================================================================

// TestNewConfigResult verifies NewConfigResult constructor creates valid ConfigResult.
//
// The implementation should:
// - Define NewConfigResult(wasAlreadyEnabled, isNowEnabled bool, defaultBranch, fullName string) *ConfigResult
func TestNewConfigResult(t *testing.T) {
	// Arrange
	wasAlreadyEnabled := false
	isNowEnabled := true
	defaultBranch := "main"
	fullName := "octocat/hello-world"

	// Act
	result := config.NewConfigResult(wasAlreadyEnabled, isNowEnabled, defaultBranch, fullName)

	// Assert
	if result == nil {
		t.Fatal("NewConfigResult() returned nil")
	}
	if result.WasAlreadyEnabled() != wasAlreadyEnabled {
		t.Errorf("WasAlreadyEnabled() = %v, expected %v", result.WasAlreadyEnabled(), wasAlreadyEnabled)
	}
	if result.IsNowEnabled() != isNowEnabled {
		t.Errorf("IsNowEnabled() = %v, expected %v", result.IsNowEnabled(), isNowEnabled)
	}
	if result.GetDefaultBranch() != defaultBranch {
		t.Errorf("GetDefaultBranch() = %q, expected %q", result.GetDefaultBranch(), defaultBranch)
	}
	if result.GetRepositoryFullName() != fullName {
		t.Errorf("GetRepositoryFullName() = %q, expected %q", result.GetRepositoryFullName(), fullName)
	}
}

// =============================================================================
// State Transition Tests
// =============================================================================
// These tests verify different configuration state scenarios.

// TestConfigResultStateNewlyEnabled verifies state when feature was just enabled.
func TestConfigResultStateNewlyEnabled(t *testing.T) {
	// Arrange - Feature was disabled, now enabled
	result := config.ConfigResult{
		AlreadyEnabled:     false,
		NowEnabled:         true,
		DefaultBranch:      "main",
		RepositoryFullName: "octocat/hello-world",
	}

	// Assert
	if result.WasAlreadyEnabled() != false {
		t.Error("WasAlreadyEnabled() should be false for newly enabled feature")
	}
	if result.IsNowEnabled() != true {
		t.Error("IsNowEnabled() should be true for newly enabled feature")
	}
}

// TestConfigResultStateAlreadyEnabled verifies state when feature was already enabled.
func TestConfigResultStateAlreadyEnabled(t *testing.T) {
	// Arrange - Feature was already enabled, still enabled
	result := config.ConfigResult{
		AlreadyEnabled:     true,
		NowEnabled:         true,
		DefaultBranch:      "main",
		RepositoryFullName: "octocat/hello-world",
	}

	// Assert
	if result.WasAlreadyEnabled() != true {
		t.Error("WasAlreadyEnabled() should be true when feature was already enabled")
	}
	if result.IsNowEnabled() != true {
		t.Error("IsNowEnabled() should be true when feature is still enabled")
	}
}

// TestConfigResultStateCheckOnlyEnabled verifies state for check-only when enabled.
func TestConfigResultStateCheckOnlyEnabled(t *testing.T) {
	// Arrange - Check-only mode, feature is enabled
	result := config.ConfigResult{
		AlreadyEnabled:     true,
		NowEnabled:         true,
		DefaultBranch:      "main",
		RepositoryFullName: "octocat/hello-world",
	}

	// Assert
	if result.WasAlreadyEnabled() != true {
		t.Error("WasAlreadyEnabled() should be true when checking already enabled feature")
	}
	if result.IsNowEnabled() != true {
		t.Error("IsNowEnabled() should be true when feature is enabled")
	}
}

// TestConfigResultStateCheckOnlyDisabled verifies state for check-only when disabled.
func TestConfigResultStateCheckOnlyDisabled(t *testing.T) {
	// Arrange - Check-only mode, feature is disabled
	result := config.ConfigResult{
		AlreadyEnabled:     false,
		NowEnabled:         false,
		DefaultBranch:      "main",
		RepositoryFullName: "octocat/hello-world",
	}

	// Assert
	if result.WasAlreadyEnabled() != false {
		t.Error("WasAlreadyEnabled() should be false when checking disabled feature")
	}
	if result.IsNowEnabled() != false {
		t.Error("IsNowEnabled() should be false when feature is disabled")
	}
}

// TestConfigResultStateDryRun verifies state for dry-run mode.
func TestConfigResultStateDryRun(t *testing.T) {
	// Arrange - Dry-run mode, feature was disabled, would be enabled but isn't
	result := config.ConfigResult{
		AlreadyEnabled:     false,
		NowEnabled:         false, // Dry-run doesn't actually enable
		DefaultBranch:      "main",
		RepositoryFullName: "octocat/hello-world",
	}

	// Assert
	if result.WasAlreadyEnabled() != false {
		t.Error("WasAlreadyEnabled() should be false in dry-run when feature was disabled")
	}
	if result.IsNowEnabled() != false {
		t.Error("IsNowEnabled() should be false in dry-run mode (no actual change)")
	}
}
