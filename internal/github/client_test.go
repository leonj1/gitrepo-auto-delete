// Package github_test provides tests for the GitHubClient.
//
// These tests verify the GitHubClient's HTTP operations and error handling
// according to the Gherkin scenarios:
//
// Scenario: Repository not found -> 404 -> error code 5
// Scenario: No access to private repository -> 404 -> error code 5
// Scenario: Insufficient permissions -> 403 -> error code 4
// Scenario: API rate limit exceeded -> 403 with rate limit headers -> error code 6
// Scenario: Network connection failure -> error code 1
// Scenario: GitHub API server error -> retry 3 times with backoff -> error code 1
//
// These tests are designed to FAIL until the implementations are properly created
// by the coder agent.
package github_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	apperrors "github.com/josejulio/ghautodelete/internal/errors"
	"github.com/josejulio/ghautodelete/internal/github"
	"github.com/josejulio/ghautodelete/pkg/interfaces"
)

// =============================================================================
// Interface Satisfaction Tests
// =============================================================================

// TestGitHubClientImplementsIGitHubClient verifies GitHubClient implements IGitHubClient.
//
// The implementation should:
// - Define GitHubClient struct in internal/github/client.go
// - Implement all IGitHubClient methods
func TestGitHubClientImplementsIGitHubClient(t *testing.T) {
	// Arrange - create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Act
	client := github.NewGitHubClient(server.Client(), server.URL, "test-token")

	// Assert - compile-time interface satisfaction check
	var _ interfaces.IGitHubClient = client
	if client == nil {
		t.Error("NewGitHubClient should return a non-nil client")
	}
}

// TestNewGitHubClientReturnsNonNil verifies the constructor returns a valid client.
func TestNewGitHubClientReturnsNonNil(t *testing.T) {
	// Arrange
	httpClient := &http.Client{}
	baseURL := "https://api.github.com"
	token := "test-token"

	// Act
	client := github.NewGitHubClient(httpClient, baseURL, token)

	// Assert
	if client == nil {
		t.Error("NewGitHubClient should return non-nil client")
	}
}

// =============================================================================
// GetRepository Success Tests
// =============================================================================

// TestGetRepositorySuccess verifies successful repository retrieval.
//
// The implementation should:
// - Send GET request to /repos/{owner}/{repo}
// - Parse JSON response into IRepository
// - Include Authorization header with token
func TestGetRepositorySuccess(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/repos/octocat/hello-world" {
			t.Errorf("Expected path /repos/octocat/hello-world, got %s", r.URL.Path)
		}
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer test-token" && authHeader != "token test-token" {
			t.Errorf("Expected Authorization header with token, got %s", authHeader)
		}

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"owner": {"login": "octocat"},
			"name": "hello-world",
			"default_branch": "main",
			"delete_branch_on_merge": true
		}`))
	}))
	defer server.Close()

	client := github.NewGitHubClient(server.Client(), server.URL, "test-token")
	ctx := context.Background()

	// Act
	repo, err := client.GetRepository(ctx, "octocat", "hello-world")

	// Assert
	if err != nil {
		t.Fatalf("GetRepository() error = %v, expected nil", err)
	}
	if repo == nil {
		t.Fatal("GetRepository() returned nil repository")
	}
	if repo.GetOwner() != "octocat" {
		t.Errorf("Owner = %q, expected %q", repo.GetOwner(), "octocat")
	}
	if repo.GetName() != "hello-world" {
		t.Errorf("Name = %q, expected %q", repo.GetName(), "hello-world")
	}
	if repo.GetDefaultBranch() != "main" {
		t.Errorf("DefaultBranch = %q, expected %q", repo.GetDefaultBranch(), "main")
	}
	if !repo.GetDeleteBranchOnMerge() {
		t.Error("DeleteBranchOnMerge = false, expected true")
	}
	if repo.GetFullName() != "octocat/hello-world" {
		t.Errorf("FullName = %q, expected %q", repo.GetFullName(), "octocat/hello-world")
	}
}

// TestGetRepositoryWithDeleteBranchOnMergeFalse verifies parsing when feature is disabled.
func TestGetRepositoryWithDeleteBranchOnMergeFalse(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"owner": {"login": "github"},
			"name": "docs",
			"default_branch": "master",
			"delete_branch_on_merge": false
		}`))
	}))
	defer server.Close()

	client := github.NewGitHubClient(server.Client(), server.URL, "test-token")
	ctx := context.Background()

	// Act
	repo, err := client.GetRepository(ctx, "github", "docs")

	// Assert
	if err != nil {
		t.Fatalf("GetRepository() error = %v, expected nil", err)
	}
	if repo.GetDeleteBranchOnMerge() {
		t.Error("DeleteBranchOnMerge = true, expected false")
	}
}

