// Package app_test provides tests for the App struct that orchestrates CLI modes.
//
// These tests verify that the App correctly implements the CLI workflow:
// - Check mode (--check / -c): Shows current status without modification
// - Dry-run mode (--dry-run / -d): Shows what would happen without modification
// - Normal mode: Actually enables auto-delete branches
//
// Gherkin Scenarios:
//   - Scenario: Check status when auto-delete is disabled
//     - Output: "Repository: octocat/hello-world", "Auto-delete branches: disabled", "To enable, run without --check flag"
//   - Scenario: Check status when auto-delete is enabled
//     - Output: "Repository: octocat/hello-world", "Auto-delete branches: enabled"
//   - Scenario: Dry-run when auto-delete is disabled
//     - Output: "[DRY-RUN] Would enable auto-delete branches", "No changes made"
//   - Scenario: Dry-run when auto-delete is already enabled
//     - Output: "Auto-delete branches already enabled", "No changes needed"
//   - Scenario: Use short form -c for check mode
//   - Scenario: Use short form -d for dry-run mode
//
// These tests are designed to FAIL until the implementation is properly created
// by the coder agent.
package app_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/josejulio/ghautodelete/internal/app"
	"github.com/josejulio/ghautodelete/pkg/interfaces"
)

// =============================================================================
// Mock Implementations for Testing
// =============================================================================

// mockGitHubClient implements IGitHubClient for testing.
type mockGitHubClient struct {
	// GetRepositoryFunc is called when GetRepository is invoked.
	GetRepositoryFunc func(ctx context.Context, owner, name string) (interfaces.IRepository, error)
	// GetRepositoryCalls tracks all calls to GetRepository.
	GetRepositoryCalls []struct {
		Ctx   context.Context
		Owner string
		Name  string
	}

	// UpdateRepositoryFunc is called when UpdateRepository is invoked.
	UpdateRepositoryFunc func(ctx context.Context, owner, name string, settings interfaces.IRepositorySettings) error
	// UpdateRepositoryCalls tracks all calls to UpdateRepository.
	UpdateRepositoryCalls []struct {
		Ctx      context.Context
		Owner    string
		Name     string
		Settings interfaces.IRepositorySettings
	}

	// ValidateTokenFunc is called when ValidateToken is invoked.
	ValidateTokenFunc func(ctx context.Context) (interfaces.ITokenInfo, error)
}

func (m *mockGitHubClient) GetRepository(ctx context.Context, owner, name string) (interfaces.IRepository, error) {
	m.GetRepositoryCalls = append(m.GetRepositoryCalls, struct {
		Ctx   context.Context
		Owner string
		Name  string
	}{ctx, owner, name})
	if m.GetRepositoryFunc != nil {
		return m.GetRepositoryFunc(ctx, owner, name)
	}
	return nil, errors.New("GetRepositoryFunc not set")
}

func (m *mockGitHubClient) UpdateRepository(ctx context.Context, owner, name string, settings interfaces.IRepositorySettings) error {
	m.UpdateRepositoryCalls = append(m.UpdateRepositoryCalls, struct {
		Ctx      context.Context
		Owner    string
		Name     string
		Settings interfaces.IRepositorySettings
	}{ctx, owner, name, settings})
	if m.UpdateRepositoryFunc != nil {
		return m.UpdateRepositoryFunc(ctx, owner, name, settings)
	}
	return errors.New("UpdateRepositoryFunc not set")
}

func (m *mockGitHubClient) ValidateToken(ctx context.Context) (interfaces.ITokenInfo, error) {
	if m.ValidateTokenFunc != nil {
		return m.ValidateTokenFunc(ctx)
	}
	return nil, errors.New("ValidateTokenFunc not set")
}

// mockOutputWriter implements IOutputWriter for testing with string capture.
type mockOutputWriter struct {
	// Messages captures all messages by type.
	SuccessCalls []string
	ErrorCalls   []string
	InfoCalls    []string
	VerboseCalls []string
	// AllMessages captures all messages in order for output verification.
	AllMessages []string
}

func (m *mockOutputWriter) Success(message string) {
	m.SuccessCalls = append(m.SuccessCalls, message)
	m.AllMessages = append(m.AllMessages, message)
}

func (m *mockOutputWriter) Error(message string) {
	m.ErrorCalls = append(m.ErrorCalls, message)
	m.AllMessages = append(m.AllMessages, message)
}

func (m *mockOutputWriter) Info(message string) {
	m.InfoCalls = append(m.InfoCalls, message)
	m.AllMessages = append(m.AllMessages, message)
}

func (m *mockOutputWriter) Verbose(message string) {
	m.VerboseCalls = append(m.VerboseCalls, message)
	m.AllMessages = append(m.AllMessages, message)
}

// GetAllOutput returns all messages concatenated for easier assertion.
func (m *mockOutputWriter) GetAllOutput() string {
	return strings.Join(m.AllMessages, "\n")
}

// mockConfigService implements IConfigService for testing.
type mockConfigService struct {
	// ConfigureFunc is called when Configure is invoked.
	ConfigureFunc func(ctx context.Context, owner, name string, dryRun bool) (interfaces.IConfigResult, error)
	// ConfigureCalls tracks all calls to Configure.
	ConfigureCalls []struct {
		Ctx    context.Context
		Owner  string
		Name   string
		DryRun bool
	}

	// CheckStatusFunc is called when CheckStatus is invoked.
	CheckStatusFunc func(ctx context.Context, owner, name string) (interfaces.IConfigResult, error)
	// CheckStatusCalls tracks all calls to CheckStatus.
	CheckStatusCalls []struct {
		Ctx   context.Context
		Owner string
		Name  string
	}
}

