// Package config_test provides tests for the ConfigService implementation.
//
// These tests verify that ConfigService correctly implements the IConfigService
// interface and handles the configuration workflow:
// - Fetching repository state
// - Checking if already enabled (returns AlreadyEnabled=true, NowEnabled=true)
// - Updating repository settings
// - Verifying settings were applied
// - Verbose logging via IOutputWriter
//
// Gherkin Scenarios:
//   - Scenario: Successfully enable auto-delete branches on a repository
//     - Setting was disabled -> enable -> verify -> success
//   - Scenario: Report when auto-delete branches is already enabled
//     - Setting already enabled -> skip update -> return AlreadyEnabled=true
//   - Scenario: Verify setting was applied after update
//     - After update, fetch again to verify
//
// These tests are designed to FAIL until the implementation is properly created
// by the coder agent.
package config_test

import (
	"context"
	"errors"
	"testing"

	"github.com/josejulio/ghautodelete/internal/config"
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

// mockOutputWriter implements IOutputWriter for testing.
type mockOutputWriter struct {
	// VerboseCalls tracks all Verbose messages.
	VerboseCalls []string
	// SuccessCalls tracks all Success messages.
	SuccessCalls []string
	// ErrorCalls tracks all Error messages.
	ErrorCalls []string
	// InfoCalls tracks all Info messages.
	InfoCalls []string
}

func (m *mockOutputWriter) Success(message string) {
	m.SuccessCalls = append(m.SuccessCalls, message)
}

func (m *mockOutputWriter) Error(message string) {
	m.ErrorCalls = append(m.ErrorCalls, message)
}

func (m *mockOutputWriter) Info(message string) {
	m.InfoCalls = append(m.InfoCalls, message)
}

func (m *mockOutputWriter) Verbose(message string) {
	m.VerboseCalls = append(m.VerboseCalls, message)
}

// mockRepository implements IRepository for testing.
type mockRepository struct {
	owner               string
	name                string
	defaultBranch       string
	deleteBranchOnMerge bool
}

func (m *mockRepository) GetOwner() string {
	return m.owner
}

func (m *mockRepository) GetName() string {
	return m.name
}

func (m *mockRepository) GetDefaultBranch() string {
	return m.defaultBranch
}

func (m *mockRepository) GetDeleteBranchOnMerge() bool {
	return m.deleteBranchOnMerge
}

func (m *mockRepository) GetFullName() string {
	return m.owner + "/" + m.name
}

// =============================================================================
// Interface Satisfaction Tests
// =============================================================================

// TestConfigServiceImplementsIConfigService verifies ConfigService implements IConfigService.
//
// The implementation should:
// - Define ConfigService struct in internal/config/service.go
// - Implement Configure(ctx, owner, name, dryRun) (IConfigResult, error)
// - Implement CheckStatus(ctx, owner, name) (IConfigResult, error)
func TestConfigServiceImplementsIConfigService(t *testing.T) {
	// Arrange
	mockClient := &mockGitHubClient{}
	mockWriter := &mockOutputWriter{}

	// Act
	service := config.NewConfigService(mockClient, mockWriter)

	// Assert - compile-time interface satisfaction check
	var _ interfaces.IConfigService = service
	if service == nil {
		t.Error("NewConfigService should return a non-nil service")
	}
}

// TestNewConfigServiceReturnsNonNil verifies constructor returns non-nil.
//
// The implementation should:
// - Return a non-nil ConfigService from NewConfigService
func TestNewConfigServiceReturnsNonNil(t *testing.T) {
	// Arrange
	mockClient := &mockGitHubClient{}
	mockWriter := &mockOutputWriter{}

	// Act
	service := config.NewConfigService(mockClient, mockWriter)

	// Assert
	if service == nil {
		t.Error("NewConfigService() should return non-nil service")
	}
}

// TestNewConfigServiceAcceptsDependencies verifies constructor accepts correct dependencies.
//
// The implementation should:
// - Accept IGitHubClient as first parameter
// - Accept IOutputWriter as second parameter
func TestNewConfigServiceAcceptsDependencies(t *testing.T) {
	// Arrange
	var client interfaces.IGitHubClient = &mockGitHubClient{}
	var writer interfaces.IOutputWriter = &mockOutputWriter{}

	// Act
	service := config.NewConfigService(client, writer)

	// Assert
	if service == nil {
		t.Error("NewConfigService should accept interface parameters")
	}
}

// =============================================================================
// Configure Method Tests - Enable Flow (disabled -> enabled)
// =============================================================================

// TestConfigureEnablesAutoDeleteWhenDisabled verifies the enable flow.
//
// Gherkin: Scenario: Successfully enable auto-delete branches on a repository
// - Setting was disabled -> enable -> verify -> success
//
// The implementation should:
// - Fetch current repository state
// - If disabled, update to enabled
// - Verify by fetching again
// - Return AlreadyEnabled=false, NowEnabled=true
func TestConfigureEnablesAutoDeleteWhenDisabled(t *testing.T) {
	// Arrange
	fetchCount := 0
	mockClient := &mockGitHubClient{
		GetRepositoryFunc: func(ctx context.Context, owner, name string) (interfaces.IRepository, error) {
			fetchCount++
			if fetchCount == 1 {
				// First fetch: disabled
				return &mockRepository{
					owner:               "octocat",
					name:                "hello-world",
					defaultBranch:       "main",
					deleteBranchOnMerge: false,
				}, nil
			}
			// Second fetch (verification): enabled
			return &mockRepository{
				owner:               "octocat",
				name:                "hello-world",
				defaultBranch:       "main",
				deleteBranchOnMerge: true,
			}, nil
		},
		UpdateRepositoryFunc: func(ctx context.Context, owner, name string, settings interfaces.IRepositorySettings) error {
			return nil
		},
	}
	mockWriter := &mockOutputWriter{}
	service := config.NewConfigService(mockClient, mockWriter)
	ctx := context.Background()

	// Act
	result, err := service.Configure(ctx, "octocat", "hello-world", false)

	// Assert - no error
	if err != nil {
		t.Fatalf("Configure() error = %v, expected nil", err)
	}

	// Assert - result not nil
	if result == nil {
		t.Fatal("Configure() returned nil result")
	}

	// Assert - WasAlreadyEnabled should be false (it was disabled before)
	if result.WasAlreadyEnabled() {
		t.Error("WasAlreadyEnabled() = true, expected false (feature was disabled)")
	}

	// Assert - IsNowEnabled should be true (feature is now enabled)
	if !result.IsNowEnabled() {
		t.Error("IsNowEnabled() = false, expected true (feature should now be enabled)")
	}

	// Assert - DefaultBranch and RepositoryFullName should be correct
	if result.GetDefaultBranch() != "main" {
		t.Errorf("GetDefaultBranch() = %q, expected %q", result.GetDefaultBranch(), "main")
	}
	if result.GetRepositoryFullName() != "octocat/hello-world" {
		t.Errorf("GetRepositoryFullName() = %q, expected %q", result.GetRepositoryFullName(), "octocat/hello-world")
	}
}

// TestConfigureCallsUpdateRepositoryWhenDisabled verifies UpdateRepository is called.
//
// The implementation should:
// - Call UpdateRepository with delete_branch_on_merge=true when disabled
func TestConfigureCallsUpdateRepositoryWhenDisabled(t *testing.T) {
	// Arrange
	fetchCount := 0
	mockClient := &mockGitHubClient{
		GetRepositoryFunc: func(ctx context.Context, owner, name string) (interfaces.IRepository, error) {
			fetchCount++
			deleteBranchOnMerge := fetchCount > 1 // enabled after update
			return &mockRepository{
				owner:               owner,
				name:                name,
				defaultBranch:       "main",
				deleteBranchOnMerge: deleteBranchOnMerge,
			}, nil
		},
		UpdateRepositoryFunc: func(ctx context.Context, owner, name string, settings interfaces.IRepositorySettings) error {
			return nil
		},
	}
	mockWriter := &mockOutputWriter{}
	service := config.NewConfigService(mockClient, mockWriter)
	ctx := context.Background()

	// Act
	_, err := service.Configure(ctx, "octocat", "hello-world", false)

	// Assert
	if err != nil {
		t.Fatalf("Configure() error = %v", err)
	}

	// Assert - UpdateRepository was called once
	if len(mockClient.UpdateRepositoryCalls) != 1 {
		t.Errorf("UpdateRepository called %d times, expected 1", len(mockClient.UpdateRepositoryCalls))
	}

	// Assert - UpdateRepository called with correct parameters
	if len(mockClient.UpdateRepositoryCalls) > 0 {
		call := mockClient.UpdateRepositoryCalls[0]
		if call.Owner != "octocat" {
			t.Errorf("UpdateRepository owner = %q, expected %q", call.Owner, "octocat")
		}
		if call.Name != "hello-world" {
			t.Errorf("UpdateRepository name = %q, expected %q", call.Name, "hello-world")
		}
		if call.Settings == nil {
			t.Error("UpdateRepository settings = nil, expected non-nil")
		} else if !call.Settings.GetDeleteBranchOnMerge() {
			t.Error("UpdateRepository settings.DeleteBranchOnMerge = false, expected true")
		}
	}
}

// TestConfigureVerifiesSettingAfterUpdate verifies the verification step.
//
// Gherkin: Scenario: Verify setting was applied after update
// - After update, fetch again to verify
//
// The implementation should:
// - Call GetRepository twice (once before, once after update)
func TestConfigureVerifiesSettingAfterUpdate(t *testing.T) {
	// Arrange
	fetchCount := 0
	mockClient := &mockGitHubClient{
		GetRepositoryFunc: func(ctx context.Context, owner, name string) (interfaces.IRepository, error) {
			fetchCount++
			deleteBranchOnMerge := fetchCount > 1
			return &mockRepository{
				owner:               owner,
				name:                name,
				defaultBranch:       "main",
				deleteBranchOnMerge: deleteBranchOnMerge,
			}, nil
		},
		UpdateRepositoryFunc: func(ctx context.Context, owner, name string, settings interfaces.IRepositorySettings) error {
			return nil
		},
	}
	mockWriter := &mockOutputWriter{}
	service := config.NewConfigService(mockClient, mockWriter)
	ctx := context.Background()

	// Act
	_, err := service.Configure(ctx, "octocat", "hello-world", false)

	// Assert
	if err != nil {
		t.Fatalf("Configure() error = %v", err)
	}

	// Assert - GetRepository called twice (fetch + verify)
	if len(mockClient.GetRepositoryCalls) != 2 {
		t.Errorf("GetRepository called %d times, expected 2 (initial fetch + verification)", len(mockClient.GetRepositoryCalls))
	}
}

// =============================================================================
// Configure Method Tests - Already Enabled Flow
// =============================================================================

// TestConfigureReturnsAlreadyEnabledWhenEnabled verifies already-enabled flow.
//
// Gherkin: Scenario: Report when auto-delete branches is already enabled
// - Setting already enabled -> skip update -> return AlreadyEnabled=true
//
// The implementation should:
// - Fetch current repository state
// - If already enabled, return AlreadyEnabled=true, NowEnabled=true
// - NOT call UpdateRepository
func TestConfigureReturnsAlreadyEnabledWhenEnabled(t *testing.T) {
	// Arrange
	mockClient := &mockGitHubClient{
		GetRepositoryFunc: func(ctx context.Context, owner, name string) (interfaces.IRepository, error) {
			return &mockRepository{
				owner:               "octocat",
				name:                "hello-world",
				defaultBranch:       "main",
				deleteBranchOnMerge: true, // Already enabled
			}, nil
		},
		UpdateRepositoryFunc: func(ctx context.Context, owner, name string, settings interfaces.IRepositorySettings) error {
			t.Error("UpdateRepository should NOT be called when already enabled")
			return nil
		},
	}
	mockWriter := &mockOutputWriter{}
	service := config.NewConfigService(mockClient, mockWriter)
	ctx := context.Background()

	// Act
	result, err := service.Configure(ctx, "octocat", "hello-world", false)

	// Assert - no error
	if err != nil {
		t.Fatalf("Configure() error = %v, expected nil", err)
	}

	// Assert - result not nil
	if result == nil {
		t.Fatal("Configure() returned nil result")
	}

	// Assert - WasAlreadyEnabled should be true
	if !result.WasAlreadyEnabled() {
		t.Error("WasAlreadyEnabled() = false, expected true (feature was already enabled)")
	}

	// Assert - IsNowEnabled should be true
	if !result.IsNowEnabled() {
		t.Error("IsNowEnabled() = false, expected true (feature is still enabled)")
	}

	// Assert - UpdateRepository was NOT called
	if len(mockClient.UpdateRepositoryCalls) != 0 {
		t.Errorf("UpdateRepository called %d times, expected 0 (should skip when already enabled)", len(mockClient.UpdateRepositoryCalls))
	}
}

// TestConfigureSkipsUpdateWhenAlreadyEnabled verifies no update when already enabled.
//
// The implementation should:
// - Only fetch repository once when already enabled (no verification needed)
func TestConfigureSkipsUpdateWhenAlreadyEnabled(t *testing.T) {
	// Arrange
	mockClient := &mockGitHubClient{
		GetRepositoryFunc: func(ctx context.Context, owner, name string) (interfaces.IRepository, error) {
			return &mockRepository{
				owner:               owner,
				name:                name,
				defaultBranch:       "main",
				deleteBranchOnMerge: true,
			}, nil
		},
	}
	mockWriter := &mockOutputWriter{}
	service := config.NewConfigService(mockClient, mockWriter)
	ctx := context.Background()

	// Act
	_, err := service.Configure(ctx, "octocat", "hello-world", false)

	// Assert
	if err != nil {
		t.Fatalf("Configure() error = %v", err)
	}

	// Assert - GetRepository called only once (no verification needed)
	if len(mockClient.GetRepositoryCalls) != 1 {
		t.Errorf("GetRepository called %d times, expected 1 (no verification when already enabled)", len(mockClient.GetRepositoryCalls))
	}

	// Assert - UpdateRepository NOT called
	if len(mockClient.UpdateRepositoryCalls) != 0 {
		t.Errorf("UpdateRepository called %d times, expected 0", len(mockClient.UpdateRepositoryCalls))
	}
}

// =============================================================================
// Configure Method Tests - Dry Run Mode
// =============================================================================

// TestConfigureDryRunDoesNotUpdate verifies dry-run mode.
//
// The implementation should:
// - NOT call UpdateRepository when dryRun=true
// - Return NowEnabled=false (no actual change)
func TestConfigureDryRunDoesNotUpdate(t *testing.T) {
	// Arrange
	mockClient := &mockGitHubClient{
		GetRepositoryFunc: func(ctx context.Context, owner, name string) (interfaces.IRepository, error) {
			return &mockRepository{
				owner:               owner,
				name:                name,
				defaultBranch:       "main",
				deleteBranchOnMerge: false, // Disabled
			}, nil
		},
		UpdateRepositoryFunc: func(ctx context.Context, owner, name string, settings interfaces.IRepositorySettings) error {
			t.Error("UpdateRepository should NOT be called in dry-run mode")
			return nil
		},
	}
	mockWriter := &mockOutputWriter{}
	service := config.NewConfigService(mockClient, mockWriter)
	ctx := context.Background()

	// Act - dry run enabled
	result, err := service.Configure(ctx, "octocat", "hello-world", true)

	// Assert - no error
	if err != nil {
		t.Fatalf("Configure() error = %v, expected nil", err)
	}

	// Assert - result not nil
	if result == nil {
		t.Fatal("Configure() returned nil result")
	}

	// Assert - WasAlreadyEnabled should be false
	if result.WasAlreadyEnabled() {
		t.Error("WasAlreadyEnabled() = true, expected false")
	}

	// Assert - IsNowEnabled should be false (no actual change in dry-run)
	if result.IsNowEnabled() {
		t.Error("IsNowEnabled() = true, expected false (dry-run should not actually enable)")
	}

	// Assert - UpdateRepository was NOT called
	if len(mockClient.UpdateRepositoryCalls) != 0 {
		t.Errorf("UpdateRepository called %d times, expected 0 in dry-run mode", len(mockClient.UpdateRepositoryCalls))
	}
}

// TestConfigureDryRunStillFetchesRepository verifies dry-run fetches repo.
//
// The implementation should:
// - Fetch repository even in dry-run mode to check current state
func TestConfigureDryRunStillFetchesRepository(t *testing.T) {
	// Arrange
	mockClient := &mockGitHubClient{
		GetRepositoryFunc: func(ctx context.Context, owner, name string) (interfaces.IRepository, error) {
			return &mockRepository{
				owner:               owner,
				name:                name,
				defaultBranch:       "main",
				deleteBranchOnMerge: false,
			}, nil
		},
	}
	mockWriter := &mockOutputWriter{}
	service := config.NewConfigService(mockClient, mockWriter)
	ctx := context.Background()

	// Act
	_, err := service.Configure(ctx, "octocat", "hello-world", true)

	// Assert
	if err != nil {
		t.Fatalf("Configure() error = %v", err)
	}

	// Assert - GetRepository called at least once
	if len(mockClient.GetRepositoryCalls) < 1 {
		t.Error("GetRepository should be called even in dry-run mode")
	}
}

// =============================================================================
// CheckStatus Method Tests
// =============================================================================

// TestCheckStatusReturnsCurrentState verifies CheckStatus returns current state.
//
// The implementation should:
// - Fetch repository state
// - Return current state without modification
func TestCheckStatusReturnsCurrentState(t *testing.T) {
	// Arrange
	mockClient := &mockGitHubClient{
		GetRepositoryFunc: func(ctx context.Context, owner, name string) (interfaces.IRepository, error) {
			return &mockRepository{
				owner:               owner,
				name:                name,
				defaultBranch:       "main",
				deleteBranchOnMerge: true,
			}, nil
		},
	}
	mockWriter := &mockOutputWriter{}
	service := config.NewConfigService(mockClient, mockWriter)
	ctx := context.Background()

	// Act
	result, err := service.CheckStatus(ctx, "octocat", "hello-world")

	// Assert - no error
	if err != nil {
		t.Fatalf("CheckStatus() error = %v, expected nil", err)
	}

	// Assert - result not nil
	if result == nil {
		t.Fatal("CheckStatus() returned nil result")
	}

	// Assert - reflects current state
	if !result.IsNowEnabled() {
		t.Error("IsNowEnabled() = false, expected true (reflects current enabled state)")
	}
}

// TestCheckStatusReturnsDisabledState verifies CheckStatus when disabled.
//
// The implementation should:
// - Return NowEnabled=false when feature is disabled
func TestCheckStatusReturnsDisabledState(t *testing.T) {
	// Arrange
	mockClient := &mockGitHubClient{
		GetRepositoryFunc: func(ctx context.Context, owner, name string) (interfaces.IRepository, error) {
			return &mockRepository{
				owner:               owner,
				name:                name,
				defaultBranch:       "main",
				deleteBranchOnMerge: false, // Disabled
			}, nil
		},
	}
	mockWriter := &mockOutputWriter{}
	service := config.NewConfigService(mockClient, mockWriter)
	ctx := context.Background()

	// Act
	result, err := service.CheckStatus(ctx, "octocat", "hello-world")

	// Assert - no error
	if err != nil {
		t.Fatalf("CheckStatus() error = %v, expected nil", err)
	}

	// Assert - reflects disabled state
	if result.IsNowEnabled() {
		t.Error("IsNowEnabled() = true, expected false (feature is disabled)")
	}
}

// TestCheckStatusDoesNotModifyRepository verifies CheckStatus is read-only.
//
// The implementation should:
// - NOT call UpdateRepository
func TestCheckStatusDoesNotModifyRepository(t *testing.T) {
	// Arrange
	mockClient := &mockGitHubClient{
		GetRepositoryFunc: func(ctx context.Context, owner, name string) (interfaces.IRepository, error) {
			return &mockRepository{
				owner:               owner,
				name:                name,
				defaultBranch:       "main",
				deleteBranchOnMerge: false,
			}, nil
		},
		UpdateRepositoryFunc: func(ctx context.Context, owner, name string, settings interfaces.IRepositorySettings) error {
			t.Error("UpdateRepository should NOT be called in CheckStatus")
			return nil
		},
	}
	mockWriter := &mockOutputWriter{}
	service := config.NewConfigService(mockClient, mockWriter)
	ctx := context.Background()

	// Act
	_, err := service.CheckStatus(ctx, "octocat", "hello-world")

	// Assert
	if err != nil {
		t.Fatalf("CheckStatus() error = %v", err)
	}

	// Assert - UpdateRepository NOT called
	if len(mockClient.UpdateRepositoryCalls) != 0 {
		t.Errorf("UpdateRepository called %d times, expected 0 (CheckStatus is read-only)", len(mockClient.UpdateRepositoryCalls))
	}
}

// TestCheckStatusReturnsRepositoryInfo verifies CheckStatus returns repo info.
//
// The implementation should:
// - Return default branch and full repository name
func TestCheckStatusReturnsRepositoryInfo(t *testing.T) {
	// Arrange
	mockClient := &mockGitHubClient{
		GetRepositoryFunc: func(ctx context.Context, owner, name string) (interfaces.IRepository, error) {
			return &mockRepository{
				owner:               "github",
				name:                "docs",
				defaultBranch:       "master",
				deleteBranchOnMerge: true,
			}, nil
		},
	}
	mockWriter := &mockOutputWriter{}
	service := config.NewConfigService(mockClient, mockWriter)
	ctx := context.Background()

	// Act
	result, err := service.CheckStatus(ctx, "github", "docs")

	// Assert
	if err != nil {
		t.Fatalf("CheckStatus() error = %v", err)
	}

	// Assert - repository info
	if result.GetDefaultBranch() != "master" {
		t.Errorf("GetDefaultBranch() = %q, expected %q", result.GetDefaultBranch(), "master")
	}
	if result.GetRepositoryFullName() != "github/docs" {
		t.Errorf("GetRepositoryFullName() = %q, expected %q", result.GetRepositoryFullName(), "github/docs")
	}
}

// =============================================================================
// Verbose Logging Tests
// =============================================================================

// TestConfigureLogsVerboseFetchingRepository verifies verbose logging.
//
// The implementation should:
// - Call writer.Verbose("Fetching repository information") when fetching
func TestConfigureLogsVerboseFetchingRepository(t *testing.T) {
	// Arrange
	fetchCount := 0
	mockClient := &mockGitHubClient{
		GetRepositoryFunc: func(ctx context.Context, owner, name string) (interfaces.IRepository, error) {
			fetchCount++
			return &mockRepository{
				owner:               owner,
				name:                name,
				defaultBranch:       "main",
				deleteBranchOnMerge: fetchCount > 1,
			}, nil
		},
		UpdateRepositoryFunc: func(ctx context.Context, owner, name string, settings interfaces.IRepositorySettings) error {
			return nil
		},
	}
	mockWriter := &mockOutputWriter{}
	service := config.NewConfigService(mockClient, mockWriter)
	ctx := context.Background()

	// Act
	_, err := service.Configure(ctx, "octocat", "hello-world", false)

	// Assert
	if err != nil {
		t.Fatalf("Configure() error = %v", err)
	}

	// Assert - Verbose called with "Fetching repository information"
	found := false
	for _, msg := range mockWriter.VerboseCalls {
		if msg == "Fetching repository information" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected Verbose('Fetching repository information'), got: %v", mockWriter.VerboseCalls)
	}
}

// TestConfigureLogsVerboseUpdatingRepository verifies update logging.
//
// The implementation should:
// - Call writer.Verbose("Updating repository settings") when updating
func TestConfigureLogsVerboseUpdatingRepository(t *testing.T) {
	// Arrange
	fetchCount := 0
	mockClient := &mockGitHubClient{
		GetRepositoryFunc: func(ctx context.Context, owner, name string) (interfaces.IRepository, error) {
			fetchCount++
			return &mockRepository{
				owner:               owner,
				name:                name,
				defaultBranch:       "main",
				deleteBranchOnMerge: fetchCount > 1,
			}, nil
		},
		UpdateRepositoryFunc: func(ctx context.Context, owner, name string, settings interfaces.IRepositorySettings) error {
			return nil
		},
	}
	mockWriter := &mockOutputWriter{}
	service := config.NewConfigService(mockClient, mockWriter)
	ctx := context.Background()

	// Act
	_, err := service.Configure(ctx, "octocat", "hello-world", false)

	// Assert
	if err != nil {
		t.Fatalf("Configure() error = %v", err)
	}

	// Assert - Verbose called with "Updating repository settings"
	found := false
	for _, msg := range mockWriter.VerboseCalls {
		if msg == "Updating repository settings" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected Verbose('Updating repository settings'), got: %v", mockWriter.VerboseCalls)
	}
}

// TestConfigureLogsVerboseVerifyingSettings verifies verification logging.
//
// The implementation should:
// - Call writer.Verbose("Verifying settings applied") after update
func TestConfigureLogsVerboseVerifyingSettings(t *testing.T) {
	// Arrange
	fetchCount := 0
	mockClient := &mockGitHubClient{
		GetRepositoryFunc: func(ctx context.Context, owner, name string) (interfaces.IRepository, error) {
			fetchCount++
			return &mockRepository{
				owner:               owner,
				name:                name,
				defaultBranch:       "main",
				deleteBranchOnMerge: fetchCount > 1,
			}, nil
		},
		UpdateRepositoryFunc: func(ctx context.Context, owner, name string, settings interfaces.IRepositorySettings) error {
			return nil
		},
	}
	mockWriter := &mockOutputWriter{}
	service := config.NewConfigService(mockClient, mockWriter)
	ctx := context.Background()

	// Act
	_, err := service.Configure(ctx, "octocat", "hello-world", false)

	// Assert
	if err != nil {
		t.Fatalf("Configure() error = %v", err)
	}

	// Assert - Verbose called with "Verifying settings applied"
	found := false
	for _, msg := range mockWriter.VerboseCalls {
		if msg == "Verifying settings applied" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected Verbose('Verifying settings applied'), got: %v", mockWriter.VerboseCalls)
	}
}

// TestConfigureNoUpdateLoggingWhenAlreadyEnabled verifies no update log when enabled.
//
// The implementation should:
// - NOT log "Updating repository settings" when already enabled
func TestConfigureNoUpdateLoggingWhenAlreadyEnabled(t *testing.T) {
	// Arrange
	mockClient := &mockGitHubClient{
		GetRepositoryFunc: func(ctx context.Context, owner, name string) (interfaces.IRepository, error) {
			return &mockRepository{
				owner:               owner,
				name:                name,
				defaultBranch:       "main",
				deleteBranchOnMerge: true, // Already enabled
			}, nil
		},
	}
	mockWriter := &mockOutputWriter{}
	service := config.NewConfigService(mockClient, mockWriter)
	ctx := context.Background()

	// Act
	_, err := service.Configure(ctx, "octocat", "hello-world", false)

	// Assert
	if err != nil {
		t.Fatalf("Configure() error = %v", err)
	}

	// Assert - "Updating repository settings" should NOT be logged
	for _, msg := range mockWriter.VerboseCalls {
		if msg == "Updating repository settings" {
			t.Error("Should NOT log 'Updating repository settings' when already enabled")
		}
	}
}

// TestCheckStatusLogsVerboseFetchingRepository verifies CheckStatus logging.
//
// The implementation should:
// - Call writer.Verbose("Fetching repository information") when checking status
func TestCheckStatusLogsVerboseFetchingRepository(t *testing.T) {
	// Arrange
	mockClient := &mockGitHubClient{
		GetRepositoryFunc: func(ctx context.Context, owner, name string) (interfaces.IRepository, error) {
			return &mockRepository{
				owner:               owner,
				name:                name,
				defaultBranch:       "main",
				deleteBranchOnMerge: true,
			}, nil
		},
	}
	mockWriter := &mockOutputWriter{}
	service := config.NewConfigService(mockClient, mockWriter)
	ctx := context.Background()

	// Act
	_, err := service.CheckStatus(ctx, "octocat", "hello-world")

	// Assert
	if err != nil {
		t.Fatalf("CheckStatus() error = %v", err)
	}

	// Assert - Verbose called with "Fetching repository information"
	found := false
	for _, msg := range mockWriter.VerboseCalls {
		if msg == "Fetching repository information" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected Verbose('Fetching repository information'), got: %v", mockWriter.VerboseCalls)
	}
}

// =============================================================================
// Error Propagation Tests
// =============================================================================

// TestConfigurePropagatesGetRepositoryError verifies error propagation from GetRepository.
//
// The implementation should:
// - Propagate errors from client.GetRepository
func TestConfigurePropagatesGetRepositoryError(t *testing.T) {
	// Arrange
	expectedError := errors.New("repository not found")
	mockClient := &mockGitHubClient{
		GetRepositoryFunc: func(ctx context.Context, owner, name string) (interfaces.IRepository, error) {
			return nil, expectedError
		},
	}
	mockWriter := &mockOutputWriter{}
	service := config.NewConfigService(mockClient, mockWriter)
	ctx := context.Background()

	// Act
	result, err := service.Configure(ctx, "octocat", "nonexistent", false)

	// Assert
	if err == nil {
		t.Fatal("Configure() should return error when GetRepository fails")
	}
	if result != nil {
		t.Error("Configure() should return nil result when GetRepository fails")
	}
}

// TestConfigurePropagatesUpdateRepositoryError verifies error propagation from UpdateRepository.
//
// The implementation should:
// - Propagate errors from client.UpdateRepository
func TestConfigurePropagatesUpdateRepositoryError(t *testing.T) {
	// Arrange
	expectedError := errors.New("insufficient permissions")
	mockClient := &mockGitHubClient{
		GetRepositoryFunc: func(ctx context.Context, owner, name string) (interfaces.IRepository, error) {
			return &mockRepository{
				owner:               owner,
				name:                name,
				defaultBranch:       "main",
				deleteBranchOnMerge: false,
			}, nil
		},
		UpdateRepositoryFunc: func(ctx context.Context, owner, name string, settings interfaces.IRepositorySettings) error {
			return expectedError
		},
	}
	mockWriter := &mockOutputWriter{}
	service := config.NewConfigService(mockClient, mockWriter)
	ctx := context.Background()

	// Act
	result, err := service.Configure(ctx, "octocat", "hello-world", false)

	// Assert
	if err == nil {
		t.Fatal("Configure() should return error when UpdateRepository fails")
	}
	if result != nil {
		t.Error("Configure() should return nil result when UpdateRepository fails")
	}
}

// TestConfigurePropagatesVerificationError verifies error propagation from verification fetch.
//
// The implementation should:
// - Propagate errors from verification GetRepository call
func TestConfigurePropagatesVerificationError(t *testing.T) {
	// Arrange
	fetchCount := 0
	mockClient := &mockGitHubClient{
		GetRepositoryFunc: func(ctx context.Context, owner, name string) (interfaces.IRepository, error) {
			fetchCount++
			if fetchCount == 1 {
				return &mockRepository{
					owner:               owner,
					name:                name,
					defaultBranch:       "main",
					deleteBranchOnMerge: false,
				}, nil
			}
			// Second call fails (verification)
			return nil, errors.New("network error during verification")
		},
		UpdateRepositoryFunc: func(ctx context.Context, owner, name string, settings interfaces.IRepositorySettings) error {
			return nil
		},
	}
	mockWriter := &mockOutputWriter{}
	service := config.NewConfigService(mockClient, mockWriter)
	ctx := context.Background()

	// Act
	result, err := service.Configure(ctx, "octocat", "hello-world", false)

	// Assert
	if err == nil {
		t.Fatal("Configure() should return error when verification fails")
	}
	if result != nil {
		t.Error("Configure() should return nil result when verification fails")
	}
}

// TestCheckStatusPropagatesGetRepositoryError verifies CheckStatus error propagation.
//
// The implementation should:
// - Propagate errors from client.GetRepository in CheckStatus
func TestCheckStatusPropagatesGetRepositoryError(t *testing.T) {
	// Arrange
	expectedError := errors.New("repository not found")
	mockClient := &mockGitHubClient{
		GetRepositoryFunc: func(ctx context.Context, owner, name string) (interfaces.IRepository, error) {
			return nil, expectedError
		},
	}
	mockWriter := &mockOutputWriter{}
	service := config.NewConfigService(mockClient, mockWriter)
	ctx := context.Background()

	// Act
	result, err := service.CheckStatus(ctx, "octocat", "nonexistent")

	// Assert
	if err == nil {
		t.Fatal("CheckStatus() should return error when GetRepository fails")
	}
	if result != nil {
		t.Error("CheckStatus() should return nil result when GetRepository fails")
	}
}

// =============================================================================
// Table-Driven Tests
// =============================================================================

// TestConfigureScenarios tests various configuration scenarios.
func TestConfigureScenarios(t *testing.T) {
	tests := []struct {
		name                    string
		initialEnabled          bool
		dryRun                  bool
		updateError             error
		expectedWasAlready      bool
		expectedNowEnabled      bool
		expectedUpdateCalls     int
		expectedGetRepoCalls    int
		expectError             bool
	}{
		{
			name:                    "disabled -> enabled",
			initialEnabled:          false,
			dryRun:                  false,
			updateError:             nil,
			expectedWasAlready:      false,
			expectedNowEnabled:      true,
			expectedUpdateCalls:     1,
			expectedGetRepoCalls:    2, // fetch + verify
			expectError:             false,
		},
		{
			name:                    "already enabled",
			initialEnabled:          true,
			dryRun:                  false,
			updateError:             nil,
			expectedWasAlready:      true,
			expectedNowEnabled:      true,
			expectedUpdateCalls:     0,
			expectedGetRepoCalls:    1, // fetch only
			expectError:             false,
		},
		{
			name:                    "dry run when disabled",
			initialEnabled:          false,
			dryRun:                  true,
			updateError:             nil,
			expectedWasAlready:      false,
			expectedNowEnabled:      false,
			expectedUpdateCalls:     0,
			expectedGetRepoCalls:    1, // fetch only
			expectError:             false,
		},
		{
			name:                    "dry run when already enabled",
			initialEnabled:          true,
			dryRun:                  true,
			updateError:             nil,
			expectedWasAlready:      true,
			expectedNowEnabled:      true,
			expectedUpdateCalls:     0,
			expectedGetRepoCalls:    1, // fetch only
			expectError:             false,
		},
		{
			name:                    "update fails",
			initialEnabled:          false,
			dryRun:                  false,
			updateError:             errors.New("update failed"),
			expectedWasAlready:      false,
			expectedNowEnabled:      false,
			expectedUpdateCalls:     1,
			expectedGetRepoCalls:    1, // fetch only (no verify after error)
			expectError:             true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			fetchCount := 0
			mockClient := &mockGitHubClient{
				GetRepositoryFunc: func(ctx context.Context, owner, name string) (interfaces.IRepository, error) {
					fetchCount++
					enabled := tt.initialEnabled
					if !tt.initialEnabled && fetchCount > 1 && tt.updateError == nil {
						enabled = true // After successful update
					}
					return &mockRepository{
						owner:               owner,
						name:                name,
						defaultBranch:       "main",
						deleteBranchOnMerge: enabled,
					}, nil
				},
				UpdateRepositoryFunc: func(ctx context.Context, owner, name string, settings interfaces.IRepositorySettings) error {
					return tt.updateError
				},
			}
			mockWriter := &mockOutputWriter{}
			service := config.NewConfigService(mockClient, mockWriter)
			ctx := context.Background()

			// Act
			result, err := service.Configure(ctx, "owner", "repo", tt.dryRun)

			// Assert - error
			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Assert - result
			if result == nil {
				t.Fatal("Expected non-nil result")
			}
			if result.WasAlreadyEnabled() != tt.expectedWasAlready {
				t.Errorf("WasAlreadyEnabled() = %v, expected %v", result.WasAlreadyEnabled(), tt.expectedWasAlready)
			}
			if result.IsNowEnabled() != tt.expectedNowEnabled {
				t.Errorf("IsNowEnabled() = %v, expected %v", result.IsNowEnabled(), tt.expectedNowEnabled)
			}

			// Assert - call counts
			if len(mockClient.UpdateRepositoryCalls) != tt.expectedUpdateCalls {
				t.Errorf("UpdateRepository called %d times, expected %d", len(mockClient.UpdateRepositoryCalls), tt.expectedUpdateCalls)
			}
			if len(mockClient.GetRepositoryCalls) != tt.expectedGetRepoCalls {
				t.Errorf("GetRepository called %d times, expected %d", len(mockClient.GetRepositoryCalls), tt.expectedGetRepoCalls)
			}
		})
	}
}

