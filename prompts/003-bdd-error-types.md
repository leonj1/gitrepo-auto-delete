---
executor: bdd
source_feature: ./tests/bdd/error_handling.feature
---

<objective>
Implement the error type hierarchy and exit codes as specified in the Draft specification.
These error types will be used throughout the application for consistent error handling.
</objective>

<gherkin>
Feature: Error Handling
  As a user of the CLI tool
  I want to receive clear and actionable error messages
  So that I can resolve issues quickly

  Background:
    Given a valid GitHub token with "repo" scope is configured

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

  Scenario: Setting verification fails after update
    Given auto-delete branches is currently disabled on the repository
    And the API update succeeds but verification shows setting not applied
    When the user runs "ghautodelete octocat/hello-world"
    Then an error should occur with code 1
    And the output should contain "Setting was not applied"

  Scenario: Invalid command line arguments
    When the user runs "ghautodelete"
    Then an error should occur with code 2
    And the output should contain "Repository argument is required"
    And the output should show usage information

  Scenario: Unknown flag provided
    When the user runs "ghautodelete --unknown-flag octocat/hello-world"
    Then an error should occur with code 2
    And the output should contain "unknown flag"
</gherkin>

<requirements>
Based on the Gherkin scenarios and Draft specification, implement:

1. AppError struct in `internal/errors/errors.go`:
   - Code field (ErrorCode type)
   - Message field (string)
   - Cause field (error, for wrapping)
   - Error() method returning formatted message
   - Unwrap() method for error chain

2. ErrorCode constants in `internal/errors/codes.go`:
   ```go
   const (
       ErrGeneral              ErrorCode = 1  // General/unexpected error
       ErrInvalidArguments     ErrorCode = 2  // Invalid CLI arguments
       ErrAuthenticationFailed ErrorCode = 3  // Token invalid/missing
       ErrInsufficientPerms    ErrorCode = 4  // Missing repo scope or admin access
       ErrRepositoryNotFound   ErrorCode = 5  // Repo doesn't exist or no access
       ErrAPIRateLimited       ErrorCode = 6  // Rate limit exceeded
   )
   ```

3. Error constructor functions:
   - `NewValidationError(message string) *AppError`
   - `NewAuthenticationError(message string, cause error) *AppError`
   - `NewAuthorizationError(message string) *AppError`
   - `NewRepositoryNotFoundError(owner, repo string) *AppError`
   - `NewRateLimitError(resetTime time.Time) *AppError`
   - `NewNetworkError(cause error) *AppError`
   - `NewAPIError(message string, cause error) *AppError`

4. Exit code mapping function:
   - `GetExitCode(err error) int` - Maps AppError to exit code

Edge Cases:
- Non-AppError errors should map to exit code 1
- Error messages should include actionable remediation advice
- Rate limit errors should include reset time
</requirements>

<context>
Draft Specification: specs/DRAFT-github-auto-delete-branches.md
Gap Analysis: specs/GAP-ANALYSIS.md

Exit codes from DRAFT spec (lines 435-444):
| Code | Meaning |
|------|---------|
| 0    | Success |
| 1    | General error |
| 2    | Invalid arguments |
| 3    | Authentication failed |
| 4    | Insufficient permissions |
| 5    | Repository not found |
| 6    | API rate limited |

Error hierarchy from DRAFT spec (lines 539-569):
- AppError base with Code, Message, Cause
- ValidationError, AuthenticationError, AuthorizationError
- ResourceError, APIError, ConfigurationError
</context>

<implementation>
Follow TDD approach:
1. Write tests for each error type constructor
2. Write tests for Error() message formatting
3. Write tests for Unwrap() error chain
4. Write tests for GetExitCode() mapping
5. Implement error types to pass tests

Architecture Guidelines:
- Keep error definitions in internal/errors/
- Error messages should follow the pattern: What happened + Why + How to fix
- Use error wrapping for cause chain
</implementation>

<verification>
All Gherkin scenarios must pass:
- [ ] Scenario: Repository not found (code 5)
- [ ] Scenario: No access to private repository (code 5)
- [ ] Scenario: Insufficient permissions to modify repository (code 4)
- [ ] Scenario: API rate limit exceeded (code 6)
- [ ] Scenario: Network connection failure (code 1)
- [ ] Scenario: GitHub API server error (code 1)
- [ ] Scenario: Setting verification fails after update (code 1)
- [ ] Scenario: Invalid command line arguments (code 2)
- [ ] Scenario: Unknown flag provided (code 2)
</verification>

<success_criteria>
- All error types implemented per spec
- All exit codes mapped correctly
- Error messages are clear and actionable
- Error wrapping works correctly
- Unit tests provide complete coverage
- All Gherkin scenarios pass
</success_criteria>