func (m *mockConfigService) Configure(ctx context.Context, owner, name string, dryRun bool) (interfaces.IConfigResult, error) {
	m.ConfigureCalls = append(m.ConfigureCalls, struct {
		Ctx    context.Context
		Owner  string
		Name   string
		DryRun bool
	}{ctx, owner, name, dryRun})
	if m.ConfigureFunc != nil {
		return m.ConfigureFunc(ctx, owner, name, dryRun)
	}
	return nil, errors.New("ConfigureFunc not set")
}

func (m *mockConfigService) CheckStatus(ctx context.Context, owner, name string) (interfaces.IConfigResult, error) {
	m.CheckStatusCalls = append(m.CheckStatusCalls, struct {
		Ctx   context.Context
		Owner string
		Name  string
	}{ctx, owner, name})
	if m.CheckStatusFunc != nil {
		return m.CheckStatusFunc(ctx, owner, name)
	}
	return nil, errors.New("CheckStatusFunc not set")
}

// mockConfigResult implements IConfigResult for testing.
type mockConfigResult struct {
	wasAlreadyEnabled  bool
	isNowEnabled       bool
	defaultBranch      string
	repositoryFullName string
}

func (m *mockConfigResult) WasAlreadyEnabled() bool {
	return m.wasAlreadyEnabled
}

func (m *mockConfigResult) IsNowEnabled() bool {
	return m.isNowEnabled
}

func (m *mockConfigResult) GetDefaultBranch() string {
	return m.defaultBranch
}

func (m *mockConfigResult) GetRepositoryFullName() string {
	return m.repositoryFullName
}

// mockRepoParser implements IRepoParser for testing.
type mockRepoParser struct {
	// ParseFunc is called when Parse is invoked.
	ParseFunc func(repoIdentifier string) (owner string, name string, err error)
	// ParseCalls tracks all calls to Parse.
	ParseCalls []string
}

func (m *mockRepoParser) Parse(repoIdentifier string) (owner string, name string, err error) {
	m.ParseCalls = append(m.ParseCalls, repoIdentifier)
	if m.ParseFunc != nil {
		return m.ParseFunc(repoIdentifier)
	}
	return "", "", errors.New("ParseFunc not set")
}

// =============================================================================
// Test Helpers
// =============================================================================

// newMockConfigResult creates a mock IConfigResult for testing.
func newMockConfigResult(wasAlready, isNow bool, defaultBranch, fullName string) *mockConfigResult {
	return &mockConfigResult{
		wasAlreadyEnabled:  wasAlready,
		isNowEnabled:       isNow,
		defaultBranch:      defaultBranch,
		repositoryFullName: fullName,
	}
}

// =============================================================================
// Interface Satisfaction Tests
// =============================================================================

// TestNewAppReturnsNonNil verifies the constructor returns a non-nil App.
//
// The implementation should:
// - Define App struct in internal/app/app.go
// - Have NewApp constructor that accepts dependencies via interfaces
func TestNewAppReturnsNonNil(t *testing.T) {
	// Arrange
	mockWriter := &mockOutputWriter{}
	mockConfigSvc := &mockConfigService{}
	mockParser := &mockRepoParser{}

	// Act
	application := app.NewApp(mockWriter, mockConfigSvc, mockParser)

	// Assert
	if application == nil {
		t.Error("NewApp() should return non-nil App")
	}
}

// TestNewAppAcceptsDependenciesAsInterfaces verifies dependency injection.
//
// The implementation should:
// - Accept IOutputWriter as first parameter
// - Accept IConfigService as second parameter
// - Accept IRepoParser as third parameter
func TestNewAppAcceptsDependenciesAsInterfaces(t *testing.T) {
	// Arrange - use interface types explicitly
	var writer interfaces.IOutputWriter = &mockOutputWriter{}
	var configSvc interfaces.IConfigService = &mockConfigService{}
	var parser interfaces.IRepoParser = &mockRepoParser{}

	// Act
	application := app.NewApp(writer, configSvc, parser)

	// Assert
	if application == nil {
		t.Error("NewApp should accept interface parameters")
	}
}

// =============================================================================
// Check Mode Tests - Disabled State
// =============================================================================

// TestRunCheckModeWhenDisabledOutputsCorrectMessages verifies check mode output when disabled.
//
// Gherkin: Scenario: Check status when auto-delete is disabled
// - Output: "Repository: octocat/hello-world", "Auto-delete branches: disabled", "To enable, run without --check flag"
//
// The implementation should:
// - Call ConfigService.CheckStatus (not Configure)
// - Output repository name
// - Output disabled status
// - Output help message to enable
func TestRunCheckModeWhenDisabledOutputsCorrectMessages(t *testing.T) {
	// Arrange
	mockWriter := &mockOutputWriter{}
	mockConfigSvc := &mockConfigService{
		CheckStatusFunc: func(ctx context.Context, owner, name string) (interfaces.IConfigResult, error) {
			return newMockConfigResult(false, false, "main", "octocat/hello-world"), nil
		},
	}
	mockParser := &mockRepoParser{
		ParseFunc: func(repoIdentifier string) (string, string, error) {
			return "octocat", "hello-world", nil
		},
	}

	application := app.NewApp(mockWriter, mockConfigSvc, mockParser)
	ctx := context.Background()
	opts := interfaces.CLIOptions{
		Repository: "octocat/hello-world",
		CheckOnly:  true,
		DryRun:     false,
	}

	// Act
	err := application.Run(ctx, opts)

	// Assert - no error
	if err != nil {
		t.Fatalf("Run() error = %v, expected nil", err)
	}

	// Assert - output contains required messages
	output := mockWriter.GetAllOutput()

	if !strings.Contains(output, "Repository: octocat/hello-world") {
		t.Errorf("Output should contain 'Repository: octocat/hello-world', got: %s", output)
	}
	if !strings.Contains(output, "Auto-delete branches: disabled") {
		t.Errorf("Output should contain 'Auto-delete branches: disabled', got: %s", output)
	}
	if !strings.Contains(output, "To enable, run without --check flag") {
		t.Errorf("Output should contain 'To enable, run without --check flag', got: %s", output)
	}
}

