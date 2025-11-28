---
executor: bdd
source_feature: interfaces
---

<objective>
Define all interface contracts and data models as specified in the Draft specification.
These interfaces will be implemented by subsequent components and enable dependency injection for testing.
</objective>

<requirements>
Based on the Draft specification, create the following interfaces in `pkg/interfaces/interfaces.go`:

1. Core Service Interfaces:
   - `IGitHubClient` - GitHub API operations (GetRepository, UpdateRepository, ValidateToken)
   - `IRepoParser` - Repository identifier parsing (Parse method)
   - `ITokenProvider` - Token retrieval (GetToken method)
   - `IOutputWriter` - CLI output formatting (Success, Error, Info, Verbose methods)
   - `IConfigService` - Configuration operations (Configure, CheckStatus methods)

2. Data Model Interfaces:
   - `IRepository` - Repository data (GetOwner, GetName, GetDefaultBranch, GetDeleteBranchOnMerge, GetFullName)
   - `IRepositorySettings` - Updateable settings (GetDeleteBranchOnMerge)
   - `ITokenInfo` - Token metadata (GetScopes, HasScope, GetUsername)
   - `IConfigResult` - Operation result (WasAlreadyEnabled, IsNowEnabled, GetDefaultBranch, GetRepositoryFullName)

3. Data Model Implementations in `internal/` packages:
   - `Repository` struct implementing `IRepository`
   - `RepositorySettings` struct implementing `IRepositorySettings`
   - `TokenInfo` struct implementing `ITokenInfo`
   - `ConfigResult` struct implementing `IConfigResult`

4. CLI Options struct:
   - `CLIOptions` with Repository, Token, Verbose, DryRun, CheckOnly fields

</requirements>

<context>
Draft Specification: specs/DRAFT-github-auto-delete-branches.md
Gap Analysis: specs/GAP-ANALYSIS.md

Interface definitions from DRAFT spec (lines 19-88):
- IGitHubClient with context-aware methods
- IRepository with getter methods
- IRepositorySettings for updates
- ITokenInfo for scope validation
- IRepoParser for input parsing
- ITokenProvider for token retrieval
- IOutputWriter for formatted output
- IConfigService for main operations
- IConfigResult for operation results

Data models from DRAFT spec (lines 96-173):
- Repository, RepositorySettings, TokenInfo, ConfigResult structs
- CLIOptions struct
</context>

<implementation>
Follow TDD approach:
1. Write tests verifying interface method signatures
2. Write tests verifying struct implementations satisfy interfaces
3. Implement interfaces and structs

Architecture Guidelines:
- Keep interfaces in `pkg/interfaces/` for external access
- Keep struct implementations in their respective `internal/` packages
- Use pointer receivers for struct methods
- Ensure all getter methods are exported
</implementation>

<verification>
Interface implementation verification:
- [ ] All interfaces defined in pkg/interfaces/interfaces.go
- [ ] Repository struct implements IRepository
- [ ] RepositorySettings struct implements IRepositorySettings
- [ ] TokenInfo struct implements ITokenInfo
- [ ] ConfigResult struct implements IConfigResult
- [ ] All interface methods have correct signatures per spec
- [ ] Compile-time interface satisfaction checks pass
</verification>

<success_criteria>
- All interfaces from spec defined correctly
- All struct implementations satisfy their interfaces
- Code compiles without errors
- Interface satisfaction verified at compile time
- Unit tests pass for all data models
</success_criteria>
