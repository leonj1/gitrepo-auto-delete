// Package config provides data models and types for configuration operations.
//
// This package contains the concrete implementations of configuration-related
// interfaces and data structures.
package config

// ConfigResult represents the outcome of a configuration operation.
// It implements the IConfigResult interface.
type ConfigResult struct {
	// AlreadyEnabled indicates whether delete-branch-on-merge was already enabled before the operation.
	AlreadyEnabled bool

	// NowEnabled indicates whether delete-branch-on-merge is currently enabled after the operation.
	NowEnabled bool

	// DefaultBranch is the repository's default branch name.
	DefaultBranch string

	// RepositoryFullName is the full repository name in "owner/name" format.
	RepositoryFullName string
}

// WasAlreadyEnabled returns true if delete-branch-on-merge was already enabled.
func (c *ConfigResult) WasAlreadyEnabled() bool {
	return c.AlreadyEnabled
}

// IsNowEnabled returns true if delete-branch-on-merge is currently enabled.
func (c *ConfigResult) IsNowEnabled() bool {
	return c.NowEnabled
}

// GetDefaultBranch returns the repository's default branch name.
func (c *ConfigResult) GetDefaultBranch() string {
	return c.DefaultBranch
}

// GetRepositoryFullName returns the full repository name in "owner/name" format.
func (c *ConfigResult) GetRepositoryFullName() string {
	return c.RepositoryFullName
}

// NewConfigResult creates a new ConfigResult instance.
// Parameters:
//   - wasAlreadyEnabled: whether delete-branch-on-merge was already enabled
//   - isNowEnabled: whether delete-branch-on-merge is currently enabled
//   - defaultBranch: the repository's default branch name
//   - fullName: the full repository name in "owner/name" format
func NewConfigResult(wasAlreadyEnabled, isNowEnabled bool, defaultBranch, fullName string) *ConfigResult {
	return &ConfigResult{
		AlreadyEnabled:     wasAlreadyEnabled,
		NowEnabled:         isNowEnabled,
		DefaultBranch:      defaultBranch,
		RepositoryFullName: fullName,
	}
}
