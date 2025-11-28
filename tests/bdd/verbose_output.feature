Feature: Verbose Output Mode
  As a developer troubleshooting issues
  I want to see detailed operation information
  So that I can understand what the tool is doing

  Background:
    Given a valid GitHub token with "repo" scope is configured
    And the repository "octocat/hello-world" exists

  Scenario: Verbose mode shows authentication details
    Given the user provides token via the --token flag
    When the user runs "ghautodelete --verbose octocat/hello-world"
    Then the output should show the authenticated username
    And the output should show token validation succeeded

  Scenario: Verbose mode shows repository fetch details
    When the user runs "ghautodelete --verbose octocat/hello-world"
    Then the output should show "Fetching repository information"
    And the output should show the repository owner and name

  Scenario: Verbose mode shows API operations
    Given auto-delete branches is currently disabled on the repository
    When the user runs "ghautodelete --verbose octocat/hello-world"
    Then the output should show "Updating repository settings"
    And the output should show "Verifying settings applied"
    And the output should show the final success message

  Scenario: Use short form -v for verbose mode
    When the user runs "ghautodelete -v octocat/hello-world"
    Then verbose output should be displayed

  Scenario: Non-verbose mode shows only essential output
    Given auto-delete branches is currently disabled on the repository
    When the user runs "ghautodelete octocat/hello-world"
    Then the output should not show "Fetching repository information"
    And the output should not show "Updating repository settings"
    And the output should show the success message
