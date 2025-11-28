---
executor: bdd
source_feature: ./tests/bdd/token_authentication.feature
---

<objective>
Implement the token provider that retrieves GitHub authentication tokens from multiple sources
with proper precedence: CLI flag > environment variable > gh CLI config.
</objective>

<gherkin>
Feature: Token Authentication
  As a developer
  I want to authenticate with GitHub using various token sources
  So that I can configure repository settings securely

  Background:
    Given a valid repository "octocat/hello-world" exists

  Scenario: Authenticate using token from command line flag
    Given the user provides token "ghp_validtoken123" via the --token flag
    And the token has the "repo" scope
    When the tool authenticates with GitHub
    Then authentication should succeed
    And the authenticated user should be identified

  Scenario: Authenticate using GITHUB_TOKEN environment variable
    Given the GITHUB_TOKEN environment variable is set to "ghp_envtoken456"
    And the token has the "repo" scope
    When the tool authenticates with GitHub
    Then authentication should succeed

  Scenario: Authenticate using gh CLI configuration
    Given the gh CLI is configured with a valid token for github.com
    And the token has the "repo" scope
    And no explicit token is provided
    And GITHUB_TOKEN environment variable is not set
    When the tool authenticates with GitHub
    Then authentication should succeed

  Scenario: Token flag takes precedence over environment variable
    Given the user provides token "ghp_flagtoken" via the --token flag
    And the GITHUB_TOKEN environment variable is set to "ghp_envtoken"
    When the tool authenticates with GitHub
    Then the tool should use the token from the --token flag

  Scenario: Environment variable takes precedence over gh CLI config
    Given the GITHUB_TOKEN environment variable is set to "ghp_envtoken"
    And the gh CLI is configured with a different token
    When the tool authenticates with GitHub
    Then the tool should use the token from GITHUB_TOKEN

  Scenario: Fail when no token source is available
    Given no token is provided via --token flag
    And GITHUB_TOKEN environment variable is not set
    And gh CLI is not configured
    When the tool attempts to authenticate
    Then an authentication error should occur
    And the error message should contain "No GitHub token found"
    And the error message should suggest "Set GITHUB_TOKEN or use --token flag"

  Scenario: Fail with invalid token
    Given the user provides token "invalid_token" via the --token flag
    When the tool attempts to authenticate with GitHub
    Then an authentication error should occur with code 3
    And the error message should contain "Authentication failed"

  Scenario: Fail with expired token
    Given the user provides an expired token via the --token flag
    When the tool attempts to authenticate with GitHub
    Then an authentication error should occur
    And the error message should contain "Token expired"

  Scenario: Fail with token missing repo scope
    Given the user provides token "ghp_limited" via the --token flag
    And the token only has the "read:user" scope
    When the tool validates token permissions
    Then an authorization error should occur with code 4
    And the error message should contain "Token lacks required permissions"
    And the error message should suggest generating a new token with "repo" scope
</gherkin>

<requirements>
Based on the Gherkin scenarios, implement in `internal/token/provider.go`:

1. TokenProvider struct implementing ITokenProvider:
   - `GetToken() (string, error)`

2. NewTokenProvider constructor:
   - `NewTokenProvider(explicitToken string) ITokenProvider`
   - Accept explicit token from CLI flag (can be empty string)

3. Token retrieval with precedence:
   - Priority 1: Explicit token (from constructor/CLI flag)
   - Priority 2: GITHUB_TOKEN environment variable
   - Priority 3: gh CLI config file (~/.config/gh/hosts.yml)

4. gh CLI config parsing:
   - Parse YAML file at `~/.config/gh/hosts.yml`
   - Extract oauth_token for github.com host
   - Handle missing file gracefully
   - Handle malformed YAML gracefully

5. Error handling:
   - Return descriptive error when no token found
   - Error message must contain "No GitHub token found"
   - Error message must suggest "Set GITHUB_TOKEN or use --token flag"

Note: Token validation (checking if token is valid/expired/has correct scopes)
is handled by the GitHub client, not the token provider. The provider only
retrieves the token string.
</requirements>

<context>
Draft Specification: specs/DRAFT-github-auto-delete-branches.md
Gap Analysis: specs/GAP-ANALYSIS.md

Token provider pseudocode from DRAFT spec (lines 349-373):
- Priority 1: Explicit token
- Priority 2: GITHUB_TOKEN env var
- Priority 3: gh CLI config file

gh CLI config format (hosts.yml):
```yaml
github.com:
    oauth_token: ghp_xxxxxxxxxxxx
    user: username
    git_protocol: https
```
</context>

<implementation>
Follow TDD approach:
1. Write tests for explicit token retrieval
2. Write tests for environment variable retrieval
3. Write tests for gh CLI config parsing
4. Write tests for precedence rules
5. Write tests for error cases
6. Implement provider to pass all tests

Architecture Guidelines:
- Keep implementation in internal/token/
- Use os.Getenv for environment variable
- Use os.UserHomeDir for home directory
- Use gopkg.in/yaml.v3 for YAML parsing
- Return *AppError with ErrAuthenticationFailed code for errors
- Allow injecting env getter and home dir for testing
</implementation>

<verification>
All Gherkin scenarios must pass:
- [ ] Scenario: Authenticate using token from command line flag
- [ ] Scenario: Authenticate using GITHUB_TOKEN environment variable
- [ ] Scenario: Authenticate using gh CLI configuration
- [ ] Scenario: Token flag takes precedence over environment variable
- [ ] Scenario: Environment variable takes precedence over gh CLI config
- [ ] Scenario: Fail when no token source is available
- [ ] Scenario: Fail with invalid token (token provider returns token, validation elsewhere)
- [ ] Scenario: Fail with expired token (token provider returns token, validation elsewhere)
- [ ] Scenario: Fail with token missing repo scope (token provider returns token, validation elsewhere)
</verification>

<success_criteria>
- TokenProvider implements ITokenProvider interface
- Token precedence works correctly (flag > env > gh CLI)
- gh CLI config parsing works correctly
- Error when no token available has correct message
- Unit tests provide >90% coverage
- All applicable Gherkin scenarios pass
</success_criteria>
