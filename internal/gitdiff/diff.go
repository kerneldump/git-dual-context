package gitdiff

import (
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/object"
)

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
	patch, err := pTree.Patch(cTree)
	if err != nil {
		return "", nil, err
	}

	var sb strings.Builder
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

	return sb.String(), files, nil
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

	fileSet := make(map[string]bool)
	for _, f := range filterFiles {
		fileSet[f] = true
	}

	var sb strings.Builder
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

	return sb.String(), nil
}

// ShouldIgnoreFile returns true if the file should be skipped during analysis
func ShouldIgnoreFile(path string) bool {
	// 1. Lock files and checksums
	if strings.HasSuffix(path, "go.sum") ||
		strings.HasSuffix(path, "package-lock.json") ||
		strings.HasSuffix(path, "yarn.lock") ||
		strings.HasSuffix(path, "Gemfile.lock") {
		return true
	}

	// 2. Test files (unless specifically debugging tests, usually noise for logic bugs)
	if strings.HasSuffix(path, "_test.go") ||
		strings.HasSuffix(path, ".test.js") ||
		strings.HasSuffix(path, ".spec.ts") {
		return true
	}

	// 3. Vendor directory
	if strings.HasPrefix(path, "vendor/") {
		return true
	}

	return false
}
