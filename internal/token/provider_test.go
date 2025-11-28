// Package token_test provides tests for the TokenProvider implementation.
//
// These tests verify that TokenProvider correctly implements the ITokenProvider
// interface and retrieves tokens with the correct precedence:
// 1. Explicit token (from CLI flag)
// 2. GITHUB_TOKEN environment variable
// 3. gh CLI configuration file (~/.config/gh/hosts.yml)
//
// These tests are designed to FAIL until the implementation is properly created
// by the coder agent.
package token_test

import (
	"strings"
	"testing"

	"github.com/josejulio/ghautodelete/internal/token"
	"github.com/josejulio/ghautodelete/pkg/interfaces"
)

// =============================================================================
// Interface Satisfaction Tests
// =============================================================================

// TestTokenProviderImplementsITokenProvider verifies TokenProvider implements ITokenProvider.
//
// The implementation should:
// - Define TokenProvider struct in internal/token/provider.go
// - Implement GetToken() (string, error) method
func TestTokenProviderImplementsITokenProvider(t *testing.T) {
	// Arrange - create with dependency injection functions
	envGetter := func(key string) string { return "" }
	homeGetter := func() (string, error) { return "/home/test", nil }
	fileReader := func(path string) ([]byte, error) { return nil, nil }

	// Act - create TokenProvider
	provider := token.NewTokenProvider("", envGetter, homeGetter, fileReader)

	// Assert - compile-time interface satisfaction check
	var _ interfaces.ITokenProvider = provider
	if provider == nil {
		t.Error("NewTokenProvider should return a non-nil provider")
	}
}

// =============================================================================
// Explicit Token Tests (CLI Flag - Priority 1)
// =============================================================================

// TestGetTokenWithExplicitToken verifies that explicit token from CLI flag is returned.
//
// Gherkin: Scenario: Authenticate using token from command line flag
//
// The implementation should:
// - Accept explicit token in NewTokenProvider constructor
// - Return the explicit token if provided (non-empty)
func TestGetTokenWithExplicitToken(t *testing.T) {
	// Arrange
	explicitToken := "ghp_explicit_token_from_cli_12345"
	envGetter := func(key string) string { return "ghp_env_token_should_not_be_used" }
	homeGetter := func() (string, error) { return "/home/test", nil }
	fileReader := func(path string) ([]byte, error) {
		return []byte("github.com:\n  oauth_token: ghp_gh_cli_token_should_not_be_used\n"), nil
	}

	provider := token.NewTokenProvider(explicitToken, envGetter, homeGetter, fileReader)

	// Act
	result, err := provider.GetToken()

	// Assert
	if err != nil {
		t.Errorf("GetToken() unexpected error: %v", err)
	}
	if result != explicitToken {
		t.Errorf("GetToken() = %q, expected %q (explicit token should be returned)", result, explicitToken)
	}
}

