# DRAFT: GitHub Auto-Delete Branches CLI Tool

**Status**: DRAFT
**Created**: 2025-11-28
**Author**: Solutions Architect

---

## Overview

A command-line tool written in Go that configures GitHub repositories to automatically delete feature branches after pull requests are merged. The tool authenticates with GitHub's API and enables the "Automatically delete head branches" repository setting.

---

## Interfaces

### Core Abstractions

```go
// IGitHubClient abstracts all GitHub API operations
type IGitHubClient interface {
    // GetRepository fetches repository metadata
    GetRepository(ctx context.Context, owner, repo string) (IRepository, error)

    // UpdateRepository updates repository settings
    UpdateRepository(ctx context.Context, owner, repo string, settings IRepositorySettings) error

    // ValidateToken checks if the current token has sufficient permissions
    ValidateToken(ctx context.Context) (ITokenInfo, error)
}

// IRepository represents a GitHub repository
type IRepository interface {
    GetOwner() string
    GetName() string
    GetDefaultBranch() string
    GetDeleteBranchOnMerge() bool
    GetFullName() string
}

// IRepositorySettings represents updateable repository settings
type IRepositorySettings interface {
    GetDeleteBranchOnMerge() bool
}

// ITokenInfo represents GitHub token information
type ITokenInfo interface {
    GetScopes() []string
    HasScope(scope string) bool
    GetUsername() string
}

// IRepoParser parses repository identifiers from various formats
type IRepoParser interface {
    // Parse accepts "owner/repo" or full GitHub URL
    Parse(input string) (owner string, repo string, err error)
}

// ITokenProvider retrieves GitHub authentication tokens
type ITokenProvider interface {
    // GetToken retrieves token from configured sources
    GetToken() (string, error)
}

// IOutputWriter handles CLI output formatting
type IOutputWriter interface {
    Success(message string)
    Error(message string)
    Info(message string)
    Verbose(message string)
}

// IConfigService manages application configuration
type IConfigService interface {
    // Configure sets up the auto-delete branch setting
    Configure(ctx context.Context, owner, repo string) (IConfigResult, error)

    // CheckStatus returns current configuration status
    CheckStatus(ctx context.Context, owner, repo string) (IConfigResult, error)
}

// IConfigResult represents the result of a configuration operation
type IConfigResult interface {
    WasAlreadyEnabled() bool
    IsNowEnabled() bool
    GetDefaultBranch() string
    GetRepositoryFullName() string
}
```

---

## Data Models

### Repository

```go
// Repository holds GitHub repository information
type Repository struct {
    Owner               string
    Name                string
    DefaultBranch       string
    DeleteBranchOnMerge bool
}

func (r *Repository) GetOwner() string              { return r.Owner }
func (r *Repository) GetName() string               { return r.Name }
func (r *Repository) GetDefaultBranch() string      { return r.DefaultBranch }
func (r *Repository) GetDeleteBranchOnMerge() bool  { return r.DeleteBranchOnMerge }
func (r *Repository) GetFullName() string           { return r.Owner + "/" + r.Name }
```

### RepositorySettings

```go
// RepositorySettings holds updateable repository configuration
type RepositorySettings struct {
    DeleteBranchOnMerge bool
}

func (s *RepositorySettings) GetDeleteBranchOnMerge() bool { return s.DeleteBranchOnMerge }
```

### TokenInfo

```go
// TokenInfo holds GitHub token metadata
type TokenInfo struct {
    Scopes   []string
    Username string
}

func (t *TokenInfo) GetScopes() []string { return t.Scopes }
func (t *TokenInfo) GetUsername() string { return t.Username }
func (t *TokenInfo) HasScope(scope string) bool {
    for _, s := range t.Scopes {
        if s == scope {
            return true
        }
    }
    return false
}
```

### ConfigResult

