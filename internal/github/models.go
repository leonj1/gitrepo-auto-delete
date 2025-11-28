// Package github provides data models and types for GitHub API interactions.
//
// This package contains the concrete implementations of repository-related
// interfaces and data structures.
package github

import "fmt"

// Repository represents a GitHub repository with its key properties.
// It implements the IRepository interface.
type Repository struct {
	// Owner is the repository owner (user or organization).
	Owner struct {
		Login string `json:"login"`
	} `json:"owner"`

	// Name is the repository name.
	Name string `json:"name"`

	// DefaultBranch is the default branch name (e.g., "main", "master").
	DefaultBranch string `json:"default_branch"`

	// DeleteBranchOnMerge indicates whether automatic branch deletion is enabled.
	DeleteBranchOnMerge bool `json:"delete_branch_on_merge"`
}

// GetOwner returns the repository owner.
func (r *Repository) GetOwner() string {
	return r.Owner.Login
}

// GetName returns the repository name.
func (r *Repository) GetName() string {
	return r.Name
}

// GetDefaultBranch returns the default branch name.
func (r *Repository) GetDefaultBranch() string {
	return r.DefaultBranch
}

// GetDeleteBranchOnMerge returns whether delete-branch-on-merge is enabled.
func (r *Repository) GetDeleteBranchOnMerge() bool {
	return r.DeleteBranchOnMerge
}

// GetFullName returns the full repository name in "owner/name" format.
func (r *Repository) GetFullName() string {
	return fmt.Sprintf("%s/%s", r.Owner.Login, r.Name)
}

// NewRepository creates a new Repository instance.
// Parameters:
//   - owner: the repository owner (user or organization)
//   - name: the repository name
//   - defaultBranch: the default branch name
//   - deleteBranchOnMerge: whether delete-branch-on-merge is enabled
func NewRepository(owner, name, defaultBranch string, deleteBranchOnMerge bool) *Repository {
	repo := &Repository{
		Name:                name,
		DefaultBranch:       defaultBranch,
		DeleteBranchOnMerge: deleteBranchOnMerge,
	}
	repo.Owner.Login = owner
	return repo
}

// RepositorySettings represents the settings that can be applied to a repository.
// It implements the IRepositorySettings interface.
type RepositorySettings struct {
	// DeleteBranchOnMerge indicates whether automatic branch deletion should be enabled.
	DeleteBranchOnMerge bool
}

// GetDeleteBranchOnMerge returns whether delete-branch-on-merge should be enabled.
func (s *RepositorySettings) GetDeleteBranchOnMerge() bool {
	return s.DeleteBranchOnMerge
}

// NewRepositorySettings creates a new RepositorySettings instance.
// Parameters:
//   - deleteBranchOnMerge: whether delete-branch-on-merge should be enabled
func NewRepositorySettings(deleteBranchOnMerge bool) *RepositorySettings {
	return &RepositorySettings{
		DeleteBranchOnMerge: deleteBranchOnMerge,
	}
}
