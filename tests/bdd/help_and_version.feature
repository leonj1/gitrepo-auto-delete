Feature: Help and Version Information
  As a new user of the CLI tool
  I want to access help and version information
  So that I can learn how to use the tool

  Scenario: Display help with --help flag
    When the user runs "ghautodelete --help"
    Then the output should contain the tool description
    And the output should contain "Usage:"
    And the output should list available flags
    And the output should show examples
    And the exit code should be 0

  Scenario: Display help with -h flag
    When the user runs "ghautodelete -h"
    Then the output should contain "Usage:"
    And the exit code should be 0

  Scenario: Display version with --version flag
    When the user runs "ghautodelete --version"
    Then the output should contain the version number
    And the exit code should be 0

  Scenario: Help shows all available flags
    When the user runs "ghautodelete --help"
    Then the output should describe the --token flag
    And the output should describe the --check flag
    And the output should describe the --dry-run flag
    And the output should describe the --verbose flag

  Scenario: Help shows supported input formats
    When the user runs "ghautodelete --help"
    Then the output should show "owner/repo" format example
    And the output should show HTTPS URL format example
    And the output should show SSH URL format example