```go
// ConfigResult holds the outcome of a configuration operation
type ConfigResult struct {
    AlreadyEnabled     bool
    NowEnabled         bool
    DefaultBranch      string
    RepositoryFullName string
}

func (c *ConfigResult) WasAlreadyEnabled() bool      { return c.AlreadyEnabled }
func (c *ConfigResult) IsNowEnabled() bool           { return c.NowEnabled }
func (c *ConfigResult) GetDefaultBranch() string     { return c.DefaultBranch }
func (c *ConfigResult) GetRepositoryFullName() string { return c.RepositoryFullName }
```

### CLI Options

```go
// CLIOptions holds parsed command-line arguments
type CLIOptions struct {
    Repository   string // owner/repo or full URL
    Token        string // GitHub token (optional, can come from env/file)
    Verbose      bool   // Enable verbose output
    DryRun       bool   // Check status without modifying
    CheckOnly    bool   // Only check current status
}
```

### Error Types

```go
// AppError represents application-specific errors
type AppError struct {
    Code    ErrorCode
    Message string
    Cause   error
}

type ErrorCode int

const (
    ErrInvalidRepository ErrorCode = iota + 1
    ErrAuthenticationFailed
    ErrInsufficientPermissions
    ErrRepositoryNotFound
    ErrAPIRateLimited
    ErrNetworkFailure
    ErrUnexpected
)

func (e *AppError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("%s: %v", e.Message, e.Cause)
    }
    return e.Message
}

func (e *AppError) Unwrap() error { return e.Cause }
```

---

## Logic Flow

### Main Application Flow

```
START
  |
  v
[Parse CLI Arguments]
  |
  v
[Validate Repository Format] --> ERROR: Invalid repository format
  |
  v
[Obtain GitHub Token]
  |-- From --token flag
  |-- From GITHUB_TOKEN env var
  |-- From gh CLI config (~/.config/gh/hosts.yml)
  |
  v
[Validate Token] --> ERROR: Authentication failed
  |
  v
[Check Token Permissions] --> ERROR: Insufficient permissions (needs 'repo' scope)
  |
  v
[Fetch Repository Info] --> ERROR: Repository not found
  |
  v
[Check if setting already enabled?]
  |-- YES --> OUTPUT: "Already configured" + EXIT(0)
  |-- NO  --> continue
  |
  v
[DryRun mode?]
  |-- YES --> OUTPUT: "Would enable setting" + EXIT(0)
  |-- NO  --> continue
  |
  v
[Update Repository Setting]
  |
  v
[Verify Setting Applied] --> ERROR: Failed to verify
  |
  v
OUTPUT: Success message with details
  |
  v
END (EXIT 0)
```

### Pseudocode: ConfigService.Configure

```
FUNCTION Configure(ctx, owner, repo):
    // Fetch current state
    repository = client.GetRepository(ctx, owner, repo)
    IF error THEN RETURN error

    // Check if already configured
    IF repository.GetDeleteBranchOnMerge() THEN
        RETURN ConfigResult{
            AlreadyEnabled: true,
            NowEnabled: true,
            DefaultBranch: repository.GetDefaultBranch(),
            RepositoryFullName: repository.GetFullName()
        }
    END IF

    // Apply setting
    settings = RepositorySettings{DeleteBranchOnMerge: true}
    error = client.UpdateRepository(ctx, owner, repo, settings)
    IF error THEN RETURN error

    // Verify change
    repository = client.GetRepository(ctx, owner, repo)
    IF error THEN RETURN error

    IF NOT repository.GetDeleteBranchOnMerge() THEN
        RETURN error("Setting was not applied")
    END IF

    RETURN ConfigResult{
        AlreadyEnabled: false,
        NowEnabled: true,
        DefaultBranch: repository.GetDefaultBranch(),
        RepositoryFullName: repository.GetFullName()
    }
END FUNCTION
```

### Pseudocode: RepoParser.Parse

