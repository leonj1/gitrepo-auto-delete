// Package parser_test provides tests for the repository parser.
//
// These tests verify that RepoParser correctly parses various repository identifier
// formats according to the Gherkin scenarios:
//
// Feature: Repository Input Parsing
//   Scenario: Parse owner/repo format -> owner="octocat", repo="hello-world"
//   Scenario: Parse HTTPS GitHub URL -> owner="octocat", repo="hello-world"
//   Scenario: Parse HTTPS GitHub URL with .git suffix -> owner="octocat", repo="hello-world"
//   Scenario: Parse SSH GitHub URL -> owner="octocat", repo="hello-world"
//   Scenario: Parse SSH GitHub URL without .git suffix -> owner="octocat", repo="hello-world"
//   Scenario: Handle input with leading and trailing whitespace -> trims whitespace
//   Scenario: Reject invalid repository format with single segment -> error "Expected format: owner/repo"
//   Scenario: Reject invalid repository format with too many segments -> error "Expected format: owner/repo"
//   Scenario: Reject empty repository input -> error "Repository identifier is required"
//   Scenario: Reject repository with invalid characters in owner -> error "Invalid repository name characters"
//   Scenario: Reject repository with invalid characters in repo name -> error "Invalid repository name characters"
//   Scenario: Reject invalid HTTPS URL format -> error "Invalid GitHub URL format"
//   Scenario: Reject non-GitHub HTTPS URL -> error "Expected format: owner/repo"
//
// These tests are designed to FAIL until the implementations are properly created
// by the coder agent.
package parser_test

import (
	"errors"
	"strings"
	"testing"

	apperrors "github.com/josejulio/ghautodelete/internal/errors"
	"github.com/josejulio/ghautodelete/internal/parser"
	"github.com/josejulio/ghautodelete/pkg/interfaces"
)

// =============================================================================
// Interface Satisfaction Tests
// =============================================================================

// TestRepoParserImplementsIRepoParser verifies RepoParser implements IRepoParser interface.
//
// The implementation should:
// - Define RepoParser struct in internal/parser/repo_parser.go
// - Implement Parse(input string) (owner string, repo string, err error) method
func TestRepoParserImplementsIRepoParser(t *testing.T) {
	// Arrange
	p := parser.NewRepoParser()

	// Act & Assert - compile-time interface satisfaction check
	var _ interfaces.IRepoParser = p
}

// TestNewRepoParserReturnsNonNil verifies the constructor returns a valid instance.
//
// The implementation should:
// - Define NewRepoParser() IRepoParser constructor function
// - Return a valid RepoParser instance
func TestNewRepoParserReturnsNonNil(t *testing.T) {
	// Act
	p := parser.NewRepoParser()

	// Assert
	if p == nil {
		t.Fatal("NewRepoParser() returned nil, expected non-nil IRepoParser")
	}
}

// =============================================================================
// Happy Path Tests - Gherkin Scenarios
// =============================================================================

