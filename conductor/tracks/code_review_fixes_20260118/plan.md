# Implementation Plan: Code Review Fixes

## Phase 1: Critical Bug Fixes (P0)
- [x] Task: Fix silently ignored error in `c.Parent(0)` call in `engine.go:91-93`
- [x] Task: Add `found` flag to detect missing text content in LLM response in `engine.go:124-143`
- [x] Task: Update `AnalyzeCommit` signature to accept `headCommit` instead of `headHash` for performance

## Phase 2: High Priority Improvements (P1)
- [x] Task: Create `internal/analyzer/retry.go` with `RetryConfig`, `IsRetryable`, and `WithRetry` functions
- [x] Task: Add context lines (case 0: Equal) to diff output in `GetStandardDiff` and `GetFullDiff`
- [x] Task: Implement ordered streaming output with `orderedPrinter` type in `main.go`

## Phase 3: Medium Priority Improvements (P2)
- [x] Task: Add `-model` flag for configurable Gemini model selection
- [x] Task: Add `-timeout` flag for configurable per-commit timeout
- [x] Task: Implement graceful shutdown with signal handling (SIGINT, SIGTERM)
- [x] Task: Expand `ShouldIgnoreFile` with more lock files, test patterns, CI/CD, build dirs
- [x] Task: Add `TruncateDiff` function with 50KB limit and truncation marker

## Phase 4: Low Priority Enhancements (P3)
- [x] Task: Add `-o` flag for output file option
- [x] Task: Add `-branch` flag for branch selection
- [x] Task: Add `Summary` type and output summary at end of analysis
- [x] Task: Add `-v` flag for verbose/debug output
- [x] Task: Add `Message` field to `JSONResult` and update `ToJSONResult` signature
- [x] Task: Remove redundant HEAD commit fetch from `AnalyzeCommit` (now passed as parameter)

## Phase 5: Testing & Quality
- [x] Task: Create `internal/gitdiff/diff_test.go` with `TestShouldIgnoreFile` test cases
- [x] Task: Add `TestTruncateDiff` and `TestTruncateDiffPreservesLineBreaks` tests
- [x] Task: Update existing `engine_test.go` tests to match new `ToJSONResult` signature
- [x] Task: Add `TestToJSONResultTruncatesLongMessage` test

## Phase 6: Security & Performance
- [x] Task: Add API key command-line warning in `main.go`
- [x] Task: Implement temp directory cleanup on fatal exit
- [x] Task: Add `sb.Grow(8192)` pre-allocation to StringBuilder in diff functions
- [x] Task: Pre-size `fileSet` map with `make(map[string]bool, len(filterFiles))`

## Phase 7: Documentation
- [x] Task: Update README.md with new CLI options and features
- [x] Task: Update conductor tracks.md with this track
- [x] Task: Verify all tests pass and binary builds successfully
