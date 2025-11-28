Feature: Repository Input Parsing
  As a developer
  I want to provide repository identifiers in various formats
  So that I can use the tool flexibly with different inputs

  Scenario: Parse owner/repo format
    Given the user provides repository input "octocat/hello-world"
    When the repository identifier is parsed
    Then the owner should be "octocat"
    And the repository name should be "hello-world"

  Scenario: Parse HTTPS GitHub URL
    Given the user provides repository input "https://github.com/octocat/hello-world"
    When the repository identifier is parsed
    Then the owner should be "octocat"
    And the repository name should be "hello-world"

  Scenario: Parse HTTPS GitHub URL with .git suffix
    Given the user provides repository input "https://github.com/octocat/hello-world.git"
    When the repository identifier is parsed
    Then the owner should be "octocat"
    And the repository name should be "hello-world"

  Scenario: Parse SSH GitHub URL
    Given the user provides repository input "git@github.com:octocat/hello-world.git"
    When the repository identifier is parsed
    Then the owner should be "octocat"
    And the repository name should be "hello-world"

  Scenario: Parse SSH GitHub URL without .git suffix
    Given the user provides repository input "git@github.com:octocat/hello-world"
    When the repository identifier is parsed
    Then the owner should be "octocat"
    And the repository name should be "hello-world"

  Scenario: Handle input with leading and trailing whitespace
    Given the user provides repository input "  octocat/hello-world  "
    When the repository identifier is parsed
    Then the owner should be "octocat"
    And the repository name should be "hello-world"

  Scenario: Reject invalid repository format with single segment
    Given the user provides repository input "hello-world"
    When the repository identifier is parsed
    Then a validation error should occur with message "Expected format: owner/repo"

  Scenario: Reject invalid repository format with too many segments
    Given the user provides repository input "octocat/hello/world"
    When the repository identifier is parsed
    Then a validation error should occur with message "Expected format: owner/repo"

  Scenario: Reject empty repository input
    Given the user provides repository input ""
    When the repository identifier is parsed
    Then a validation error should occur with message "Repository identifier is required"

  Scenario: Reject repository with invalid characters in owner
    Given the user provides repository input "octo cat/hello-world"
    When the repository identifier is parsed
    Then a validation error should occur with message "Invalid repository name characters"

  Scenario: Reject repository with invalid characters in repo name
    Given the user provides repository input "octocat/hello world"
    When the repository identifier is parsed
    Then a validation error should occur with message "Invalid repository name characters"

  Scenario: Reject invalid HTTPS URL format
    Given the user provides repository input "https://github.com/octocat"
    When the repository identifier is parsed
    Then a validation error should occur with message "Invalid GitHub URL format"

  Scenario: Reject non-GitHub HTTPS URL
    Given the user provides repository input "https://gitlab.com/octocat/hello-world"
    When the repository identifier is parsed
    Then a validation error should occur with message "Expected format: owner/repo"