```
FUNCTION Parse(input):
    // Trim whitespace
    input = strings.TrimSpace(input)

    // Handle full GitHub URLs
    IF strings.HasPrefix(input, "https://github.com/") THEN
        path = strings.TrimPrefix(input, "https://github.com/")
        path = strings.TrimSuffix(path, ".git")
        parts = strings.Split(path, "/")
        IF len(parts) != 2 THEN
            RETURN error("Invalid GitHub URL format")
        END IF
        RETURN parts[0], parts[1], nil
    END IF

    // Handle git@ URLs
    IF strings.HasPrefix(input, "git@github.com:") THEN
        path = strings.TrimPrefix(input, "git@github.com:")
        path = strings.TrimSuffix(path, ".git")
        parts = strings.Split(path, "/")
        IF len(parts) != 2 THEN
            RETURN error("Invalid git URL format")
        END IF
        RETURN parts[0], parts[1], nil
    END IF

    // Handle owner/repo format
    parts = strings.Split(input, "/")
    IF len(parts) != 2 THEN
        RETURN error("Expected format: owner/repo")
    END IF

    owner = parts[0]
    repo = parts[1]

    // Validate owner and repo names
    IF NOT isValidGitHubName(owner) OR NOT isValidGitHubName(repo) THEN
        RETURN error("Invalid repository name characters")
    END IF

    RETURN owner, repo, nil
END FUNCTION
```

### Pseudocode: TokenProvider.GetToken

```
FUNCTION GetToken():
    // Priority 1: Explicit token (from CLI)
    IF explicitToken != "" THEN
        RETURN explicitToken, nil
    END IF

    // Priority 2: Environment variable
    envToken = os.Getenv("GITHUB_TOKEN")
    IF envToken != "" THEN
        RETURN envToken, nil
    END IF

    // Priority 3: GitHub CLI config
    ghConfigPath = filepath.Join(os.UserHomeDir(), ".config", "gh", "hosts.yml")
    IF fileExists(ghConfigPath) THEN
        token = parseGHConfig(ghConfigPath, "github.com")
        IF token != "" THEN
            RETURN token, nil
        END IF
    END IF

    RETURN "", error("No GitHub token found. Set GITHUB_TOKEN or use --token flag")
END FUNCTION
```

---

## Constructor Signatures

All constructors follow the rule of maximum 3 parameters. Dependencies are injected via interfaces.

```go
// NewGitHubClient creates a new GitHub API client
// Parameters: httpClient for HTTP operations, baseURL for API endpoint, token for auth
func NewGitHubClient(httpClient IHTTPClient, baseURL string, token string) IGitHubClient

// NewRepoParser creates a repository identifier parser
// No dependencies required
func NewRepoParser() IRepoParser

// NewTokenProvider creates a token provider
// Parameters: explicitToken from CLI (can be empty)
func NewTokenProvider(explicitToken string) ITokenProvider

// NewConfigService creates the main configuration service
// Parameters: client for GitHub operations, writer for output
func NewConfigService(client IGitHubClient, writer IOutputWriter) IConfigService

// NewOutputWriter creates a CLI output writer
// Parameters: verbose flag, output destination
func NewOutputWriter(verbose bool, out io.Writer) IOutputWriter

// NewApp creates the main application
// Parameters: configService for operations, parser for input, writer for output
func NewApp(configService IConfigService, parser IRepoParser, writer IOutputWriter) *App
```

---

## CLI Interface Design

### Command Structure

```
ghautodelete [flags] <repository>

Arguments:
  repository    GitHub repository (owner/repo or full URL)

Flags:
  -t, --token string    GitHub personal access token (or set GITHUB_TOKEN)
  -c, --check           Only check current status, don't modify
  -d, --dry-run         Show what would be done without making changes
  -v, --verbose         Enable verbose output
  -h, --help            Show help message
      --version         Show version information

Examples:
  ghautodelete octocat/hello-world
  ghautodelete https://github.com/octocat/hello-world
  ghautodelete git@github.com:octocat/hello-world.git
  ghautodelete --check octocat/hello-world
  ghautodelete --token ghp_xxxx octocat/hello-world
```

### Exit Codes