// TestCheckStatusScenarios tests various check status scenarios.
func TestCheckStatusScenarios(t *testing.T) {
	tests := []struct {
		name               string
		enabled            bool
		defaultBranch      string
		repoFullName       string
		fetchError         error
		expectedNowEnabled bool
		expectError        bool
	}{
		{
			name:               "enabled repository",
			enabled:            true,
			defaultBranch:      "main",
			repoFullName:       "octocat/hello-world",
			fetchError:         nil,
			expectedNowEnabled: true,
			expectError:        false,
		},
		{
			name:               "disabled repository",
			enabled:            false,
			defaultBranch:      "master",
			repoFullName:       "github/docs",
			fetchError:         nil,
			expectedNowEnabled: false,
			expectError:        false,
		},
		{
			name:               "fetch fails",
			enabled:            false,
			defaultBranch:      "",
			repoFullName:       "",
			fetchError:         errors.New("not found"),
			expectedNowEnabled: false,
			expectError:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockClient := &mockGitHubClient{
				GetRepositoryFunc: func(ctx context.Context, owner, name string) (interfaces.IRepository, error) {
					if tt.fetchError != nil {
						return nil, tt.fetchError
					}
					return &mockRepository{
						owner:               owner,
						name:                name,
						defaultBranch:       tt.defaultBranch,
						deleteBranchOnMerge: tt.enabled,
					}, nil
				},
			}
			mockWriter := &mockOutputWriter{}
			service := config.NewConfigService(mockClient, mockWriter)
			ctx := context.Background()

			// Act
			result, err := service.CheckStatus(ctx, "owner", "repo")

			// Assert - error
			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Assert - result
			if result == nil {
				t.Fatal("Expected non-nil result")
			}
			if result.IsNowEnabled() != tt.expectedNowEnabled {
				t.Errorf("IsNowEnabled() = %v, expected %v", result.IsNowEnabled(), tt.expectedNowEnabled)
			}
		})
	}
}

