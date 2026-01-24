package gitdiff

import (
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/object"
)

const (
	// MaxDiffSize is the maximum size of a diff in characters
	MaxDiffSize = 50000
	// TruncationMarker is appended when diffs are truncated
	TruncationMarker = "\n... [truncated: diff too large] ...\n"
	// defaultDiffBufferSize is the pre-allocation size for diff string builders
	defaultDiffBufferSize = 8192
)

// TruncateDiff limits diff size to prevent context window overflow
func TruncateDiff(diff string, maxSize int) string {
	if len(diff) <= maxSize {
		return diff
	}

	// Ensure we have enough room for the marker
	markerLen := len(TruncationMarker)
	if maxSize <= markerLen {
		// If maxSize is too small, just return what we can
		if maxSize <= 0 {
			return ""
		}
		return diff[:maxSize]
	}

	// Try to truncate at a line boundary
	truncateAt := maxSize - markerLen
	lastNewline := strings.LastIndex(diff[:truncateAt], "\n")
	if lastNewline > truncateAt/2 {
		truncateAt = lastNewline
	}

	return diff[:truncateAt] + TruncationMarker
}

// GetStandardDiff returns the diff string and a list of modified file paths
func GetStandardDiff(c, parent *object.Commit) (string, []string, error) {
	cTree, err := c.Tree()
	if err != nil {
		return "", nil, err
	}

	var pTree *object.Tree
	if parent != nil {
		pTree, err = parent.Tree()
		if err != nil {
			return "", nil, err
		}
	}

	// Diff parent -> commit
	// For the first commit (no parent), pTree will be nil
	// Use DiffTree which handles nil trees correctly (treats as empty tree)
	changes, err := object.DiffTree(pTree, cTree)
	if err != nil {
		return "", nil, fmt.Errorf("failed to diff trees: %w", err)
	}

	patch, err := changes.Patch()
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate patch: %w", err)
	}

	var sb strings.Builder
	sb.Grow(defaultDiffBufferSize)
	var files []string

	for _, fp := range patch.FilePatches() {
		if fp.IsBinary() {
			continue
		}
		from, to := fp.Files()
		path := ""
		if from != nil {
			path = from.Path()
		}
		if to != nil {
			path = to.Path()
		}

		// Filter out irrelevant files to save tokens and reduce noise
		if ShouldIgnoreFile(path) {
			continue
		}

		if path != "" {
			files = append(files, path)
			sb.WriteString(fmt.Sprintf("--- %s\n", path))
			for _, chunk := range fp.Chunks() {
				content := chunk.Content()
				if len(content) == 0 {
					continue
				}
				op := " "
				switch chunk.Type() {
				case 0: // Equal (context)
					op = " "
				case 1: // Add
					op = "+"
				case 2: // Delete
					op = "-"
				}
				lines := strings.Split(content, "\n")
				for _, line := range lines {
					if line == "" {
						continue
					}
					sb.WriteString(fmt.Sprintf("%s%s\n", op, line))
				}
			}
		}
	}

	result := sb.String()
	return TruncateDiff(result, MaxDiffSize), files, nil
}

// GetFullDiff returns the diff between the commit and HEAD, restricted to the provided files
func GetFullDiff(c, head *object.Commit, filterFiles []string) (string, error) {
	cTree, err := c.Tree()
	if err != nil {
		return "", err
	}
	headTree, err := head.Tree()
	if err != nil {
		return "", err
	}

	// Diff commit -> head (shows what happened *after* the commit)
	patch, err := cTree.Patch(headTree)
	if err != nil {
		return "", err
	}

	// Pre-size the map
	fileSet := make(map[string]bool, len(filterFiles))
	for _, f := range filterFiles {
		fileSet[f] = true
	}

	var sb strings.Builder
	sb.Grow(defaultDiffBufferSize)

	for _, fp := range patch.FilePatches() {
		from, to := fp.Files()
		path := ""
		if from != nil {
			path = from.Path()
		}
		if to != nil {
			path = to.Path()
		}

		if fileSet[path] && !fp.IsBinary() {
			sb.WriteString(fmt.Sprintf("--- %s (Evolution to HEAD)\n", path))
			for _, chunk := range fp.Chunks() {
				content := chunk.Content()
				if len(content) == 0 {
					continue
				}
				op := " "
				switch chunk.Type() {
				case 0: // Equal (context)
					op = " "
				case 1: // Add
					op = "+"
				case 2: // Delete
					op = "-"
				}
				lines := strings.Split(content, "\n")
				for _, line := range lines {
					if line == "" {
						continue
					}
					sb.WriteString(fmt.Sprintf("%s%s\n", op, line))
				}
			}
		}
	}

	if sb.Len() == 0 {
		return "No further changes to these files since this commit.", nil
	}

	result := sb.String()
	return TruncateDiff(result, MaxDiffSize), nil
}

// ShouldIgnoreFile returns true if the file should be skipped during analysis
func ShouldIgnoreFile(path string) bool {
	// Normalize path separators
	path = strings.ReplaceAll(path, "\\", "/")

	// 1. Lock files and checksums
	lockFiles := []string{
		"go.sum", "package-lock.json", "yarn.lock", "Gemfile.lock",
		"poetry.lock", "pnpm-lock.yaml", "Cargo.lock", "composer.lock",
		"Pipfile.lock", "shrinkwrap.yaml",
	}
	for _, lf := range lockFiles {
		if strings.HasSuffix(path, lf) {
			return true
		}
	}

	// 2. Test files
	testPatterns := []string{
		"_test.go", ".test.js", ".test.ts", ".spec.js", ".spec.ts",
		"_test.py", "_spec.rb",
	}
	for _, tp := range testPatterns {
		if strings.HasSuffix(path, tp) {
			return true
		}
	}
	// Python test files with test_ prefix
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		filename := parts[len(parts)-1]
		if strings.HasPrefix(filename, "test_") && strings.HasSuffix(filename, ".py") {
			return true
		}
	}

	// 3. Directories to ignore
	ignoreDirs := []string{
		"vendor/", "node_modules/", "dist/", "build/", "out/",
		".idea/", ".vscode/", ".git/",
		"__pycache__/", ".pytest_cache/", ".tox/",
	}
	for _, dir := range ignoreDirs {
		if strings.Contains(path, dir) {
			return true
		}
	}

	// 4. CI/CD files (usually don't cause runtime bugs)
	if strings.HasPrefix(path, ".github/") ||
		strings.HasPrefix(path, ".gitlab/") ||
		strings.HasPrefix(path, ".circleci/") ||
		path == ".gitlab-ci.yml" ||
		path == ".travis.yml" {
		return true
	}

	return false
}
