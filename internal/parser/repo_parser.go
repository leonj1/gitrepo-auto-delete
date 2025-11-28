// Package parser provides functionality for parsing repository identifiers.
//
// The repository parser handles multiple input formats:
// - Simple format: "owner/repo"
// - HTTPS GitHub URL: "https://github.com/owner/repo[.git]"
// - SSH GitHub URL: "git@github.com:owner/repo[.git]"
package parser

import (
	"regexp"
	"strings"

	"github.com/josejulio/ghautodelete/internal/errors"
	"github.com/josejulio/ghautodelete/pkg/interfaces"
)

// RepoParser implements the IRepoParser interface for parsing repository identifiers.
//
// It handles multiple repository identifier formats and validates owner/repo names
// according to GitHub's naming rules.
type RepoParser struct{}

// NewRepoParser creates a new RepoParser instance.
//
// Returns a RepoParser that implements the IRepoParser interface.
func NewRepoParser() interfaces.IRepoParser {
	return &RepoParser{}
}

// validNamePattern is a regex pattern for valid GitHub owner and repository names.
// GitHub allows alphanumeric characters, hyphens, underscores, and dots.
var validNamePattern = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

// Parse extracts owner and repository name from a repository identifier.
//
// Supported formats:
//   - Simple: "owner/repo"
//   - HTTPS: "https://github.com/owner/repo[.git]"
//   - SSH: "git@github.com:owner/repo[.git]"
//
// The function trims whitespace and validates owner/repo names.
//
// Returns:
//   - owner: The repository owner (user or organization)
//   - repo: The repository name
//   - err: An AppError if parsing or validation fails
//
// Errors:
//   - "Repository identifier is required" for empty input
//   - "Expected format: owner/repo" for invalid format or non-GitHub URLs
//   - "Invalid repository name characters" for invalid characters
//   - "Invalid GitHub URL format" for malformed GitHub URLs
func (p *RepoParser) Parse(input string) (owner string, repo string, err error) {
	// Trim whitespace
	input = strings.TrimSpace(input)

	// Check for empty input
	if input == "" {
		return "", "", errors.NewValidationError("Repository identifier is required")
	}

	// Try to parse as HTTPS URL
	if strings.HasPrefix(strings.ToLower(input), "https://") {
		return p.parseHTTPSURL(input)
	}

	// Reject HTTP URLs (not HTTPS)
	if strings.HasPrefix(strings.ToLower(input), "http://") {
		return "", "", errors.NewValidationError("Expected format: owner/repo")
	}

	// Try to parse as SSH URL (must be git@ not any other user)
	if strings.HasPrefix(input, "git@") {
		return p.parseSSHURL(input)
	}

	// Reject other SSH-like formats (e.g., user@github.com:)
	// Only check this if not an HTTP/HTTPS URL
	if strings.Contains(input, "@") && strings.Contains(input, ":") {
		return "", "", errors.NewValidationError("Expected format: owner/repo")
	}

	// Try to parse as simple owner/repo format
	return p.parseSimpleFormat(input)
}

// parseSimpleFormat parses "owner/repo" format.
func (p *RepoParser) parseSimpleFormat(input string) (owner string, repo string, err error) {
	parts := strings.Split(input, "/")

	// Check if any part is empty (catches double slashes, leading/trailing slashes)
	// This needs to be checked before the count check
	for _, part := range parts {
		if part == "" {
			return "", "", errors.NewValidationError("Invalid repository name characters")
		}
	}

	// Must have exactly 2 parts
	if len(parts) != 2 {
		return "", "", errors.NewValidationError("Expected format: owner/repo")
	}

	owner = parts[0]
	repo = parts[1]

	// Validate owner and repo
	if err := p.validateName(owner); err != nil {
		return "", "", err
	}
	if err := p.validateName(repo); err != nil {
		return "", "", err
	}

	return owner, repo, nil
}

// parseHTTPSURL parses HTTPS GitHub URLs.
//
// Supported formats:
//   - https://github.com/owner/repo
//   - https://github.com/owner/repo.git
//   - https://github.com/owner/repo/
//   - HTTPS://github.com/owner/repo (case-insensitive scheme)
//   - https://GitHub.com/owner/repo (case-insensitive domain)
func (p *RepoParser) parseHTTPSURL(input string) (owner string, repo string, err error) {
	// Remove scheme (case-insensitive)
	lowerInput := strings.ToLower(input)
	if !strings.HasPrefix(lowerInput, "https://") {
		return "", "", errors.NewValidationError("Expected format: owner/repo")
	}

	// Extract the part after the scheme
	afterScheme := input[8:] // len("https://") = 8

	// Check for github.com domain (case-insensitive)
	lowerAfterScheme := strings.ToLower(afterScheme)
	if !strings.HasPrefix(lowerAfterScheme, "github.com/") {
		return "", "", errors.NewValidationError("Expected format: owner/repo")
	}

	// Extract the path after github.com/
	path := afterScheme[11:] // len("github.com/") = 11

	// Remove trailing slash if present
	path = strings.TrimSuffix(path, "/")

	// Remove .git suffix if present
	path = strings.TrimSuffix(path, ".git")

	// Split path into parts
	parts := strings.Split(path, "/")

	// Must have exactly 2 parts (owner and repo)
	if len(parts) != 2 {
		return "", "", errors.NewValidationError("Invalid GitHub URL format")
	}

	owner = parts[0]
	repo = parts[1]

	// Validate owner and repo
	if err := p.validateName(owner); err != nil {
		return "", "", err
	}
	if err := p.validateName(repo); err != nil {
		return "", "", err
	}

	return owner, repo, nil
}

// parseSSHURL parses SSH GitHub URLs.
//
// Supported formats:
//   - git@github.com:owner/repo
//   - git@github.com:owner/repo.git
func (p *RepoParser) parseSSHURL(input string) (owner string, repo string, err error) {
	// Check for correct SSH format
	if !strings.HasPrefix(input, "git@github.com:") {
		return "", "", errors.NewValidationError("Expected format: owner/repo")
	}

	// Extract the path after git@github.com:
	path := input[15:] // len("git@github.com:") = 15

	// Remove .git suffix if present
	path = strings.TrimSuffix(path, ".git")

	// Split path into parts
	parts := strings.Split(path, "/")

	// Must have exactly 2 parts (owner and repo)
	if len(parts) != 2 {
		return "", "", errors.NewValidationError("Invalid GitHub URL format")
	}

	owner = parts[0]
	repo = parts[1]

	// Validate owner and repo
	if err := p.validateName(owner); err != nil {
		return "", "", err
	}
	if err := p.validateName(repo); err != nil {
		return "", "", err
	}

	return owner, repo, nil
}

// validateName validates that a name (owner or repo) contains only valid characters.
//
// Valid characters are: alphanumeric, hyphen, underscore, and dot.
// The name must not be empty.
func (p *RepoParser) validateName(name string) error {
	if name == "" {
		return errors.NewValidationError("Invalid repository name characters")
	}

	if !validNamePattern.MatchString(name) {
		return errors.NewValidationError("Invalid repository name characters")
	}

	return nil
}
