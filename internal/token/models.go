// Package token provides data models and types for GitHub token information.
//
// This package contains the concrete implementations of token-related
// interfaces and data structures.
package token

// TokenInfo represents GitHub API token metadata.
// It implements the ITokenInfo interface.
type TokenInfo struct {
	// Scopes is the list of OAuth scopes granted to the token.
	Scopes []string

	// Username is the username associated with the token.
	Username string
}

// GetScopes returns the list of OAuth scopes granted to the token.
func (t *TokenInfo) GetScopes() []string {
	return t.Scopes
}

// HasScope checks if the token has a specific OAuth scope.
// The check is case-sensitive and requires an exact match.
// Returns true if the scope is present, false otherwise.
func (t *TokenInfo) HasScope(scope string) bool {
	if t.Scopes == nil {
		return false
	}

	for _, s := range t.Scopes {
		if s == scope {
			return true
		}
	}

	return false
}

// GetUsername returns the username associated with the token.
func (t *TokenInfo) GetUsername() string {
	return t.Username
}

// NewTokenInfo creates a new TokenInfo instance.
// Parameters:
//   - username: the username associated with the token
//   - scopes: the list of OAuth scopes granted to the token
func NewTokenInfo(username string, scopes []string) *TokenInfo {
	return &TokenInfo{
		Username: username,
		Scopes:   scopes,
	}
}