// TestParseValidInputFormats tests parsing of all valid input formats.
//
// Gherkin Scenarios:
// - Parse owner/repo format -> owner="octocat", repo="hello-world"
// - Parse HTTPS GitHub URL -> owner="octocat", repo="hello-world"
// - Parse HTTPS GitHub URL with .git suffix -> owner="octocat", repo="hello-world"
// - Parse SSH GitHub URL -> owner="octocat", repo="hello-world"
// - Parse SSH GitHub URL without .git suffix -> owner="octocat", repo="hello-world"
// - Handle input with leading and trailing whitespace -> trims whitespace
func TestParseValidInputFormats(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedOwner string
		expectedRepo  string
	}{
		// Scenario: Parse owner/repo format
		{
			name:          "parse owner/repo format",
			input:         "octocat/hello-world",
			expectedOwner: "octocat",
			expectedRepo:  "hello-world",
		},
		// Scenario: Parse HTTPS GitHub URL
		{
			name:          "parse HTTPS GitHub URL",
			input:         "https://github.com/octocat/hello-world",
			expectedOwner: "octocat",
			expectedRepo:  "hello-world",
		},
		// Scenario: Parse HTTPS GitHub URL with .git suffix
		{
			name:          "parse HTTPS GitHub URL with .git suffix",
			input:         "https://github.com/octocat/hello-world.git",
			expectedOwner: "octocat",
			expectedRepo:  "hello-world",
		},
		// Scenario: Parse SSH GitHub URL (with .git suffix)
		{
			name:          "parse SSH GitHub URL with .git suffix",
			input:         "git@github.com:octocat/hello-world.git",
			expectedOwner: "octocat",
			expectedRepo:  "hello-world",
		},
		// Scenario: Parse SSH GitHub URL without .git suffix
		{
			name:          "parse SSH GitHub URL without .git suffix",
			input:         "git@github.com:octocat/hello-world",
			expectedOwner: "octocat",
			expectedRepo:  "hello-world",
		},
		// Scenario: Handle input with leading whitespace
		{
			name:          "handle input with leading whitespace",
			input:         "  octocat/hello-world",
			expectedOwner: "octocat",
			expectedRepo:  "hello-world",
		},
		// Scenario: Handle input with trailing whitespace
		{
			name:          "handle input with trailing whitespace",
			input:         "octocat/hello-world  ",
			expectedOwner: "octocat",
			expectedRepo:  "hello-world",
		},
		// Scenario: Handle input with leading and trailing whitespace
		{
			name:          "handle input with leading and trailing whitespace",
			input:         "  octocat/hello-world  ",
			expectedOwner: "octocat",
			expectedRepo:  "hello-world",
		},
		// Additional valid cases
		{
			name:          "parse owner/repo with underscores",
			input:         "my_owner/my_repo",
			expectedOwner: "my_owner",
			expectedRepo:  "my_repo",
		},
		{
			name:          "parse owner/repo with dots in repo name",
			input:         "owner/repo.js",
			expectedOwner: "owner",
			expectedRepo:  "repo.js",
		},
		{
			name:          "parse owner/repo with hyphens",
			input:         "my-owner/my-repo",
			expectedOwner: "my-owner",
			expectedRepo:  "my-repo",
		},
		{
			name:          "parse HTTPS URL with whitespace",
			input:         "  https://github.com/octocat/hello-world  ",
			expectedOwner: "octocat",
			expectedRepo:  "hello-world",
		},
		{
			name:          "parse SSH URL with whitespace",
			input:         "  git@github.com:octocat/hello-world.git  ",
			expectedOwner: "octocat",
			expectedRepo:  "hello-world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			p := parser.NewRepoParser()

			// Act
			owner, repo, err := p.Parse(tt.input)

			// Assert
			if err != nil {
				t.Fatalf("Parse(%q) returned unexpected error: %v", tt.input, err)
			}
			if owner != tt.expectedOwner {
				t.Errorf("Parse(%q) owner = %q, expected %q", tt.input, owner, tt.expectedOwner)
			}
			if repo != tt.expectedRepo {
				t.Errorf("Parse(%q) repo = %q, expected %q", tt.input, repo, tt.expectedRepo)
			}
		})
	}
}

// =============================================================================
// Error Case Tests - Gherkin Scenarios
// =============================================================================