| Code | Meaning |
|------|---------|
| 0    | Success (setting enabled or already enabled) |
| 1    | General error |
| 2    | Invalid arguments |
| 3    | Authentication failed |
| 4    | Insufficient permissions |
| 5    | Repository not found |
| 6    | API rate limited |

### Output Examples

**Success (newly configured)**:
```
Successfully enabled auto-delete branches for octocat/hello-world
  Default branch: main
  Feature branches will now be deleted after PR merge
```

**Success (already configured)**:
```
Auto-delete branches already enabled for octocat/hello-world
  Default branch: main
  No changes needed
```

**Check mode output**:
```
Repository: octocat/hello-world
Default branch: main
Auto-delete branches: disabled

To enable, run without --check flag
```

**Dry-run output**:
```
[DRY-RUN] Would enable auto-delete branches for octocat/hello-world
  Default branch: main
  No changes made
```

**Error output**:
```
Error: Repository not found: octocat/nonexistent
  Ensure the repository exists and you have access to it
```

---

## GitHub API Integration Approach

### Endpoints Used

1. **GET /repos/{owner}/{repo}** - Fetch repository information
   - Used to check current `delete_branch_on_merge` setting
   - Used to verify repository exists and user has access
   - Returns `default_branch` (main/master)

2. **PATCH /repos/{owner}/{repo}** - Update repository settings
   - Payload: `{"delete_branch_on_merge": true}`
   - Requires `repo` scope on token

3. **GET /user** - Validate token and get user info
   - Used to verify authentication works
   - Response headers contain token scopes

### Required Token Scopes

- `repo` - Full control of private repositories (required)
- For public repos only: `public_repo` may suffice

### Rate Limiting Strategy

```go
type IRateLimiter interface {
    // WaitIfNeeded blocks if rate limit is close to exceeded
    WaitIfNeeded(ctx context.Context, remaining int, resetTime time.Time) error

    // ShouldRetry determines if request should be retried after 403
    ShouldRetry(response IHTTPResponse) (bool, time.Duration)
}
```

- Check `X-RateLimit-Remaining` header after each request
- If remaining < 10, log warning
- On 403 with rate limit exceeded, wait until reset time or fail with clear message
- Implement exponential backoff for transient failures (max 3 retries)

### HTTP Client Configuration

```go
type HTTPClientConfig struct {
    Timeout         time.Duration // Default: 30 seconds
    MaxRetries      int           // Default: 3
    RetryBackoff    time.Duration // Default: 1 second (exponential)
    UserAgent       string        // "ghautodelete/1.0"
}
```

---

## Error Handling Strategy

### Error Hierarchy

```
AppError (base)
  |
  +-- ValidationError (invalid input)
  |     +-- InvalidRepositoryFormat
  |     +-- InvalidTokenFormat
  |
  +-- AuthenticationError
  |     +-- TokenMissing
  |     +-- TokenInvalid
  |     +-- TokenExpired
  |
  +-- AuthorizationError
  |     +-- InsufficientScopes
  |     +-- NoRepoAccess
  |
  +-- ResourceError
  |     +-- RepositoryNotFound
  |     +-- BranchNotFound
  |
  +-- APIError
  |     +-- RateLimited
  |     +-- ServerError
  |     +-- NetworkError
  |
  +-- ConfigurationError
        +-- SettingNotApplied
```

### Error Messages

All errors include:
1. **What happened** - Clear description of the failure
2. **Why it happened** - Context about the cause
3. **How to fix it** - Actionable remediation steps

Example:
```go
&AppError{
    Code:    ErrInsufficientPermissions,
    Message: "Token lacks required permissions",
    Details: "The 'repo' scope is required to modify repository settings. " +
             "Generate a new token at https://github.com/settings/tokens " +
             "with the 'repo' scope selected.",
}
```

### Recovery Strategies

| Error Type | Strategy |
|------------|----------|
| Network timeout | Retry with exponential backoff (3 attempts) |
| Rate limited | Wait until reset time, then retry once |
| 5xx errors | Retry with backoff (3 attempts) |
| 4xx errors | Fail immediately with descriptive message |
| Invalid input | Fail immediately with usage help |