// =============================================================================
// GetRepository Error Tests - Gherkin Scenarios
// =============================================================================

// TestGetRepositoryNotFound verifies 404 error handling.
//
// Gherkin: Scenario: Repository not found -> 404 -> error code 5
//
// The implementation should:
// - Return error with ErrRepositoryNotFound code (5) for 404 responses
func TestGetRepositoryNotFound(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message": "Not Found"}`))
	}))
	defer server.Close()

	client := github.NewGitHubClient(server.Client(), server.URL, "test-token")
	ctx := context.Background()

	// Act
	repo, err := client.GetRepository(ctx, "octocat", "nonexistent")

	// Assert
	if repo != nil {
		t.Error("GetRepository() should return nil repository for 404")
	}
	if err == nil {
		t.Fatal("GetRepository() should return error for 404")
	}

	var appErr *apperrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Error should be *AppError, got %T", err)
	}
	if appErr.Code != apperrors.ErrRepositoryNotFound {
		t.Errorf("Error code = %v, expected %v (ErrRepositoryNotFound)", appErr.Code, apperrors.ErrRepositoryNotFound)
	}

	exitCode := apperrors.GetExitCode(err)
	if exitCode != 5 {
		t.Errorf("Exit code = %d, expected 5", exitCode)
	}
}

// TestGetRepositoryPrivateNoAccess verifies 404 for private repos without access.
//
// Gherkin: Scenario: No access to private repository -> 404 -> error code 5
//
// Note: GitHub returns 404 (not 403) for private repos the user can't access,
// to avoid leaking existence information.
func TestGetRepositoryPrivateNoAccess(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// GitHub returns 404 for private repos without access (not 403)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message": "Not Found"}`))
	}))
	defer server.Close()

	client := github.NewGitHubClient(server.Client(), server.URL, "test-token")
	ctx := context.Background()

	// Act
	repo, err := client.GetRepository(ctx, "private-org", "private-repo")

	// Assert
	if repo != nil {
		t.Error("GetRepository() should return nil repository for private repo without access")
	}
	if err == nil {
		t.Fatal("GetRepository() should return error for private repo without access")
	}

	var appErr *apperrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Error should be *AppError, got %T", err)
	}
	if appErr.Code != apperrors.ErrRepositoryNotFound {
		t.Errorf("Error code = %v, expected %v (ErrRepositoryNotFound)", appErr.Code, apperrors.ErrRepositoryNotFound)
	}

	exitCode := apperrors.GetExitCode(err)
	if exitCode != 5 {
		t.Errorf("Exit code = %d, expected 5", exitCode)
	}
}

// TestGetRepositoryInsufficientPermissions verifies 403 error handling.
//
// Gherkin: Scenario: Insufficient permissions -> 403 -> error code 4
//
// The implementation should:
// - Return error with ErrInsufficientPerms code (4) for 403 responses without rate limit
func TestGetRepositoryInsufficientPermissions(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// No rate limit headers = permission issue
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message": "Must have admin rights to Repository."}`))
	}))
	defer server.Close()

	client := github.NewGitHubClient(server.Client(), server.URL, "test-token")
	ctx := context.Background()

	// Act
	repo, err := client.GetRepository(ctx, "octocat", "restricted")

	// Assert
	if repo != nil {
		t.Error("GetRepository() should return nil repository for 403")
	}
	if err == nil {
		t.Fatal("GetRepository() should return error for 403")
	}

	var appErr *apperrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Error should be *AppError, got %T", err)
	}
	if appErr.Code != apperrors.ErrInsufficientPerms {
		t.Errorf("Error code = %v, expected %v (ErrInsufficientPerms)", appErr.Code, apperrors.ErrInsufficientPerms)
	}

	exitCode := apperrors.GetExitCode(err)
	if exitCode != 4 {
		t.Errorf("Exit code = %d, expected 4", exitCode)
	}
}

