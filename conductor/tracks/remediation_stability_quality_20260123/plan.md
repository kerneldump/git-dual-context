# Implementation Plan - Code Quality & Stability Remediation

## Phase 1: Critical Stability & Quick Wins ✅
- [ ] Task: Fix path slicing panics in `pkg/config/config.go`
    - [ ] Write tests in `pkg/config/config_test.go` that pass short paths (e.g., "~", "/", "a") to `LoadConfig` and `SaveConfig`
    - [ ] Implement length checks before slicing `path[:2]`
- [ ] Task: Fix nil pointer in `pkg/gitdiff/diff.go`
    - [ ] Create a test case in `pkg/gitdiff/diff_test.go` that simulates a first commit (no parent)
    - [ ] Implement `object.DiffTree(nil, cTree)` fallback when `pTree` is nil
- [ ] Task: Standardize constants and magic numbers
    - [ ] Create `pkg/analyzer/constants.go`
    - [ ] Move magic numbers for truncation (80), buffer sizes (8192), and default model names into constants
    - [ ] Update `pkg/analyzer/engine.go`, `pkg/gitdiff/diff.go`, and `cmd/` files to use these constants
- [ ] Task: Conductor - User Manual Verification 'Phase 1: Critical Stability & Quick Wins' (Protocol in workflow.md)

## Phase 2: Architectural Foundation & Concurrency ✅
- [ ] Task: Implement Repository Pool for safe concurrency
    - [ ] Create `pkg/analyzer/pool.go` to manage multiple `*git.Repository` instances
    - [ ] Implement `Get()` and `Put()` logic to reuse repository handles safely across goroutines
- [ ] Task: Extract LLM Prompt to embedded file
    - [ ] Create `pkg/analyzer/prompts/analysis.txt`
    - [ ] Use `//go:embed` to load the prompt in `pkg/analyzer/engine.go`
- [ ] Task: Extract Shared Orchestrator
    - [ ] Create `pkg/analyzer/orchestrator.go`
    - [ ] Implement `RunAnalysis` function that handles commit iteration, worker pooling, and progress reporting
    - [ ] Ensure the orchestrator uses the Repository Pool
- [ ] Task: Conductor - User Manual Verification 'Phase 2: Architectural Foundation & Concurrency' (Protocol in workflow.md)

## Phase 3: Configuration Integration & Robustness ✅
- [ ] Task: Integrate `pkg/config` into CLI
    - [ ] Update `cmd/git-commit-analysis/main.go` to use `config.FindConfigFile()` and `LoadConfig()`
    - [ ] Implement the precedence logic (Config -> Env -> Flags)
- [ ] Task: Integrate `pkg/config` into MCP Server
    - [ ] Update `cmd/mcp-server/internal/tools/rootcause.go` to use the config package
- [ ] Task: Improve JSON Parsing Robustness
    - [ ] Refactor `FindJSONBlock` in `pkg/analyzer/engine.go`
    - [ ] Add regex-based fallback for extracting JSON when markers are missing or malformed
- [ ] Task: Conductor - User Manual Verification 'Phase 3: Configuration Integration & Robustness' (Protocol in workflow.md)

## Phase 4: Testing, Documentation & Cleanup ✅
- [ ] Task: Add missing unit tests for core engine
    - [ ] Implement `pkg/analyzer/engine_test.go` covering `AnalyzeCommit` with mock/in-memory git repos
    - [ ] Add tests for `GetStandardDiff` and `GetFullDiff` in `pkg/gitdiff/diff_test.go`
- [ ] Task: Ensure thread safety for CLI output
    - [ ] Update `orderedPrinter` in `main.go` to ensure `json.Encoder` access is strictly protected by the mutex
- [ ] Task: Add Godoc and final cleanup
    - [ ] Add documentation to all exported functions in `pkg/`
    - [ ] Run `go test -race ./...` to verify no regressions
- [ ] Task: Conductor - User Manual Verification 'Phase 4: Testing, Documentation & Cleanup' (Protocol in workflow.md)
