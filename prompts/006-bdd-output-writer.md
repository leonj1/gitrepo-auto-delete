---
executor: bdd
source_feature: ./tests/bdd/verbose_output.feature
---

<objective>
Implement the output writer that handles CLI output formatting with support for
verbose mode and different message types (success, error, info, verbose).
</objective>

<gherkin>
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
</gherkin>

<requirements>
Based on the Gherkin scenarios, implement in `internal/output/writer.go`:

1. OutputWriter struct implementing IOutputWriter:
   - `Success(message string)` - Print success message (always shown)
   - `Error(message string)` - Print error message (always shown, to stderr)
   - `Info(message string)` - Print informational message (always shown)
   - `Verbose(message string)` - Print verbose message (only in verbose mode)

2. NewOutputWriter constructor:
   - `NewOutputWriter(verbose bool, out io.Writer, errOut io.Writer) IOutputWriter`
   - Accept verbose flag to control verbose output
   - Accept output writer (usually os.Stdout)
   - Accept error output writer (usually os.Stderr)

3. Output formatting:
   - Success messages: Prefixed with checkmark or "Success: "
   - Error messages: Prefixed with "Error: " and written to stderr
   - Info messages: Plain text output
   - Verbose messages: Prefixed with "[verbose]" or similar marker

4. Verbose mode behavior:
   - When verbose=false, Verbose() calls are no-ops
   - When verbose=true, Verbose() calls write to output
   - Success, Error, and Info always write regardless of verbose flag

5. Message formatting for scenarios:
   - Authentication success: Show username if verbose
   - Repository fetch: Show "Fetching repository information" if verbose
   - API operations: Show "Updating repository settings" if verbose
   - Verification: Show "Verifying settings applied" if verbose
</requirements>

<context>
Draft Specification: specs/DRAFT-github-auto-delete-branches.md
Gap Analysis: specs/GAP-ANALYSIS.md

Output interface from DRAFT spec (lines 65-71):
```go
type IOutputWriter interface {
    Success(message string)
    Error(message string)
    Info(message string)
    Verbose(message string)
}
```

Output examples from DRAFT spec (lines 447-483):
- Success messages with repository name and default branch
- Error messages with context and suggestions
- Check mode output format
- Dry-run output format
</context>

<implementation>
Follow TDD approach:
1. Write tests for Success output
2. Write tests for Error output (verify stderr)
3. Write tests for Info output
4. Write tests for Verbose output (both modes)
5. Write tests verifying verbose suppression
6. Implement writer to pass all tests

Architecture Guidelines:
- Keep implementation in internal/output/
- Use io.Writer for testability
- Use consistent formatting prefixes
- Consider colors for terminal output (optional, check if TTY)
- Keep methods simple and focused
</implementation>

<verification>
All Gherkin scenarios must pass:
- [ ] Scenario: Verbose mode shows authentication details
- [ ] Scenario: Verbose mode shows repository fetch details
- [ ] Scenario: Verbose mode shows API operations
- [ ] Scenario: Use short form -v for verbose mode
- [ ] Scenario: Non-verbose mode shows only essential output
</verification>

<success_criteria>
- OutputWriter implements IOutputWriter interface
- Verbose mode shows detailed output
- Non-verbose mode suppresses verbose messages
- Error messages go to stderr
- Success/Info messages go to stdout
- Unit tests provide >90% coverage
- All Gherkin scenarios pass
</success_criteria>
