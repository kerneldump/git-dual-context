// Package analyzer constants for git-dual-context
package analyzer

import "time"

// Default configuration values
const (
	// DefaultModel is the default LLM model name (without "models/" prefix)
	DefaultModel = "gemini-flash-latest"

	// DefaultNumCommits is the default number of commits to analyze
	DefaultNumCommits = 5

	// DefaultNumWorkers is the default number of concurrent workers
	DefaultNumWorkers = 3

	// DefaultTimeout is the default timeout per commit analysis
	DefaultTimeout = 10 * time.Minute

	// DefaultTemperature is the default LLM temperature for deterministic output
	DefaultTemperature float32 = 0.1

	// DefaultCommitMessageMaxLength is the max length for truncated commit messages
	DefaultCommitMessageMaxLength = 80
)

// Diff-related constants
const (
	// DefaultDiffBufferSize is the pre-allocation size for diff string builders
	DefaultDiffBufferSize = 8192
)