// TestRunCheckModeWhenDisabledUsesCheckStatus verifies CheckStatus is called.
//
// The implementation should:
// - Call CheckStatus (not Configure) when --check flag is set
func TestRunCheckModeWhenDisabledUsesCheckStatus(t *testing.T) {
	// Arrange
	mockWriter := &mockOutputWriter{}
	mockConfigSvc := &mockConfigService{
		CheckStatusFunc: func(ctx context.Context, owner, name string) (interfaces.IConfigResult, error) {
			return newMockConfigResult(false, false, "main", "octocat/hello-world"), nil
		},
		ConfigureFunc: func(ctx context.Context, owner, name string, dryRun bool) (interfaces.IConfigResult, error) {
			t.Error("Configure should NOT be called in check mode")
			return nil, nil
		},
	}
	mockParser := &mockRepoParser{
		ParseFunc: func(repoIdentifier string) (string, string, error) {
			return "octocat", "hello-world", nil
		},
	}

	application := app.NewApp(mockWriter, mockConfigSvc, mockParser)
	ctx := context.Background()
	opts := interfaces.CLIOptions{
		Repository: "octocat/hello-world",
		CheckOnly:  true,
	}

	// Act
	err := application.Run(ctx, opts)

	// Assert
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	// Assert - CheckStatus was called
	if len(mockConfigSvc.CheckStatusCalls) != 1 {
		t.Errorf("CheckStatus called %d times, expected 1", len(mockConfigSvc.CheckStatusCalls))
	}

	// Assert - Configure was NOT called
	if len(mockConfigSvc.ConfigureCalls) != 0 {
		t.Errorf("Configure called %d times, expected 0 in check mode", len(mockConfigSvc.ConfigureCalls))
	}
}

// TestRunCheckModeNeverModifiesRepository verifies no UpdateRepository calls.
//
// The implementation should:
// - Never call any method that modifies the repository in check mode
func TestRunCheckModeNeverModifiesRepository(t *testing.T) {
	// Arrange
	mockWriter := &mockOutputWriter{}
	mockClient := &mockGitHubClient{} // Track UpdateRepository calls
	mockConfigSvc := &mockConfigService{
		CheckStatusFunc: func(ctx context.Context, owner, name string) (interfaces.IConfigResult, error) {
			return newMockConfigResult(false, false, "main", "octocat/hello-world"), nil
		},
	}
	mockParser := &mockRepoParser{
		ParseFunc: func(repoIdentifier string) (string, string, error) {
			return "octocat", "hello-world", nil
		},
	}

	application := app.NewApp(mockWriter, mockConfigSvc, mockParser)
	ctx := context.Background()
	opts := interfaces.CLIOptions{
		Repository: "octocat/hello-world",
		CheckOnly:  true,
	}

	// Act
	_ = application.Run(ctx, opts)

	// Assert - no UpdateRepository calls should have been made
	// Since we're using mockConfigSvc, no direct client calls happen
	// But Configure should not be called either
	if len(mockConfigSvc.ConfigureCalls) != 0 {
		t.Errorf("Configure should not be called in check mode, but was called %d times", len(mockConfigSvc.ConfigureCalls))
	}
	// Ensure client wasn't used directly
	if len(mockClient.UpdateRepositoryCalls) != 0 {
		t.Errorf("UpdateRepository should not be called in check mode")
	}
}

// =============================================================================
// Check Mode Tests - Enabled State
// =============================================================================

// TestRunCheckModeWhenEnabledOutputsCorrectMessages verifies check mode output when enabled.
//
// Gherkin: Scenario: Check status when auto-delete is enabled
// - Output: "Repository: octocat/hello-world", "Auto-delete branches: enabled"
//
// The implementation should:
// - Output repository name
// - Output enabled status
// - NOT output "To enable" message (since already enabled)
func TestRunCheckModeWhenEnabledOutputsCorrectMessages(t *testing.T) {
	// Arrange
	mockWriter := &mockOutputWriter{}
	mockConfigSvc := &mockConfigService{
		CheckStatusFunc: func(ctx context.Context, owner, name string) (interfaces.IConfigResult, error) {
			return newMockConfigResult(true, true, "main", "octocat/hello-world"), nil
		},
	}
	mockParser := &mockRepoParser{
		ParseFunc: func(repoIdentifier string) (string, string, error) {
			return "octocat", "hello-world", nil
		},
	}

	application := app.NewApp(mockWriter, mockConfigSvc, mockParser)
	ctx := context.Background()
	opts := interfaces.CLIOptions{
		Repository: "octocat/hello-world",
		CheckOnly:  true,
	}

	// Act
	err := application.Run(ctx, opts)

	// Assert - no error
	if err != nil {
		t.Fatalf("Run() error = %v, expected nil", err)
	}

	// Assert - output contains required messages
	output := mockWriter.GetAllOutput()

	if !strings.Contains(output, "Repository: octocat/hello-world") {
		t.Errorf("Output should contain 'Repository: octocat/hello-world', got: %s", output)
	}
	if !strings.Contains(output, "Auto-delete branches: enabled") {
		t.Errorf("Output should contain 'Auto-delete branches: enabled', got: %s", output)
	}

	// Assert - should NOT contain "To enable" message
	if strings.Contains(output, "To enable") {
		t.Errorf("Output should NOT contain 'To enable' when already enabled, got: %s", output)
	}
}

// =============================================================================
// Dry-Run Mode Tests - Disabled State
// =============================================================================

