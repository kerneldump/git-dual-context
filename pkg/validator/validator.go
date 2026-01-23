// Package validator provides input validation utilities for security and safety
package validator

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	// MaxCommits is the maximum number of commits that can be analyzed in one run
	MaxCommits = 1000
	// MaxWorkers is the maximum number of concurrent workers allowed
	MaxWorkers = 50
)

var (
	// branchNameRegex validates branch names according to git conventions
	// Allows alphanumeric, hyphens, underscores, forward slashes (for feature branches), and dots (for versions)
	branchNameRegex = regexp.MustCompile(`^[a-zA-Z0-9/_.-]+$`)
)

// ValidateNumCommits checks if the number of commits is within reasonable bounds
func ValidateNumCommits(n int) error {
	if n <= 0 {
		return fmt.Errorf("number of commits must be positive, got %d", n)
	}
	if n > MaxCommits {
		return fmt.Errorf("number of commits exceeds maximum of %d, got %d", MaxCommits, n)
	}
	return nil
}

// ValidateNumWorkers checks if the number of workers is within reasonable bounds
func ValidateNumWorkers(n int) error {
	if n <= 0 {
		return fmt.Errorf("number of workers must be positive, got %d", n)
	}
	if n > MaxWorkers {
		return fmt.Errorf("number of workers exceeds maximum of %d, got %d", MaxWorkers, n)
	}
	return nil
}

// ValidateBranchName checks if a branch name is valid and safe
func ValidateBranchName(branch string) error {
	if branch == "" {
		return nil // Empty is allowed (means use HEAD)
	}

	// Check for suspicious patterns
	if strings.Contains(branch, "..") {
		return fmt.Errorf("branch name contains suspicious pattern '..'")
	}
	if strings.HasPrefix(branch, "-") {
		return fmt.Errorf("branch name cannot start with '-'")
	}
	if strings.HasPrefix(branch, "/") || strings.HasSuffix(branch, "/") {
		return fmt.Errorf("branch name cannot start or end with '/'")
	}

	// Validate against regex
	if !branchNameRegex.MatchString(branch) {
		return fmt.Errorf("branch name contains invalid characters: %s", branch)
	}

	return nil
}

// ValidateRepoPath validates that a repository path is safe to use
// It checks for directory traversal attempts and other suspicious patterns
func ValidateRepoPath(path string) error {
	if path == "" {
		return fmt.Errorf("repository path cannot be empty")
	}

	// Allow remote URLs
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") || strings.HasPrefix(path, "git@") {
		return nil // Remote URLs are handled by git itself
	}

	// Clean the path to resolve any . or .. components
	cleanPath := filepath.Clean(path)

	// Check for suspicious patterns
	if strings.Contains(path, "..") {
		return fmt.Errorf("repository path contains suspicious pattern '..'")
	}

	// Warn about sensitive directories (but don't block - user might legitimately analyze these)
	sensitivePaths := []string{"/etc", "/sys", "/proc", "/dev"}
	for _, sensitive := range sensitivePaths {
		if strings.HasPrefix(cleanPath, sensitive) {
			return fmt.Errorf("repository path points to sensitive system directory: %s", cleanPath)
		}
	}

	return nil
}

// ValidateErrorMessage ensures the error message is not empty
func ValidateErrorMessage(msg string) error {
	if strings.TrimSpace(msg) == "" {
		return fmt.Errorf("error message cannot be empty")
	}
	return nil
}
