# Implementation Plan - Refactor and Code Health

## Phase 1: Critical Test Coverage & Cleanup
- [ ] Task: Create `pkg/analyzer/retry_test.go` and implement comprehensive tests for retry logic.
- [ ] Task: Replace magic numbers in `pkg/gitdiff/diff.go` with `ChunkType` constants.
- [ ] Task: Implement consistent error wrapping (`%w`) across `pkg/analyzer` and `pkg/gitdiff`.

## Phase 2: Core Refactoring (The Engine)
- [ ] Task: Define `LLMProvider` interface and `GeminiProvider` implementation in `pkg/analyzer`.
- [ ] Task: Create `pkg/analyzer/runner.go` and implement `RunAnalysis` logic.
- [ ] Task: Update `RunAnalysis` callback to support detailed result objects (streaming).
- [ ] Task: Refactor `cmd/git-commit-analysis` to use the shared `RunAnalysis` engine.
- [ ] Task: Refactor `cmd/mcp-server/internal/tools/rootcause.go` to use the shared `RunAnalysis` engine with detailed streaming logs.
- [ ] Task: Remove redundant ordering and aggregation logic from command packages.

## Phase 3: Integration & Quality
- [ ] Task: Add early input validation to `cmd/git-commit-analysis/main.go`.
- [ ] Task: Implement basic integration tests in `pkg/analyzer/engine_test.go` using the `LLMProvider` interface.
- [ ] Task: Update documentation (GoDoc) for key public functions.
- [ ] Task: Conductor - User Manual Verification 'Refactor and Code Health' (Protocol in workflow.md)
