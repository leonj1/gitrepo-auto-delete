// Package token_test provides tests for the token package data models.
//
// These tests verify that TokenInfo struct correctly implements the ITokenInfo interface.
//
// These tests are designed to FAIL until the implementations are properly created
// by the coder agent.
package token_test

import (
	"testing"

	"github.com/josejulio/ghautodelete/internal/token"
	"github.com/josejulio/ghautodelete/pkg/interfaces"
)

// =============================================================================
// TokenInfo Struct Tests
// =============================================================================

// TestTokenInfoImplementsITokenInfo verifies TokenInfo implements ITokenInfo.
//
// The implementation should:
// - Define TokenInfo struct in internal/token/models.go
// - Implement all ITokenInfo methods
func TestTokenInfoImplementsITokenInfo(t *testing.T) {
	// Arrange & Act - compile-time interface satisfaction check
	var tokenInfo interfaces.ITokenInfo = &token.TokenInfo{}

	// Assert
	if tokenInfo == nil {
		t.Error("TokenInfo should implement ITokenInfo interface")
	}
}

// TestTokenInfoGetScopes verifies TokenInfo.GetScopes() returns the scopes.
//
// The implementation should:
// - Store scopes slice in TokenInfo struct
// - Return the scopes []string from GetScopes()
func TestTokenInfoGetScopes(t *testing.T) {
	tests := []struct {
		name     string
		scopes   []string
		expected []string
	}{
		{
			name:     "returns single scope",
			scopes:   []string{"repo"},
			expected: []string{"repo"},
		},
		{
			name:     "returns multiple scopes",
			scopes:   []string{"repo", "user", "admin:org"},
			expected: []string{"repo", "user", "admin:org"},
		},
		{
			name:     "returns empty slice when no scopes",
			scopes:   []string{},
			expected: []string{},
		},
		{
			name:     "returns nil as nil",
			scopes:   nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			tokenInfo := token.TokenInfo{
				Scopes: tt.scopes,
			}

			// Act
			result := tokenInfo.GetScopes()

			// Assert
			if len(result) != len(tt.expected) {
				t.Errorf("GetScopes() length = %d, expected %d", len(result), len(tt.expected))
				return
			}
			for i, scope := range result {
				if scope != tt.expected[i] {
					t.Errorf("GetScopes()[%d] = %q, expected %q", i, scope, tt.expected[i])
				}
			}
		})
	}
}

// TestTokenInfoHasScope verifies TokenInfo.HasScope() checks for scope presence.
//
// The implementation should:
// - Return true if the scope is present in the Scopes slice
// - Return false if the scope is not present
func TestTokenInfoHasScope(t *testing.T) {
	tests := []struct {
		name     string
		scopes   []string
		scope    string
		expected bool
	}{
		{
			name:     "returns true when scope exists",
			scopes:   []string{"repo", "user", "admin:org"},
			scope:    "repo",
			expected: true,
		},
		{
			name:     "returns true for scope in middle of list",
			scopes:   []string{"repo", "user", "admin:org"},
			scope:    "user",
			expected: true,
		},
		{
			name:     "returns true for scope at end of list",
			scopes:   []string{"repo", "user", "admin:org"},
			scope:    "admin:org",
			expected: true,
		},
		{
			name:     "returns false when scope does not exist",
			scopes:   []string{"repo", "user"},
			scope:    "admin:org",
			expected: false,
		},
		{
			name:     "returns false when scopes is empty",
			scopes:   []string{},
			scope:    "repo",
			expected: false,
		},
		{
			name:     "returns false when scopes is nil",
			scopes:   nil,
			scope:    "repo",
			expected: false,
		},
		{
			name:     "returns false for partial scope match",
			scopes:   []string{"repo:status"},
			scope:    "repo",
			expected: false,
		},
		{
			name:     "is case sensitive",
			scopes:   []string{"Repo"},
			scope:    "repo",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			tokenInfo := token.TokenInfo{
				Scopes: tt.scopes,
			}

			// Act
			result := tokenInfo.HasScope(tt.scope)

			// Assert
			if result != tt.expected {
				t.Errorf("HasScope(%q) = %v, expected %v", tt.scope, result, tt.expected)
			}
		})
	}
}