// TestParseInvalidInputFormats tests that invalid inputs return appropriate errors.
//
// Gherkin Scenarios:
// - Reject invalid repository format with single segment -> error "Expected format: owner/repo"
// - Reject invalid repository format with too many segments -> error "Expected format: owner/repo"
// - Reject empty repository input -> error "Repository identifier is required"
// - Reject repository with invalid characters in owner -> error "Invalid repository name characters"
// - Reject repository with invalid characters in repo name -> error "Invalid repository name characters"
// - Reject invalid HTTPS URL format -> error "Invalid GitHub URL format"
// - Reject non-GitHub HTTPS URL -> error "Expected format: owner/repo"
func TestParseInvalidInputFormats(t *testing.T) {
	tests := []struct {
		name                   string
		input                  string
		expectedErrorContains  string
		expectedCode           apperrors.ErrorCode
	}{
		// Scenario: Reject empty repository input
		{
			name:                   "reject empty repository input",
			input:                  "",
			expectedErrorContains:  "Repository identifier is required",
			expectedCode:           apperrors.ErrInvalidArguments,
		},
		// Scenario: Reject whitespace-only input
		{
			name:                   "reject whitespace-only input",
			input:                  "   ",
			expectedErrorContains:  "Repository identifier is required",
			expectedCode:           apperrors.ErrInvalidArguments,
		},
		// Scenario: Reject invalid repository format with single segment
		{
			name:                   "reject single segment format",
			input:                  "octocat",
			expectedErrorContains:  "Expected format: owner/repo",
			expectedCode:           apperrors.ErrInvalidArguments,
		},
		// Scenario: Reject invalid repository format with too many segments
		{
			name:                   "reject too many segments format",
			input:                  "octocat/hello/world",
			expectedErrorContains:  "Expected format: owner/repo",
			expectedCode:           apperrors.ErrInvalidArguments,
		},
		// Scenario: Reject repository with invalid characters in owner
		{
			name:                   "reject invalid characters in owner (spaces)",
			input:                  "octo cat/hello-world",
			expectedErrorContains:  "Invalid repository name characters",
			expectedCode:           apperrors.ErrInvalidArguments,
		},
		// Scenario: Reject repository with invalid characters in owner (special chars)
		{
			name:                   "reject invalid characters in owner (special chars)",
			input:                  "octo@cat/hello-world",
			expectedErrorContains:  "Invalid repository name characters",
			expectedCode:           apperrors.ErrInvalidArguments,
		},
		// Scenario: Reject repository with invalid characters in repo name
		{
			name:                   "reject invalid characters in repo (spaces)",
			input:                  "octocat/hello world",
			expectedErrorContains:  "Invalid repository name characters",
			expectedCode:           apperrors.ErrInvalidArguments,
		},
		// Scenario: Reject repository with invalid characters in repo name (special chars)
		{
			name:                   "reject invalid characters in repo (special chars)",
			input:                  "octocat/hello@world",
			expectedErrorContains:  "Invalid repository name characters",
			expectedCode:           apperrors.ErrInvalidArguments,
		},
		// Scenario: Reject invalid HTTPS URL format (missing path)
		{
			name:                   "reject invalid HTTPS URL format (missing repo path)",
			input:                  "https://github.com/",
			expectedErrorContains:  "Invalid GitHub URL format",
			expectedCode:           apperrors.ErrInvalidArguments,
		},
		// Scenario: Reject invalid HTTPS URL format (only owner)
		{
			name:                   "reject invalid HTTPS URL format (only owner)",
			input:                  "https://github.com/octocat",
			expectedErrorContains:  "Invalid GitHub URL format",
			expectedCode:           apperrors.ErrInvalidArguments,
		},
		// Scenario: Reject non-GitHub HTTPS URL
		{
			name:                   "reject non-GitHub HTTPS URL (gitlab)",
			input:                  "https://gitlab.com/octocat/hello-world",
			expectedErrorContains:  "Expected format: owner/repo",
			expectedCode:           apperrors.ErrInvalidArguments,
		},
		// Scenario: Reject non-GitHub HTTPS URL (bitbucket)
		{
			name:                   "reject non-GitHub HTTPS URL (bitbucket)",
			input:                  "https://bitbucket.org/octocat/hello-world",
			expectedErrorContains:  "Expected format: owner/repo",
			expectedCode:           apperrors.ErrInvalidArguments,
		},
		// Additional error cases
		{
			name:                   "reject invalid HTTPS URL format (too many segments)",
			input:                  "https://github.com/octocat/hello/world/extra",
			expectedErrorContains:  "Invalid GitHub URL format",
			expectedCode:           apperrors.ErrInvalidArguments,
		},
		{
			name:                   "reject invalid SSH URL format (missing repo)",
			input:                  "git@github.com:octocat",
			expectedErrorContains:  "Invalid GitHub URL format",
			expectedCode:           apperrors.ErrInvalidArguments,
		},
		{
			name:                   "reject non-GitHub SSH URL",
			input:                  "git@gitlab.com:octocat/hello-world.git",
			expectedErrorContains:  "Expected format: owner/repo",
			expectedCode:           apperrors.ErrInvalidArguments,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			p := parser.NewRepoParser()

			// Act
			owner, repo, err := p.Parse(tt.input)

			// Assert
			if err == nil {
				t.Fatalf("Parse(%q) expected error containing %q, got owner=%q repo=%q",
					tt.input, tt.expectedErrorContains, owner, repo)
			}

			// Verify error message contains expected text
			if !strings.Contains(err.Error(), tt.expectedErrorContains) {
				t.Errorf("Parse(%q) error = %q, expected to contain %q",
					tt.input, err.Error(), tt.expectedErrorContains)
			}

			// Verify error is *AppError with correct code
			var appErr *apperrors.AppError
			if !errors.As(err, &appErr) {
				t.Errorf("Parse(%q) error should be *AppError, got %T", tt.input, err)
			} else if appErr.Code != tt.expectedCode {
				t.Errorf("Parse(%q) error code = %v, expected %v",
					tt.input, appErr.Code, tt.expectedCode)
			}

			// Verify empty return values on error
			if owner != "" {
				t.Errorf("Parse(%q) owner should be empty on error, got %q", tt.input, owner)
			}
			if repo != "" {
				t.Errorf("Parse(%q) repo should be empty on error, got %q", tt.input, repo)
			}
		})
	}
}

