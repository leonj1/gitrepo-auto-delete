---
executor: bdd
source_feature: ./tests/bdd/error_handling.feature
---

<objective>
Implement the GitHub API client that handles repository operations, token validation,
rate limiting, retries, and error mapping.
</objective>

<gherkin>
Feature: Error Handling (GitHub Client Scenarios)
  As a user of the CLI tool
  I want to receive clear and actionable error messages
  So that I can resolve issues quickly

  Scenario: Repository not found
    Given the repository "octocat/nonexistent-repo" does not exist
    When the user runs "ghautodelete octocat/nonexistent-repo"
    Then an error should occur with code 5
    And the output should contain "Repository not found: octocat/nonexistent-repo"
    And the output should contain "Ensure the repository exists and you have access to it"

  Scenario: No access to private repository
    Given the repository "privateorg/private-repo" exists
    And the authenticated user does not have access to the repository
    When the user runs "ghautodelete privateorg/private-repo"
    Then an error should occur with code 5
    And the output should contain "Repository not found"

  Scenario: Insufficient permissions to modify repository
    Given the repository "octocat/hello-world" exists
    And the authenticated user has read-only access
    When the user runs "ghautodelete octocat/hello-world"
    Then an error should occur with code 4
    And the output should contain "Insufficient permissions"
    And the output should suggest "Admin access is required"

  Scenario: API rate limit exceeded
    Given the GitHub API rate limit has been exceeded
    When the user runs "ghautodelete octocat/hello-world"
    Then an error should occur with code 6
    And the output should contain "API rate limit exceeded"
    And the output should show when the rate limit will reset

  Scenario: Network connection failure
    Given the network connection to GitHub is unavailable
    When the user runs "ghautodelete octocat/hello-world"
    Then an error should occur with code 1
    And the output should contain "Network error"
    And the output should suggest checking internet connectivity

  Scenario: GitHub API server error
    Given the GitHub API returns a 500 Internal Server Error
    When the user runs "ghautodelete octocat/hello-world"
    Then the tool should retry the request up to 3 times
    And if all retries fail an error should occur with code 1
    And the output should contain "GitHub API error"
</gherkin>

<requirements>
Based on the Gherkin scenarios, implement in `internal/github/client.go`:

1. GitHubClient struct implementing IGitHubClient:
   - `GetRepository(ctx context.Context, owner, repo string) (IRepository, error)`
   - `UpdateRepository(ctx context.Context, owner, repo string, settings IRepositorySettings) error`
   - `ValidateToken(ctx context.Context) (ITokenInfo, error)`

2. NewGitHubClient constructor:
   - `NewGitHubClient(httpClient *http.Client, baseURL string, token string) IGitHubClient`
   - Accept HTTP client for testability
   - Accept base URL (default: https://api.github.com)
   - Accept authentication token

3. HTTP operations:
   - GET /repos/{owner}/{repo} - Fetch repository
   - PATCH /repos/{owner}/{repo} - Update repository
   - GET /user - Validate token and get user info
   - Set Authorization header with token
   - Set Accept header for GitHub API v3
   - Set User-Agent header

4. Response handling:
   - Parse JSON responses into model structs
   - Extract token scopes from X-OAuth-Scopes header
   - Extract rate limit info from headers

5. Retry logic:
   - Retry on 5xx errors up to 3 times
   - Use exponential backoff (1s, 2s, 4s)
   - Do NOT retry on 4xx errors
   - Respect context cancellation

6. Error mapping:
   - 401 -> ErrAuthenticationFailed (code 3)
   - 403 with rate limit -> ErrAPIRateLimited (code 6)
   - 403 without rate limit -> ErrInsufficientPerms (code 4)
   - 404 -> ErrRepositoryNotFound (code 5)
   - 5xx -> ErrGeneral (code 1) after retries
   - Network errors -> ErrGeneral (code 1)

7. Rate limiting:
   - Parse X-RateLimit-Remaining header
   - Parse X-RateLimit-Reset header (Unix timestamp)
   - Include reset time in rate limit errors
</requirements>

<context>
Draft Specification: specs/DRAFT-github-auto-delete-branches.md
Gap Analysis: specs/GAP-ANALYSIS.md

GitHub client interface from DRAFT spec (lines 21-30):
```go
type IGitHubClient interface {
    GetRepository(ctx context.Context, owner, repo string) (IRepository, error)
    UpdateRepository(ctx context.Context, owner, repo string, settings IRepositorySettings) error
    ValidateToken(ctx context.Context) (ITokenInfo, error)
}
```

API endpoints from DRAFT spec (lines 489-499):
- GET /repos/{owner}/{repo}
- PATCH /repos/{owner}/{repo}
- GET /user

Rate limiting strategy from DRAFT spec (lines 510-524):
- Check X-RateLimit-Remaining
- Exponential backoff for transient failures
- Max 3 retries
</context>

<implementation>
Follow TDD approach:
1. Write tests using httptest.Server for mock responses
2. Test successful operations (get, update, validate)
3. Test error scenarios (404, 401, 403, 5xx)
4. Test retry logic with transient failures
5. Test rate limit handling
6. Implement client to pass all tests

Architecture Guidelines:
- Keep implementation in internal/github/
- Use context for cancellation and timeouts
- Use interface types from pkg/interfaces/
- Return *AppError for all error cases
- Keep methods focused, extract helpers for HTTP operations
</implementation>

<verification>
All Gherkin scenarios must pass:
- [ ] Scenario: Repository not found
- [ ] Scenario: No access to private repository
- [ ] Scenario: Insufficient permissions to modify repository
- [ ] Scenario: API rate limit exceeded
- [ ] Scenario: Network connection failure
- [ ] Scenario: GitHub API server error (with retry logic)
</verification>

<success_criteria>
- GitHubClient implements IGitHubClient interface
- All API operations work correctly
- Retry logic works with exponential backoff
- Rate limiting is properly handled
- All error codes map correctly
- Unit tests provide >90% coverage
- All Gherkin scenarios pass
</success_criteria>