// TestGetRepositoryRateLimitExceeded verifies 403 with rate limit headers.
//
// Gherkin: Scenario: API rate limit exceeded -> 403 with rate limit headers -> error code 6
//
// The implementation should:
// - Check X-RateLimit-Remaining header
// - Return error with ErrAPIRateLimited code (6) when rate limited
// - Parse X-RateLimit-Reset header for reset time
func TestGetRepositoryRateLimitExceeded(t *testing.T) {
	// Arrange
	resetTime := time.Now().Add(1 * time.Hour).Unix()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Rate limit headers indicating exhausted limit
		w.Header().Set("X-RateLimit-Limit", "5000")
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.Header().Set("X-RateLimit-Reset", time.Unix(resetTime, 0).Format("2006-01-02T15:04:05Z"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message": "API rate limit exceeded"}`))
	}))
	defer server.Close()

	client := github.NewGitHubClient(server.Client(), server.URL, "test-token")
	ctx := context.Background()

	// Act
	repo, err := client.GetRepository(ctx, "octocat", "hello-world")

	// Assert
	if repo != nil {
		t.Error("GetRepository() should return nil repository when rate limited")
	}
	if err == nil {
		t.Fatal("GetRepository() should return error when rate limited")
	}

	var appErr *apperrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Error should be *AppError, got %T", err)
	}
	if appErr.Code != apperrors.ErrAPIRateLimited {
		t.Errorf("Error code = %v, expected %v (ErrAPIRateLimited)", appErr.Code, apperrors.ErrAPIRateLimited)
	}

	exitCode := apperrors.GetExitCode(err)
	if exitCode != 6 {
		t.Errorf("Exit code = %d, expected 6", exitCode)
	}
}

// TestGetRepositoryRateLimitWithNumericReset verifies rate limit reset time parsing.
//
// The implementation should handle X-RateLimit-Reset as Unix timestamp.
func TestGetRepositoryRateLimitWithNumericReset(t *testing.T) {
	// Arrange
	resetUnix := time.Now().Add(30 * time.Minute).Unix()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Limit", "5000")
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.Header().Set("X-RateLimit-Reset", string(rune(resetUnix)))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message": "API rate limit exceeded"}`))
	}))
	defer server.Close()

	client := github.NewGitHubClient(server.Client(), server.URL, "test-token")
	ctx := context.Background()

	// Act
	_, err := client.GetRepository(ctx, "octocat", "hello-world")

	// Assert
	if err == nil {
		t.Fatal("GetRepository() should return error when rate limited")
	}

	var appErr *apperrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Error should be *AppError, got %T", err)
	}
	if appErr.Code != apperrors.ErrAPIRateLimited {
		t.Errorf("Error code = %v, expected %v (ErrAPIRateLimited)", appErr.Code, apperrors.ErrAPIRateLimited)
	}
}

// TestGetRepositoryAuthenticationFailed verifies 401 error handling.
//
// The implementation should:
// - Return error with ErrAuthenticationFailed code (3) for 401 responses
func TestGetRepositoryAuthenticationFailed(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message": "Bad credentials"}`))
	}))
	defer server.Close()

	client := github.NewGitHubClient(server.Client(), server.URL, "invalid-token")
	ctx := context.Background()

	// Act
	repo, err := client.GetRepository(ctx, "octocat", "hello-world")

	// Assert
	if repo != nil {
		t.Error("GetRepository() should return nil repository for 401")
	}
	if err == nil {
		t.Fatal("GetRepository() should return error for 401")
	}

	var appErr *apperrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Error should be *AppError, got %T", err)
	}
	if appErr.Code != apperrors.ErrAuthenticationFailed {
		t.Errorf("Error code = %v, expected %v (ErrAuthenticationFailed)", appErr.Code, apperrors.ErrAuthenticationFailed)
	}

	exitCode := apperrors.GetExitCode(err)
	if exitCode != 3 {
		t.Errorf("Exit code = %d, expected 3", exitCode)
	}
}

// TestGetRepositoryServerErrorWithRetry verifies 5xx error handling with retry.
//
// Gherkin: Scenario: GitHub API server error -> retry 3 times with backoff -> error code 1
//
// The implementation should:
// - Retry up to 3 times on 5xx errors
// - Return error with ErrGeneral code (1) after all retries exhausted
func TestGetRepositoryServerErrorWithRetry(t *testing.T) {
	// Arrange
	var requestCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message": "Internal Server Error"}`))
	}))
	defer server.Close()

	client := github.NewGitHubClient(server.Client(), server.URL, "test-token")
	ctx := context.Background()

	// Act
	repo, err := client.GetRepository(ctx, "octocat", "hello-world")

	// Assert
	if repo != nil {
		t.Error("GetRepository() should return nil repository for server error")
	}
	if err == nil {
		t.Fatal("GetRepository() should return error for server error")
	}

	// Verify retry occurred (initial + 3 retries = 4 total, or initial + 2 retries = 3 total)
	count := atomic.LoadInt32(&requestCount)
	if count < 3 {
		t.Errorf("Expected at least 3 requests (with retries), got %d", count)
	}

	var appErr *apperrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Error should be *AppError, got %T", err)
	}
	if appErr.Code != apperrors.ErrGeneral {
		t.Errorf("Error code = %v, expected %v (ErrGeneral)", appErr.Code, apperrors.ErrGeneral)
	}

	exitCode := apperrors.GetExitCode(err)
	if exitCode != 1 {
		t.Errorf("Exit code = %d, expected 1", exitCode)
	}
}

