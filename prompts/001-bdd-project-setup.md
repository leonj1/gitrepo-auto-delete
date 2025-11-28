---
executor: bdd
source_feature: project-setup
---

<objective>
Initialize the Go project with proper module structure, Makefile, and test infrastructure.
This is the foundational setup that all other components will build upon.
</objective>

<requirements>
Based on the Draft specification and BDD requirements, set up:

1. Go Module Initialization
   - Create `go.mod` with module `github.com/josejulio/ghautodelete`
   - Set Go version to 1.21 or later
   - Add required dependencies:
     - `github.com/spf13/cobra` v1.8+
     - `gopkg.in/yaml.v3` v3.0+

2. Project Directory Structure
   ```
   ghautodelete/
   |-- cmd/
   |   +-- ghautodelete/
   |       +-- main.go
   |-- internal/
   |   |-- app/
   |   |-- config/
   |   |-- github/
   |   |-- parser/
   |   |-- token/
   |   |-- output/
   |   +-- errors/
   |-- pkg/
   |   +-- interfaces/
   |-- Makefile
   |-- Dockerfile.test
   +-- README.md
   ```

3. Makefile with targets:
   - `build` - Build the binary
   - `test` - Run all tests
   - `lint` - Run linter (golangci-lint)
   - `clean` - Clean build artifacts
   - `coverage` - Generate test coverage report

4. Dockerfile.test for containerized testing:
   - Based on golang:1.21-alpine or similar
   - Copy source code
   - Run `make test`

5. Basic main.go placeholder:
   - Import cobra
   - Set up root command stub
   - Version variable for ldflags injection

</requirements>

<context>
Draft Specification: specs/DRAFT-github-auto-delete-branches.md
Gap Analysis: specs/GAP-ANALYSIS.md

This is a greenfield project - all code must be created from scratch.

Project Structure from DRAFT spec (lines 604-651):
- cmd/ghautodelete/main.go - Entry point
- internal/ - Private implementation packages
- pkg/interfaces/ - Public interface definitions
</context>

<implementation>
Follow TDD approach:
1. Create tests that verify project structure exists
2. Create tests that verify Makefile targets work
3. Implement the structure to pass tests

Key Points:
- Use Go modules (go mod init)
- Follow Go project layout conventions
- Ensure all directories have .gitkeep or placeholder files
- Makefile should be self-documenting with help target
</implementation>

<verification>
Project setup verification checklist:
- [ ] go.mod exists with correct module name
- [ ] go.sum exists after dependency installation
- [ ] All directories in structure exist
- [ ] Makefile exists with required targets
- [ ] `make build` compiles successfully
- [ ] `make test` runs (even with no tests yet)
- [ ] Dockerfile.test can build
- [ ] main.go compiles and shows version
</verification>

<success_criteria>
- Go module properly initialized
- All directories created per spec
- Makefile functional with all targets
- Dockerfile.test buildable
- `go build ./...` succeeds
- `go test ./...` runs without error
</success_criteria>