// TestRunDryRunModeWhenDisabledOutputsCorrectMessages verifies dry-run output when disabled.
//
// Gherkin: Scenario: Dry-run when auto-delete is disabled
// - Output: "[DRY-RUN] Would enable auto-delete branches", "No changes made"
//
// The implementation should:
// - Call ConfigService.Configure with dryRun=true
// - Output dry-run prefix message
// - Output "No changes made" message
func TestRunDryRunModeWhenDisabledOutputsCorrectMessages(t *testing.T) {
	// Arrange
	mockWriter := &mockOutputWriter{}
	mockConfigSvc := &mockConfigService{
		ConfigureFunc: func(ctx context.Context, owner, name string, dryRun bool) (interfaces.IConfigResult, error) {
			if !dryRun {
				t.Error("Configure should be called with dryRun=true")
			}
			return newMockConfigResult(false, false, "main", "octocat/hello-world"), nil
		},
	}
	mockParser := &mockRepoParser{
		ParseFunc: func(repoIdentifier string) (string, string, error) {
			return "octocat", "hello-world", nil
		},
	}

	application := app.NewApp(mockWriter, mockConfigSvc, mockParser)
	ctx := context.Background()
	opts := interfaces.CLIOptions{
		Repository: "octocat/hello-world",
		DryRun:     true,
		CheckOnly:  false,
	}

	// Act
	err := application.Run(ctx, opts)

	// Assert - no error
	if err != nil {
		t.Fatalf("Run() error = %v, expected nil", err)
	}

	// Assert - output contains required messages
	output := mockWriter.GetAllOutput()

	if !strings.Contains(output, "[DRY-RUN] Would enable auto-delete branches") {
		t.Errorf("Output should contain '[DRY-RUN] Would enable auto-delete branches', got: %s", output)
	}
	if !strings.Contains(output, "No changes made") {
		t.Errorf("Output should contain 'No changes made', got: %s", output)
	}
}

// TestRunDryRunModeCallsConfigureWithDryRunTrue verifies dryRun parameter is passed.
//
// The implementation should:
// - Call Configure with dryRun=true when --dry-run flag is set
func TestRunDryRunModeCallsConfigureWithDryRunTrue(t *testing.T) {
	// Arrange
	mockWriter := &mockOutputWriter{}
	mockConfigSvc := &mockConfigService{
		ConfigureFunc: func(ctx context.Context, owner, name string, dryRun bool) (interfaces.IConfigResult, error) {
			return newMockConfigResult(false, false, "main", "octocat/hello-world"), nil
		},
	}
	mockParser := &mockRepoParser{
		ParseFunc: func(repoIdentifier string) (string, string, error) {
			return "octocat", "hello-world", nil
		},
	}

	application := app.NewApp(mockWriter, mockConfigSvc, mockParser)
	ctx := context.Background()
	opts := interfaces.CLIOptions{
		Repository: "octocat/hello-world",
		DryRun:     true,
	}

	// Act
	err := application.Run(ctx, opts)

	// Assert
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	// Assert - Configure was called with dryRun=true
	if len(mockConfigSvc.ConfigureCalls) != 1 {
		t.Fatalf("Configure called %d times, expected 1", len(mockConfigSvc.ConfigureCalls))
	}
	if !mockConfigSvc.ConfigureCalls[0].DryRun {
		t.Error("Configure should be called with dryRun=true")
	}
}

// TestRunDryRunModeNeverModifiesRepository verifies no actual changes in dry-run.
//
// The implementation should:
// - Pass dryRun=true to ConfigService.Configure
// - ConfigService ensures no UpdateRepository is called
func TestRunDryRunModeNeverModifiesRepository(t *testing.T) {
	// Arrange
	mockWriter := &mockOutputWriter{}
	mockConfigSvc := &mockConfigService{
		ConfigureFunc: func(ctx context.Context, owner, name string, dryRun bool) (interfaces.IConfigResult, error) {
			if !dryRun {
				t.Error("Configure called with dryRun=false, expected true")
			}
			return newMockConfigResult(false, false, "main", "octocat/hello-world"), nil
		},
	}
	mockParser := &mockRepoParser{
		ParseFunc: func(repoIdentifier string) (string, string, error) {
			return "octocat", "hello-world", nil
		},
	}

	application := app.NewApp(mockWriter, mockConfigSvc, mockParser)
	ctx := context.Background()
	opts := interfaces.CLIOptions{
		Repository: "octocat/hello-world",
		DryRun:     true,
	}

	// Act
	err := application.Run(ctx, opts)

	// Assert
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	// Assert - Configure was called with dryRun=true
	if len(mockConfigSvc.ConfigureCalls) != 1 {
		t.Fatalf("Configure called %d times, expected 1", len(mockConfigSvc.ConfigureCalls))
	}
	if !mockConfigSvc.ConfigureCalls[0].DryRun {
		t.Error("DryRun should be true to prevent repository modification")
	}
}

// =============================================================================
// Dry-Run Mode Tests - Already Enabled State
// =============================================================================

// TestRunDryRunModeWhenEnabledOutputsCorrectMessages verifies dry-run output when enabled.
//
// Gherkin: Scenario: Dry-run when auto-delete is already enabled
// - Output: "Auto-delete branches already enabled", "No changes needed"
//
// The implementation should:
// - Output already enabled message
// - Output no changes needed message
func TestRunDryRunModeWhenEnabledOutputsCorrectMessages(t *testing.T) {
	// Arrange
	mockWriter := &mockOutputWriter{}
	mockConfigSvc := &mockConfigService{
		ConfigureFunc: func(ctx context.Context, owner, name string, dryRun bool) (interfaces.IConfigResult, error) {
			return newMockConfigResult(true, true, "main", "octocat/hello-world"), nil
		},
	}
	mockParser := &mockRepoParser{
		ParseFunc: func(repoIdentifier string) (string, string, error) {
			return "octocat", "hello-world", nil
		},
	}

	application := app.NewApp(mockWriter, mockConfigSvc, mockParser)
	ctx := context.Background()
	opts := interfaces.CLIOptions{
		Repository: "octocat/hello-world",
		DryRun:     true,
	}

	// Act
	err := application.Run(ctx, opts)

	// Assert - no error
	if err != nil {
		t.Fatalf("Run() error = %v, expected nil", err)
	}

	// Assert - output contains required messages
	output := mockWriter.GetAllOutput()

	if !strings.Contains(output, "Auto-delete branches already enabled") {
		t.Errorf("Output should contain 'Auto-delete branches already enabled', got: %s", output)
	}
	if !strings.Contains(output, "No changes needed") {
		t.Errorf("Output should contain 'No changes needed', got: %s", output)
	}
}