// TestGetRepositoryServerErrorRecovery verifies successful retry after transient failure.
//
// The implementation should:
// - Succeed if server recovers during retry attempts
func TestGetRepositoryServerErrorRecovery(t *testing.T) {
	// Arrange
	var requestCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&requestCount, 1)
		if count < 3 {
			// First two requests fail
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"message": "Service Unavailable"}`))
			return
		}
		// Third request succeeds
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"owner": {"login": "octocat"},
			"name": "hello-world",
			"default_branch": "main",
			"delete_branch_on_merge": true
		}`))
	}))
	defer server.Close()

	client := github.NewGitHubClient(server.Client(), server.URL, "test-token")
	ctx := context.Background()

	// Act
	repo, err := client.GetRepository(ctx, "octocat", "hello-world")

	// Assert
	if err != nil {
		t.Fatalf("GetRepository() should succeed after recovery, got error: %v", err)
	}
	if repo == nil {
		t.Fatal("GetRepository() should return repository after recovery")
	}
	if repo.GetOwner() != "octocat" {
		t.Errorf("Owner = %q, expected %q", repo.GetOwner(), "octocat")
	}
}

// TestGetRepositoryNoRetryOn4xx verifies no retry on 4xx errors.
//
// The implementation should:
// - NOT retry on 4xx errors (except rate limiting which is handled separately)
func TestGetRepositoryNoRetryOn4xx(t *testing.T) {
	// Arrange
	var requestCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message": "Not Found"}`))
	}))
	defer server.Close()

	client := github.NewGitHubClient(server.Client(), server.URL, "test-token")
	ctx := context.Background()

	// Act
	_, err := client.GetRepository(ctx, "octocat", "nonexistent")

	// Assert
	if err == nil {
		t.Fatal("GetRepository() should return error for 404")
	}

	count := atomic.LoadInt32(&requestCount)
	if count != 1 {
		t.Errorf("Expected exactly 1 request (no retry for 4xx), got %d", count)
	}
}

// =============================================================================
// UpdateRepository Tests
// =============================================================================

// TestUpdateRepositorySuccess verifies successful repository update.
//
// The implementation should:
// - Send PATCH request to /repos/{owner}/{repo}
// - Include settings in JSON body
// - Include Authorization header with token
func TestUpdateRepositorySuccess(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != http.MethodPatch {
			t.Errorf("Expected PATCH request, got %s", r.Method)
		}
		if r.URL.Path != "/repos/octocat/hello-world" {
			t.Errorf("Expected path /repos/octocat/hello-world, got %s", r.URL.Path)
		}
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer test-token" && authHeader != "token test-token" {
			t.Errorf("Expected Authorization header with token, got %s", authHeader)
		}

		// Return success
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"owner": {"login": "octocat"},
			"name": "hello-world",
			"default_branch": "main",
			"delete_branch_on_merge": true
		}`))
	}))
	defer server.Close()

	client := github.NewGitHubClient(server.Client(), server.URL, "test-token")
	ctx := context.Background()
	settings := github.NewRepositorySettings(true)

	// Act
	err := client.UpdateRepository(ctx, "octocat", "hello-world", settings)

	// Assert
	if err != nil {
		t.Fatalf("UpdateRepository() error = %v, expected nil", err)
	}
}

// TestUpdateRepositoryNotFound verifies 404 error handling.
//
// Gherkin: Scenario: Repository not found -> 404 -> error code 5
func TestUpdateRepositoryNotFound(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message": "Not Found"}`))
	}))
	defer server.Close()

	client := github.NewGitHubClient(server.Client(), server.URL, "test-token")
	ctx := context.Background()
	settings := github.NewRepositorySettings(true)

	// Act
	err := client.UpdateRepository(ctx, "octocat", "nonexistent", settings)

	// Assert
	if err == nil {
		t.Fatal("UpdateRepository() should return error for 404")
	}

	var appErr *apperrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Error should be *AppError, got %T", err)
	}
	if appErr.Code != apperrors.ErrRepositoryNotFound {
		t.Errorf("Error code = %v, expected %v (ErrRepositoryNotFound)", appErr.Code, apperrors.ErrRepositoryNotFound)
	}

	exitCode := apperrors.GetExitCode(err)
	if exitCode != 5 {
		t.Errorf("Exit code = %d, expected 5", exitCode)
	}
}

