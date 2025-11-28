---
executor: bdd
source_feature: ./tests/bdd/auto_delete_configuration.feature
---

<objective>
Implement the configuration service that orchestrates the auto-delete branch setting
operations including enabling the setting, checking status, and verifying changes.
</objective>

<gherkin>
Feature: Auto-Delete Branch Configuration
  As a repository administrator
  I want to enable automatic deletion of feature branches after merge
  So that my repository stays clean without manual cleanup

  Background:
    Given a valid GitHub token with "repo" scope is configured
    And the repository "octocat/hello-world" exists

  Scenario: Successfully enable auto-delete branches on a repository
    Given auto-delete branches is currently disabled on the repository
    When the user runs "ghautodelete octocat/hello-world"
    Then the tool should enable the delete_branch_on_merge setting
    And the output should contain "Successfully enabled auto-delete branches for octocat/hello-world"
    And the output should show the default branch name
    And the output should contain "Feature branches will now be deleted after PR merge"
    And the exit code should be 0

  Scenario: Report when auto-delete branches is already enabled
    Given auto-delete branches is already enabled on the repository
    When the user runs "ghautodelete octocat/hello-world"
    Then the tool should not modify repository settings
    And the output should contain "Auto-delete branches already enabled for octocat/hello-world"
    And the output should contain "No changes needed"
    And the exit code should be 0

  Scenario: Configure repository using HTTPS URL
    Given auto-delete branches is currently disabled on the repository
    When the user runs "ghautodelete https://github.com/octocat/hello-world"
    Then the tool should enable the delete_branch_on_merge setting
    And the output should contain "Successfully enabled auto-delete branches for octocat/hello-world"
    And the exit code should be 0

  Scenario: Configure repository using SSH URL
    Given auto-delete branches is currently disabled on the repository
    When the user runs "ghautodelete git@github.com:octocat/hello-world.git"
    Then the tool should enable the delete_branch_on_merge setting
    And the output should contain "Successfully enabled auto-delete branches for octocat/hello-world"
    And the exit code should be 0

  Scenario: Verify setting was applied after update
    Given auto-delete branches is currently disabled on the repository
    When the user runs "ghautodelete octocat/hello-world"
    Then the tool should fetch the repository settings after update
    And verify the delete_branch_on_merge setting is now true
    And the exit code should be 0
</gherkin>

<requirements>
Based on the Gherkin scenarios, implement in `internal/config/service.go`:

1. ConfigService struct implementing IConfigService:
   - `Configure(ctx context.Context, owner, repo string) (IConfigResult, error)`
   - `CheckStatus(ctx context.Context, owner, repo string) (IConfigResult, error)`

2. NewConfigService constructor:
   - `NewConfigService(client IGitHubClient, writer IOutputWriter) IConfigService`
   - Accept GitHub client for API operations
   - Accept output writer for logging

3. Configure method flow:
   a. Fetch current repository state via client.GetRepository
   b. Check if delete_branch_on_merge is already true
      - If yes: return result with AlreadyEnabled=true, skip update
   c. Create settings with DeleteBranchOnMerge=true
   d. Call client.UpdateRepository with settings
   e. Fetch repository again to verify setting applied
   f. If verification fails, return error
   g. Return ConfigResult with success state

4. CheckStatus method flow:
   a. Fetch repository state via client.GetRepository
   b. Return ConfigResult with current state (no modification)

5. ConfigResult implementation in `internal/config/result.go`:
   - AlreadyEnabled bool
   - NowEnabled bool
   - DefaultBranch string
   - RepositoryFullName string
   - Implement IConfigResult interface

6. Verbose logging:
   - Log "Fetching repository information" before get
   - Log "Updating repository settings" before update
   - Log "Verifying settings applied" before verification
</requirements>

<context>
Draft Specification: specs/DRAFT-github-auto-delete-branches.md
Gap Analysis: specs/GAP-ANALYSIS.md

ConfigService interface from DRAFT spec (lines 73-80):
```go
type IConfigService interface {
    Configure(ctx context.Context, owner, repo string) (IConfigResult, error)
    CheckStatus(ctx context.Context, owner, repo string) (IConfigResult, error)
}
```

Configure pseudocode from DRAFT spec (lines 261-297):
- Fetch current state
- Check if already configured
- Apply setting
- Verify change

ConfigResult from DRAFT spec (lines 147-159):
- AlreadyEnabled, NowEnabled, DefaultBranch, RepositoryFullName
</context>

<implementation>
Follow TDD approach:
1. Write tests with mock IGitHubClient
2. Test successful enable flow
3. Test already-enabled flow
4. Test verification flow
5. Test check status flow
6. Test error propagation
7. Implement service to pass all tests

Architecture Guidelines:
- Keep implementation in internal/config/
- Inject dependencies via constructor
- Use interfaces for all dependencies
- Return *AppError for all error cases
- Use output writer for verbose logging
</implementation>

<verification>
All Gherkin scenarios must pass:
- [ ] Scenario: Successfully enable auto-delete branches on a repository
- [ ] Scenario: Report when auto-delete branches is already enabled
- [ ] Scenario: Configure repository using HTTPS URL
- [ ] Scenario: Configure repository using SSH URL
- [ ] Scenario: Verify setting was applied after update
</verification>

<success_criteria>
- ConfigService implements IConfigService interface
- Configure enables setting when disabled
- Configure reports when already enabled
- Verification catches failed updates
- CheckStatus returns current state
- Unit tests provide >90% coverage
- All Gherkin scenarios pass
</success_criteria>
