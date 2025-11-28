---
executor: bdd
source_feature: integration
---

<objective>
Wire together all components in the main function and App struct to create a
fully functional CLI application. This integrates all previously built components.
</objective>

<requirements>
Based on all BDD scenarios and the Draft specification, complete the integration:

1. App struct in `internal/app/app.go`:
   ```go
   type App struct {
       configService IConfigService
       parser        IRepoParser
       writer        IOutputWriter
       tokenProvider ITokenProvider
       options       *CLIOptions
   }
   ```

2. NewApp constructor:
   - `NewApp(configService IConfigService, parser IRepoParser, writer IOutputWriter) *App`
   - Wire all dependencies

3. App.Run method:
   - Parse repository input
   - Determine operation mode (check, dry-run, or configure)
   - Execute appropriate operation
   - Return error (or nil on success)

4. Main function wiring in `cmd/ghautodelete/main.go`:
   ```go
   func main() {
       if err := run(); err != nil {
           os.Exit(errors.GetExitCode(err))
       }
   }

   func run() error {
       // Parse CLI flags with Cobra
       // Create dependencies
       // Wire up App
       // Execute App.Run
       return nil
   }
   ```

5. Dependency creation order:
   a. Parse CLI flags into CLIOptions
   b. Create TokenProvider with explicit token
   c. Get token from provider
   d. Create HTTP client with timeout
   e. Create GitHubClient with token
   f. Create OutputWriter with verbose flag
   g. Create ConfigService with client and writer
   h. Create RepoParser
   i. Create App with all dependencies
   j. Run App

6. Error handling flow:
   - Cobra handles --help and --version
   - Token provider errors -> exit 3
   - Parser errors -> exit 2
   - GitHub client errors -> appropriate exit code
   - ConfigService errors -> appropriate exit code

7. Context handling:
   - Create context with timeout (30 seconds default)
   - Pass context through all operations
   - Handle context cancellation gracefully
</requirements>

<context>
Draft Specification: specs/DRAFT-github-auto-delete-branches.md
Gap Analysis: specs/GAP-ANALYSIS.md

Constructor signatures from DRAFT spec (lines 378-404):
- NewGitHubClient(httpClient, baseURL, token)
- NewRepoParser()
- NewTokenProvider(explicitToken)
- NewConfigService(client, writer)
- NewOutputWriter(verbose, out)
- NewApp(configService, parser, writer)

Main application flow from DRAFT spec (lines 213-258):
- Parse CLI -> Validate repo -> Get token -> Validate token
- Fetch repo -> Check setting -> Update if needed -> Verify
</context>

<implementation>
Follow TDD approach:
1. Write integration tests for full flow
2. Test successful configuration flow
3. Test already-enabled flow
4. Test check mode flow
5. Test dry-run flow
6. Test error flows (auth, not found, etc.)
7. Implement main wiring to pass tests

Architecture Guidelines:
- Keep main.go minimal (just wiring)
- Put application logic in App struct
- Use dependency injection for testability
- Handle signals for graceful shutdown (optional)
- Use context for cancellation
</implementation>

<verification>
Full application verification checklist:
- [ ] `make build` produces working binary
- [ ] `./ghautodelete --help` shows help
- [ ] `./ghautodelete --version` shows version
- [ ] Mock GitHub API tests pass for all modes
- [ ] Exit codes are correct for all scenarios
- [ ] Error messages are clear and actionable
- [ ] Verbose mode shows detailed output
- [ ] Non-verbose mode shows only essential output
</verification>

<success_criteria>
- All components properly wired together
- Full end-to-end flow works
- All BDD scenarios from all feature files pass
- `make test` passes with >80% coverage
- `make build` produces working binary
- Binary runs correctly with mock data
- Error handling is comprehensive
- Code follows project standards
</success_criteria>