// TestUpdateRepositoryInsufficientPermissions verifies 403 error handling.
//
// Gherkin: Scenario: Insufficient permissions -> 403 -> error code 4
func TestUpdateRepositoryInsufficientPermissions(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message": "Must have admin rights to Repository."}`))
	}))
	defer server.Close()

	client := github.NewGitHubClient(server.Client(), server.URL, "test-token")
	ctx := context.Background()
	settings := github.NewRepositorySettings(true)

	// Act
	err := client.UpdateRepository(ctx, "octocat", "restricted", settings)

	// Assert
	if err == nil {
		t.Fatal("UpdateRepository() should return error for 403")
	}

	var appErr *apperrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Error should be *AppError, got %T", err)
	}
	if appErr.Code != apperrors.ErrInsufficientPerms {
		t.Errorf("Error code = %v, expected %v (ErrInsufficientPerms)", appErr.Code, apperrors.ErrInsufficientPerms)
	}

	exitCode := apperrors.GetExitCode(err)
	if exitCode != 4 {
		t.Errorf("Exit code = %d, expected 4", exitCode)
	}
}

// TestUpdateRepositoryRateLimitExceeded verifies rate limit handling.
//
// Gherkin: Scenario: API rate limit exceeded -> 403 with rate limit headers -> error code 6
func TestUpdateRepositoryRateLimitExceeded(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Limit", "5000")
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.Header().Set("X-RateLimit-Reset", time.Now().Add(1*time.Hour).Format("2006-01-02T15:04:05Z"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message": "API rate limit exceeded"}`))
	}))
	defer server.Close()

	client := github.NewGitHubClient(server.Client(), server.URL, "test-token")
	ctx := context.Background()
	settings := github.NewRepositorySettings(true)

	// Act
	err := client.UpdateRepository(ctx, "octocat", "hello-world", settings)

	// Assert
	if err == nil {
		t.Fatal("UpdateRepository() should return error when rate limited")
	}

	var appErr *apperrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Error should be *AppError, got %T", err)
	}
	if appErr.Code != apperrors.ErrAPIRateLimited {
		t.Errorf("Error code = %v, expected %v (ErrAPIRateLimited)", appErr.Code, apperrors.ErrAPIRateLimited)
	}

	exitCode := apperrors.GetExitCode(err)
	if exitCode != 6 {
		t.Errorf("Exit code = %d, expected 6", exitCode)
	}
}

// TestUpdateRepositoryServerErrorWithRetry verifies 5xx error handling with retry.
//
// Gherkin: Scenario: GitHub API server error -> retry 3 times with backoff -> error code 1
func TestUpdateRepositoryServerErrorWithRetry(t *testing.T) {
	// Arrange
	var requestCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"message": "Bad Gateway"}`))
	}))
	defer server.Close()

	client := github.NewGitHubClient(server.Client(), server.URL, "test-token")
	ctx := context.Background()
	settings := github.NewRepositorySettings(true)

	// Act
	err := client.UpdateRepository(ctx, "octocat", "hello-world", settings)

	// Assert
	if err == nil {
		t.Fatal("UpdateRepository() should return error for server error")
	}

	count := atomic.LoadInt32(&requestCount)
	if count < 3 {
		t.Errorf("Expected at least 3 requests (with retries), got %d", count)
	}

	var appErr *apperrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Error should be *AppError, got %T", err)
	}
	if appErr.Code != apperrors.ErrGeneral {
		t.Errorf("Error code = %v, expected %v (ErrGeneral)", appErr.Code, apperrors.ErrGeneral)
	}
}

// =============================================================================
// ValidateToken Tests
// =============================================================================

// TestValidateTokenSuccess verifies successful token validation.
//
// The implementation should:
// - Send GET request to /user
// - Parse scopes from X-OAuth-Scopes header
// - Return ITokenInfo with scopes and username
func TestValidateTokenSuccess(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/user" {
			t.Errorf("Expected path /user, got %s", r.URL.Path)
		}

		// Return mock response with scopes header
		w.Header().Set("X-OAuth-Scopes", "repo, user")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"login": "octocat",
			"id": 1,
			"type": "User"
		}`))
	}))
	defer server.Close()

	client := github.NewGitHubClient(server.Client(), server.URL, "test-token")
	ctx := context.Background()

	// Act
	tokenInfo, err := client.ValidateToken(ctx)

	// Assert
	if err != nil {
		t.Fatalf("ValidateToken() error = %v, expected nil", err)
	}
	if tokenInfo == nil {
		t.Fatal("ValidateToken() returned nil tokenInfo")
	}
	if tokenInfo.GetUsername() != "octocat" {
		t.Errorf("Username = %q, expected %q", tokenInfo.GetUsername(), "octocat")
	}

	scopes := tokenInfo.GetScopes()
	if len(scopes) < 1 {
		t.Error("Expected at least 1 scope")
	}

	if !tokenInfo.HasScope("repo") {
		t.Error("HasScope('repo') = false, expected true")
	}
}

