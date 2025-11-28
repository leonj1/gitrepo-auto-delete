# Gap Analysis: GitHub Auto-Delete Branches CLI

> Created: 2025-11-28
> Status: Complete
> Source: BDD Specification and Draft Spec

## Project Status

This is a **greenfield project** with no existing Go code. All components must be built from scratch.

## Codebase Scan Results

### Existing Code
- **Go source files**: 0
- **Test files**: 0
- **go.mod**: Does not exist
- **Makefile**: Does not exist

### Project Structure
```
gitrepo-auto-delete-branches/
|-- .claude/              # Claude Code configuration
|-- specs/                # Specifications
|   |-- DRAFT-github-auto-delete-branches.md
|   |-- BDD-SPEC-github-auto-delete-branches.md
|   +-- GAP-ANALYSIS.md (this file)
|-- tests/
|   +-- bdd/              # Gherkin feature files (8 files)
+-- prompts/              # TDD prompt files (to be created)
```

## Components to Build

Based on the BDD scenarios and Draft specification, the following components need to be implemented:

### 1. Project Setup (Foundational)
- `go.mod` with module name `github.com/josejulio/ghautodelete`
- Project structure as defined in DRAFT spec
- Makefile with build, test, lint targets
- Dockerfile.test for containerized testing

### 2. Interfaces Package (`pkg/interfaces/`)
- `IGitHubClient` - GitHub API abstraction
- `IRepository` - Repository data model interface
- `IRepositorySettings` - Settings interface
- `ITokenInfo` - Token metadata interface
- `IRepoParser` - Repository identifier parser
- `ITokenProvider` - Token source abstraction
- `IOutputWriter` - CLI output formatting
- `IConfigService` - Main configuration service
- `IConfigResult` - Configuration result interface

### 3. Repository Parser (`internal/parser/`)
- Parse `owner/repo` format
- Parse HTTPS GitHub URLs
- Parse SSH GitHub URLs
- Handle whitespace trimming
- Validate owner/repo names
- Return proper error messages

### 4. Token Provider (`internal/token/`)
- Support `--token` flag (highest priority)
- Support `GITHUB_TOKEN` env var (medium priority)
- Support gh CLI config file (lowest priority)
- Return descriptive error when no token found

### 5. GitHub Client (`internal/github/`)
- `GetRepository(ctx, owner, repo)` - Fetch repo info
- `UpdateRepository(ctx, owner, repo, settings)` - Update settings
- `ValidateToken(ctx)` - Check token validity
- HTTP client with timeout, retries, backoff
- Rate limit handling
- Error mapping to AppError

### 6. Config Service (`internal/config/`)
- `Configure(ctx, owner, repo)` - Enable auto-delete
- `CheckStatus(ctx, owner, repo)` - Check current status
- Verify settings after update
- Return ConfigResult with status

### 7. Output Writer (`internal/output/`)
- `Success(message)` - Success output
- `Error(message)` - Error output
- `Info(message)` - Informational output
- `Verbose(message)` - Debug/verbose output
- Respect verbose flag

### 8. Error Types (`internal/errors/`)
- `AppError` base type with Code, Message, Cause
- Error codes: 0-6 as defined in spec
- Error hierarchy for different failure types

### 9. CLI Interface (`cmd/ghautodelete/`)
- Cobra-based CLI
- Flags: `--token/-t`, `--check/-c`, `--dry-run/-d`, `--verbose/-v`, `--help/-h`, `--version`
- Repository argument parsing
- Help text with examples
- Version display

### 10. Main Application (`internal/app/`)
- Wire up all dependencies
- Execute main flow
- Handle exit codes

## Reuse Opportunities

Since this is a greenfield project, there are **no existing code reuse opportunities**.

However, the implementation can leverage:
- Standard library `net/http` for HTTP client
- `encoding/json` for JSON handling
- `gopkg.in/yaml.v3` for gh CLI config parsing
- `github.com/spf13/cobra` for CLI framework

## Refactoring Requirements

**No refactoring needed** - this is a new project.

## Implementation Order Recommendation

The prompts should be executed sequentially in this order:

1. **Project Setup** - Foundation (go.mod, structure, Makefile)
2. **Interfaces** - Define all interface contracts
3. **Error Types** - Foundation for error handling
4. **Repository Parser** - Independent component
5. **Token Provider** - Independent component
6. **Output Writer** - Independent component
7. **GitHub Client** - Depends on interfaces
8. **Config Service** - Depends on GitHub client, output writer
9. **CLI Interface** - Depends on all components
10. **Main Integration** - Wire everything together

## Test Strategy

Each component should have:
- Unit tests with mock dependencies
- Table-driven tests for edge cases
- Integration tests where applicable

The BDD scenarios will drive test creation through the TDD process.

## GO Signal

**Status: GO**

This is a greenfield project. No refactoring needed. Proceed with prompt creation for TDD implementation.
