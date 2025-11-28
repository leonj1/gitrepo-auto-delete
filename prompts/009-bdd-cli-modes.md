---
executor: bdd
source_feature: ./tests/bdd/check_and_dryrun_modes.feature
---

<objective>
Implement the check and dry-run CLI modes that allow users to preview changes
without modifying repository settings.
</objective>

<gherkin>
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
</gherkin>

<requirements>
Based on the Gherkin scenarios, extend the application logic:

1. CLI mode flags in CLIOptions:
   - `CheckOnly bool` - Only check status, don't modify
   - `DryRun bool` - Show what would happen, don't modify

2. Check mode behavior (`--check` / `-c`):
   - Fetch repository information only
   - Display current status:
     - "Repository: owner/repo"
     - "Default branch: main"
     - "Auto-delete branches: enabled|disabled"
   - If disabled, show: "To enable, run without --check flag"
   - Never call UpdateRepository
   - Always exit 0 on success

3. Dry-run mode behavior (`--dry-run` / `-d`):
   - Fetch repository information
   - If disabled, show: "[DRY-RUN] Would enable auto-delete branches for owner/repo"
   - If enabled, show: "Auto-delete branches already enabled"
   - Show "No changes made" or "No changes needed"
   - Never call UpdateRepository
   - Always exit 0 on success

4. Output formatting for modes:
   ```
   Check mode (disabled):
   Repository: octocat/hello-world
   Default branch: main
   Auto-delete branches: disabled

   To enable, run without --check flag

   Check mode (enabled):
   Repository: octocat/hello-world
   Default branch: main
   Auto-delete branches: enabled

   Dry-run mode (disabled):
   [DRY-RUN] Would enable auto-delete branches for octocat/hello-world
     Default branch: main
     No changes made

   Dry-run mode (enabled):
   Auto-delete branches already enabled for octocat/hello-world
     No changes needed
   ```

5. Mode precedence:
   - If both --check and --dry-run provided, --check takes precedence
   - Error if conflicting flags (optional, or just pick one)
</requirements>

<context>
Draft Specification: specs/DRAFT-github-auto-delete-branches.md
Gap Analysis: specs/GAP-ANALYSIS.md

CLI flags from DRAFT spec (lines 419-424):
- `-c, --check` - Only check current status, don't modify
- `-d, --dry-run` - Show what would be done without making changes

Output examples from DRAFT spec (lines 462-477):
- Check mode output format
- Dry-run output format
</context>

<implementation>
Follow TDD approach:
1. Write tests for check mode with disabled setting
2. Write tests for check mode with enabled setting
3. Write tests for dry-run mode with disabled setting
4. Write tests for dry-run mode with enabled setting
5. Write tests for short form flags (-c, -d)
6. Implement modes to pass all tests

Architecture Guidelines:
- Add mode handling to the App struct or create dedicated handler
- Use ConfigService.CheckStatus for check mode
- Create DryRun method or handle in App
- Format output according to spec examples
</implementation>

<verification>
All Gherkin scenarios must pass:
- [ ] Scenario: Check status when auto-delete is disabled
- [ ] Scenario: Check status when auto-delete is enabled
- [ ] Scenario: Dry-run when auto-delete is disabled
- [ ] Scenario: Dry-run when auto-delete is already enabled
- [ ] Scenario: Use short form -c for check mode
- [ ] Scenario: Use short form -d for dry-run mode
</verification>

<success_criteria>
- Check mode shows current status without modification
- Dry-run mode shows what would happen without modification
- Both short and long form flags work
- Output format matches specification
- No API update calls in check/dry-run modes
- Unit tests provide >90% coverage
- All Gherkin scenarios pass
</success_criteria>
