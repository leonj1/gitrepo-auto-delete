---
executor: bdd
source_feature: ./tests/bdd/repository_parsing.feature
---

<objective>
Implement the repository identifier parser that handles owner/repo format, HTTPS URLs, and SSH URLs.
The parser validates input and extracts owner and repository name.
</objective>

<gherkin>
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
</gherkin>

<requirements>
Based on the Gherkin scenarios, implement in `internal/parser/repo_parser.go`:

1. RepoParser struct implementing IRepoParser:
   - `Parse(input string) (owner string, repo string, err error)`

2. NewRepoParser constructor:
   - `NewRepoParser() IRepoParser`

3. Parsing logic:
   - Trim leading/trailing whitespace
   - Detect and parse HTTPS URLs: `https://github.com/owner/repo[.git]`
   - Detect and parse SSH URLs: `git@github.com:owner/repo[.git]`
   - Parse owner/repo format: `owner/repo`
   - Remove `.git` suffix if present

4. Validation:
   - Empty input returns error
   - Single segment (no `/`) returns error
   - More than 2 segments returns error
   - Invalid characters in owner/repo returns error
   - Non-GitHub URLs (gitlab, etc.) return error
   - Invalid HTTPS URL format returns error

5. Valid GitHub name characters:
   - Alphanumeric (a-z, A-Z, 0-9)
   - Hyphens (-)
   - Underscores (_) for repo names
   - Periods (.) for repo names
   - Cannot start with hyphen or period

Edge Cases:
- Whitespace handling (trim)
- Case preservation (GitHub is case-insensitive but preserve original)
- .git suffix removal
- URL path traversal prevention
</requirements>

<context>
Draft Specification: specs/DRAFT-github-auto-delete-branches.md
Gap Analysis: specs/GAP-ANALYSIS.md

Parsing pseudocode from DRAFT spec (lines 300-344):
- Handle https://github.com/ prefix
- Handle git@github.com: prefix
- Handle owner/repo format
- Validate GitHub name format
</context>

<implementation>
Follow TDD approach:
1. Write table-driven tests for all 13 scenarios
2. Write tests for edge cases (special characters, unicode, etc.)
3. Implement Parse method to pass all tests

Architecture Guidelines:
- Keep implementation in internal/parser/
- Return *AppError with ErrInvalidArguments code for validation errors
- Use regex for validation where appropriate
- Keep function under 50 lines, extract helpers as needed
</implementation>

<verification>
All Gherkin scenarios must pass:
- [ ] Scenario: Parse owner/repo format
- [ ] Scenario: Parse HTTPS GitHub URL
- [ ] Scenario: Parse HTTPS GitHub URL with .git suffix
- [ ] Scenario: Parse SSH GitHub URL
- [ ] Scenario: Parse SSH GitHub URL without .git suffix
- [ ] Scenario: Handle input with leading and trailing whitespace
- [ ] Scenario: Reject invalid repository format with single segment
- [ ] Scenario: Reject invalid repository format with too many segments
- [ ] Scenario: Reject empty repository input
- [ ] Scenario: Reject repository with invalid characters in owner
- [ ] Scenario: Reject repository with invalid characters in repo name
- [ ] Scenario: Reject invalid HTTPS URL format
- [ ] Scenario: Reject non-GitHub HTTPS URL
</verification>

<success_criteria>
- RepoParser implements IRepoParser interface
- All 13 Gherkin scenarios pass
- All valid input formats parse correctly
- All invalid inputs return appropriate errors
- Error messages match expected text
- Unit tests provide >90% coverage
</success_criteria>
