// Package config provides services for configuring repository settings.
//
// This package contains the implementation of the ConfigService which orchestrates
// the process of checking and updating the delete-branch-on-merge setting.
package config

import (
	"context"

	"github.com/josejulio/ghautodelete/internal/github"
	"github.com/josejulio/ghautodelete/pkg/interfaces"
)

// ConfigService implements the IConfigService interface.
// It orchestrates repository configuration operations using a GitHub client.
type ConfigService struct {
	client interfaces.IGitHubClient
	writer interfaces.IOutputWriter
}

// NewConfigService creates a new ConfigService instance.
// Parameters:
//   - client: the GitHub client for API operations
//   - writer: the output writer for logging messages
func NewConfigService(client interfaces.IGitHubClient, writer interfaces.IOutputWriter) *ConfigService {
	return &ConfigService{
		client: client,
		writer: writer,
	}
}

// Configure applies the delete-branch-on-merge setting to a repository.
// It follows this workflow:
//  1. Fetch current repository state
//  2. If already enabled: return AlreadyEnabled=true, NowEnabled=true
//  3. If dryRun: return AlreadyEnabled=false, NowEnabled=false
//  4. Otherwise: update settings, verify, return AlreadyEnabled=false, NowEnabled=true
//
// Parameters:
//   - ctx: the context for cancellation and deadlines
//   - owner: the repository owner (user or organization)
//   - name: the repository name
//   - dryRun: if true, don't actually update the repository
//
// Returns:
//   - IConfigResult: the outcome of the configuration operation
//   - error: any error that occurred during the operation
func (s *ConfigService) Configure(ctx context.Context, owner, name string, dryRun bool) (interfaces.IConfigResult, error) {
	// Step 1: Fetch current repository state
	s.writer.Verbose("Fetching repository information")
	repo, err := s.client.GetRepository(ctx, owner, name)
	if err != nil {
		return nil, err
	}

	// Check if already enabled
	alreadyEnabled := repo.GetDeleteBranchOnMerge()

	// Step 2: If already enabled, return early
	if alreadyEnabled {
		return NewConfigResult(
			true,  // wasAlreadyEnabled
			true,  // isNowEnabled
			repo.GetDefaultBranch(),
			repo.GetFullName(),
		), nil
	}

	// Step 3: If dry run, return without updating
	if dryRun {
		return NewConfigResult(
			false, // wasAlreadyEnabled
			false, // isNowEnabled
			repo.GetDefaultBranch(),
			repo.GetFullName(),
		), nil
	}

	// Step 4: Update repository settings
	s.writer.Verbose("Updating repository settings")
	settings := github.NewRepositorySettings(true)
	err = s.client.UpdateRepository(ctx, owner, name, settings)
	if err != nil {
		return nil, err
	}

	// Step 5: Verify settings were applied
	s.writer.Verbose("Verifying settings applied")
	verifiedRepo, err := s.client.GetRepository(ctx, owner, name)
	if err != nil {
		return nil, err
	}

	// Step 6: Return result
	return NewConfigResult(
		false, // wasAlreadyEnabled
		verifiedRepo.GetDeleteBranchOnMerge(), // isNowEnabled
		verifiedRepo.GetDefaultBranch(),
		verifiedRepo.GetFullName(),
	), nil
}

// CheckStatus checks the current delete-branch-on-merge setting status.
// It fetches the repository state and returns the current configuration.
//
// Parameters:
//   - ctx: the context for cancellation and deadlines
//   - owner: the repository owner (user or organization)
//   - name: the repository name
//
// Returns:
//   - IConfigResult: the current configuration state
//   - error: any error that occurred during the operation
func (s *ConfigService) CheckStatus(ctx context.Context, owner, name string) (interfaces.IConfigResult, error) {
	// Fetch current repository state
	s.writer.Verbose("Fetching repository information")
	repo, err := s.client.GetRepository(ctx, owner, name)
	if err != nil {
		return nil, err
	}

	// Return current state without modification
	currentlyEnabled := repo.GetDeleteBranchOnMerge()
	return NewConfigResult(
		currentlyEnabled, // wasAlreadyEnabled (same as current state for read-only operation)
		currentlyEnabled, // isNowEnabled
		repo.GetDefaultBranch(),
		repo.GetFullName(),
	), nil
}
