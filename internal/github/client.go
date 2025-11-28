// Package github provides GitHub API client implementation.
//
// This package implements the IGitHubClient interface and handles:
// - Repository retrieval and updates
// - Token validation
// - Error mapping (401->3, 403->4/6, 404->5, 5xx->1)
// - Retry logic for 5xx errors
// - Rate limit handling
package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	apperrors "github.com/josejulio/ghautodelete/internal/errors"
	"github.com/josejulio/ghautodelete/internal/token"
	"github.com/josejulio/ghautodelete/pkg/interfaces"
)

const (
	maxRetries     = 3
	retryBaseDelay = 500 * time.Millisecond
)

// GitHubClient implements the IGitHubClient interface for GitHub API operations.
type GitHubClient struct {
	httpClient *http.Client
	baseURL    string
	token      string
}

// NewGitHubClient creates a new GitHubClient instance.
// Parameters:
//   - httpClient: the HTTP client to use for API requests
//   - baseURL: the base URL for the GitHub API (e.g., "https://api.github.com")
//   - token: the GitHub API token for authentication
func NewGitHubClient(httpClient *http.Client, baseURL string, token string) *GitHubClient {
	return &GitHubClient{
		httpClient: httpClient,
		baseURL:    baseURL,
		token:      token,
	}
}

// GetRepository retrieves repository information from GitHub.
// Returns an IRepository containing the repository details.
func (c *GitHubClient) GetRepository(ctx context.Context, owner, name string) (interfaces.IRepository, error) {
	url := fmt.Sprintf("%s/repos/%s/%s", c.baseURL, owner, name)

	var repo *Repository
	err := c.doRequestWithRetry(ctx, http.MethodGet, url, nil, &repo)
	if err != nil {
		return nil, err
	}

	return repo, nil
}

// UpdateRepository updates repository settings on GitHub.
// Takes repository settings as IRepositorySettings and applies them.
func (c *GitHubClient) UpdateRepository(ctx context.Context, owner, name string, settings interfaces.IRepositorySettings) error {
	url := fmt.Sprintf("%s/repos/%s/%s", c.baseURL, owner, name)

	body := map[string]interface{}{
		"delete_branch_on_merge": settings.GetDeleteBranchOnMerge(),
	}

	return c.doRequestWithRetry(ctx, http.MethodPatch, url, body, nil)
}

// ValidateToken validates the GitHub API token and returns token information.
// Returns ITokenInfo containing scopes and user details.
func (c *GitHubClient) ValidateToken(ctx context.Context) (interfaces.ITokenInfo, error) {
	url := fmt.Sprintf("%s/user", c.baseURL)

	var response struct {
		Login string `json:"login"`
	}

	var scopes []string
	err := c.doRequestWithRetry(ctx, http.MethodGet, url, nil, &response, func(resp *http.Response) {
		// Parse scopes from X-OAuth-Scopes header
		scopesHeader := resp.Header.Get("X-OAuth-Scopes")
		if scopesHeader != "" {
			// Split by comma and trim whitespace
			for _, scope := range strings.Split(scopesHeader, ",") {
				trimmed := strings.TrimSpace(scope)
				if trimmed != "" {
					scopes = append(scopes, trimmed)
				}
			}
		}
	})

	if err != nil {
		return nil, err
	}

	tokenInfo := token.NewTokenInfo(response.Login, scopes)
	return tokenInfo, nil
}

// doRequestWithRetry executes an HTTP request with retry logic for 5xx errors.
// It handles error mapping and response parsing.
func (c *GitHubClient) doRequestWithRetry(
	ctx context.Context,
	method, url string,
	body interface{},
	result interface{},
	responseHandlers ...func(*http.Response),
) error {
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			delay := retryBaseDelay * time.Duration(attempt)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}

		err := c.doRequest(ctx, method, url, body, result, responseHandlers...)
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable (5xx)
		// We retry on server errors (ErrGeneral with Code 1) but not on other errors
		// Network errors and server errors both map to ErrGeneral
		if apperrors.GetExitCode(err) == 1 {
			// Only retry if we haven't exhausted retries
			continue
		}

		// Don't retry on 4xx errors
		return lastErr
	}

	return lastErr
}

// doRequest executes a single HTTP request without retry logic.
func (c *GitHubClient) doRequest(
	ctx context.Context,
	method, url string,
	bodyData interface{},
	result interface{},
	responseHandlers ...func(*http.Response),
) error {
	var bodyReader io.Reader
	if bodyData != nil {
		jsonBody, err := json.Marshal(bodyData)
		if err != nil {
			return apperrors.NewAPIError("failed to marshal request body", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return apperrors.NewNetworkError(err)
	}

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "ghautodelete")
	if bodyData != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return apperrors.NewNetworkError(err)
	}
	defer resp.Body.Close()

	// Execute response handlers (for extracting headers, etc.)
	for _, handler := range responseHandlers {
		handler(resp)
	}

	// Handle HTTP status codes
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if result != nil {
			if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
				return apperrors.NewAPIError("failed to decode response", err)
			}
		}
		return nil
	}

	// Read response body for error messages
	bodyBytes, _ := io.ReadAll(resp.Body)

	// Map HTTP status codes to application errors
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return apperrors.NewAuthenticationError("authentication failed", nil)

	case http.StatusForbidden:
		// Check if it's a rate limit error
		if c.isRateLimited(resp) {
			resetTime := c.parseResetTime(resp)
			return apperrors.NewRateLimitError(resetTime)
		}
		// Otherwise it's a permissions error
		return apperrors.NewAuthorizationError("insufficient permissions")

	case http.StatusNotFound:
		// Extract owner/repo from URL if possible
		// URL format: /repos/{owner}/{repo}
		owner := ""
		repo := ""
		parts := strings.Split(url, "/")
		for i, part := range parts {
			if part == "repos" && i+2 < len(parts) {
				owner = parts[i+1]
				repo = parts[i+2]
				break
			}
		}
		return apperrors.NewRepositoryNotFoundError(owner, repo)

	case http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		// Server errors - will be retried
		return apperrors.NewAPIError(
			fmt.Sprintf("GitHub API server error: %d %s", resp.StatusCode, string(bodyBytes)),
			nil,
		)

	default:
		// Other errors
		return apperrors.NewAPIError(
			fmt.Sprintf("GitHub API error: %d %s", resp.StatusCode, string(bodyBytes)),
			nil,
		)
	}
}

// isRateLimited checks if the response indicates a rate limit error.
func (c *GitHubClient) isRateLimited(resp *http.Response) bool {
	remaining := resp.Header.Get("X-RateLimit-Remaining")
	if remaining == "" {
		return false
	}

	remainingInt, err := strconv.Atoi(remaining)
	if err != nil {
		return false
	}

	return remainingInt == 0
}

// parseResetTime extracts the rate limit reset time from response headers.
func (c *GitHubClient) parseResetTime(resp *http.Response) time.Time {
	resetHeader := resp.Header.Get("X-RateLimit-Reset")
	if resetHeader == "" {
		return time.Now().Add(1 * time.Hour)
	}

	// Try to parse as Unix timestamp
	resetUnix, err := strconv.ParseInt(resetHeader, 10, 64)
	if err == nil {
		return time.Unix(resetUnix, 0)
	}

	// Try to parse as RFC3339 timestamp
	resetTime, err := time.Parse(time.RFC3339, resetHeader)
	if err == nil {
		return resetTime
	}

	// Fallback: 1 hour from now
	return time.Now().Add(1 * time.Hour)
}