// =============================================================================
// Edge Case Tests
// =============================================================================

// TestParseEdgeCases tests various edge cases for repository parsing.
func TestParseEdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedOwner string
		expectedRepo  string
		expectError   bool
		errorContains string
	}{
		// Valid edge cases
		{
			name:          "single character owner and repo",
			input:         "a/b",
			expectedOwner: "a",
			expectedRepo:  "b",
			expectError:   false,
		},
		{
			name:          "numbers in owner and repo",
			input:         "user123/repo456",
			expectedOwner: "user123",
			expectedRepo:  "repo456",
			expectError:   false,
		},
		{
			name:          "mixed case owner and repo",
			input:         "MyOwner/MyRepo",
			expectedOwner: "MyOwner",
			expectedRepo:  "MyRepo",
			expectError:   false,
		},
		{
			name:          "repo name starting with dot",
			input:         "owner/.repo",
			expectedOwner: "owner",
			expectedRepo:  ".repo",
			expectError:   false,
		},
		{
			name:          "repo name with multiple dots",
			input:         "owner/repo.name.js",
			expectedOwner: "owner",
			expectedRepo:  "repo.name.js",
			expectError:   false,
		},
		// Invalid edge cases
		{
			name:          "empty owner",
			input:         "/repo",
			expectedOwner: "",
			expectedRepo:  "",
			expectError:   true,
			errorContains: "Invalid repository name characters",
		},
		{
			name:          "empty repo",
			input:         "owner/",
			expectedOwner: "",
			expectedRepo:  "",
			expectError:   true,
			errorContains: "Invalid repository name characters",
		},
		{
			name:          "only slash",
			input:         "/",
			expectedOwner: "",
			expectedRepo:  "",
			expectError:   true,
			errorContains: "Invalid repository name characters",
		},
		{
			name:          "tab in input",
			input:         "owner\t/repo",
			expectedOwner: "",
			expectedRepo:  "",
			expectError:   true,
			errorContains: "Invalid repository name characters",
		},
		{
			name:          "newline in input",
			input:         "owner\n/repo",
			expectedOwner: "",
			expectedRepo:  "",
			expectError:   true,
			errorContains: "Invalid repository name characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			p := parser.NewRepoParser()

			// Act
			owner, repo, err := p.Parse(tt.input)

			// Assert
			if tt.expectError {
				if err == nil {
					t.Fatalf("Parse(%q) expected error containing %q, got owner=%q repo=%q",
						tt.input, tt.errorContains, owner, repo)
				}
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Parse(%q) error = %q, expected to contain %q",
						tt.input, err.Error(), tt.errorContains)
				}

				// Verify error is *AppError
				var appErr *apperrors.AppError
				if !errors.As(err, &appErr) {
					t.Errorf("Parse(%q) error should be *AppError, got %T", tt.input, err)
				}
			} else {
				if err != nil {
					t.Fatalf("Parse(%q) unexpected error: %v", tt.input, err)
				}
				if owner != tt.expectedOwner {
					t.Errorf("Parse(%q) owner = %q, expected %q", tt.input, owner, tt.expectedOwner)
				}
				if repo != tt.expectedRepo {
					t.Errorf("Parse(%q) repo = %q, expected %q", tt.input, repo, tt.expectedRepo)
				}
			}
		})
	}
}

// =============================================================================
// Unicode and Special Character Tests
// =============================================================================

