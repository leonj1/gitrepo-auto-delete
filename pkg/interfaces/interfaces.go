// Package interfaces defines the core interfaces and types for the GitHub Auto Delete application.
//
// This package provides all the interface contracts that allow for dependency injection
// and testability throughout the application.
package interfaces

import "context"

// IGitHubClient provides methods for interacting with the GitHub API.
// It abstracts repository operations and token validation.
type IGitHubClient interface {
	// GetRepository retrieves repository information from GitHub.
	// Returns an IRepository containing the repository details.
	GetRepository(ctx context.Context, owner, name string) (IRepository, error)

	// UpdateRepository updates repository settings on GitHub.
	// Takes repository settings as IRepositorySettings and applies them.
	UpdateRepository(ctx context.Context, owner, name string, settings IRepositorySettings) error

	// ValidateToken validates the GitHub API token and returns token information.
	// Returns ITokenInfo containing scopes and user details.
	ValidateToken(ctx context.Context) (ITokenInfo, error)
}

// IRepoParser provides methods for parsing repository identifiers.
// It handles various repository identifier formats (e.g., "owner/repo").
type IRepoParser interface {
	// Parse extracts owner and repository name from a repository identifier.
	// Returns owner, name, and error if parsing fails.
	Parse(repoIdentifier string) (owner string, name string, err error)
}

// ITokenProvider provides methods for retrieving GitHub API tokens.
// It abstracts token retrieval from various sources (environment, config, etc.).
type ITokenProvider interface {
	// GetToken retrieves the GitHub API token.
	// Returns the token string or an error if unavailable.
	GetToken() (string, error)
}

// IOutputWriter provides methods for writing output messages.
// It abstracts output operations for different verbosity levels.
type IOutputWriter interface {
	// Success writes a success message.
	Success(message string)

	// Error writes an error message.
	Error(message string)

	// Info writes an informational message.
	Info(message string)

	// Verbose writes a verbose/debug message.
	Verbose(message string)
}

// IConfigService provides methods for configuring repository settings.
// It orchestrates the process of checking and updating repository configuration.
type IConfigService interface {
	// Configure applies the delete-branch-on-merge setting to a repository.
	// Returns IConfigResult containing the configuration outcome.
	Configure(ctx context.Context, owner, name string, dryRun bool) (IConfigResult, error)

	// CheckStatus checks the current delete-branch-on-merge setting status.
	// Returns IConfigResult containing the current configuration state.
	CheckStatus(ctx context.Context, owner, name string) (IConfigResult, error)
}

// IRepository provides methods for accessing repository information.
// It represents a GitHub repository with its key properties.
type IRepository interface {
	// GetOwner returns the repository owner (user or organization).
	GetOwner() string

	// GetName returns the repository name.
	GetName() string

	// GetDefaultBranch returns the default branch name (e.g., "main", "master").
	GetDefaultBranch() string

	// GetDeleteBranchOnMerge returns whether delete-branch-on-merge is enabled.
	GetDeleteBranchOnMerge() bool

	// GetFullName returns the full repository name in "owner/name" format.
	GetFullName() string
}

// IRepositorySettings provides methods for accessing repository settings.
// It represents the settings that can be applied to a repository.
type IRepositorySettings interface {
	// GetDeleteBranchOnMerge returns whether delete-branch-on-merge should be enabled.
	GetDeleteBranchOnMerge() bool
}

// ITokenInfo provides methods for accessing GitHub token information.
// It represents the metadata associated with a GitHub API token.
type ITokenInfo interface {
	// GetScopes returns the list of OAuth scopes granted to the token.
	GetScopes() []string

	// HasScope checks if the token has a specific OAuth scope.
	HasScope(scope string) bool

	// GetUsername returns the username associated with the token.
	GetUsername() string
}

// IConfigResult provides methods for accessing configuration operation results.
// It represents the outcome of a configuration check or update operation.
type IConfigResult interface {
	// WasAlreadyEnabled returns true if delete-branch-on-merge was already enabled.
	WasAlreadyEnabled() bool

	// IsNowEnabled returns true if delete-branch-on-merge is currently enabled.
	IsNowEnabled() bool

	// GetDefaultBranch returns the repository's default branch name.
	GetDefaultBranch() string

	// GetRepositoryFullName returns the full repository name in "owner/name" format.
	GetRepositoryFullName() string
}

// CLIOptions represents the command-line options for the application.
// It contains all user-configurable parameters passed via CLI flags.
type CLIOptions struct {
	// Repository is the repository identifier in "owner/name" format.
	Repository string

	// Token is the GitHub API token for authentication.
	Token string

	// Verbose enables verbose/debug output.
	Verbose bool

	// DryRun enables dry-run mode (no actual changes made).
	DryRun bool

	// CheckOnly enables check-only mode (only check status, don't update).
	CheckOnly bool
}