// =============================================================================
// Normal Mode Tests - Actually Enable
// =============================================================================

// TestRunNormalModeActuallyEnables verifies normal mode enables the feature.
//
// The implementation should:
// - Call Configure with dryRun=false
// - Actually enable the feature (via ConfigService)
func TestRunNormalModeActuallyEnables(t *testing.T) {
	// Arrange
	mockWriter := &mockOutputWriter{}
	mockConfigSvc := &mockConfigService{
		ConfigureFunc: func(ctx context.Context, owner, name string, dryRun bool) (interfaces.IConfigResult, error) {
			if dryRun {
				t.Error("Configure called with dryRun=true in normal mode, expected false")
			}
			return newMockConfigResult(false, true, "main", "octocat/hello-world"), nil
		},
	}
	mockParser := &mockRepoParser{
		ParseFunc: func(repoIdentifier string) (string, string, error) {
			return "octocat", "hello-world", nil
		},
	}

	application := app.NewApp(mockWriter, mockConfigSvc, mockParser)
	ctx := context.Background()
	opts := interfaces.CLIOptions{
		Repository: "octocat/hello-world",
		DryRun:     false,
		CheckOnly:  false,
	}

	// Act
	err := application.Run(ctx, opts)

	// Assert
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	// Assert - Configure was called with dryRun=false
	if len(mockConfigSvc.ConfigureCalls) != 1 {
		t.Fatalf("Configure called %d times, expected 1", len(mockConfigSvc.ConfigureCalls))
	}
	if mockConfigSvc.ConfigureCalls[0].DryRun {
		t.Error("Configure should be called with dryRun=false in normal mode")
	}
}

// TestRunNormalModeOutputsSuccessMessage verifies success output.
//
// The implementation should:
// - Output success message when feature is enabled
func TestRunNormalModeOutputsSuccessMessage(t *testing.T) {
	// Arrange
	mockWriter := &mockOutputWriter{}
	mockConfigSvc := &mockConfigService{
		ConfigureFunc: func(ctx context.Context, owner, name string, dryRun bool) (interfaces.IConfigResult, error) {
			return newMockConfigResult(false, true, "main", "octocat/hello-world"), nil
		},
	}
	mockParser := &mockRepoParser{
		ParseFunc: func(repoIdentifier string) (string, string, error) {
			return "octocat", "hello-world", nil
		},
	}

	application := app.NewApp(mockWriter, mockConfigSvc, mockParser)
	ctx := context.Background()
	opts := interfaces.CLIOptions{
		Repository: "octocat/hello-world",
		DryRun:     false,
		CheckOnly:  false,
	}

	// Act
	err := application.Run(ctx, opts)

	// Assert
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	// Assert - success message was output
	if len(mockWriter.SuccessCalls) == 0 {
		t.Error("Expected at least one Success() call for successful enable")
	}
}

// TestRunNormalModeWhenAlreadyEnabledOutputsMessage verifies already enabled output.
//
// The implementation should:
// - Output "already enabled" message when feature was already enabled
func TestRunNormalModeWhenAlreadyEnabledOutputsMessage(t *testing.T) {
	// Arrange
	mockWriter := &mockOutputWriter{}
	mockConfigSvc := &mockConfigService{
		ConfigureFunc: func(ctx context.Context, owner, name string, dryRun bool) (interfaces.IConfigResult, error) {
			return newMockConfigResult(true, true, "main", "octocat/hello-world"), nil
		},
	}
	mockParser := &mockRepoParser{
		ParseFunc: func(repoIdentifier string) (string, string, error) {
			return "octocat", "hello-world", nil
		},
	}

	application := app.NewApp(mockWriter, mockConfigSvc, mockParser)
	ctx := context.Background()
	opts := interfaces.CLIOptions{
		Repository: "octocat/hello-world",
		DryRun:     false,
		CheckOnly:  false,
	}

	// Act
	err := application.Run(ctx, opts)

	// Assert
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	// Assert - output mentions already enabled
	output := mockWriter.GetAllOutput()
	if !strings.Contains(output, "already enabled") {
		t.Errorf("Output should mention 'already enabled' when feature was already on, got: %s", output)
	}
}

// =============================================================================
// Short Flag Tests
// =============================================================================

// TestRunShortFlagCWorksLikeCheckOnly verifies -c is equivalent to --check.
//
// Gherkin: Scenario: Use short form -c for check mode
//
// The implementation should:
// - Treat CheckOnly=true same regardless of how it was set (--check or -c)
// Note: The CLI parsing handles -c -> CheckOnly=true, App just sees CheckOnly=true
func TestRunShortFlagCWorksLikeCheckOnly(t *testing.T) {
	// Arrange
	mockWriter := &mockOutputWriter{}
	mockConfigSvc := &mockConfigService{
		CheckStatusFunc: func(ctx context.Context, owner, name string) (interfaces.IConfigResult, error) {
			return newMockConfigResult(false, false, "main", "octocat/hello-world"), nil
		},
	}
	mockParser := &mockRepoParser{
		ParseFunc: func(repoIdentifier string) (string, string, error) {
			return "octocat", "hello-world", nil
		},
	}

	application := app.NewApp(mockWriter, mockConfigSvc, mockParser)
	ctx := context.Background()

	// -c flag sets CheckOnly=true (CLI handles the flag parsing)
	opts := interfaces.CLIOptions{
		Repository: "octocat/hello-world",
		CheckOnly:  true, // This is what -c sets
	}

	// Act
	err := application.Run(ctx, opts)

	// Assert
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	// Assert - CheckStatus was called (not Configure)
	if len(mockConfigSvc.CheckStatusCalls) != 1 {
		t.Errorf("CheckStatus should be called once for -c flag, called %d times", len(mockConfigSvc.CheckStatusCalls))
	}
	if len(mockConfigSvc.ConfigureCalls) != 0 {
		t.Errorf("Configure should NOT be called for -c flag, called %d times", len(mockConfigSvc.ConfigureCalls))
	}
}