---

## Project Structure

```
ghautodelete/
|-- cmd/
|   +-- ghautodelete/
|       +-- main.go           # Entry point, DI wiring
|
|-- internal/
|   |-- app/
|   |   +-- app.go            # Main application logic
|   |   +-- app_test.go
|   |
|   |-- config/
|   |   +-- service.go        # IConfigService implementation
|   |   +-- service_test.go
|   |   +-- result.go         # ConfigResult
|   |
|   |-- github/
|   |   +-- client.go         # IGitHubClient implementation
|   |   +-- client_test.go
|   |   +-- repository.go     # Repository model
|   |   +-- settings.go       # RepositorySettings model
|   |   +-- token.go          # TokenInfo model
|   |
|   |-- parser/
|   |   +-- repo_parser.go    # IRepoParser implementation
|   |   +-- repo_parser_test.go
|   |
|   |-- token/
|   |   +-- provider.go       # ITokenProvider implementation
|   |   +-- provider_test.go
|   |
|   |-- output/
|   |   +-- writer.go         # IOutputWriter implementation
|   |   +-- writer_test.go
|   |
|   +-- errors/
|       +-- errors.go         # AppError and error types
|       +-- codes.go          # ErrorCode constants
|
|-- pkg/
|   +-- interfaces/
|       +-- interfaces.go     # All interface definitions
|
|-- go.mod
|-- go.sum
|-- README.md
|-- Makefile
+-- .goreleaser.yml           # Release configuration
```

---

## Testing Strategy

### Unit Tests

- Mock all interfaces for isolated testing
- Test each component independently
- Use table-driven tests for parser edge cases
- Target 80%+ code coverage

### Integration Tests

- Use `httptest` for GitHub API simulation
- Test full flow with mock HTTP responses
- Test error scenarios (rate limiting, auth failures)

### Test Doubles

```go
// MockGitHubClient for testing
type MockGitHubClient struct {
    GetRepositoryFunc    func(ctx context.Context, owner, repo string) (IRepository, error)
    UpdateRepositoryFunc func(ctx context.Context, owner, repo string, settings IRepositorySettings) error
    ValidateTokenFunc    func(ctx context.Context) (ITokenInfo, error)
}
```

---

## Dependencies

### External Libraries

| Library | Purpose | Version |
|---------|---------|---------|
| `github.com/spf13/cobra` | CLI framework | v1.8+ |
| `github.com/spf13/viper` | Configuration | v1.18+ |
| `gopkg.in/yaml.v3` | YAML parsing (gh config) | v3.0+ |

### Standard Library Usage

- `net/http` - HTTP client
- `encoding/json` - JSON marshaling
- `context` - Request cancellation
- `os` - Environment variables
- `path/filepath` - File path handling
- `testing` - Unit tests

---

## Security Considerations

1. **Token Storage**: Never log or display tokens
2. **Token Transmission**: HTTPS only for API calls
3. **Token Sources**: Support env vars to avoid CLI history exposure
4. **Permissions**: Request minimum required scopes
5. **Input Validation**: Sanitize repository names before API calls

---

## Future Enhancements (Out of Scope)

- Batch mode for multiple repositories
- GitHub App authentication (vs PAT)
- Organization-wide configuration
- Undo/disable functionality
- Interactive mode with prompts
- Configuration file support

---

## Acceptance Criteria

1. CLI accepts repository in `owner/repo` format
2. CLI accepts full GitHub URLs (HTTPS and SSH)
3. Tool authenticates via token (flag, env var, or gh CLI)
4. Tool enables `delete_branch_on_merge` setting
5. Tool reports if setting was already enabled
6. Tool validates token has required permissions
7. Tool provides clear error messages
8. Tool supports `--check` mode for status only
9. Tool supports `--dry-run` mode
10. Tool exits with appropriate codes

---

**END OF DRAFT SPECIFICATION**
