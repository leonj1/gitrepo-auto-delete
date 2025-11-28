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