// TestRunShortFlagDWorksLikeDryRun verifies -d is equivalent to --dry-run.
//
// Gherkin: Scenario: Use short form -d for dry-run mode
//
// The implementation should:
// - Treat DryRun=true same regardless of how it was set (--dry-run or -d)
// Note: The CLI parsing handles -d -> DryRun=true, App just sees DryRun=true
func TestRunShortFlagDWorksLikeDryRun(t *testing.T) {
	// Arrange
	mockWriter := &mockOutputWriter{}
	mockConfigSvc := &mockConfigService{
		ConfigureFunc: func(ctx context.Context, owner, name string, dryRun bool) (interfaces.IConfigResult, error) {
			if !dryRun {
				t.Error("Configure should receive dryRun=true for -d flag")
			}
			return newMockConfigResult(false, false, "main", "octocat/hello-world"), nil
		},
	}
	mockParser := &mockRepoParser{
		ParseFunc: func(repoIdentifier string) (string, string, error) {
			return "octocat", "hello-world", nil
		},
	}

	application := app.NewApp(mockWriter, mockConfigSvc, mockParser)
	ctx := context.Background()

	// -d flag sets DryRun=true (CLI handles the flag parsing)
	opts := interfaces.CLIOptions{
		Repository: "octocat/hello-world",
		DryRun:     true, // This is what -d sets
	}

	// Act
	err := application.Run(ctx, opts)

	// Assert
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	// Assert - Configure was called with dryRun=true
	if len(mockConfigSvc.ConfigureCalls) != 1 {
		t.Fatalf("Configure should be called once for -d flag, called %d times", len(mockConfigSvc.ConfigureCalls))
	}
	if !mockConfigSvc.ConfigureCalls[0].DryRun {
		t.Error("Configure should receive dryRun=true for -d flag")
	}
}

// =============================================================================
// Repository Parsing Tests
// =============================================================================

// TestRunParsesRepositoryIdentifier verifies repository parsing.
//
// The implementation should:
// - Use IRepoParser to parse the repository identifier
// - Pass owner and name to ConfigService methods
func TestRunParsesRepositoryIdentifier(t *testing.T) {
	// Arrange
	mockWriter := &mockOutputWriter{}
	mockConfigSvc := &mockConfigService{
		CheckStatusFunc: func(ctx context.Context, owner, name string) (interfaces.IConfigResult, error) {
			if owner != "github" || name != "docs" {
				t.Errorf("Expected owner='github', name='docs', got owner='%s', name='%s'", owner, name)
			}
			return newMockConfigResult(true, true, "main", "github/docs"), nil
		},
	}
	mockParser := &mockRepoParser{
		ParseFunc: func(repoIdentifier string) (string, string, error) {
			if repoIdentifier != "github/docs" {
				t.Errorf("Expected to parse 'github/docs', got '%s'", repoIdentifier)
			}
			return "github", "docs", nil
		},
	}

	application := app.NewApp(mockWriter, mockConfigSvc, mockParser)
	ctx := context.Background()
	opts := interfaces.CLIOptions{
		Repository: "github/docs",
		CheckOnly:  true,
	}

	// Act
	err := application.Run(ctx, opts)

	// Assert
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	// Assert - parser was called with correct identifier
	if len(mockParser.ParseCalls) != 1 {
		t.Fatalf("Parse called %d times, expected 1", len(mockParser.ParseCalls))
	}
	if mockParser.ParseCalls[0] != "github/docs" {
		t.Errorf("Parse called with '%s', expected 'github/docs'", mockParser.ParseCalls[0])
	}
}

// TestRunReturnsErrorForInvalidRepository verifies error handling for invalid repo.
//
// The implementation should:
// - Return error if IRepoParser.Parse fails
func TestRunReturnsErrorForInvalidRepository(t *testing.T) {
	// Arrange
	mockWriter := &mockOutputWriter{}
	mockConfigSvc := &mockConfigService{}
	mockParser := &mockRepoParser{
		ParseFunc: func(repoIdentifier string) (string, string, error) {
			return "", "", errors.New("invalid repository format")
		},
	}

	application := app.NewApp(mockWriter, mockConfigSvc, mockParser)
	ctx := context.Background()
	opts := interfaces.CLIOptions{
		Repository: "invalid-repo-format",
		CheckOnly:  true,
	}

	// Act
	err := application.Run(ctx, opts)

	// Assert - error should be returned
	if err == nil {
		t.Error("Run() should return error for invalid repository format")
	}
}

// =============================================================================
// Error Propagation Tests
// =============================================================================

// TestRunPropagatesCheckStatusError verifies error propagation from CheckStatus.
//
// The implementation should:
// - Propagate errors from ConfigService.CheckStatus
func TestRunPropagatesCheckStatusError(t *testing.T) {
	// Arrange
	expectedError := errors.New("check status failed")
	mockWriter := &mockOutputWriter{}
	mockConfigSvc := &mockConfigService{
		CheckStatusFunc: func(ctx context.Context, owner, name string) (interfaces.IConfigResult, error) {
			return nil, expectedError
		},
	}
	mockParser := &mockRepoParser{
		ParseFunc: func(repoIdentifier string) (string, string, error) {
			return "octocat", "hello-world", nil
		},
	}

	application := app.NewApp(mockWriter, mockConfigSvc, mockParser)
	ctx := context.Background()
	opts := interfaces.CLIOptions{
		Repository: "octocat/hello-world",
		CheckOnly:  true,
	}

	// Act
	err := application.Run(ctx, opts)

	// Assert
	if err == nil {
		t.Fatal("Run() should return error when CheckStatus fails")
	}
}

