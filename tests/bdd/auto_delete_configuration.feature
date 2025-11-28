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
