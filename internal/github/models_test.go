// Package github_test provides tests for the github package data models.
//
// These tests verify that Repository and RepositorySettings structs correctly
// implement their respective interfaces (IRepository and IRepositorySettings).
//
// These tests are designed to FAIL until the implementations are properly created
// by the coder agent.
package github_test

import (
	"testing"

	"github.com/josejulio/ghautodelete/internal/github"
	"github.com/josejulio/ghautodelete/pkg/interfaces"
)

// =============================================================================
// Repository Struct Tests
// =============================================================================

// TestRepositoryImplementsIRepository verifies Repository implements IRepository.
//
// The implementation should:
// - Define Repository struct in internal/github/models.go
// - Implement all IRepository methods
func TestRepositoryImplementsIRepository(t *testing.T) {
	// Arrange & Act - compile-time interface satisfaction check
	var repo interfaces.IRepository = &github.Repository{}

	// Assert
	if repo == nil {
		t.Error("Repository should implement IRepository interface")
	}
}

// TestRepositoryGetOwner verifies Repository.GetOwner() returns the owner.
//
// The implementation should:
// - Store owner in Repository struct
// - Return the owner string from GetOwner()
func TestRepositoryGetOwner(t *testing.T) {
	tests := []struct {
		name     string
		owner    string
		expected string
	}{
		{
			name:     "returns owner for standard repository",
			owner:    "octocat",
			expected: "octocat",
		},
		{
			name:     "returns owner for organization repository",
			owner:    "github",
			expected: "github",
		},
		{
			name:     "returns empty string for empty owner",
			owner:    "",
			expected: "",
		},
		{
			name:     "handles owner with special characters",
			owner:    "my-org-123",
			expected: "my-org-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := github.NewRepository(tt.owner, "", "", false)

			// Act
			result := repo.GetOwner()

			// Assert
			if result != tt.expected {
				t.Errorf("GetOwner() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

// TestRepositoryGetName verifies Repository.GetName() returns the repository name.
//
// The implementation should:
// - Store name in Repository struct
// - Return the name string from GetName()
func TestRepositoryGetName(t *testing.T) {
	tests := []struct {
		name         string
		repoName     string
		expected     string
	}{
		{
			name:     "returns name for standard repository",
			repoName: "hello-world",
			expected: "hello-world",
		},
		{
			name:     "returns name with dots",
			repoName: "my.repo.name",
			expected: "my.repo.name",
		},
		{
			name:     "returns empty string for empty name",
			repoName: "",
			expected: "",
		},
		{
			name:     "handles name with underscores",
			repoName: "my_repo_name",
			expected: "my_repo_name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := github.Repository{
				Name: tt.repoName,
			}

			// Act
			result := repo.GetName()

			// Assert
			if result != tt.expected {
				t.Errorf("GetName() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

// TestRepositoryGetDefaultBranch verifies Repository.GetDefaultBranch() returns the default branch.
//
// The implementation should:
// - Store default branch in Repository struct
// - Return the default branch string from GetDefaultBranch()
func TestRepositoryGetDefaultBranch(t *testing.T) {
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
			repo := github.Repository{
				DefaultBranch: tt.defaultBranch,
			}

			// Act
			result := repo.GetDefaultBranch()

			// Assert
			if result != tt.expected {
				t.Errorf("GetDefaultBranch() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

// TestRepositoryGetDeleteBranchOnMerge verifies Repository.GetDeleteBranchOnMerge() returns the setting.
//
// The implementation should:
// - Store delete_branch_on_merge setting in Repository struct
// - Return the boolean value from GetDeleteBranchOnMerge()
func TestRepositoryGetDeleteBranchOnMerge(t *testing.T) {
	tests := []struct {
		name                string
		deleteBranchOnMerge bool
		expected            bool
	}{
		{
			name:                "returns true when enabled",
			deleteBranchOnMerge: true,
			expected:            true,
		},
		{
			name:                "returns false when disabled",
			deleteBranchOnMerge: false,
			expected:            false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := github.Repository{
				DeleteBranchOnMerge: tt.deleteBranchOnMerge,
			}

			// Act
			result := repo.GetDeleteBranchOnMerge()

			// Assert
			if result != tt.expected {
				t.Errorf("GetDeleteBranchOnMerge() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// TestRepositoryGetFullName verifies Repository.GetFullName() returns owner/name format.
//
// The implementation should:
// - Return the full name in "owner/name" format from GetFullName()
func TestRepositoryGetFullName(t *testing.T) {
	tests := []struct {
		name     string
		owner    string
		repoName string
		expected string
	}{
		{
			name:     "returns full name for standard repository",
			owner:    "octocat",
			repoName: "hello-world",
			expected: "octocat/hello-world",
		},
		{
			name:     "returns full name for organization repository",
			owner:    "github",
			repoName: "docs",
			expected: "github/docs",
		},
		{
			name:     "handles empty owner",
			owner:    "",
			repoName: "repo",
			expected: "/repo",
		},
		{
			name:     "handles empty name",
			owner:    "owner",
			repoName: "",
			expected: "owner/",
		},
		{
			name:     "handles both empty",
			owner:    "",
			repoName: "",
			expected: "/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := github.NewRepository(tt.owner, tt.repoName, "", false)

			// Act
			result := repo.GetFullName()

			// Assert
			if result != tt.expected {
				t.Errorf("GetFullName() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

// TestRepositoryAllFieldsSet verifies Repository works with all fields set.
//
// The implementation should correctly return all values when all fields are set.
func TestRepositoryAllFieldsSet(t *testing.T) {
	// Arrange
	repo := github.NewRepository("octocat", "hello-world", "main", true)

	// Act & Assert
	if owner := repo.GetOwner(); owner != "octocat" {
		t.Errorf("GetOwner() = %q, expected %q", owner, "octocat")
	}
	if name := repo.GetName(); name != "hello-world" {
		t.Errorf("GetName() = %q, expected %q", name, "hello-world")
	}
	if branch := repo.GetDefaultBranch(); branch != "main" {
		t.Errorf("GetDefaultBranch() = %q, expected %q", branch, "main")
	}
	if enabled := repo.GetDeleteBranchOnMerge(); enabled != true {
		t.Errorf("GetDeleteBranchOnMerge() = %v, expected %v", enabled, true)
	}
	if fullName := repo.GetFullName(); fullName != "octocat/hello-world" {
		t.Errorf("GetFullName() = %q, expected %q", fullName, "octocat/hello-world")
	}
}

// =============================================================================
// RepositorySettings Struct Tests
// =============================================================================

// TestRepositorySettingsImplementsIRepositorySettings verifies RepositorySettings implements IRepositorySettings.
//
// The implementation should:
// - Define RepositorySettings struct in internal/github/models.go
// - Implement all IRepositorySettings methods
func TestRepositorySettingsImplementsIRepositorySettings(t *testing.T) {
	// Arrange & Act - compile-time interface satisfaction check
	var settings interfaces.IRepositorySettings = &github.RepositorySettings{}

	// Assert
	if settings == nil {
		t.Error("RepositorySettings should implement IRepositorySettings interface")
	}
}

// TestRepositorySettingsGetDeleteBranchOnMerge verifies RepositorySettings.GetDeleteBranchOnMerge().
//
// The implementation should:
// - Store delete_branch_on_merge setting in RepositorySettings struct
// - Return the boolean value from GetDeleteBranchOnMerge()
func TestRepositorySettingsGetDeleteBranchOnMerge(t *testing.T) {
	tests := []struct {
		name                string
		deleteBranchOnMerge bool
		expected            bool
	}{
		{
			name:                "returns true when enabled",
			deleteBranchOnMerge: true,
			expected:            true,
		},
		{
			name:                "returns false when disabled",
			deleteBranchOnMerge: false,
			expected:            false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			settings := github.RepositorySettings{
				DeleteBranchOnMerge: tt.deleteBranchOnMerge,
			}

			// Act
			result := settings.GetDeleteBranchOnMerge()

			// Assert
			if result != tt.expected {
				t.Errorf("GetDeleteBranchOnMerge() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// TestRepositorySettingsAsIRepositorySettings verifies RepositorySettings can be used as IRepositorySettings.
//
// The implementation should allow RepositorySettings to be passed where IRepositorySettings is expected.
func TestRepositorySettingsAsIRepositorySettings(t *testing.T) {
	// Arrange
	settings := &github.RepositorySettings{
		DeleteBranchOnMerge: true,
	}

	// Act - function that accepts IRepositorySettings
	result := acceptIRepositorySettings(settings)

	// Assert
	if result != true {
		t.Errorf("acceptIRepositorySettings() = %v, expected %v", result, true)
	}
}

// Helper function that accepts IRepositorySettings interface
func acceptIRepositorySettings(settings interfaces.IRepositorySettings) bool {
	return settings.GetDeleteBranchOnMerge()
}

// =============================================================================
// Constructor Tests
// =============================================================================

// TestNewRepository verifies NewRepository constructor creates valid Repository.
//
// The implementation should:
// - Define NewRepository(owner, name, defaultBranch string, deleteBranchOnMerge bool) *Repository
func TestNewRepository(t *testing.T) {
	// Arrange
	owner := "octocat"
	name := "hello-world"
	defaultBranch := "main"
	deleteBranchOnMerge := true

	// Act
	repo := github.NewRepository(owner, name, defaultBranch, deleteBranchOnMerge)

	// Assert
	if repo == nil {
		t.Fatal("NewRepository() returned nil")
	}
	if repo.GetOwner() != owner {
		t.Errorf("GetOwner() = %q, expected %q", repo.GetOwner(), owner)
	}
	if repo.GetName() != name {
		t.Errorf("GetName() = %q, expected %q", repo.GetName(), name)
	}
	if repo.GetDefaultBranch() != defaultBranch {
		t.Errorf("GetDefaultBranch() = %q, expected %q", repo.GetDefaultBranch(), defaultBranch)
	}
	if repo.GetDeleteBranchOnMerge() != deleteBranchOnMerge {
		t.Errorf("GetDeleteBranchOnMerge() = %v, expected %v", repo.GetDeleteBranchOnMerge(), deleteBranchOnMerge)
	}
}

// TestNewRepositorySettings verifies NewRepositorySettings constructor creates valid RepositorySettings.
//
// The implementation should:
// - Define NewRepositorySettings(deleteBranchOnMerge bool) *RepositorySettings
func TestNewRepositorySettings(t *testing.T) {
	tests := []struct {
		name                string
		deleteBranchOnMerge bool
	}{
		{
			name:                "creates settings with delete enabled",
			deleteBranchOnMerge: true,
		},
		{
			name:                "creates settings with delete disabled",
			deleteBranchOnMerge: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			settings := github.NewRepositorySettings(tt.deleteBranchOnMerge)

			// Assert
			if settings == nil {
				t.Fatal("NewRepositorySettings() returned nil")
			}
			if settings.GetDeleteBranchOnMerge() != tt.deleteBranchOnMerge {
				t.Errorf("GetDeleteBranchOnMerge() = %v, expected %v", settings.GetDeleteBranchOnMerge(), tt.deleteBranchOnMerge)
			}
		})
	}
}