// TestRunPropagatesConfigureError verifies error propagation from Configure.
//
// The implementation should:
// - Propagate errors from ConfigService.Configure
func TestRunPropagatesConfigureError(t *testing.T) {
	// Arrange
	expectedError := errors.New("configure failed")
	mockWriter := &mockOutputWriter{}
	mockConfigSvc := &mockConfigService{
		ConfigureFunc: func(ctx context.Context, owner, name string, dryRun bool) (interfaces.IConfigResult, error) {
			return nil, expectedError
		},
	}
	mockParser := &mockRepoParser{
		ParseFunc: func(repoIdentifier string) (string, string, error) {
			return "octocat", "hello-world", nil
		},
	}

	application := app.NewApp(mockWriter, mockConfigSvc, mockParser)
	ctx := context.Background()
	opts := interfaces.CLIOptions{
		Repository: "octocat/hello-world",
		DryRun:     false,
		CheckOnly:  false,
	}

	// Act
	err := application.Run(ctx, opts)

	// Assert
	if err == nil {
		t.Fatal("Run() should return error when Configure fails")
	}
}

// =============================================================================
// Context Handling Tests
// =============================================================================

// TestRunPassesContextToConfigService verifies context is passed through.
//
// The implementation should:
// - Pass the context to ConfigService methods
func TestRunPassesContextToConfigService(t *testing.T) {
	// Arrange
	type contextKey string
	ctx := context.WithValue(context.Background(), contextKey("key"), "value")
	var receivedCtx context.Context

	mockWriter := &mockOutputWriter{}
	mockConfigSvc := &mockConfigService{
		CheckStatusFunc: func(ctx context.Context, owner, name string) (interfaces.IConfigResult, error) {
			receivedCtx = ctx
			return newMockConfigResult(true, true, "main", "octocat/hello-world"), nil
		},
	}
	mockParser := &mockRepoParser{
		ParseFunc: func(repoIdentifier string) (string, string, error) {
			return "octocat", "hello-world", nil
		},
	}

	application := app.NewApp(mockWriter, mockConfigSvc, mockParser)
	opts := interfaces.CLIOptions{
		Repository: "octocat/hello-world",
		CheckOnly:  true,
	}

	// Act
	err := application.Run(ctx, opts)

	// Assert
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if receivedCtx == nil {
		t.Fatal("Context was not passed to CheckStatus")
	}
	if receivedCtx.Value(contextKey("key")) != "value" {
		t.Error("Context value was not preserved")
	}
}

// =============================================================================
// Table-Driven Tests
// =============================================================================

// TestRunModeScenarios tests various mode combinations.
func TestRunModeScenarios(t *testing.T) {
	tests := []struct {
		name                  string
		checkOnly             bool
		dryRun                bool
		alreadyEnabled        bool
		expectedCheckCalls    int
		expectedConfigCalls   int
		expectedDryRunInCall  bool
		expectedOutputStrings []string
	}{
		{
			name:                  "check mode disabled",
			checkOnly:             true,
			dryRun:                false,
			alreadyEnabled:        false,
			expectedCheckCalls:    1,
			expectedConfigCalls:   0,
			expectedDryRunInCall:  false,
			expectedOutputStrings: []string{"Repository:", "disabled", "To enable"},
		},
		{
			name:                  "check mode enabled",
			checkOnly:             true,
			dryRun:                false,
			alreadyEnabled:        true,
			expectedCheckCalls:    1,
			expectedConfigCalls:   0,
			expectedDryRunInCall:  false,
			expectedOutputStrings: []string{"Repository:", "enabled"},
		},
		{
			name:                  "dry-run mode disabled",
			checkOnly:             false,
			dryRun:                true,
			alreadyEnabled:        false,
			expectedCheckCalls:    0,
			expectedConfigCalls:   1,
			expectedDryRunInCall:  true,
			expectedOutputStrings: []string{"[DRY-RUN]", "No changes made"},
		},
		{
			name:                  "dry-run mode enabled",
			checkOnly:             false,
			dryRun:                true,
			alreadyEnabled:        true,
			expectedCheckCalls:    0,
			expectedConfigCalls:   1,
			expectedDryRunInCall:  true,
			expectedOutputStrings: []string{"already enabled", "No changes needed"},
		},
		{
			name:                  "normal mode enable",
			checkOnly:             false,
			dryRun:                false,
			alreadyEnabled:        false,
			expectedCheckCalls:    0,
			expectedConfigCalls:   1,
			expectedDryRunInCall:  false,
			expectedOutputStrings: []string{}, // Success message varies
		},
		{
			name:                  "normal mode already enabled",
			checkOnly:             false,
			dryRun:                false,
			alreadyEnabled:        true,
			expectedCheckCalls:    0,
			expectedConfigCalls:   1,
			expectedDryRunInCall:  false,
			expectedOutputStrings: []string{"already enabled"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockWriter := &mockOutputWriter{}
			mockConfigSvc := &mockConfigService{
				CheckStatusFunc: func(ctx context.Context, owner, name string) (interfaces.IConfigResult, error) {
					return newMockConfigResult(tt.alreadyEnabled, tt.alreadyEnabled, "main", "owner/repo"), nil
				},
				ConfigureFunc: func(ctx context.Context, owner, name string, dryRun bool) (interfaces.IConfigResult, error) {
					if tt.alreadyEnabled {
						return newMockConfigResult(true, true, "main", "owner/repo"), nil
					}
					if dryRun {
						return newMockConfigResult(false, false, "main", "owner/repo"), nil
					}
					return newMockConfigResult(false, true, "main", "owner/repo"), nil
				},
			}
			mockParser := &mockRepoParser{
				ParseFunc: func(repoIdentifier string) (string, string, error) {
					return "owner", "repo", nil
				},
			}

			application := app.NewApp(mockWriter, mockConfigSvc, mockParser)
			ctx := context.Background()
			opts := interfaces.CLIOptions{
				Repository: "owner/repo",
				CheckOnly:  tt.checkOnly,
				DryRun:     tt.dryRun,
			}

			// Act
			err := application.Run(ctx, opts)

			// Assert - no error
			if err != nil {
				t.Fatalf("Run() error = %v", err)
			}

			// Assert - correct method calls
			if len(mockConfigSvc.CheckStatusCalls) != tt.expectedCheckCalls {
				t.Errorf("CheckStatus called %d times, expected %d", len(mockConfigSvc.CheckStatusCalls), tt.expectedCheckCalls)
			}
			if len(mockConfigSvc.ConfigureCalls) != tt.expectedConfigCalls {
				t.Errorf("Configure called %d times, expected %d", len(mockConfigSvc.ConfigureCalls), tt.expectedConfigCalls)
			}

			// Assert - dryRun parameter if Configure was called
			if tt.expectedConfigCalls > 0 && len(mockConfigSvc.ConfigureCalls) > 0 {
				if mockConfigSvc.ConfigureCalls[0].DryRun != tt.expectedDryRunInCall {
					t.Errorf("Configure dryRun = %v, expected %v", mockConfigSvc.ConfigureCalls[0].DryRun, tt.expectedDryRunInCall)
				}
			}

			// Assert - output contains expected strings
			output := mockWriter.GetAllOutput()
			for _, expected := range tt.expectedOutputStrings {
				if !strings.Contains(output, expected) {
					t.Errorf("Output should contain '%s', got: %s", expected, output)
				}
			}
		})
	}
}

