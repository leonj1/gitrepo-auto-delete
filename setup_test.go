// Package setup_test provides integration-style tests to verify the Go project
// structure is correctly initialized. These tests check for the existence and
// correctness of go.mod, directory structure, Makefile, Dockerfile.test, and main.go.
//
// These tests are designed to FAIL until the project structure is properly created
// by the coder agent. They serve as a specification for what the coder needs to implement.
package setup_test

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// getProjectRoot returns the absolute path to the project root directory.
// The project root is the directory containing go.mod.
func getProjectRoot(t *testing.T) string {
	t.Helper()
	// Get the directory of this test file
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	return wd
}

// TestGoModExists verifies that go.mod exists in the project root.
//
// The implementation should:
// - Create go.mod with module github.com/josejulio/ghautodelete
// - Set Go version to 1.21 or later
func TestGoModExists(t *testing.T) {
	// Arrange
	projectRoot := getProjectRoot(t)
	goModPath := filepath.Join(projectRoot, "go.mod")

	// Act
	_, err := os.Stat(goModPath)

	// Assert
	if os.IsNotExist(err) {
		t.Fatalf("go.mod does not exist at %s - implementation required", goModPath)
	}
	if err != nil {
		t.Fatalf("Error checking go.mod: %v", err)
	}
}

// TestGoModModuleName verifies that go.mod contains the correct module name.
//
// The implementation should:
// - Set module name to github.com/josejulio/ghautodelete
func TestGoModModuleName(t *testing.T) {
	// Arrange
	projectRoot := getProjectRoot(t)
	goModPath := filepath.Join(projectRoot, "go.mod")
	expectedModule := "github.com/josejulio/ghautodelete"

	// Act
	content, err := os.ReadFile(goModPath)
	if err != nil {
		t.Fatalf("Failed to read go.mod: %v", err)
	}

	// Assert
	moduleRegex := regexp.MustCompile(`^module\s+(.+)$`)
	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	var foundModule string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if matches := moduleRegex.FindStringSubmatch(line); matches != nil {
			foundModule = matches[1]
			break
		}
	}

	if foundModule == "" {
		t.Fatalf("No module declaration found in go.mod")
	}
	if foundModule != expectedModule {
		t.Errorf("Module name mismatch: expected %q, got %q", expectedModule, foundModule)
	}
}

// TestGoModVersion verifies that go.mod specifies Go version 1.21 or later.
//
// The implementation should:
// - Set Go version to 1.21 or later in go.mod
func TestGoModVersion(t *testing.T) {
	// Arrange
	projectRoot := getProjectRoot(t)
	goModPath := filepath.Join(projectRoot, "go.mod")
	minMajor := 1
	minMinor := 21

	// Act
	content, err := os.ReadFile(goModPath)
	if err != nil {
		t.Fatalf("Failed to read go.mod: %v", err)
	}

	// Assert
	versionRegex := regexp.MustCompile(`^go\s+(\d+)\.(\d+)`)
	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	var foundVersion bool
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if matches := versionRegex.FindStringSubmatch(line); matches != nil {
			foundVersion = true
			major := parseInt(t, matches[1])
			minor := parseInt(t, matches[2])

			if major < minMajor || (major == minMajor && minor < minMinor) {
				t.Errorf("Go version too old: expected >= %d.%d, got %d.%d",
					minMajor, minMinor, major, minor)
			}
			break
		}
	}

	if !foundVersion {
		t.Fatalf("No Go version declaration found in go.mod")
	}
}