// =============================================================================
// Context Handling Tests
// =============================================================================

// TestConfigurePassesContextToClient verifies context is passed to client.
//
// The implementation should:
// - Pass context to all client method calls
func TestConfigurePassesContextToClient(t *testing.T) {
	// Arrange
	type contextKey string
	ctx := context.WithValue(context.Background(), contextKey("key"), "value")
	var receivedCtx context.Context

	mockClient := &mockGitHubClient{
		GetRepositoryFunc: func(ctx context.Context, owner, name string) (interfaces.IRepository, error) {
			receivedCtx = ctx
			return &mockRepository{
				owner:               owner,
				name:                name,
				defaultBranch:       "main",
				deleteBranchOnMerge: true,
			}, nil
		},
	}
	mockWriter := &mockOutputWriter{}
	service := config.NewConfigService(mockClient, mockWriter)

	// Act
	_, err := service.Configure(ctx, "owner", "repo", false)

	// Assert
	if err != nil {
		t.Fatalf("Configure() error = %v", err)
	}
	if receivedCtx == nil {
		t.Error("Context was not passed to GetRepository")
	}
	if receivedCtx.Value(contextKey("key")) != "value" {
		t.Error("Context value was not preserved")
	}
}

// TestCheckStatusPassesContextToClient verifies context is passed to client.
//
// The implementation should:
// - Pass context to GetRepository call
func TestCheckStatusPassesContextToClient(t *testing.T) {
	// Arrange
	type contextKey string
	ctx := context.WithValue(context.Background(), contextKey("key"), "value")
	var receivedCtx context.Context

	mockClient := &mockGitHubClient{
		GetRepositoryFunc: func(ctx context.Context, owner, name string) (interfaces.IRepository, error) {
			receivedCtx = ctx
			return &mockRepository{
				owner:               owner,
				name:                name,
				defaultBranch:       "main",
				deleteBranchOnMerge: true,
			}, nil
		},
	}
	mockWriter := &mockOutputWriter{}
	service := config.NewConfigService(mockClient, mockWriter)

	// Act
	_, err := service.CheckStatus(ctx, "owner", "repo")

	// Assert
	if err != nil {
		t.Fatalf("CheckStatus() error = %v", err)
	}
	if receivedCtx == nil {
		t.Error("Context was not passed to GetRepository")
	}
	if receivedCtx.Value(contextKey("key")) != "value" {
		t.Error("Context value was not preserved")
	}
}
