# Specification - Refactor and Code Health

## Overview
This track addresses technical debt and architectural improvements identified during a comprehensive code review on January 22, 2026. The focus is on increasing test coverage for critical logic, unifying duplicated code between the CLI and MCP server, and improving code maintainability.

## Problem Statement
- **Critical Logic Unprotected:** `pkg/analyzer/retry.go` handles API failures but has no unit tests.
- **Code Duplication:** The CLI and MCP server independently implement commit iteration, filtering, and result aggregation logic.
- **Tight Coupling:** The analysis engine is tightly coupled to the Google Gemini implementation, making testing and future provider expansion difficult.
- **Magic Numbers:** `pkg/gitdiff` uses magic numbers for diff chunk types.

## Goals
- [ ] Achieve >90% test coverage for `pkg/analyzer/retry.go`.
- [ ] Consolidate core analysis logic into a reusable `pkg/analyzer/runner.go`.
- [ ] Decouple the analysis engine from the LLM provider using an interface.
- [ ] Improve code readability and standards (constants, validation, error wrapping).
- [ ] Enable real-time streaming of detailed analysis results (Reasoning, Probability) via MCP logs.

## Proposed Changes

### 1. Test Coverage
- Create `pkg/analyzer/retry_test.go`.
- Mock various error scenarios (429, 5xx, context cancellation).

### 2. Core Logic Consolidation
- Create `pkg/analyzer/runner.go` to handle:
    - Repository iteration.
    - Commit filtering (skip merges).
    - Concurrent analysis execution.
    - Result aggregation and summary generation.
- Update `cmd/git-commit-analysis` and `cmd/mcp-server` to use this shared runner.

### 3. Abstraction & Cleanliness
- Introduce `LLMProvider` interface in `pkg/analyzer`.
- Replace magic numbers in `pkg/gitdiff` with named constants.
- Implement early input validation in CLI.
- Standardize error wrapping with `%w`.

## Success Criteria
- All tests pass.
- `pkg/analyzer` coverage increases.
- No functional regressions in CLI or MCP server.
- Shared logic reduced by at least 150 lines across command packages.