// TestGoModDependencies verifies that required dependencies are declared in go.mod.
//
// The implementation should include:
// - github.com/spf13/cobra v1.8+
// - gopkg.in/yaml.v3 v3.0+
func TestGoModDependencies(t *testing.T) {
	// Arrange
	projectRoot := getProjectRoot(t)
	goModPath := filepath.Join(projectRoot, "go.mod")

	requiredDeps := []struct {
		name       string
		minVersion string
	}{
		{"github.com/spf13/cobra", "v1.8"},
		{"gopkg.in/yaml.v3", "v3.0"},
	}

	// Act
	content, err := os.ReadFile(goModPath)
	if err != nil {
		t.Fatalf("Failed to read go.mod: %v", err)
	}
	contentStr := string(content)

	// Assert
	for _, dep := range requiredDeps {
		t.Run(dep.name, func(t *testing.T) {
			// Check if dependency is present
			if !strings.Contains(contentStr, dep.name) {
				t.Errorf("Required dependency %q not found in go.mod", dep.name)
				return
			}

			// Verify minimum version (basic check - dependency should be present with version >= minVersion)
			depRegex := regexp.MustCompile(regexp.QuoteMeta(dep.name) + `\s+(v[\d.]+)`)
			matches := depRegex.FindStringSubmatch(contentStr)
			if matches == nil {
				t.Errorf("Could not find version for dependency %q", dep.name)
				return
			}

			foundVersion := matches[1]
			if !versionAtLeast(foundVersion, dep.minVersion) {
				t.Errorf("Dependency %q version too old: expected >= %s, got %s",
					dep.name, dep.minVersion, foundVersion)
			}
		})
	}
}

// TestRequiredDirectoriesExist verifies that all required project directories exist.
//
// The implementation should create:
// - cmd/ghautodelete/
// - internal/app/
// - internal/config/
// - internal/github/
// - internal/parser/
// - internal/token/
// - internal/output/
// - internal/errors/
// - pkg/interfaces/
func TestRequiredDirectoriesExist(t *testing.T) {
	// Arrange
	projectRoot := getProjectRoot(t)
	requiredDirs := []string{
		"cmd/ghautodelete",
		"internal/app",
		"internal/config",
		"internal/github",
		"internal/parser",
		"internal/token",
		"internal/output",
		"internal/errors",
		"pkg/interfaces",
	}

	// Act & Assert
	for _, dir := range requiredDirs {
		t.Run(dir, func(t *testing.T) {
			dirPath := filepath.Join(projectRoot, dir)
			info, err := os.Stat(dirPath)

			if os.IsNotExist(err) {
				t.Errorf("Required directory does not exist: %s", dirPath)
				return
			}
			if err != nil {
				t.Errorf("Error checking directory %s: %v", dirPath, err)
				return
			}
			if !info.IsDir() {
				t.Errorf("Path exists but is not a directory: %s", dirPath)
			}
		})
	}
}

// TestMakefileExists verifies that Makefile exists in the project root.
//
// The implementation should:
// - Create a Makefile with standard Go project targets
func TestMakefileExists(t *testing.T) {
	// Arrange
	projectRoot := getProjectRoot(t)
	makefilePath := filepath.Join(projectRoot, "Makefile")

	// Act
	_, err := os.Stat(makefilePath)

	// Assert
	if os.IsNotExist(err) {
		t.Fatalf("Makefile does not exist at %s - implementation required", makefilePath)
	}
	if err != nil {
		t.Fatalf("Error checking Makefile: %v", err)
	}
}

// TestMakefileTargets verifies that Makefile contains all required targets.
//
// The implementation should include targets:
// - build: compile the application
// - test: run tests
// - lint: run linting tools
// - clean: remove build artifacts
// - coverage: generate test coverage report
func TestMakefileTargets(t *testing.T) {
	// Arrange
	projectRoot := getProjectRoot(t)
	makefilePath := filepath.Join(projectRoot, "Makefile")
	requiredTargets := []string{"build", "test", "lint", "clean", "coverage"}

	// Act
	content, err := os.ReadFile(makefilePath)
	if err != nil {
		t.Fatalf("Failed to read Makefile: %v", err)
	}
	contentStr := string(content)

	// Assert
	for _, target := range requiredTargets {
		t.Run(target, func(t *testing.T) {
			// Look for target definition pattern: "target:" at the beginning of a line
			targetRegex := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(target) + `\s*:`)
			if !targetRegex.MatchString(contentStr) {
				t.Errorf("Required Makefile target %q not found", target)
			}
		})
	}
}

// TestDockerfileTestExists verifies that Dockerfile.test exists in the project root.
//
// The implementation should:
// - Create Dockerfile.test for containerized testing
func TestDockerfileTestExists(t *testing.T) {
	// Arrange
	projectRoot := getProjectRoot(t)
	dockerfilePath := filepath.Join(projectRoot, "Dockerfile.test")

	// Act
	_, err := os.Stat(dockerfilePath)

	// Assert
	if os.IsNotExist(err) {
		t.Fatalf("Dockerfile.test does not exist at %s - implementation required", dockerfilePath)
	}
	if err != nil {
		t.Fatalf("Error checking Dockerfile.test: %v", err)
	}
}

