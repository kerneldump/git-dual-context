# Track Specification: Code Quality & Stability Remediation

### Overview
This track addresses a comprehensive set of 12 issues identified in the code analysis report, ranging from critical stability bugs (panics, nil pointers) to architectural improvements (shared orchestration, configuration integration). The goal is to move the project from a proof-of-concept to a robust, maintainable library and toolset.

### Functional Requirements
- **Stability Fixes:**
    - Prevent panics in `pkg/config/config.go` by adding length checks before path slicing.
    - Handle nil parent trees in `pkg/gitdiff/diff.go` to support first-commit analysis.
    - Ensure `json.Encoder` thread safety in the CLI `orderedPrinter`.
- **Architectural Improvements:**
    - **Safe Concurrency:** Implement a Repository Pool in the analyzer to allow safe parallel processing across both CLI and MCP server.
    - **Shared Orchestration:** Extract the commit iteration and analysis logic from `main.go` and `rootcause.go` into a reusable `pkg/analyzer/orchestrator.go`.
    - **Configuration Integration:** Implement a hybrid precedence model (Defaults < Config File < Env Vars < Flags) for both CLI and MCP server.
- **Refactoring & Standards:**
    - **Prompt Management:** Extract the LLM prompt from `pkg/analyzer/engine.go` into an embedded text file.
    - **Constants:** Replace magic numbers (buffer sizes, truncation lengths) with named constants.
    - **Model Naming:** Standardize model name formats across all entry points.
    - **Documentation:** Add missing Godoc comments to all exported functions.
- **Robustness:**
    - Improve JSON block extraction from LLM responses with fallback regex parsing.

### Non-Functional Requirements
- **Test Coverage:** Maintain or exceed 80% coverage for all new and modified packages.
- **Concurrency:** Ensure no race conditions occur when using `go-git` in parallel.
- **Performance:** Concurrency should be enabled by default for both CLI and MCP server using the Repository Pool.

### Acceptance Criteria
- [ ] No panics occur when providing short or malformed config paths (e.g., `~`).
- [ ] Analysis successfully runs on repositories with only a single commit.
- [ ] `go test -race ./...` passes without any data races detected.
- [ ] The CLI and MCP server both successfully load settings from a `.git-dual-context.yaml` file.
- [ ] Unit tests exist for `AnalyzeCommit`, `GetStandardDiff`, and `GetFullDiff`.
- [ ] All exported functions have descriptive Godoc comments.

### Out of Scope
- Implementing support for new LLM providers beyond Gemini.
- Adding a GUI or Web interface.