// TestParseUnicodeAndSpecialCharacters tests handling of unicode and special characters.
func TestParseUnicodeAndSpecialCharacters(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectError   bool
		errorContains string
	}{
		{
			name:          "unicode in owner name",
			input:         "usér/repo",
			expectError:   true,
			errorContains: "Invalid repository name characters",
		},
		{
			name:          "unicode in repo name",
			input:         "owner/repö",
			expectError:   true,
			errorContains: "Invalid repository name characters",
		},
		{
			name:          "emoji in owner name",
			input:         "owner😀/repo",
			expectError:   true,
			errorContains: "Invalid repository name characters",
		},
		{
			name:          "hash in owner name",
			input:         "owner#name/repo",
			expectError:   true,
			errorContains: "Invalid repository name characters",
		},
		{
			name:          "percent in repo name",
			input:         "owner/repo%name",
			expectError:   true,
			errorContains: "Invalid repository name characters",
		},
		{
			name:          "backslash instead of forward slash",
			input:         "owner\\repo",
			expectError:   true,
			errorContains: "Expected format: owner/repo",
		},
		{
			name:          "question mark in input",
			input:         "owner/repo?",
			expectError:   true,
			errorContains: "Invalid repository name characters",
		},
		{
			name:          "asterisk in input",
			input:         "owner/repo*",
			expectError:   true,
			errorContains: "Invalid repository name characters",
		},
		{
			name:          "double slash in owner/repo format",
			input:         "owner//repo",
			expectError:   true,
			errorContains: "Invalid repository name characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			p := parser.NewRepoParser()

			// Act
			owner, repo, err := p.Parse(tt.input)

			// Assert
			if tt.expectError {
				if err == nil {
					t.Fatalf("Parse(%q) expected error containing %q, got owner=%q repo=%q",
						tt.input, tt.errorContains, owner, repo)
				}
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Parse(%q) error = %q, expected to contain %q",
						tt.input, err.Error(), tt.errorContains)
				}

				// Verify error is *AppError with ErrInvalidArguments code
				var appErr *apperrors.AppError
				if !errors.As(err, &appErr) {
					t.Errorf("Parse(%q) error should be *AppError, got %T", tt.input, err)
				} else if appErr.Code != apperrors.ErrInvalidArguments {
					t.Errorf("Parse(%q) error code = %v, expected %v",
						tt.input, appErr.Code, apperrors.ErrInvalidArguments)
				}
			} else {
				if err != nil {
					t.Fatalf("Parse(%q) unexpected error: %v", tt.input, err)
				}
			}
		})
	}
}

// =============================================================================
// URL Format Tests
// =============================================================================

// TestParseURLFormats tests various URL formats more thoroughly.
func TestParseURLFormats(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedOwner string
		expectedRepo  string
		expectError   bool
		errorContains string
	}{
		// Valid HTTPS URLs
		{
			name:          "HTTPS URL with www prefix",
			input:         "https://www.github.com/octocat/hello-world",
			expectedOwner: "",
			expectedRepo:  "",
			expectError:   true,
			errorContains: "Expected format: owner/repo",
		},
		{
			name:          "HTTP URL (not HTTPS)",
			input:         "http://github.com/octocat/hello-world",
			expectedOwner: "",
			expectedRepo:  "",
			expectError:   true,
			errorContains: "Expected format: owner/repo",
		},
		{
			name:          "HTTPS URL with trailing slash",
			input:         "https://github.com/octocat/hello-world/",
			expectedOwner: "octocat",
			expectedRepo:  "hello-world",
			expectError:   false,
		},
		{
			name:          "HTTPS URL with uppercase scheme",
			input:         "HTTPS://github.com/octocat/hello-world",
			expectedOwner: "octocat",
			expectedRepo:  "hello-world",
			expectError:   false,
		},
		{
			name:          "HTTPS URL with mixed case domain",
			input:         "https://GitHub.com/octocat/hello-world",
			expectedOwner: "octocat",
			expectedRepo:  "hello-world",
			expectError:   false,
		},
		// Invalid SSH URLs
		{
			name:          "SSH URL with wrong user",
			input:         "user@github.com:octocat/hello-world.git",
			expectedOwner: "",
			expectedRepo:  "",
			expectError:   true,
			errorContains: "Expected format: owner/repo",
		},
		{
			name:          "SSH URL with wrong separator",
			input:         "git@github.com/octocat/hello-world.git",
			expectedOwner: "",
			expectedRepo:  "",
			expectError:   true,
			errorContains: "Expected format: owner/repo",
		},
		// Malformed URLs
		{
			name:          "URL with query parameters",
			input:         "https://github.com/octocat/hello-world?ref=main",
			expectedOwner: "",
			expectedRepo:  "",
			expectError:   true,
			errorContains: "Invalid repository name characters",
		},
		{
			name:          "URL with fragment",
			input:         "https://github.com/octocat/hello-world#readme",
			expectedOwner: "",
			expectedRepo:  "",
			expectError:   true,
			errorContains: "Invalid repository name characters",
		},
		{
			name:          "URL with port number",
			input:         "https://github.com:443/octocat/hello-world",
			expectedOwner: "",
			expectedRepo:  "",
			expectError:   true,
			errorContains: "Expected format: owner/repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			p := parser.NewRepoParser()

			// Act
			owner, repo, err := p.Parse(tt.input)

			// Assert
			if tt.expectError {
				if err == nil {
					t.Fatalf("Parse(%q) expected error containing %q, got owner=%q repo=%q",
						tt.input, tt.errorContains, owner, repo)
				}
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Parse(%q) error = %q, expected to contain %q",
						tt.input, err.Error(), tt.errorContains)
				}
			} else {
				if err != nil {
					t.Fatalf("Parse(%q) unexpected error: %v", tt.input, err)
				}
				if owner != tt.expectedOwner {
					t.Errorf("Parse(%q) owner = %q, expected %q", tt.input, owner, tt.expectedOwner)
				}
				if repo != tt.expectedRepo {
					t.Errorf("Parse(%q) repo = %q, expected %q", tt.input, repo, tt.expectedRepo)
				}
			}
		})
	}
}