// TestDockerfileTestValidity verifies that Dockerfile.test has valid structure.
//
// The implementation should include:
// - FROM instruction with Go image
// - WORKDIR instruction
// - COPY instruction for source files
// - RUN or CMD instruction for running tests
func TestDockerfileTestValidity(t *testing.T) {
	// Arrange
	projectRoot := getProjectRoot(t)
	dockerfilePath := filepath.Join(projectRoot, "Dockerfile.test")

	requiredInstructions := []struct {
		name    string
		pattern string
	}{
		{"FROM", `(?m)^FROM\s+`},
		{"WORKDIR", `(?m)^WORKDIR\s+`},
		{"COPY", `(?m)^COPY\s+`},
	}

	// Act
	content, err := os.ReadFile(dockerfilePath)
	if err != nil {
		t.Fatalf("Failed to read Dockerfile.test: %v", err)
	}
	contentStr := string(content)

	// Assert
	for _, instr := range requiredInstructions {
		t.Run(instr.name, func(t *testing.T) {
			instrRegex := regexp.MustCompile(instr.pattern)
			if !instrRegex.MatchString(contentStr) {
				t.Errorf("Required Dockerfile instruction %q not found", instr.name)
			}
		})
	}

	// Check for either RUN or CMD instruction for test execution
	runOrCmdRegex := regexp.MustCompile(`(?m)^(RUN|CMD)\s+`)
	if !runOrCmdRegex.MatchString(contentStr) {
		t.Errorf("Dockerfile.test must have RUN or CMD instruction for test execution")
	}
}

// TestDockerfileTestUsesGoImage verifies that Dockerfile.test uses a Go base image.
//
// The implementation should:
// - Use golang:1.21 or later as base image
func TestDockerfileTestUsesGoImage(t *testing.T) {
	// Arrange
	projectRoot := getProjectRoot(t)
	dockerfilePath := filepath.Join(projectRoot, "Dockerfile.test")

	// Act
	content, err := os.ReadFile(dockerfilePath)
	if err != nil {
		t.Fatalf("Failed to read Dockerfile.test: %v", err)
	}
	contentStr := string(content)

	// Assert
	goImageRegex := regexp.MustCompile(`(?m)^FROM\s+golang:`)
	if !goImageRegex.MatchString(contentStr) {
		t.Errorf("Dockerfile.test should use a golang base image (e.g., golang:1.21)")
	}
}

// TestMainGoExists verifies that main.go exists in cmd/ghautodelete directory.
//
// The implementation should:
// - Create main.go as the entry point for the application
func TestMainGoExists(t *testing.T) {
	// Arrange
	projectRoot := getProjectRoot(t)
	mainGoPath := filepath.Join(projectRoot, "cmd", "ghautodelete", "main.go")

	// Act
	_, err := os.Stat(mainGoPath)

	// Assert
	if os.IsNotExist(err) {
		t.Fatalf("main.go does not exist at %s - implementation required", mainGoPath)
	}
	if err != nil {
		t.Fatalf("Error checking main.go: %v", err)
	}
}

// TestMainGoHasPackageMain verifies that main.go declares package main.
//
// The implementation should:
// - Declare package main at the top of the file
func TestMainGoHasPackageMain(t *testing.T) {
	// Arrange
	projectRoot := getProjectRoot(t)
	mainGoPath := filepath.Join(projectRoot, "cmd", "ghautodelete", "main.go")

	// Act
	content, err := os.ReadFile(mainGoPath)
	if err != nil {
		t.Fatalf("Failed to read main.go: %v", err)
	}
	contentStr := string(content)

	// Assert
	packageRegex := regexp.MustCompile(`(?m)^package\s+main\s*$`)
	if !packageRegex.MatchString(contentStr) {
		t.Errorf("main.go must declare 'package main'")
	}
}

