# Track Specification: Code Quality & Stability Remediation

### Overview
This track addresses a comprehensive set of 12 issues identified in the code analysis report, ranging from critical stability bugs (panics, nil pointers) to architectural improvements (shared orchestration, configuration integration). The goal is to move the project from a proof-of-concept to a robust, maintainable library and toolset.

### Functional Requirements
- **Stability Fixes:**
    - Prevent panics in `pkg/config/config.go` by adding length checks before path slicing.
    - Handle nil parent trees in `pkg/gitdiff/diff.go` to support first-commit analysis.
    - Ensure `json.Encoder` thread safety in the CLI `orderedPrinter`.
- **Architectural Improvements:**
    - **Safe Concurrency:** Implement Two-Phase design (sequential git ops, parallel LLM calls) for safe parallel processing.
    - **Shared Orchestration:** Extract the commit iteration and analysis logic into reusable `pkg/analyzer/orchestrator.go`.
    - **Configuration Integration:** Implement a hybrid precedence model (Defaults < Config File < Env Vars < Flags) for both CLI and MCP server.
- **Refactoring & Standards:**
    - **Prompt Management:** Extract the LLM prompt from `pkg/analyzer/engine.go` into an embedded text file.
    - **Constants:** Replace magic numbers (buffer sizes, truncation lengths) with named constants.
    - **Model Naming:** Standardize model name formats across all entry points.
    - **Documentation:** Add missing Godoc comments to all exported functions.
- **Robustness:**
    - Improve JSON block extraction from LLM responses with fallback regex parsing.
- **Testability:**
    - Add `LLMModel` interface to enable mocking in unit tests.

### Non-Functional Requirements
- **Test Coverage:** Maintain or exceed 80% coverage for all new and modified packages.
- **Concurrency:** Ensure no race conditions occur when using `go-git` in parallel.
- **Performance:** Concurrency enabled for LLM calls using Two-Phase design with configurable workers.

### Acceptance Criteria
- [x] No panics occur when providing short or malformed config paths (e.g., `~`).
- [x] Analysis successfully runs on repositories with only a single commit.
- [x] `go test ./...` passes without any failures.
- [x] The CLI and MCP server both successfully load settings from config files.
- [x] Unit tests exist for JSON parsing, prompt loading, and orchestrator functions.
- [x] All exported functions have descriptive Godoc comments.
- [x] `LLMModel` interface enables mocking for unit tests.

### Out of Scope
- Implementing support for new LLM providers beyond Gemini.
- Adding a GUI or Web interface.

### Implementation Notes

**Key Design Decision:** Instead of Repository Pool, implemented Two-Phase design:
1. Phase 1: Extract diffs sequentially (git operations - NOT thread-safe)
2. Phase 2: Call LLM in parallel (API calls - thread-safe)

This approach guarantees thread safety while maximizing parallelism for expensive LLM calls.