// TestValidateTokenAuthenticationFailed verifies 401 error handling.
//
// The implementation should:
// - Return error with ErrAuthenticationFailed code (3) for 401 responses
func TestValidateTokenAuthenticationFailed(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message": "Bad credentials"}`))
	}))
	defer server.Close()

	client := github.NewGitHubClient(server.Client(), server.URL, "invalid-token")
	ctx := context.Background()

	// Act
	tokenInfo, err := client.ValidateToken(ctx)

	// Assert
	if tokenInfo != nil {
		t.Error("ValidateToken() should return nil tokenInfo for 401")
	}
	if err == nil {
		t.Fatal("ValidateToken() should return error for 401")
	}

	var appErr *apperrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Error should be *AppError, got %T", err)
	}
	if appErr.Code != apperrors.ErrAuthenticationFailed {
		t.Errorf("Error code = %v, expected %v (ErrAuthenticationFailed)", appErr.Code, apperrors.ErrAuthenticationFailed)
	}

	exitCode := apperrors.GetExitCode(err)
	if exitCode != 3 {
		t.Errorf("Exit code = %d, expected 3", exitCode)
	}
}

// TestValidateTokenRateLimitExceeded verifies rate limit handling.
//
// Gherkin: Scenario: API rate limit exceeded -> 403 with rate limit headers -> error code 6
func TestValidateTokenRateLimitExceeded(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Limit", "5000")
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.Header().Set("X-RateLimit-Reset", time.Now().Add(1*time.Hour).Format("2006-01-02T15:04:05Z"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message": "API rate limit exceeded"}`))
	}))
	defer server.Close()

	client := github.NewGitHubClient(server.Client(), server.URL, "test-token")
	ctx := context.Background()

	// Act
	tokenInfo, err := client.ValidateToken(ctx)

	// Assert
	if tokenInfo != nil {
		t.Error("ValidateToken() should return nil tokenInfo when rate limited")
	}
	if err == nil {
		t.Fatal("ValidateToken() should return error when rate limited")
	}

	var appErr *apperrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Error should be *AppError, got %T", err)
	}
	if appErr.Code != apperrors.ErrAPIRateLimited {
		t.Errorf("Error code = %v, expected %v (ErrAPIRateLimited)", appErr.Code, apperrors.ErrAPIRateLimited)
	}
}

