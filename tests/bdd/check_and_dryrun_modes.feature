Feature: Check and Dry-Run Modes
  As a cautious developer
  I want to preview changes before applying them
  So that I can understand the impact without modifying settings

  Background:
    Given a valid GitHub token with "repo" scope is configured
    And the repository "octocat/hello-world" exists

  Scenario: Check status when auto-delete is disabled
    Given auto-delete branches is currently disabled on the repository
    When the user runs "ghautodelete --check octocat/hello-world"
    Then the output should contain "Repository: octocat/hello-world"
    And the output should contain the default branch name
    And the output should contain "Auto-delete branches: disabled"
    And the output should contain "To enable, run without --check flag"
    And the repository settings should not be modified
    And the exit code should be 0

  Scenario: Check status when auto-delete is enabled
    Given auto-delete branches is already enabled on the repository
    When the user runs "ghautodelete --check octocat/hello-world"
    Then the output should contain "Repository: octocat/hello-world"
    And the output should contain "Auto-delete branches: enabled"
    And the exit code should be 0

  Scenario: Dry-run when auto-delete is disabled
    Given auto-delete branches is currently disabled on the repository
    When the user runs "ghautodelete --dry-run octocat/hello-world"
    Then the output should contain "[DRY-RUN] Would enable auto-delete branches for octocat/hello-world"
    And the output should contain the default branch name
    And the output should contain "No changes made"
    And the repository settings should not be modified
    And the exit code should be 0

  Scenario: Dry-run when auto-delete is already enabled
    Given auto-delete branches is already enabled on the repository
    When the user runs "ghautodelete --dry-run octocat/hello-world"
    Then the output should contain "Auto-delete branches already enabled"
    And the output should contain "No changes needed"
    And the exit code should be 0

  Scenario: Use short form -c for check mode
    Given auto-delete branches is currently disabled on the repository
    When the user runs "ghautodelete -c octocat/hello-world"
    Then the output should contain "Auto-delete branches: disabled"
    And the exit code should be 0

  Scenario: Use short form -d for dry-run mode
    Given auto-delete branches is currently disabled on the repository
    When the user runs "ghautodelete -d octocat/hello-world"
    Then the output should contain "[DRY-RUN]"
    And the repository settings should not be modified
    And the exit code should be 0
