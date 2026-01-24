# Implementation Plan - Code Quality & Stability Remediation

## Phase 1: Critical Stability & Quick Wins ✅
- [x] Task: Fix path slicing panics in `pkg/config/config.go`
    - [x] Write tests in `pkg/config/config_test.go` that pass short paths (e.g., "~", "/", "a") to `LoadConfig` and `SaveConfig`
    - [x] Implement length checks before slicing `path[:2]`
- [x] Task: Fix nil pointer in `pkg/gitdiff/diff.go`
    - [x] Create a test case in `pkg/gitdiff/diff_test.go` that simulates a first commit (no parent)
    - [x] Implement `object.DiffTree(nil, cTree)` fallback when `pTree` is nil
- [x] Task: Standardize constants and magic numbers
    - [x] Create `pkg/analyzer/constants.go`
    - [x] Move magic numbers for truncation (80), buffer sizes (8192), and default model names into constants
    - [x] Update `pkg/analyzer/engine.go`, `pkg/gitdiff/diff.go`, and `cmd/` files to use these constants
- [x] Task: Conductor - User Manual Verification 'Phase 1: Critical Stability & Quick Wins' (Protocol in workflow.md)

## Phase 2: Architectural Foundation & Concurrency ✅
- [x] Task: Implement safe concurrency for parallel LLM calls
    - [x] **Alternative approach**: Implemented Two-Phase design instead of Repository Pool
    - [x] Phase 1: Extract diffs sequentially (git operations - NOT thread-safe)
    - [x] Phase 2: Call LLM in parallel (API calls - thread-safe)
    - [x] Created `ExtractDiffs()` and `AnalyzeWithDiffs()` functions in `pkg/analyzer/engine.go`
- [x] Task: Extract LLM Prompt to embedded file
    - [x] Create `pkg/analyzer/prompts/analysis.txt`
    - [x] Use `//go:embed` to load the prompt in `pkg/analyzer/engine.go`
- [x] Task: Extract Shared Orchestrator
    - [x] Create `pkg/analyzer/orchestrator.go`
    - [x] Implement `CollectCommits`, `AnalyzeCommitSequential`, `CalculateSummary` functions
    - [x] Document the Two-Phase approach for concurrency
- [x] Task: Conductor - User Manual Verification 'Phase 2: Architectural Foundation & Concurrency' (Protocol in workflow.md)

## Phase 3: Configuration Integration & Robustness ✅
- [x] Task: Integrate `pkg/config` into CLI
    - [x] CLI uses `config.FindConfigFile()` and `LoadConfig()`
    - [x] Implements precedence logic (Config -> Env -> Flags)
- [x] Task: Integrate `pkg/config` into MCP Server
    - [x] MCP server in `cmd/mcp-server/internal/tools/rootcause.go` uses config package
- [x] Task: Improve JSON Parsing Robustness
    - [x] Refactor `FindJSONBlock` in `pkg/analyzer/engine.go`
    - [x] Add regex-based fallback for extracting JSON when markers are missing or malformed
    - [x] Added `jsonFallbackRegex` with two-strategy approach
- [x] Task: Conductor - User Manual Verification 'Phase 3: Configuration Integration & Robustness' (Protocol in workflow.md)

## Phase 4: Testing, Documentation & Cleanup ✅
- [x] Task: Add missing unit tests for core engine
    - [x] Added `TestFindJSONBlockRegexFallback` covering regex fallback scenarios
    - [x] Added `TestAnalysisPromptTemplateLoaded` verifying embedded prompt
    - [x] Tests exist for `GetStandardDiff` and `GetFullDiff` in `pkg/gitdiff/diff_test.go`
- [x] Task: Add `LLMModel` interface for testability
    - [x] Created `LLMModel` interface in `pkg/analyzer/engine.go`
    - [x] Updated `AnalyzeCommit` and `AnalyzeWithDiffs` to accept interface
    - [x] Enables mocking for unit tests
- [x] Task: Ensure thread safety for MCP server output
    - [x] Two-Phase design ensures git operations are sequential
    - [x] LLM calls run in parallel with semaphore control
- [x] Task: Add Godoc and final cleanup
    - [x] Documentation added to exported functions
    - [x] All tests pass: `go test ./...`
- [x] Task: Conductor - User Manual Verification 'Phase 4: Testing, Documentation & Cleanup' (Protocol in workflow.md)

---

## Summary of Implementation

### Key Design Decision: Two-Phase vs Repository Pool

Instead of implementing a Repository Pool (which has uncertain thread-safety with go-git), we implemented a **Two-Phase Design**:

| Phase | Operation | Thread Safety |
|-------|-----------|---------------|
| Phase 1 | `ExtractDiffs()` - git operations | Sequential (guaranteed safe) |
| Phase 2 | `AnalyzeWithDiffs()` - LLM calls | Parallel (N workers) |

This approach:
- **Guarantees** thread safety for git operations
- **Maximizes** parallelism for expensive LLM API calls
- **Simpler** than pool management
- **Lower memory** (single repo handle)

### Additional Improvements (Hybrid Approach)

Combined best practices from code review:
1. **`LLMModel` interface** - Enables mocking for unit tests
2. **`//go:embed` for prompts** - External prompt file for easy modification
3. **Regex fallback in JSON parsing** - More robust extraction
4. **Constants centralized** - All magic numbers in `constants.go`