// =============================================================================
// Verbose Mode Tests
// =============================================================================

// TestRunVerboseModePassesToWriter verifies verbose flag is used.
//
// The implementation should:
// - Pass verbose messages through the IOutputWriter
// Note: The App receives a pre-configured IOutputWriter that handles verbose internally
func TestRunVerboseModeOutputsVerboseMessages(t *testing.T) {
	// Arrange
	mockWriter := &mockOutputWriter{}
	mockConfigSvc := &mockConfigService{
		CheckStatusFunc: func(ctx context.Context, owner, name string) (interfaces.IConfigResult, error) {
			return newMockConfigResult(true, true, "main", "octocat/hello-world"), nil
		},
	}
	mockParser := &mockRepoParser{
		ParseFunc: func(repoIdentifier string) (string, string, error) {
			return "octocat", "hello-world", nil
		},
	}

	application := app.NewApp(mockWriter, mockConfigSvc, mockParser)
	ctx := context.Background()
	opts := interfaces.CLIOptions{
		Repository: "octocat/hello-world",
		CheckOnly:  true,
		Verbose:    true,
	}

	// Act
	err := application.Run(ctx, opts)

	// Assert
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	// Note: Verbose output is controlled by the IOutputWriter implementation
	// The App may or may not emit verbose messages based on its internal logic
	// This test verifies no error occurs with verbose=true
}

// =============================================================================
// Edge Case Tests
// =============================================================================

// TestRunWithEmptyRepository verifies error for empty repository.
//
// The implementation should:
// - Handle empty repository identifier gracefully
func TestRunWithEmptyRepository(t *testing.T) {
	// Arrange
	mockWriter := &mockOutputWriter{}
	mockConfigSvc := &mockConfigService{}
	mockParser := &mockRepoParser{
		ParseFunc: func(repoIdentifier string) (string, string, error) {
			if repoIdentifier == "" {
				return "", "", errors.New("empty repository identifier")
			}
			return "owner", "repo", nil
		},
	}

	application := app.NewApp(mockWriter, mockConfigSvc, mockParser)
	ctx := context.Background()
	opts := interfaces.CLIOptions{
		Repository: "",
		CheckOnly:  true,
	}

	// Act
	err := application.Run(ctx, opts)

	// Assert - error expected for empty repository
	if err == nil {
		t.Error("Run() should return error for empty repository")
	}
}

// TestRunCheckOnlyTakesPrecedenceOverDryRun verifies mode priority.
//
// The implementation should:
// - If both CheckOnly and DryRun are true, CheckOnly takes precedence
//   (Check mode is read-only, dry-run is preview of write operation)
func TestRunCheckOnlyTakesPrecedenceOverDryRun(t *testing.T) {
	// Arrange
	mockWriter := &mockOutputWriter{}
	mockConfigSvc := &mockConfigService{
		CheckStatusFunc: func(ctx context.Context, owner, name string) (interfaces.IConfigResult, error) {
			return newMockConfigResult(false, false, "main", "octocat/hello-world"), nil
		},
		ConfigureFunc: func(ctx context.Context, owner, name string, dryRun bool) (interfaces.IConfigResult, error) {
			t.Error("Configure should NOT be called when CheckOnly=true")
			return nil, nil
		},
	}
	mockParser := &mockRepoParser{
		ParseFunc: func(repoIdentifier string) (string, string, error) {
			return "octocat", "hello-world", nil
		},
	}

	application := app.NewApp(mockWriter, mockConfigSvc, mockParser)
	ctx := context.Background()
	opts := interfaces.CLIOptions{
		Repository: "octocat/hello-world",
		CheckOnly:  true,
		DryRun:     true, // Both flags set
	}

	// Act
	err := application.Run(ctx, opts)

	// Assert
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	// Assert - CheckStatus was called (not Configure)
	if len(mockConfigSvc.CheckStatusCalls) != 1 {
		t.Errorf("CheckStatus should be called when CheckOnly=true, called %d times", len(mockConfigSvc.CheckStatusCalls))
	}
	if len(mockConfigSvc.ConfigureCalls) != 0 {
		t.Errorf("Configure should NOT be called when CheckOnly=true, called %d times", len(mockConfigSvc.ConfigureCalls))
	}
}
