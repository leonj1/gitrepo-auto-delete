---
executor: bdd
source_feature: ./tests/bdd/help_and_version.feature
---

<objective>
Implement the Cobra-based CLI interface with help, version, and all command-line flags.
This provides the user-facing interface for the tool.
</objective>

<gherkin>
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
</gherkin>

<requirements>
Based on the Gherkin scenarios, implement in `cmd/ghautodelete/`:

1. Cobra root command setup:
   - Command name: `ghautodelete`
   - Short description: "Enable auto-delete branches on GitHub repositories"
   - Long description with full explanation
   - Usage pattern: `ghautodelete [flags] <repository>`

2. Command-line flags:
   - `-t, --token string` - GitHub personal access token
   - `-c, --check` - Only check current status, don't modify
   - `-d, --dry-run` - Show what would be done without making changes
   - `-v, --verbose` - Enable verbose output
   - `--version` - Show version information

3. Help text content:
   - Tool description explaining purpose
   - Usage section with command pattern
   - Flags section with all available options
   - Examples section showing:
     - `ghautodelete octocat/hello-world`
     - `ghautodelete https://github.com/octocat/hello-world`
     - `ghautodelete git@github.com:octocat/hello-world.git`
     - `ghautodelete --check octocat/hello-world`
     - `ghautodelete --token ghp_xxxx octocat/hello-world`

4. Version display:
   - Show version number (injected via ldflags)
   - Format: `ghautodelete version X.Y.Z`
   - Default version for development: "dev"

5. Argument parsing:
   - Accept exactly one positional argument (repository)
   - Error with code 2 if no repository provided
   - Error with code 2 if too many arguments

6. Help/version special cases:
   - --help and -h should not require repository argument
   - --version should not require repository argument
   - Both should exit with code 0
</requirements>

<context>
Draft Specification: specs/DRAFT-github-auto-delete-branches.md
Gap Analysis: specs/GAP-ANALYSIS.md

CLI interface design from DRAFT spec (lines 409-432):
```
ghautodelete [flags] <repository>

Arguments:
  repository    GitHub repository (owner/repo or full URL)

Flags:
  -t, --token string    GitHub personal access token
  -c, --check           Only check current status
  -d, --dry-run         Show what would be done
  -v, --verbose         Enable verbose output
  -h, --help            Show help message
      --version         Show version information
```

Dependencies from DRAFT spec:
- github.com/spf13/cobra v1.8+
</context>

<implementation>
Follow TDD approach:
1. Write tests for help flag output
2. Write tests for version flag output
3. Write tests for argument parsing
4. Write tests for flag parsing
5. Write tests for error cases (missing repo)
6. Implement CLI to pass all tests

Architecture Guidelines:
- Use Cobra for CLI framework
- Keep CLI setup in cmd/ghautodelete/
- Use ldflags for version injection: `-ldflags "-X main.version=1.0.0"`
- Parse flags into CLIOptions struct
- Handle errors with appropriate exit codes
</implementation>

<verification>
All Gherkin scenarios must pass:
- [ ] Scenario: Display help with --help flag
- [ ] Scenario: Display help with -h flag
- [ ] Scenario: Display version with --version flag
- [ ] Scenario: Help shows all available flags
- [ ] Scenario: Help shows supported input formats
</verification>

<success_criteria>
- Cobra CLI properly configured
- All flags documented in help
- Version display works
- Examples shown in help
- Argument validation works
- Exit codes correct (0 for help/version)
- Unit tests provide >90% coverage
- All Gherkin scenarios pass
</success_criteria>
