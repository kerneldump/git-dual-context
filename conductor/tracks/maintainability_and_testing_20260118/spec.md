# Specification: Maintainability & Testing Improvements

## Overview
This track addresses the findings from the code review dated January 18, 2026. The focus is on improving system reliability (high priority), code standards (medium priority), and minor optimizations (low priority).

## Goals
1.  **Reliability:** Eliminate silent failures in worker goroutines and ensure retry logic is fully tested.
2.  **Code Quality:** Remove magic numbers, enforce error handling for JSON operations, and clean up global state.
3.  **Testing:** Increase test coverage for edge cases and critical paths.

## Detailed Requirements

### High Priority (Reliability)
1.  **Retry Logic Tests:**
    -   Must cover `IsRetryable` with various error types (network, 5xx, 4xx, etc.).
    -   Must verify exponential backoff timing.
    -   Must verify `MaxRetries` cap.
2.  **Panic Recovery:**
    -   Worker goroutines in `main.go` must have a `defer recover()` block.
    -   Panics should be caught and reported as errors in the `commitResult` rather than crashing the process.

### Medium Priority (Standards & Robustness)
1.  **JSON Error Handling:**
    -   All calls to `encoder.Encode()` must have their errors checked.
    -   Errors should be logged to `stderr` to avoid corrupting the JSON stdout stream.
2.  **Constants for Diff Types:**
    -   Replace `0`, `1`, `2` in `diff.go` with named constants (e.g., `ChunkEqual`, `ChunkAdd`).
3.  **JSON Edge Case Tests:**
    -   Add unit tests for `findJSONBlock` covering empty strings, missing braces, nested objects, etc.
4.  **Temp Directory Management:**
    -   Refactor the global `tempDir` variable in `main.go` into a struct-based manager (e.g., `CleanupManager`).

### Low Priority (Polish)
1.  **Struct Alignment:** Optimize `orderedPrinter` for memory alignment.
2.  **Benchmarks:** Add benchmarks for `TruncateDiff` and `ShouldIgnoreFile`.
3.  **Documentation:** Fix scenario references in `GitCommitAnalysis.md`.
4.  **Error Context:** Add context to errors returned by `GetStandardDiff`.

## Constraints
-   All changes must be backward compatible.
-   No new external libraries should be introduced for these fixes.