// =============================================================================
// Error Type Verification Tests
// =============================================================================

// TestParseReturnsAppErrorType verifies that all errors are *AppError type.
func TestParseReturnsAppErrorType(t *testing.T) {
	invalidInputs := []string{
		"",
		"   ",
		"single",
		"a/b/c",
		"owner@/repo",
		"owner/ repo",
		"https://gitlab.com/owner/repo",
	}

	p := parser.NewRepoParser()

	for _, input := range invalidInputs {
		t.Run(input, func(t *testing.T) {
			// Act
			_, _, err := p.Parse(input)

			// Assert
			if err == nil {
				t.Fatalf("Parse(%q) expected error, got nil", input)
			}

			var appErr *apperrors.AppError
			if !errors.As(err, &appErr) {
				t.Errorf("Parse(%q) error type = %T, expected *apperrors.AppError", input, err)
			}

			if appErr != nil && appErr.Code != apperrors.ErrInvalidArguments {
				t.Errorf("Parse(%q) error code = %d, expected %d (ErrInvalidArguments)",
					input, appErr.Code, apperrors.ErrInvalidArguments)
			}
		})
	}
}

// =============================================================================
// Consistency Tests
// =============================================================================

// TestParseSameResultForEquivalentInputs verifies consistent results for equivalent inputs.
func TestParseSameResultForEquivalentInputs(t *testing.T) {
	equivalentInputs := []struct {
		name   string
		inputs []string
	}{
		{
			name: "octocat/hello-world in all formats",
			inputs: []string{
				"octocat/hello-world",
				"  octocat/hello-world",
				"octocat/hello-world  ",
				"  octocat/hello-world  ",
				"https://github.com/octocat/hello-world",
				"https://github.com/octocat/hello-world.git",
				"git@github.com:octocat/hello-world.git",
				"git@github.com:octocat/hello-world",
			},
		},
	}

	p := parser.NewRepoParser()

	for _, tt := range equivalentInputs {
		t.Run(tt.name, func(t *testing.T) {
			// Get expected result from first input
			expectedOwner, expectedRepo, err := p.Parse(tt.inputs[0])
			if err != nil {
				t.Fatalf("Parse(%q) failed: %v", tt.inputs[0], err)
			}

			// Verify all other inputs produce same result
			for _, input := range tt.inputs[1:] {
				owner, repo, err := p.Parse(input)
				if err != nil {
					t.Errorf("Parse(%q) failed: %v", input, err)
					continue
				}
				if owner != expectedOwner || repo != expectedRepo {
					t.Errorf("Parse(%q) = (%q, %q), expected (%q, %q)",
						input, owner, repo, expectedOwner, expectedRepo)
				}
			}
		})
	}
}

// TestParseMultipleCallsReturnConsistentResults verifies parser is stateless.
func TestParseMultipleCallsReturnConsistentResults(t *testing.T) {
	// Arrange
	p := parser.NewRepoParser()
	input := "octocat/hello-world"

	// Act - call multiple times
	results := make([]struct {
		owner string
		repo  string
		err   error
	}, 5)

	for i := range results {
		results[i].owner, results[i].repo, results[i].err = p.Parse(input)
	}

	// Assert - all results should be identical
	for i := 1; i < len(results); i++ {
		if results[i].owner != results[0].owner {
			t.Errorf("Call %d owner = %q, expected %q", i, results[i].owner, results[0].owner)
		}
		if results[i].repo != results[0].repo {
			t.Errorf("Call %d repo = %q, expected %q", i, results[i].repo, results[0].repo)
		}
		if results[i].err != results[0].err {
			t.Errorf("Call %d err = %v, expected %v", i, results[i].err, results[0].err)
		}
	}
}