// TestValidateTokenServerError verifies 5xx error handling.
//
// Gherkin: Scenario: GitHub API server error -> retry 3 times with backoff -> error code 1
func TestValidateTokenServerError(t *testing.T) {
	// Arrange
	var requestCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"message": "Service Unavailable"}`))
	}))
	defer server.Close()

	client := github.NewGitHubClient(server.Client(), server.URL, "test-token")
	ctx := context.Background()

	// Act
	tokenInfo, err := client.ValidateToken(ctx)

	// Assert
	if tokenInfo != nil {
		t.Error("ValidateToken() should return nil tokenInfo for server error")
	}
	if err == nil {
		t.Fatal("ValidateToken() should return error for server error")
	}

	count := atomic.LoadInt32(&requestCount)
	if count < 3 {
		t.Errorf("Expected at least 3 requests (with retries), got %d", count)
	}

	var appErr *apperrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Error should be *AppError, got %T", err)
	}
	if appErr.Code != apperrors.ErrGeneral {
		t.Errorf("Error code = %v, expected %v (ErrGeneral)", appErr.Code, apperrors.ErrGeneral)
	}
}

// =============================================================================
// Network Error Tests
// =============================================================================

// TestGetRepositoryNetworkError verifies network connection failure handling.
//
// Gherkin: Scenario: Network connection failure -> error code 1
//
// The implementation should:
// - Return error with ErrGeneral code (1) for network errors
func TestGetRepositoryNetworkError(t *testing.T) {
	// Arrange - create client with invalid URL to simulate network error
	httpClient := &http.Client{
		Timeout: 100 * time.Millisecond,
	}
	client := github.NewGitHubClient(httpClient, "http://localhost:1", "test-token")
	ctx := context.Background()

	// Act
	repo, err := client.GetRepository(ctx, "octocat", "hello-world")

	// Assert
	if repo != nil {
		t.Error("GetRepository() should return nil repository for network error")
	}
	if err == nil {
		t.Fatal("GetRepository() should return error for network error")
	}

	var appErr *apperrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Error should be *AppError, got %T", err)
	}
	if appErr.Code != apperrors.ErrGeneral {
		t.Errorf("Error code = %v, expected %v (ErrGeneral)", appErr.Code, apperrors.ErrGeneral)
	}

	exitCode := apperrors.GetExitCode(err)
	if exitCode != 1 {
		t.Errorf("Exit code = %d, expected 1", exitCode)
	}
}

// TestUpdateRepositoryNetworkError verifies network failure during update.
//
// Gherkin: Scenario: Network connection failure -> error code 1
func TestUpdateRepositoryNetworkError(t *testing.T) {
	// Arrange - create client with invalid URL
	httpClient := &http.Client{
		Timeout: 100 * time.Millisecond,
	}
	client := github.NewGitHubClient(httpClient, "http://localhost:1", "test-token")
	ctx := context.Background()
	settings := github.NewRepositorySettings(true)

	// Act
	err := client.UpdateRepository(ctx, "octocat", "hello-world", settings)

	// Assert
	if err == nil {
		t.Fatal("UpdateRepository() should return error for network error")
	}

	var appErr *apperrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Error should be *AppError, got %T", err)
	}
	if appErr.Code != apperrors.ErrGeneral {
		t.Errorf("Error code = %v, expected %v (ErrGeneral)", appErr.Code, apperrors.ErrGeneral)
	}
}

// TestValidateTokenNetworkError verifies network failure during validation.
//
// Gherkin: Scenario: Network connection failure -> error code 1
func TestValidateTokenNetworkError(t *testing.T) {
	// Arrange - create client with invalid URL
	httpClient := &http.Client{
		Timeout: 100 * time.Millisecond,
	}
	client := github.NewGitHubClient(httpClient, "http://localhost:1", "test-token")
	ctx := context.Background()

	// Act
	tokenInfo, err := client.ValidateToken(ctx)

	// Assert
	if tokenInfo != nil {
		t.Error("ValidateToken() should return nil tokenInfo for network error")
	}
	if err == nil {
		t.Fatal("ValidateToken() should return error for network error")
	}

	var appErr *apperrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Error should be *AppError, got %T", err)
	}
	if appErr.Code != apperrors.ErrGeneral {
		t.Errorf("Error code = %v, expected %v (ErrGeneral)", appErr.Code, apperrors.ErrGeneral)
	}
}

// =============================================================================
// Context Cancellation Tests
// =============================================================================

// TestGetRepositoryContextCancellation verifies context cancellation handling.
func TestGetRepositoryContextCancellation(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(1 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := github.NewGitHubClient(server.Client(), server.URL, "test-token")
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Act
	repo, err := client.GetRepository(ctx, "octocat", "hello-world")

	// Assert
	if repo != nil {
		t.Error("GetRepository() should return nil repository for cancelled context")
	}
	if err == nil {
		t.Fatal("GetRepository() should return error for cancelled context")
	}
}

// =============================================================================
// Authorization Header Tests
// =============================================================================

// TestGetRepositoryAuthorizationHeader verifies Authorization header format.
func TestGetRepositoryAuthorizationHeader(t *testing.T) {
	// Arrange
	var receivedAuthHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuthHeader = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"owner": {"login": "octocat"},
			"name": "hello-world",
			"default_branch": "main",
			"delete_branch_on_merge": true
		}`))
	}))
	defer server.Close()

	client := github.NewGitHubClient(server.Client(), server.URL, "ghp_test123token")
	ctx := context.Background()

	// Act
	_, err := client.GetRepository(ctx, "octocat", "hello-world")

	// Assert
	if err != nil {
		t.Fatalf("GetRepository() error = %v", err)
	}

	// Verify Authorization header contains the token
	if receivedAuthHeader == "" {
		t.Error("Authorization header should not be empty")
	}
	if receivedAuthHeader != "Bearer ghp_test123token" && receivedAuthHeader != "token ghp_test123token" {
		t.Errorf("Authorization header = %q, expected 'Bearer ghp_test123token' or 'token ghp_test123token'", receivedAuthHeader)
	}
}

// =============================================================================
// Table-Driven Error Code Tests
// =============================================================================

