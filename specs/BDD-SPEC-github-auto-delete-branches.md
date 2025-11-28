# BDD Specification Summary: GitHub Auto-Delete Branches

> Created: 2025-11-28
> Status: Approved
> Source: DRAFT-github-auto-delete-branches.md

## Overview

This document summarizes the Behavior-Driven Development (BDD) scenarios for the GitHub Auto-Delete Branches CLI tool. These scenarios capture the expected behavior of the tool from the user's perspective and will drive test-first development.

## Feature Coverage

The BDD specification covers 8 major feature areas with a total of 67 scenarios:

### 1. Repository Input Parsing (13 scenarios)
**File**: `/home/jose/src/gitrepo-auto-delete-branches/tests/bdd/repository_parsing.feature`

Validates the tool's ability to parse various repository identifier formats and reject invalid inputs.

**Positive Scenarios (6)**:
- Parse owner/repo format
- Parse HTTPS GitHub URL
- Parse HTTPS GitHub URL with .git suffix
- Parse SSH GitHub URL
- Parse SSH GitHub URL without .git suffix
- Handle input with leading and trailing whitespace

**Negative Scenarios (7)**:
- Reject invalid repository format with single segment
- Reject invalid repository format with too many segments
- Reject empty repository input
- Reject repository with invalid characters in owner
- Reject repository with invalid characters in repo name
- Reject invalid HTTPS URL format
- Reject non-GitHub HTTPS URL

### 2. Token Authentication (10 scenarios)
**File**: `/home/jose/src/gitrepo-auto-delete-branches/tests/bdd/token_authentication.feature`

Tests authentication mechanisms and token precedence rules.

**Positive Scenarios (5)**:
- Authenticate using token from command line flag
- Authenticate using GITHUB_TOKEN environment variable
- Authenticate using gh CLI configuration
- Token flag takes precedence over environment variable
- Environment variable takes precedence over gh CLI config

**Negative Scenarios (5)**:
- Fail when no token source is available
- Fail with invalid token
- Fail with expired token
- Fail with token missing repo scope

### 3. Auto-Delete Branch Configuration (5 scenarios)
**File**: `/home/jose/src/gitrepo-auto-delete-branches/tests/bdd/auto_delete_configuration.feature`

Validates the core functionality of enabling auto-delete branches.

**Scenarios**:
- Successfully enable auto-delete branches on a repository
- Report when auto-delete branches is already enabled
- Configure repository using HTTPS URL
- Configure repository using SSH URL
- Verify setting was applied after update

### 4. Check and Dry-Run Modes (6 scenarios)
**File**: `/home/jose/src/gitrepo-auto-delete-branches/tests/bdd/check_and_dryrun_modes.feature`

Tests preview modes that don't modify repository settings.

**Scenarios**:
- Check status when auto-delete is disabled
- Check status when auto-delete is enabled
- Dry-run when auto-delete is disabled
- Dry-run when auto-delete is already enabled
- Use short form -c for check mode
- Use short form -d for dry-run mode

### 5. Verbose Output Mode (5 scenarios)
**File**: `/home/jose/src/gitrepo-auto-delete-branches/tests/bdd/verbose_output.feature`

Validates verbose logging functionality.

**Scenarios**:
- Verbose mode shows authentication details
- Verbose mode shows repository fetch details
- Verbose mode shows API operations
- Use short form -v for verbose mode
- Non-verbose mode shows only essential output

### 6. Error Handling (9 scenarios)
**File**: `/home/jose/src/gitrepo-auto-delete-branches/tests/bdd/error_handling.feature`

Comprehensive error handling and recovery scenarios.

**Scenarios**:
- Repository not found
- No access to private repository
- Insufficient permissions to modify repository
- API rate limit exceeded
- Network connection failure
- GitHub API server error (with retry logic)
- Setting verification fails after update
- Invalid command line arguments
- Unknown flag provided

### 7. Help and Version Information (5 scenarios)
**File**: `/home/jose/src/gitrepo-auto-delete-branches/tests/bdd/help_and_version.feature`

Tests help and version display functionality.

**Scenarios**:
- Display help with --help flag
- Display help with -h flag
- Display version with --version flag
- Help shows all available flags
- Help shows supported input formats

### 8. Exit Codes (9 scenarios)
**File**: `/home/jose/src/gitrepo-auto-delete-branches/tests/bdd/exit_codes.feature`

Validates proper exit code handling for scripting integration.

**Exit Code Coverage**:
- Exit code 0: Success (4 scenarios)
- Exit code 2: Invalid arguments (2 scenarios)
- Exit code 3: Authentication failure (1 scenario)
- Exit code 4: Insufficient permissions (1 scenario)
- Exit code 5: Repository not found (1 scenario)
- Exit code 6: API rate limit (1 scenario)

## Coverage Analysis

### Functional Coverage
- **Input Validation**: 13 scenarios covering all input formats
- **Authentication**: 10 scenarios covering all auth mechanisms
- **Core Functionality**: 5 scenarios for main feature
- **Operating Modes**: 6 scenarios for check/dry-run modes
- **Output/Logging**: 5 scenarios for verbose mode
- **Error Handling**: 9 scenarios covering major failure cases
- **User Experience**: 5 scenarios for help/version
- **Integration**: 9 scenarios for exit codes

### Exit Code Coverage
All documented exit codes (0, 2, 3, 4, 5, 6) are covered with scenarios.

### Command-Line Flag Coverage
All flags are tested:
- `--token` / `-t`
- `--check` / `-c`
- `--dry-run` / `-d`
- `--verbose` / `-v`
- `--help` / `-h`
- `--version`

### Input Format Coverage
All supported repository identifier formats:
- owner/repo
- https://github.com/owner/repo
- https://github.com/owner/repo.git
- git@github.com:owner/repo.git
- git@github.com:owner/repo

### Edge Cases Covered
- Whitespace handling
- Invalid character validation
- Token precedence rules
- Already-enabled state handling
- API retry logic
- Setting verification

## Scenario Breakdown by Type

**Total Scenarios**: 67

- **Happy Path**: 25 scenarios (37%)
- **Error Handling**: 24 scenarios (36%)
- **Edge Cases**: 18 scenarios (27%)

## Test Implementation Strategy

These Gherkin scenarios will be converted to:

1. **Unit Tests**: Testing individual components
   - Repository parser
   - Token resolver
   - GitHub API client wrapper
   - Output formatter

2. **Integration Tests**: Testing component interactions
   - Authentication flow
   - API configuration flow
   - Error propagation

3. **End-to-End Tests**: Testing complete workflows
   - Full CLI execution
   - Mock GitHub API responses
   - Exit code validation

## Related Documentation

- **Original Spec**: DRAFT-github-auto-delete-branches.md (in prompts directory)
- **Feature Files**: /home/jose/src/gitrepo-auto-delete-branches/tests/bdd/*.feature

## Next Steps

1. Convert Gherkin scenarios to TDD test prompts
2. Implement tests in Go using testing framework
3. Implement features to make tests pass
4. Run BDD test infrastructure validation

## Notes

- All scenarios have been reviewed and approved
- Scenarios provide comprehensive coverage of requirements
- Each feature file is self-contained and testable
- Background sections reduce duplication in scenarios
- Error messages are specified for better UX