// TestMainGoHasVersionVariable verifies that main.go contains a version variable.
//
// The implementation should:
// - Declare a version variable (e.g., var version string or var Version string)
// - This allows version to be set at build time via ldflags
func TestMainGoHasVersionVariable(t *testing.T) {
	// Arrange
	projectRoot := getProjectRoot(t)
	mainGoPath := filepath.Join(projectRoot, "cmd", "ghautodelete", "main.go")

	// Act
	content, err := os.ReadFile(mainGoPath)
	if err != nil {
		t.Fatalf("Failed to read main.go: %v", err)
	}
	contentStr := string(content)

	// Assert
	// Look for version variable declaration (case-insensitive for variable name)
	versionRegex := regexp.MustCompile(`(?m)^\s*var\s+[vV]ersion\s+`)
	if !versionRegex.MatchString(contentStr) {
		t.Errorf("main.go must declare a version variable (e.g., 'var version string')")
	}
}

// TestMainGoHasMainFunction verifies that main.go contains a main function.
//
// The implementation should:
// - Define func main() as the entry point
func TestMainGoHasMainFunction(t *testing.T) {
	// Arrange
	projectRoot := getProjectRoot(t)
	mainGoPath := filepath.Join(projectRoot, "cmd", "ghautodelete", "main.go")

	// Act
	content, err := os.ReadFile(mainGoPath)
	if err != nil {
		t.Fatalf("Failed to read main.go: %v", err)
	}
	contentStr := string(content)

	// Assert
	mainFuncRegex := regexp.MustCompile(`(?m)^func\s+main\s*\(\s*\)`)
	if !mainFuncRegex.MatchString(contentStr) {
		t.Errorf("main.go must define 'func main()'")
	}
}

// TestMainGoImportsCobra verifies that main.go imports the cobra library.
//
// The implementation should:
// - Import github.com/spf13/cobra for CLI functionality
func TestMainGoImportsCobra(t *testing.T) {
	// Arrange
	projectRoot := getProjectRoot(t)
	mainGoPath := filepath.Join(projectRoot, "cmd", "ghautodelete", "main.go")

	// Act
	content, err := os.ReadFile(mainGoPath)
	if err != nil {
		t.Fatalf("Failed to read main.go: %v", err)
	}
	contentStr := string(content)

	// Assert
	if !strings.Contains(contentStr, "github.com/spf13/cobra") {
		t.Errorf("main.go must import github.com/spf13/cobra")
	}
}

// TestReadmeExists verifies that README.md exists in the project root.
//
// The implementation should:
// - Create README.md with project documentation
func TestReadmeExists(t *testing.T) {
	// Arrange
	projectRoot := getProjectRoot(t)
	readmePath := filepath.Join(projectRoot, "README.md")

	// Act
	_, err := os.Stat(readmePath)

	// Assert
	if os.IsNotExist(err) {
		t.Fatalf("README.md does not exist at %s - implementation required", readmePath)
	}
	if err != nil {
		t.Fatalf("Error checking README.md: %v", err)
	}
}

// Helper function: parseInt converts a string to int
func parseInt(t *testing.T, s string) int {
	t.Helper()
	var result int
	for _, c := range s {
		if c < '0' || c > '9' {
			break
		}
		result = result*10 + int(c-'0')
	}
	return result
}

// Helper function: versionAtLeast checks if version a is >= version b
// Versions are expected in format "v1.2" or "v1.2.3"
func versionAtLeast(a, b string) bool {
	// Strip 'v' prefix if present
	a = strings.TrimPrefix(a, "v")
	b = strings.TrimPrefix(b, "v")

	partsA := strings.Split(a, ".")
	partsB := strings.Split(b, ".")

	for i := 0; i < len(partsB); i++ {
		if i >= len(partsA) {
			return false
		}

		// Extract numeric part only (handle versions like "1.8.0")
		numA := extractNumber(partsA[i])
		numB := extractNumber(partsB[i])

		if numA > numB {
			return true
		}
		if numA < numB {
			return false
		}
	}
	return true
}

// Helper function: extractNumber extracts the leading numeric part of a string
func extractNumber(s string) int {
	var result int
	for _, c := range s {
		if c < '0' || c > '9' {
			break
		}
		result = result*10 + int(c-'0')
	}
	return result
}
