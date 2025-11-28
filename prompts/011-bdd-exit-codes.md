---
executor: bdd
source_feature: ./tests/bdd/exit_codes.feature
---

<objective>
Implement proper exit code handling throughout the application to enable
scripting integration and programmatic result handling.
</objective>

<gherkin>
Feature: Exit Codes
  As a developer integrating the tool into scripts
  I want consistent and meaningful exit codes
  So that I can handle different outcomes programmatically

  Background:
    Given a valid GitHub token with "repo" scope is configured

  Scenario: Exit code 0 on successful configuration
    Given the repository "octocat/hello-world" exists
    And auto-delete branches is currently disabled on the repository
    When the user runs "ghautodelete octocat/hello-world"
    Then the exit code should be 0

  Scenario: Exit code 0 when already configured
    Given the repository "octocat/hello-world" exists
    And auto-delete branches is already enabled on the repository
    When the user runs "ghautodelete octocat/hello-world"
    Then the exit code should be 0

  Scenario: Exit code 0 on successful check
    Given the repository "octocat/hello-world" exists
    When the user runs "ghautodelete --check octocat/hello-world"
    Then the exit code should be 0

  Scenario: Exit code 0 on successful dry-run
    Given the repository "octocat/hello-world" exists
    When the user runs "ghautodelete --dry-run octocat/hello-world"
    Then the exit code should be 0

  Scenario: Exit code 2 on invalid arguments
    When the user runs "ghautodelete invalid/repo/format/extra"
    Then the exit code should be 2

  Scenario: Exit code 3 on authentication failure
    Given an invalid token is provided
    When the user runs "ghautodelete octocat/hello-world"
    Then the exit code should be 3

  Scenario: Exit code 4 on insufficient permissions
    Given the token lacks the "repo" scope
    When the user runs "ghautodelete octocat/hello-world"
    Then the exit code should be 4

  Scenario: Exit code 5 on repository not found
    Given the repository "octocat/nonexistent" does not exist
    When the user runs "ghautodelete octocat/nonexistent"
    Then the exit code should be 5

  Scenario: Exit code 6 on API rate limit
    Given the GitHub API rate limit has been exceeded
    When the user runs "ghautodelete octocat/hello-world"
    Then the exit code should be 6
</gherkin>

<requirements>
Based on the Gherkin scenarios, implement exit code handling:

1. Exit code constants (already in errors package):
   ```go
   const (
       ExitSuccess             = 0  // Success
       ExitGeneral             = 1  // General/unexpected error
       ExitInvalidArguments    = 2  // Invalid CLI arguments
       ExitAuthenticationFailed = 3  // Token invalid/missing
       ExitInsufficientPerms   = 4  // Missing permissions
       ExitRepositoryNotFound  = 5  // Repo doesn't exist
       ExitRateLimited         = 6  // Rate limit exceeded
   )
   ```

2. Exit code mapping function (GetExitCode in errors package):
   - Map AppError.Code to exit code
   - Map non-AppError to exit code 1
   - Map nil error to exit code 0

3. Main function exit handling:
   - Capture error from App.Run()
   - Map error to exit code
   - Call os.Exit with appropriate code

4. Success cases (exit 0):
   - Successful configuration
   - Already configured (no-op)
   - Successful check mode
   - Successful dry-run mode
   - Help display
   - Version display

5. Error cases with specific codes:
   - Invalid arguments (exit 2): repo format, missing repo, unknown flag
   - Authentication (exit 3): invalid token, expired token
   - Authorization (exit 4): missing scope, no admin access
   - Not found (exit 5): repo doesn't exist, no access
   - Rate limit (exit 6): API rate limit exceeded
   - General (exit 1): network error, server error, unexpected

6. Integration in main.go:
   ```go
   func main() {
       if err := run(); err != nil {
           os.Exit(errors.GetExitCode(err))
       }
   }
   ```
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
</context>

<implementation>
Follow TDD approach:
1. Write integration tests that verify exit codes
2. Test all success scenarios (exit 0)
3. Test invalid argument scenarios (exit 2)
4. Test authentication failure scenarios (exit 3)
5. Test authorization failure scenarios (exit 4)
6. Test not found scenarios (exit 5)
7. Test rate limit scenarios (exit 6)
8. Ensure proper exit code propagation

Architecture Guidelines:
- Use errors.GetExitCode for mapping
- Handle exit in main, not in App
- Test via subprocess execution or by testing App.Run error returns
- Ensure all error paths return appropriate AppError
</implementation>

<verification>
All Gherkin scenarios must pass:
- [ ] Scenario: Exit code 0 on successful configuration
- [ ] Scenario: Exit code 0 when already configured
- [ ] Scenario: Exit code 0 on successful check
- [ ] Scenario: Exit code 0 on successful dry-run
- [ ] Scenario: Exit code 2 on invalid arguments
- [ ] Scenario: Exit code 3 on authentication failure
- [ ] Scenario: Exit code 4 on insufficient permissions
- [ ] Scenario: Exit code 5 on repository not found
- [ ] Scenario: Exit code 6 on API rate limit
</verification>

<success_criteria>
- All exit codes documented and implemented
- Error to exit code mapping is consistent
- Success cases all exit with code 0
- Error cases exit with appropriate codes
- Exit codes enable scripting integration
- Integration tests verify exit codes
- All Gherkin scenarios pass
</success_criteria>