// TestGetTokenWithExplicitTokenTrimsWhitespace verifies whitespace handling.
//
// The implementation should trim any surrounding whitespace from the token.
func TestGetTokenWithExplicitTokenTrimsWhitespace(t *testing.T) {
	tests := []struct {
		name          string
		explicitToken string
		expected      string
	}{
		{
			name:          "token with leading whitespace",
			explicitToken: "  ghp_token_12345",
			expected:      "ghp_token_12345",
		},
		{
			name:          "token with trailing whitespace",
			explicitToken: "ghp_token_12345  ",
			expected:      "ghp_token_12345",
		},
		{
			name:          "token with both leading and trailing whitespace",
			explicitToken: "  ghp_token_12345  ",
			expected:      "ghp_token_12345",
		},
		{
			name:          "token with newline",
			explicitToken: "ghp_token_12345\n",
			expected:      "ghp_token_12345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			envGetter := func(key string) string { return "" }
			homeGetter := func() (string, error) { return "/home/test", nil }
			fileReader := func(path string) ([]byte, error) { return nil, nil }

			provider := token.NewTokenProvider(tt.explicitToken, envGetter, homeGetter, fileReader)

			// Act
			result, err := provider.GetToken()

			// Assert
			if err != nil {
				t.Errorf("GetToken() unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("GetToken() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

// =============================================================================
// Environment Variable Tests (GITHUB_TOKEN - Priority 2)
// =============================================================================

// TestGetTokenWithEnvironmentVariable verifies GITHUB_TOKEN retrieval.
//
// Gherkin: Scenario: Authenticate using GITHUB_TOKEN environment variable
//
// The implementation should:
// - Check GITHUB_TOKEN environment variable if no explicit token
// - Return the environment variable value if set
func TestGetTokenWithEnvironmentVariable(t *testing.T) {
	// Arrange
	envToken := "ghp_environment_token_67890"
	envGetter := func(key string) string {
		if key == "GITHUB_TOKEN" {
			return envToken
		}
		return ""
	}
	homeGetter := func() (string, error) { return "/home/test", nil }
	fileReader := func(path string) ([]byte, error) {
		return []byte("github.com:\n  oauth_token: ghp_gh_cli_token_should_not_be_used\n"), nil
	}

	// No explicit token - empty string
	provider := token.NewTokenProvider("", envGetter, homeGetter, fileReader)

	// Act
	result, err := provider.GetToken()

	// Assert
	if err != nil {
		t.Errorf("GetToken() unexpected error: %v", err)
	}
	if result != envToken {
		t.Errorf("GetToken() = %q, expected %q (GITHUB_TOKEN should be returned)", result, envToken)
	}
}

// TestGetTokenEnvironmentVariableTrimsWhitespace verifies whitespace handling for env var.
func TestGetTokenEnvironmentVariableTrimsWhitespace(t *testing.T) {
	// Arrange
	envGetter := func(key string) string {
		if key == "GITHUB_TOKEN" {
			return "  ghp_token_with_whitespace  \n"
		}
		return ""
	}
	homeGetter := func() (string, error) { return "/home/test", nil }
	fileReader := func(path string) ([]byte, error) { return nil, nil }

	provider := token.NewTokenProvider("", envGetter, homeGetter, fileReader)

	// Act
	result, err := provider.GetToken()

	// Assert
	if err != nil {
		t.Errorf("GetToken() unexpected error: %v", err)
	}
	if result != "ghp_token_with_whitespace" {
		t.Errorf("GetToken() = %q, expected %q", result, "ghp_token_with_whitespace")
	}
}

// =============================================================================
// gh CLI Configuration Tests (Priority 3)
// =============================================================================

// TestGetTokenWithGhCLIConfig verifies gh CLI config file retrieval.
//
// Gherkin: Scenario: Authenticate using gh CLI configuration
//
// The implementation should:
// - Parse ~/.config/gh/hosts.yml file
// - Extract oauth_token for github.com host
func TestGetTokenWithGhCLIConfig(t *testing.T) {
	// Arrange
	ghConfigToken := "ghp_gh_cli_config_token_abcde"
	envGetter := func(key string) string { return "" }
	homeGetter := func() (string, error) { return "/home/testuser", nil }
	fileReader := func(path string) ([]byte, error) {
		if path == "/home/testuser/.config/gh/hosts.yml" {
			yamlContent := `github.com:
  oauth_token: ` + ghConfigToken + `
  user: testuser
  git_protocol: https`
			return []byte(yamlContent), nil
		}
		return nil, &mockFileNotFoundError{}
	}

	// No explicit token, no environment variable
	provider := token.NewTokenProvider("", envGetter, homeGetter, fileReader)

	// Act
	result, err := provider.GetToken()

	// Assert
	if err != nil {
		t.Errorf("GetToken() unexpected error: %v", err)
	}
	if result != ghConfigToken {
		t.Errorf("GetToken() = %q, expected %q (gh CLI token should be returned)", result, ghConfigToken)
	}
}

// TestGetTokenWithGhCLIConfigMultipleHosts verifies parsing with multiple hosts.
func TestGetTokenWithGhCLIConfigMultipleHosts(t *testing.T) {
	// Arrange
	expectedToken := "ghp_github_com_token"
	envGetter := func(key string) string { return "" }
	homeGetter := func() (string, error) { return "/home/testuser", nil }
	fileReader := func(path string) ([]byte, error) {
		if path == "/home/testuser/.config/gh/hosts.yml" {
			yamlContent := `github.enterprise.com:
  oauth_token: ghp_enterprise_token
  user: enterpriseuser
  git_protocol: https
github.com:
  oauth_token: ` + expectedToken + `
  user: testuser
  git_protocol: https`
			return []byte(yamlContent), nil
		}
		return nil, &mockFileNotFoundError{}
	}

	provider := token.NewTokenProvider("", envGetter, homeGetter, fileReader)

	// Act
	result, err := provider.GetToken()

	// Assert
	if err != nil {
		t.Errorf("GetToken() unexpected error: %v", err)
	}
	if result != expectedToken {
		t.Errorf("GetToken() = %q, expected %q (github.com token should be returned)", result, expectedToken)
	}
}

// TestGetTokenGhCLIConfigTrimsWhitespace verifies whitespace handling for gh config.
func TestGetTokenGhCLIConfigTrimsWhitespace(t *testing.T) {
	// Arrange
	envGetter := func(key string) string { return "" }
	homeGetter := func() (string, error) { return "/home/testuser", nil }
	fileReader := func(path string) ([]byte, error) {
		if path == "/home/testuser/.config/gh/hosts.yml" {
			yamlContent := `github.com:
  oauth_token: "  ghp_token_with_spaces  "`
			return []byte(yamlContent), nil
		}
		return nil, &mockFileNotFoundError{}
	}

	provider := token.NewTokenProvider("", envGetter, homeGetter, fileReader)

	// Act
	result, err := provider.GetToken()

	// Assert
	if err != nil {
		t.Errorf("GetToken() unexpected error: %v", err)
	}
	if result != "ghp_token_with_spaces" {
		t.Errorf("GetToken() = %q, expected %q", result, "ghp_token_with_spaces")
	}
}

// =============================================================================
// Precedence Tests
// =============================================================================

// TestGetTokenExplicitTakesPrecedenceOverEnvironment verifies explicit > env precedence.
//
// Gherkin: Scenario: Token flag takes precedence over environment variable
//
// The implementation should:
// - Return explicit token even when GITHUB_TOKEN is set
func TestGetTokenExplicitTakesPrecedenceOverEnvironment(t *testing.T) {
	// Arrange
	explicitToken := "ghp_explicit_wins"
	envToken := "ghp_env_should_be_ignored"
	envGetter := func(key string) string {
		if key == "GITHUB_TOKEN" {
			return envToken
		}
		return ""
	}
	homeGetter := func() (string, error) { return "/home/test", nil }
	fileReader := func(path string) ([]byte, error) {
		return []byte("github.com:\n  oauth_token: ghp_gh_cli_should_be_ignored\n"), nil
	}

	provider := token.NewTokenProvider(explicitToken, envGetter, homeGetter, fileReader)

	// Act
	result, err := provider.GetToken()

	// Assert
	if err != nil {
		t.Errorf("GetToken() unexpected error: %v", err)
	}
	if result != explicitToken {
		t.Errorf("GetToken() = %q, expected %q (explicit should take precedence over env)", result, explicitToken)
	}
}

// TestGetTokenEnvironmentTakesPrecedenceOverGhConfig verifies env > gh config precedence.
//
// Gherkin: Scenario: Environment variable takes precedence over gh CLI config
//
// The implementation should:
// - Return GITHUB_TOKEN even when gh CLI config exists
func TestGetTokenEnvironmentTakesPrecedenceOverGhConfig(t *testing.T) {
	// Arrange
	envToken := "ghp_env_wins"
	ghConfigToken := "ghp_gh_cli_should_be_ignored"
	envGetter := func(key string) string {
		if key == "GITHUB_TOKEN" {
			return envToken
		}
		return ""
	}
	homeGetter := func() (string, error) { return "/home/test", nil }
	fileReader := func(path string) ([]byte, error) {
		return []byte("github.com:\n  oauth_token: " + ghConfigToken + "\n"), nil
	}

	// No explicit token
	provider := token.NewTokenProvider("", envGetter, homeGetter, fileReader)

	// Act
	result, err := provider.GetToken()

	// Assert
	if err != nil {
		t.Errorf("GetToken() unexpected error: %v", err)
	}
	if result != envToken {
		t.Errorf("GetToken() = %q, expected %q (env should take precedence over gh config)", result, envToken)
	}
}

// TestGetTokenExplicitTakesPrecedenceOverAll verifies explicit > all precedence.
//
// The implementation should:
// - Return explicit token when all sources are available
func TestGetTokenExplicitTakesPrecedenceOverAll(t *testing.T) {
	// Arrange
	explicitToken := "ghp_explicit_wins_over_all"
	envToken := "ghp_env_should_be_ignored"
	ghConfigToken := "ghp_gh_cli_should_be_ignored"
	envGetter := func(key string) string {
		if key == "GITHUB_TOKEN" {
			return envToken
		}
		return ""
	}
	homeGetter := func() (string, error) { return "/home/test", nil }
	fileReader := func(path string) ([]byte, error) {
		return []byte("github.com:\n  oauth_token: " + ghConfigToken + "\n"), nil
	}

	provider := token.NewTokenProvider(explicitToken, envGetter, homeGetter, fileReader)

	// Act
	result, err := provider.GetToken()

	// Assert
	if err != nil {
		t.Errorf("GetToken() unexpected error: %v", err)
	}
	if result != explicitToken {
		t.Errorf("GetToken() = %q, expected %q (explicit should take precedence over all)", result, explicitToken)
	}
}

// =============================================================================
// Error Case Tests
// =============================================================================

// TestGetTokenNoTokenSourceAvailable verifies error when no token found.
//
// Gherkin: Scenario: Fail when no token source is available
//   -> error "No GitHub token found" + "Set GITHUB_TOKEN or use --token flag"
//
// The implementation should:
// - Return an error when no token source is available
// - Error message should contain "No GitHub token found"
// - Error message should mention "Set GITHUB_TOKEN or use --token flag"
func TestGetTokenNoTokenSourceAvailable(t *testing.T) {
	// Arrange
	envGetter := func(key string) string { return "" }
	homeGetter := func() (string, error) { return "/home/test", nil }
	fileReader := func(path string) ([]byte, error) {
		return nil, &mockFileNotFoundError{}
	}

	provider := token.NewTokenProvider("", envGetter, homeGetter, fileReader)

	// Act
	result, err := provider.GetToken()

	// Assert
	if err == nil {
		t.Fatal("GetToken() expected error when no token source available, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "No GitHub token found") {
		t.Errorf("Error message should contain 'No GitHub token found', got: %q", errMsg)
	}
	if !strings.Contains(errMsg, "GITHUB_TOKEN") && !strings.Contains(errMsg, "--token") {
		t.Errorf("Error message should mention GITHUB_TOKEN or --token flag, got: %q", errMsg)
	}
	if result != "" {
		t.Errorf("GetToken() result should be empty string on error, got: %q", result)
	}
}

// TestGetTokenEmptyExplicitTokenFallsThrough verifies empty explicit falls to next source.
//
// The implementation should:
// - Treat empty explicit token as "not provided"
// - Fall through to environment variable
func TestGetTokenEmptyExplicitTokenFallsThrough(t *testing.T) {
	// Arrange
	envToken := "ghp_fallback_to_env"
	envGetter := func(key string) string {
		if key == "GITHUB_TOKEN" {
			return envToken
		}
		return ""
	}
	homeGetter := func() (string, error) { return "/home/test", nil }
	fileReader := func(path string) ([]byte, error) { return nil, &mockFileNotFoundError{} }

	// Empty explicit token
	provider := token.NewTokenProvider("", envGetter, homeGetter, fileReader)

	// Act
	result, err := provider.GetToken()

	// Assert
	if err != nil {
		t.Errorf("GetToken() unexpected error: %v", err)
	}
	if result != envToken {
		t.Errorf("GetToken() = %q, expected %q (should fall through to env)", result, envToken)
	}
}

// TestGetTokenWhitespaceOnlyExplicitTokenFallsThrough verifies whitespace-only explicit falls through.
//
// The implementation should:
// - Treat whitespace-only explicit token as "not provided"
// - Fall through to environment variable
func TestGetTokenWhitespaceOnlyExplicitTokenFallsThrough(t *testing.T) {
	// Arrange
	envToken := "ghp_fallback_to_env_after_whitespace"
	envGetter := func(key string) string {
		if key == "GITHUB_TOKEN" {
			return envToken
		}
		return ""
	}
	homeGetter := func() (string, error) { return "/home/test", nil }
	fileReader := func(path string) ([]byte, error) { return nil, &mockFileNotFoundError{} }

	// Whitespace-only explicit token
	provider := token.NewTokenProvider("   \n\t  ", envGetter, homeGetter, fileReader)

	// Act
	result, err := provider.GetToken()

	// Assert
	if err != nil {
		t.Errorf("GetToken() unexpected error: %v", err)
	}
	if result != envToken {
		t.Errorf("GetToken() = %q, expected %q (whitespace-only should fall through to env)", result, envToken)
	}
}

// =============================================================================
// gh CLI Config File Parsing Tests
// =============================================================================

// TestGetTokenGhConfigInvalidYAML verifies handling of invalid YAML.
//
// The implementation should:
// - Handle invalid YAML gracefully
// - Return error when no other source available and gh config is invalid
func TestGetTokenGhConfigInvalidYAML(t *testing.T) {
	// Arrange
	envGetter := func(key string) string { return "" }
	homeGetter := func() (string, error) { return "/home/test", nil }
	fileReader := func(path string) ([]byte, error) {
		// Invalid YAML content
		return []byte("this is not: valid yaml:\n  missing: proper:\n structure"), nil
	}

	provider := token.NewTokenProvider("", envGetter, homeGetter, fileReader)

	// Act
	result, err := provider.GetToken()

	// Assert
	if err == nil {
		t.Fatal("GetToken() expected error for invalid YAML, got nil")
	}
	if result != "" {
		t.Errorf("GetToken() result should be empty string on error, got: %q", result)
	}
}

// TestGetTokenGhConfigMissingFile verifies handling of missing config file.
//
// The implementation should:
// - Handle missing file gracefully
// - Return error when no other source available
func TestGetTokenGhConfigMissingFile(t *testing.T) {
	// Arrange
	envGetter := func(key string) string { return "" }
	homeGetter := func() (string, error) { return "/home/test", nil }
	fileReader := func(path string) ([]byte, error) {
		return nil, &mockFileNotFoundError{}
	}

	provider := token.NewTokenProvider("", envGetter, homeGetter, fileReader)

	// Act
	result, err := provider.GetToken()

	// Assert
	if err == nil {
		t.Fatal("GetToken() expected error when gh config missing, got nil")
	}
	errMsg := err.Error()
	if !strings.Contains(errMsg, "No GitHub token found") {
		t.Errorf("Error message should contain 'No GitHub token found', got: %q", errMsg)
	}
	if result != "" {
		t.Errorf("GetToken() result should be empty string on error, got: %q", result)
	}
}

// TestGetTokenGhConfigMissingGitHubComHost verifies handling when github.com host not found.
//
// The implementation should:
// - Handle missing github.com host entry
// - Return error when no other source available
func TestGetTokenGhConfigMissingGitHubComHost(t *testing.T) {
	// Arrange
	envGetter := func(key string) string { return "" }
	homeGetter := func() (string, error) { return "/home/test", nil }
	fileReader := func(path string) ([]byte, error) {
		yamlContent := `github.enterprise.com:
  oauth_token: ghp_enterprise_only
  user: enterpriseuser`
		return []byte(yamlContent), nil
	}

	provider := token.NewTokenProvider("", envGetter, homeGetter, fileReader)

	// Act
	result, err := provider.GetToken()

	// Assert
	if err == nil {
		t.Fatal("GetToken() expected error when github.com host not in config, got nil")
	}
	if result != "" {
		t.Errorf("GetToken() result should be empty string on error, got: %q", result)
	}
}

// TestGetTokenGhConfigMissingOAuthToken verifies handling when oauth_token is missing.
//
// The implementation should:
// - Handle missing oauth_token field for github.com
// - Return error when no other source available
func TestGetTokenGhConfigMissingOAuthToken(t *testing.T) {
	// Arrange
	envGetter := func(key string) string { return "" }
	homeGetter := func() (string, error) { return "/home/test", nil }
	fileReader := func(path string) ([]byte, error) {
		yamlContent := `github.com:
  user: testuser
  git_protocol: https`
		return []byte(yamlContent), nil
	}

	provider := token.NewTokenProvider("", envGetter, homeGetter, fileReader)

	// Act
	result, err := provider.GetToken()

	// Assert
	if err == nil {
		t.Fatal("GetToken() expected error when oauth_token missing, got nil")
	}
	if result != "" {
		t.Errorf("GetToken() result should be empty string on error, got: %q", result)
	}
}

// TestGetTokenGhConfigEmptyOAuthToken verifies handling when oauth_token is empty.
//
// The implementation should:
// - Handle empty oauth_token value
// - Return error when no other source available
func TestGetTokenGhConfigEmptyOAuthToken(t *testing.T) {
	// Arrange
	envGetter := func(key string) string { return "" }
	homeGetter := func() (string, error) { return "/home/test", nil }
	fileReader := func(path string) ([]byte, error) {
		yamlContent := `github.com:
  oauth_token: ""
  user: testuser`
		return []byte(yamlContent), nil
	}

	provider := token.NewTokenProvider("", envGetter, homeGetter, fileReader)

	// Act
	result, err := provider.GetToken()

	// Assert
	if err == nil {
		t.Fatal("GetToken() expected error when oauth_token is empty, got nil")
	}
	if result != "" {
		t.Errorf("GetToken() result should be empty string on error, got: %q", result)
	}
}

// TestGetTokenGhConfigValidYAMLFormats verifies various valid YAML formats.
func TestGetTokenGhConfigValidYAMLFormats(t *testing.T) {
	tests := []struct {
		name        string
		yamlContent string
		expected    string
	}{
		{
			name: "standard format",
			yamlContent: `github.com:
  oauth_token: ghp_standard_format
  user: testuser`,
			expected: "ghp_standard_format",
		},
		{
			name: "quoted token",
			yamlContent: `github.com:
  oauth_token: "ghp_quoted_format"
  user: testuser`,
			expected: "ghp_quoted_format",
		},
		{
			name: "single quoted token",
			yamlContent: `github.com:
  oauth_token: 'ghp_single_quoted'
  user: testuser`,
			expected: "ghp_single_quoted",
		},
		{
			name: "minimal format",
			yamlContent: `github.com:
  oauth_token: ghp_minimal`,
			expected: "ghp_minimal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			envGetter := func(key string) string { return "" }
			homeGetter := func() (string, error) { return "/home/test", nil }
			fileReader := func(path string) ([]byte, error) {
				return []byte(tt.yamlContent), nil
			}

			provider := token.NewTokenProvider("", envGetter, homeGetter, fileReader)

			// Act
			result, err := provider.GetToken()

			// Assert
			if err != nil {
				t.Errorf("GetToken() unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("GetToken() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

// =============================================================================
// Home Directory Error Tests
// =============================================================================

// TestGetTokenHomeDirectoryError verifies error handling when home dir fails.
//
// The implementation should:
// - Handle home directory retrieval failure gracefully
// - Return error when no other source available
func TestGetTokenHomeDirectoryError(t *testing.T) {
	// Arrange
	envGetter := func(key string) string { return "" }
	homeGetter := func() (string, error) {
		return "", &mockHomeDirectoryError{}
	}
	fileReader := func(path string) ([]byte, error) {
		return nil, &mockFileNotFoundError{}
	}

	provider := token.NewTokenProvider("", envGetter, homeGetter, fileReader)

	// Act
	result, err := provider.GetToken()

	// Assert
	if err == nil {
		t.Fatal("GetToken() expected error when home directory fails, got nil")
	}
	if result != "" {
		t.Errorf("GetToken() result should be empty string on error, got: %q", result)
	}
}

// TestGetTokenHomeDirectoryErrorWithEnvFallback verifies env var works despite home error.
//
// The implementation should:
// - Return env var token even if home directory retrieval fails
func TestGetTokenHomeDirectoryErrorWithEnvFallback(t *testing.T) {
	// Arrange
	envToken := "ghp_env_works_despite_home_error"
	envGetter := func(key string) string {
		if key == "GITHUB_TOKEN" {
			return envToken
		}
		return ""
	}
	homeGetter := func() (string, error) {
		return "", &mockHomeDirectoryError{}
	}
	fileReader := func(path string) ([]byte, error) {
		return nil, &mockFileNotFoundError{}
	}

	provider := token.NewTokenProvider("", envGetter, homeGetter, fileReader)

	// Act
	result, err := provider.GetToken()

	// Assert
	if err != nil {
		t.Errorf("GetToken() unexpected error: %v", err)
	}
	if result != envToken {
		t.Errorf("GetToken() = %q, expected %q", result, envToken)
	}
}

// =============================================================================
// Constructor Tests
// =============================================================================

// TestNewTokenProviderReturnsNonNil verifies constructor returns non-nil.
//
// The implementation should:
// - Return a non-nil TokenProvider from NewTokenProvider
func TestNewTokenProviderReturnsNonNil(t *testing.T) {
	// Arrange
	envGetter := func(key string) string { return "" }
	homeGetter := func() (string, error) { return "/home/test", nil }
	fileReader := func(path string) ([]byte, error) { return nil, nil }

	// Act
	provider := token.NewTokenProvider("", envGetter, homeGetter, fileReader)

	// Assert
	if provider == nil {
		t.Error("NewTokenProvider() should return non-nil provider")
	}
}

// TestNewTokenProviderAcceptsAllDependencies verifies constructor accepts all deps.
//
// The implementation should:
// - Accept explicit token, env getter, home getter, and file reader
func TestNewTokenProviderAcceptsAllDependencies(t *testing.T) {
	// Arrange
	explicitToken := "ghp_test_token"
	envGetterCalled := false
	homeGetterCalled := false
	fileReaderCalled := false

	envGetter := func(key string) string {
		envGetterCalled = true
		return ""
	}
	homeGetter := func() (string, error) {
		homeGetterCalled = true
		return "/home/test", nil
	}
	fileReader := func(path string) ([]byte, error) {
		fileReaderCalled = true
		return nil, &mockFileNotFoundError{}
	}

	// Act - create provider and get token (with explicit token, deps shouldn't be called)
	provider := token.NewTokenProvider(explicitToken, envGetter, homeGetter, fileReader)
	result, err := provider.GetToken()

	// Assert
	if err != nil {
		t.Errorf("GetToken() unexpected error: %v", err)
	}
	if result != explicitToken {
		t.Errorf("GetToken() = %q, expected %q", result, explicitToken)
	}
	// With explicit token, other getters should not be called
	if envGetterCalled {
		t.Error("envGetter should not be called when explicit token is provided")
	}
	if homeGetterCalled {
		t.Error("homeGetter should not be called when explicit token is provided")
	}
	if fileReaderCalled {
		t.Error("fileReader should not be called when explicit token is provided")
	}
}

// TestNewTokenProviderDependenciesCalledInOrder verifies fallback order.
//
// The implementation should:
// - Call dependencies in order: explicit -> env -> gh config
func TestNewTokenProviderDependenciesCalledInOrder(t *testing.T) {
	// Arrange
	callOrder := []string{}
	ghConfigToken := "ghp_final_fallback"

	envGetter := func(key string) string {
		callOrder = append(callOrder, "envGetter")
		return ""
	}
	homeGetter := func() (string, error) {
		callOrder = append(callOrder, "homeGetter")
		return "/home/test", nil
	}
	fileReader := func(path string) ([]byte, error) {
		callOrder = append(callOrder, "fileReader")
		return []byte("github.com:\n  oauth_token: " + ghConfigToken + "\n"), nil
	}

	// No explicit token - should fallback through all sources
	provider := token.NewTokenProvider("", envGetter, homeGetter, fileReader)

	// Act
	result, err := provider.GetToken()

	// Assert
	if err != nil {
		t.Errorf("GetToken() unexpected error: %v", err)
	}
	if result != ghConfigToken {
		t.Errorf("GetToken() = %q, expected %q", result, ghConfigToken)
	}
	// Verify call order: first check env, then home, then file
	if len(callOrder) != 3 {
		t.Errorf("Expected 3 dependency calls, got %d: %v", len(callOrder), callOrder)
	}
	if len(callOrder) >= 1 && callOrder[0] != "envGetter" {
		t.Errorf("First call should be envGetter, got %v", callOrder)
	}
	if len(callOrder) >= 2 && callOrder[1] != "homeGetter" {
		t.Errorf("Second call should be homeGetter, got %v", callOrder)
	}
	if len(callOrder) >= 3 && callOrder[2] != "fileReader" {
		t.Errorf("Third call should be fileReader, got %v", callOrder)
	}
}

// =============================================================================
// Table-Driven Complete Scenario Tests
// =============================================================================

// TestGetTokenCompleteScenarios tests all authentication scenarios in one table.
func TestGetTokenCompleteScenarios(t *testing.T) {
	tests := []struct {
		name          string
		explicitToken string
		envToken      string
		ghConfigYAML  string
		ghConfigError error
		homeError     error
		expected      string
		expectError   bool
		errorContains string
	}{
		{
			name:          "explicit token only",
			explicitToken: "ghp_explicit_only",
			envToken:      "",
			ghConfigYAML:  "",
			expected:      "ghp_explicit_only",
			expectError:   false,
		},
		{
			name:          "env token only",
			explicitToken: "",
			envToken:      "ghp_env_only",
			ghConfigYAML:  "",
			expected:      "ghp_env_only",
			expectError:   false,
		},
		{
			name:          "gh config token only",
			explicitToken: "",
			envToken:      "",
			ghConfigYAML:  "github.com:\n  oauth_token: ghp_gh_config_only\n",
			expected:      "ghp_gh_config_only",
			expectError:   false,
		},
		{
			name:          "explicit over env",
			explicitToken: "ghp_explicit_wins",
			envToken:      "ghp_env_loses",
			ghConfigYAML:  "",
			expected:      "ghp_explicit_wins",
			expectError:   false,
		},
		{
			name:          "explicit over gh config",
			explicitToken: "ghp_explicit_wins_again",
			envToken:      "",
			ghConfigYAML:  "github.com:\n  oauth_token: ghp_gh_loses\n",
			expected:      "ghp_explicit_wins_again",
			expectError:   false,
		},
		{
			name:          "env over gh config",
			explicitToken: "",
			envToken:      "ghp_env_wins",
			ghConfigYAML:  "github.com:\n  oauth_token: ghp_gh_loses\n",
			expected:      "ghp_env_wins",
			expectError:   false,
		},
		{
			name:          "no token source - error",
			explicitToken: "",
			envToken:      "",
			ghConfigYAML:  "",
			ghConfigError: &mockFileNotFoundError{},
			expected:      "",
			expectError:   true,
			errorContains: "No GitHub token found",
		},
		{
			name:          "empty env falls through to gh config",
			explicitToken: "",
			envToken:      "",
			ghConfigYAML:  "github.com:\n  oauth_token: ghp_fallback_gh\n",
			expected:      "ghp_fallback_gh",
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			envGetter := func(key string) string {
				if key == "GITHUB_TOKEN" {
					return tt.envToken
				}
				return ""
			}
			homeGetter := func() (string, error) {
				if tt.homeError != nil {
					return "", tt.homeError
				}
				return "/home/test", nil
			}
			fileReader := func(path string) ([]byte, error) {
				if tt.ghConfigError != nil {
					return nil, tt.ghConfigError
				}
				if tt.ghConfigYAML == "" {
					return nil, &mockFileNotFoundError{}
				}
				return []byte(tt.ghConfigYAML), nil
			}

			provider := token.NewTokenProvider(tt.explicitToken, envGetter, homeGetter, fileReader)

			// Act
			result, err := provider.GetToken()

			// Assert
			if tt.expectError {
				if err == nil {
					t.Fatal("Expected error, got nil")
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Error message should contain %q, got: %q", tt.errorContains, err.Error())
				}
				if result != "" {
					t.Errorf("Result should be empty on error, got: %q", result)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("GetToken() = %q, expected %q", result, tt.expected)
				}
			}
		})
	}
}

// =============================================================================
// Mock Error Types for Testing
// =============================================================================

// mockFileNotFoundError simulates file not found errors.
type mockFileNotFoundError struct{}

func (e *mockFileNotFoundError) Error() string {
	return "no such file or directory"
}

// mockHomeDirectoryError simulates home directory retrieval errors.
type mockHomeDirectoryError struct{}

func (e *mockHomeDirectoryError) Error() string {
	return "could not determine home directory"
}