// TestTokenInfoGetUsername verifies TokenInfo.GetUsername() returns the username.
//
// The implementation should:
// - Store username in TokenInfo struct
// - Return the username string from GetUsername()
func TestTokenInfoGetUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		expected string
	}{
		{
			name:     "returns username for standard user",
			username: "octocat",
			expected: "octocat",
		},
		{
			name:     "returns username with numbers",
			username: "user123",
			expected: "user123",
		},
		{
			name:     "returns username with hyphens",
			username: "my-user-name",
			expected: "my-user-name",
		},
		{
			name:     "returns empty string for empty username",
			username: "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			tokenInfo := token.TokenInfo{
				Username: tt.username,
			}

			// Act
			result := tokenInfo.GetUsername()

			// Assert
			if result != tt.expected {
				t.Errorf("GetUsername() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

// TestTokenInfoAllFieldsSet verifies TokenInfo works with all fields set.
//
// The implementation should correctly return all values when all fields are set.
func TestTokenInfoAllFieldsSet(t *testing.T) {
	// Arrange
	tokenInfo := token.TokenInfo{
		Scopes:   []string{"repo", "user", "admin:org"},
		Username: "octocat",
	}

	// Act & Assert - GetScopes
	scopes := tokenInfo.GetScopes()
	if len(scopes) != 3 {
		t.Errorf("GetScopes() length = %d, expected %d", len(scopes), 3)
	}

	// Act & Assert - HasScope
	if !tokenInfo.HasScope("repo") {
		t.Error("HasScope(\"repo\") = false, expected true")
	}
	if !tokenInfo.HasScope("user") {
		t.Error("HasScope(\"user\") = false, expected true")
	}
	if !tokenInfo.HasScope("admin:org") {
		t.Error("HasScope(\"admin:org\") = false, expected true")
	}
	if tokenInfo.HasScope("nonexistent") {
		t.Error("HasScope(\"nonexistent\") = true, expected false")
	}

	// Act & Assert - GetUsername
	if username := tokenInfo.GetUsername(); username != "octocat" {
		t.Errorf("GetUsername() = %q, expected %q", username, "octocat")
	}
}

// TestTokenInfoAsITokenInfo verifies TokenInfo can be used as ITokenInfo.
//
// The implementation should allow TokenInfo to be passed where ITokenInfo is expected.
func TestTokenInfoAsITokenInfo(t *testing.T) {
	// Arrange
	tokenInfo := &token.TokenInfo{
		Scopes:   []string{"repo"},
		Username: "testuser",
	}

	// Act - function that accepts ITokenInfo
	username := acceptITokenInfo(tokenInfo)

	// Assert
	if username != "testuser" {
		t.Errorf("acceptITokenInfo() = %q, expected %q", username, "testuser")
	}
}

// Helper function that accepts ITokenInfo interface
func acceptITokenInfo(tokenInfo interfaces.ITokenInfo) string {
	return tokenInfo.GetUsername()
}

// =============================================================================
// Constructor Tests
// =============================================================================

// TestNewTokenInfo verifies NewTokenInfo constructor creates valid TokenInfo.
//
// The implementation should:
// - Define NewTokenInfo(username string, scopes []string) *TokenInfo
func TestNewTokenInfo(t *testing.T) {
	// Arrange
	username := "octocat"
	scopes := []string{"repo", "user"}

	// Act
	tokenInfo := token.NewTokenInfo(username, scopes)

	// Assert
	if tokenInfo == nil {
		t.Fatal("NewTokenInfo() returned nil")
	}
	if tokenInfo.GetUsername() != username {
		t.Errorf("GetUsername() = %q, expected %q", tokenInfo.GetUsername(), username)
	}
	resultScopes := tokenInfo.GetScopes()
	if len(resultScopes) != len(scopes) {
		t.Errorf("GetScopes() length = %d, expected %d", len(resultScopes), len(scopes))
	}
	for i, scope := range resultScopes {
		if scope != scopes[i] {
			t.Errorf("GetScopes()[%d] = %q, expected %q", i, scope, scopes[i])
		}
	}
}

// TestNewTokenInfoWithEmptyScopes verifies NewTokenInfo handles empty scopes.
func TestNewTokenInfoWithEmptyScopes(t *testing.T) {
	// Arrange
	username := "octocat"
	scopes := []string{}

	// Act
	tokenInfo := token.NewTokenInfo(username, scopes)

	// Assert
	if tokenInfo == nil {
		t.Fatal("NewTokenInfo() returned nil")
	}
	if len(tokenInfo.GetScopes()) != 0 {
		t.Errorf("GetScopes() length = %d, expected 0", len(tokenInfo.GetScopes()))
	}
}

// TestNewTokenInfoWithNilScopes verifies NewTokenInfo handles nil scopes.
func TestNewTokenInfoWithNilScopes(t *testing.T) {
	// Arrange
	username := "octocat"
	var scopes []string = nil

	// Act
	tokenInfo := token.NewTokenInfo(username, scopes)

	// Assert
	if tokenInfo == nil {
		t.Fatal("NewTokenInfo() returned nil")
	}
	// Implementation may return nil or empty slice - both are acceptable
	resultScopes := tokenInfo.GetScopes()
	if resultScopes != nil && len(resultScopes) != 0 {
		t.Errorf("GetScopes() = %v, expected nil or empty slice", resultScopes)
	}
}

// =============================================================================
// Common OAuth Scope Tests
// =============================================================================

// TestTokenInfoCommonGitHubScopes tests common GitHub OAuth scopes.
//
// These tests verify HasScope works correctly with common GitHub scopes.
func TestTokenInfoCommonGitHubScopes(t *testing.T) {
	// Common GitHub scopes reference:
	// repo - Full control of private repositories
	// repo:status - Access commit status
	// repo_deployment - Access deployment status
	// public_repo - Access public repositories
	// repo:invite - Access repository invitations
	// security_events - Read and write security events
	// admin:repo_hook - Full control of repository hooks
	// write:repo_hook - Write repository hooks
	// read:repo_hook - Read repository hooks
	// admin:org - Full control of orgs and teams
	// write:org - Read and write org and team membership
	// read:org - Read org and team membership
	// admin:public_key - Full control of user public keys
	// write:public_key - Write user public keys
	// read:public_key - Read user public keys
	// admin:org_hook - Full control of organization hooks
	// gist - Create gists
	// notifications - Access notifications
	// user - Update all user data
	// read:user - Read all user profile data
	// user:email - Access user email addresses (read-only)
	// user:follow - Follow and unfollow users
	// delete_repo - Delete repositories
	// write:discussion - Read and write team discussions
	// read:discussion - Read team discussions
	// write:packages - Upload packages to GitHub Package Registry
	// read:packages - Download packages from GitHub Package Registry
	// delete:packages - Delete packages from GitHub Package Registry
	// admin:gpg_key - Full control of user GPG keys
	// write:gpg_key - Write user GPG keys
	// read:gpg_key - Read user GPG keys
	// workflow - Update GitHub Action workflows

	tests := []struct {
		name     string
		scopes   []string
		testFor  string
		expected bool
	}{
		{
			name:     "has repo scope",
			scopes:   []string{"repo"},
			testFor:  "repo",
			expected: true,
		},
		{
			name:     "has public_repo scope",
			scopes:   []string{"public_repo"},
			testFor:  "public_repo",
			expected: true,
		},
		{
			name:     "has admin:org scope",
			scopes:   []string{"admin:org"},
			testFor:  "admin:org",
			expected: true,
		},
		{
			name:     "has workflow scope",
			scopes:   []string{"workflow"},
			testFor:  "workflow",
			expected: true,
		},
		{
			name:     "does not have delete_repo scope when only repo",
			scopes:   []string{"repo"},
			testFor:  "delete_repo",
			expected: false,
		},
		{
			name:     "has multiple scopes and finds specific one",
			scopes:   []string{"repo", "user", "workflow", "delete_repo"},
			testFor:  "delete_repo",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			tokenInfo := token.TokenInfo{
				Scopes:   tt.scopes,
				Username: "testuser",
			}

			// Act
			result := tokenInfo.HasScope(tt.testFor)

			// Assert
			if result != tt.expected {
				t.Errorf("HasScope(%q) = %v, expected %v (scopes: %v)", tt.testFor, result, tt.expected, tt.scopes)
			}
		})
	}
}