// TestHTTPStatusToErrorCodeMapping verifies correct error code mapping for all HTTP statuses.
func TestHTTPStatusToErrorCodeMapping(t *testing.T) {
	tests := []struct {
		name             string
		statusCode       int
		rateLimitHeaders bool
		expectedCode     apperrors.ErrorCode
		expectedExitCode int
	}{
		{
			name:             "401 maps to ErrAuthenticationFailed (code 3)",
			statusCode:       http.StatusUnauthorized,
			rateLimitHeaders: false,
			expectedCode:     apperrors.ErrAuthenticationFailed,
			expectedExitCode: 3,
		},
		{
			name:             "403 without rate limit maps to ErrInsufficientPerms (code 4)",
			statusCode:       http.StatusForbidden,
			rateLimitHeaders: false,
			expectedCode:     apperrors.ErrInsufficientPerms,
			expectedExitCode: 4,
		},
		{
			name:             "403 with rate limit maps to ErrAPIRateLimited (code 6)",
			statusCode:       http.StatusForbidden,
			rateLimitHeaders: true,
			expectedCode:     apperrors.ErrAPIRateLimited,
			expectedExitCode: 6,
		},
		{
			name:             "404 maps to ErrRepositoryNotFound (code 5)",
			statusCode:       http.StatusNotFound,
			rateLimitHeaders: false,
			expectedCode:     apperrors.ErrRepositoryNotFound,
			expectedExitCode: 5,
		},
		{
			name:             "500 maps to ErrGeneral (code 1)",
			statusCode:       http.StatusInternalServerError,
			rateLimitHeaders: false,
			expectedCode:     apperrors.ErrGeneral,
			expectedExitCode: 1,
		},
		{
			name:             "502 maps to ErrGeneral (code 1)",
			statusCode:       http.StatusBadGateway,
			rateLimitHeaders: false,
			expectedCode:     apperrors.ErrGeneral,
			expectedExitCode: 1,
		},
		{
			name:             "503 maps to ErrGeneral (code 1)",
			statusCode:       http.StatusServiceUnavailable,
			rateLimitHeaders: false,
			expectedCode:     apperrors.ErrGeneral,
			expectedExitCode: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.rateLimitHeaders {
					w.Header().Set("X-RateLimit-Limit", "5000")
					w.Header().Set("X-RateLimit-Remaining", "0")
					w.Header().Set("X-RateLimit-Reset", time.Now().Add(1*time.Hour).Format("2006-01-02T15:04:05Z"))
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(`{"message": "Error"}`))
			}))
			defer server.Close()

			client := github.NewGitHubClient(server.Client(), server.URL, "test-token")
			ctx := context.Background()

			// Act
			_, err := client.GetRepository(ctx, "owner", "repo")

			// Assert
			if err == nil {
				t.Fatal("Expected error, got nil")
			}

			var appErr *apperrors.AppError
			if !errors.As(err, &appErr) {
				t.Fatalf("Error should be *AppError, got %T", err)
			}
			if appErr.Code != tt.expectedCode {
				t.Errorf("Error code = %v, expected %v", appErr.Code, tt.expectedCode)
			}

			exitCode := apperrors.GetExitCode(err)
			if exitCode != tt.expectedExitCode {
				t.Errorf("Exit code = %d, expected %d", exitCode, tt.expectedExitCode)
			}
		})
	}
}

// =============================================================================
// Retry Behavior Tests
// =============================================================================

// TestRetryBehaviorTable tests retry behavior for different status codes.
func TestRetryBehaviorTable(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		shouldRetry    bool
		expectedCalls  int // Expected number of calls (1 = no retry, >1 = with retry)
		rateLimitExhausted bool
	}{
		{
			name:          "401 should not retry",
			statusCode:    http.StatusUnauthorized,
			shouldRetry:   false,
			expectedCalls: 1,
		},
		{
			name:               "403 without rate limit should not retry",
			statusCode:         http.StatusForbidden,
			shouldRetry:        false,
			expectedCalls:      1,
			rateLimitExhausted: false,
		},
		{
			name:               "403 with rate limit should not retry",
			statusCode:         http.StatusForbidden,
			shouldRetry:        false,
			expectedCalls:      1,
			rateLimitExhausted: true,
		},
		{
			name:          "404 should not retry",
			statusCode:    http.StatusNotFound,
			shouldRetry:   false,
			expectedCalls: 1,
		},
		{
			name:          "500 should retry",
			statusCode:    http.StatusInternalServerError,
			shouldRetry:   true,
			expectedCalls: 3, // At least 3 with retries
		},
		{
			name:          "502 should retry",
			statusCode:    http.StatusBadGateway,
			shouldRetry:   true,
			expectedCalls: 3,
		},
		{
			name:          "503 should retry",
			statusCode:    http.StatusServiceUnavailable,
			shouldRetry:   true,
			expectedCalls: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			var requestCount int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				atomic.AddInt32(&requestCount, 1)
				if tt.rateLimitExhausted {
					w.Header().Set("X-RateLimit-Limit", "5000")
					w.Header().Set("X-RateLimit-Remaining", "0")
					w.Header().Set("X-RateLimit-Reset", time.Now().Add(1*time.Hour).Format("2006-01-02T15:04:05Z"))
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(`{"message": "Error"}`))
			}))
			defer server.Close()

			client := github.NewGitHubClient(server.Client(), server.URL, "test-token")
			ctx := context.Background()

			// Act
			_, _ = client.GetRepository(ctx, "owner", "repo")

			// Assert
			count := atomic.LoadInt32(&requestCount)
			if tt.shouldRetry {
				if count < int32(tt.expectedCalls) {
					t.Errorf("Expected at least %d calls (with retry), got %d", tt.expectedCalls, count)
				}
			} else {
				if count != 1 {
					t.Errorf("Expected exactly 1 call (no retry), got %d", count)
				}
			}
		})
	}
}
