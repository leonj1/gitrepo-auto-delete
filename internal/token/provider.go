// Package token provides functionality for retrieving GitHub API tokens.
//
// TokenProvider implements the ITokenProvider interface and retrieves tokens
// with the following precedence:
// 1. Explicit token (from CLI flag)
// 2. GITHUB_TOKEN environment variable
// 3. gh CLI configuration file (~/.config/gh/hosts.yml)
package token

import (
	"path/filepath"
	"strings"

	"github.com/josejulio/ghautodelete/internal/errors"
	"github.com/josejulio/ghautodelete/pkg/interfaces"
	"gopkg.in/yaml.v3"
)

// TokenProvider retrieves GitHub API tokens from various sources.
//
// It uses dependency injection for all external dependencies (environment variables,
// home directory, and file system access) to enable comprehensive testing.
type TokenProvider struct {
	explicitToken string
	envGetter     func(string) string
	homeGetter    func() (string, error)
	fileReader    func(string) ([]byte, error)
}

// NewTokenProvider creates a new TokenProvider with dependency injection.
//
// Parameters:
//   - explicitToken: Token passed via CLI flag (highest priority)
//   - envGetter: Function to retrieve environment variables
//   - homeGetter: Function to retrieve the user's home directory
//   - fileReader: Function to read file contents
//
// Returns a TokenProvider that implements the ITokenProvider interface.
func NewTokenProvider(
	explicitToken string,
	envGetter func(string) string,
	homeGetter func() (string, error),
	fileReader func(string) ([]byte, error),
) *TokenProvider {
	return &TokenProvider{
		explicitToken: explicitToken,
		envGetter:     envGetter,
		homeGetter:    homeGetter,
		fileReader:    fileReader,
	}
}

// GetToken retrieves the GitHub API token using the precedence order.
//
// Precedence (highest to lowest):
// 1. Explicit token from CLI flag (if non-empty after trimming)
// 2. GITHUB_TOKEN environment variable (if non-empty after trimming)
// 3. gh CLI configuration file at ~/.config/gh/hosts.yml
//
// Returns:
//   - The token string (trimmed of whitespace)
//   - An error if no token source is available
//
// The error returned when no token is found is an *AppError with code
// ErrAuthenticationFailed and includes actionable guidance.
func (p *TokenProvider) GetToken() (string, error) {
	// Priority 1: Explicit token from CLI flag
	trimmedExplicit := strings.TrimSpace(p.explicitToken)
	if trimmedExplicit != "" {
		return trimmedExplicit, nil
	}

	// Priority 2: GITHUB_TOKEN environment variable
	envToken := p.envGetter("GITHUB_TOKEN")
	trimmedEnv := strings.TrimSpace(envToken)
	if trimmedEnv != "" {
		return trimmedEnv, nil
	}

	// Priority 3: gh CLI configuration file
	ghToken, err := p.getTokenFromGhConfig()
	if err == nil && ghToken != "" {
		return ghToken, nil
	}

	// No token found from any source
	return "", errors.NewAuthenticationError(
		"No GitHub token found. Set GITHUB_TOKEN environment variable or use --token flag",
		nil,
	)
}

// getTokenFromGhConfig retrieves the token from gh CLI configuration.
//
// It reads ~/.config/gh/hosts.yml and extracts the oauth_token for github.com.
//
// Returns:
//   - The token string (trimmed of whitespace)
//   - An error if the file cannot be read, parsed, or doesn't contain the token
func (p *TokenProvider) getTokenFromGhConfig() (string, error) {
	// Get home directory
	homeDir, err := p.homeGetter()
	if err != nil {
		return "", err
	}

	// Construct path to gh config file
	ghConfigPath := filepath.Join(homeDir, ".config", "gh", "hosts.yml")

	// Read the file
	content, err := p.fileReader(ghConfigPath)
	if err != nil {
		return "", err
	}

	// Parse YAML
	var hosts map[string]map[string]interface{}
	if err := yaml.Unmarshal(content, &hosts); err != nil {
		return "", err
	}

	// Extract token for github.com
	githubHost, ok := hosts["github.com"]
	if !ok {
		return "", nil // github.com not found, return empty token
	}

	oauthToken, ok := githubHost["oauth_token"]
	if !ok {
		return "", nil // oauth_token not found, return empty token
	}

	// Convert to string and trim
	tokenStr, ok := oauthToken.(string)
	if !ok {
		return "", nil // oauth_token is not a string, return empty token
	}

	return strings.TrimSpace(tokenStr), nil
}

// Ensure TokenProvider implements ITokenProvider interface
var _ interfaces.ITokenProvider = (*TokenProvider)(nil)
