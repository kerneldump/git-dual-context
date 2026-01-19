# Implementation Plan - Maintainability & Testing Improvements

## Phase 1: High Priority (Reliability & Stability)
- [ ] Task: Create `internal/analyzer/retry_test.go` file skeleton.
- [ ] Task: Implement `TestIsRetryable` with table-driven tests covering all specified error cases.
- [ ] Task: Implement `TestWithRetryExponentialBackoff` to verify delay logic.
- [ ] Task: Implement `TestWithRetryMaxRetries` to verify failure cap.
- [ ] Task: Add `recover()` logic to worker goroutines in `cmd/git-commit-analysis/main.go` to safely handle panics.

## Phase 2: Medium Priority (Code Quality & Tests)
- [ ] Task: Define `ChunkEqual`, `ChunkAdd`, `ChunkDelete` constants in `internal/gitdiff/diff.go`.
- [ ] Task: Replace magic numbers (0, 1, 2) in `internal/gitdiff/diff.go` with the new constants.
- [ ] Task: Create helper method `encode(v interface{})` in `orderedPrinter` to handle JSON errors and log to stderr.
- [ ] Task: Replace all direct `p.encoder.Encode` calls with `p.encode` in `cmd/git-commit-analysis/main.go`.
- [ ] Task: Refactor `tempDir` global variable into a `CleanupManager` struct in `cmd/git-commit-analysis/main.go` and update usage.
- [ ] Task: Add `TestFindJSONBlockEdgeCases` to `internal/analyzer/engine_test.go` with comprehensive scenarios.

## Phase 3: Low Priority (Polish & Benchmarks)
- [ ] Task: Reorder fields in `orderedPrinter` struct in `cmd/git-commit-analysis/main.go` for better memory alignment.
- [ ] Task: Add `BenchmarkTruncateDiff` and `BenchmarkShouldIgnoreFile` to `internal/gitdiff/diff_test.go`.
- [ ] Task: Add `BenchmarkFindJSONBlock` to `internal/analyzer/engine_test.go`.
- [ ] Task: Fix documentation scenario references in `docs/GitCommitAnalysis.md` (change B, C, D to 2, 3, 4).
- [ ] Task: Wrap error in `GetStandardDiff` in `internal/gitdiff/diff.go` with `fmt.Errorf` and `%w` for better context.
