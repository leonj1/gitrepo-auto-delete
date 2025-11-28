// Package app provides the main application logic for the GitHub Auto Delete CLI.
//
// This package orchestrates the CLI workflow by coordinating between the repository parser,
// configuration service, and output writer to handle the three operational modes:
// - Check mode: Shows current status without making changes
// - Dry-run mode: Shows what would happen without making changes
// - Normal mode: Actually enables auto-delete branches
package app

import (
	"context"
	"fmt"

	"github.com/josejulio/ghautodelete/pkg/interfaces"
)

// App orchestrates the CLI application logic.
// It coordinates between the repository parser, configuration service, and output writer
// to handle different operational modes (check, dry-run, normal).
type App struct {
	writer    interfaces.IOutputWriter
	configSvc interfaces.IConfigService
	parser    interfaces.IRepoParser
}

// NewApp creates a new App with the provided dependencies.
// All dependencies are injected via interfaces for testability.
func NewApp(writer interfaces.IOutputWriter, configSvc interfaces.IConfigService, parser interfaces.IRepoParser) *App {
	return &App{
		writer:    writer,
		configSvc: configSvc,
		parser:    parser,
	}
}

// Run executes the application logic based on the provided CLI options.
// It handles three modes:
// - Check mode (opts.CheckOnly): Shows current status without modification
// - Dry-run mode (opts.DryRun): Shows what would happen without modification
// - Normal mode: Actually enables auto-delete branches
//
// Returns an error if repository parsing fails or if the configuration service fails.
func (a *App) Run(ctx context.Context, opts interfaces.CLIOptions) error {
	// Parse repository identifier to extract owner and name
	owner, name, err := a.parser.Parse(opts.Repository)
	if err != nil {
		return fmt.Errorf("failed to parse repository: %w", err)
	}

	// Check mode takes precedence over dry-run mode
	if opts.CheckOnly {
		return a.handleCheckMode(ctx, owner, name)
	}

	// Dry-run mode
	if opts.DryRun {
		return a.handleDryRunMode(ctx, owner, name)
	}

	// Normal mode - actually enable the feature
	return a.handleNormalMode(ctx, owner, name)
}

// handleCheckMode handles check-only mode.
// It retrieves the current status and outputs it without making any changes.
func (a *App) handleCheckMode(ctx context.Context, owner, name string) error {
	result, err := a.configSvc.CheckStatus(ctx, owner, name)
	if err != nil {
		return fmt.Errorf("check status failed: %w", err)
	}

	// Output repository information
	a.writer.Info(fmt.Sprintf("Repository: %s", result.GetRepositoryFullName()))
	a.writer.Info(fmt.Sprintf("Default branch: %s", result.GetDefaultBranch()))

	// Output current status
	if result.IsNowEnabled() {
		a.writer.Info("Auto-delete branches: enabled")
	} else {
		a.writer.Info("Auto-delete branches: disabled")
		a.writer.Info("To enable, run without --check flag")
	}

	return nil
}

// handleDryRunMode handles dry-run mode.
// It shows what would happen without making any actual changes.
func (a *App) handleDryRunMode(ctx context.Context, owner, name string) error {
	result, err := a.configSvc.Configure(ctx, owner, name, true)
	if err != nil {
		return fmt.Errorf("dry-run configure failed: %w", err)
	}

	// If feature was already enabled
	if result.WasAlreadyEnabled() {
		a.writer.Info(fmt.Sprintf("Auto-delete branches already enabled for %s", result.GetRepositoryFullName()))
		a.writer.Info("No changes needed")
		return nil
	}

	// If feature would be enabled
	a.writer.Info(fmt.Sprintf("[DRY-RUN] Would enable auto-delete branches for %s", result.GetRepositoryFullName()))
	a.writer.Info("No changes made")

	return nil
}

// handleNormalMode handles normal mode.
// It actually enables auto-delete branches on the repository.
func (a *App) handleNormalMode(ctx context.Context, owner, name string) error {
	result, err := a.configSvc.Configure(ctx, owner, name, false)
	if err != nil {
		return fmt.Errorf("configure failed: %w", err)
	}

	// If feature was already enabled
	if result.WasAlreadyEnabled() {
		a.writer.Success(fmt.Sprintf("Auto-delete branches already enabled for %s", result.GetRepositoryFullName()))
		return nil
	}

	// If feature was successfully enabled
	if result.IsNowEnabled() {
		a.writer.Success(fmt.Sprintf("Successfully enabled auto-delete branches for %s", result.GetRepositoryFullName()))
		return nil
	}

	// This shouldn't happen, but handle it just in case
	return fmt.Errorf("unexpected state: feature was not enabled")
}
