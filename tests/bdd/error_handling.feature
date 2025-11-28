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
