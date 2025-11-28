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
