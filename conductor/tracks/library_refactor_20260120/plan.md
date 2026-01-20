# Implementation Plan - Library Refactor

## Phase 1: Structure Migration & Refactoring [checkpoint: ea90cdb]
- [x] Task: Move `internal/analyzer` to `pkg/analyzer` and `internal/gitdiff` to `pkg/gitdiff`. [3aef27a]
- [x] Task: Update import paths in `cmd/git-commit-analysis/main.go` and across `pkg/` packages. [35ffd7b]
- [x] Task: Update `engine_test.go` and `diff_test.go` to use the new package paths. [35ffd7b]
- [x] Task: Verify successful build of CLI and passing tests (`go test ./...`). [35ffd7b]
- [x] Task: Conductor - User Manual Verification 'Phase 1' (Protocol in workflow.md) [ea90cdb]

## Phase 2: API Surface Audit & Polish [checkpoint: 3412c3e]
- [x] Task: Audit `pkg/analyzer` for unexported symbols that should be public; rename and add docs. [f0320eb]
- [x] Task: Audit `pkg/gitdiff` for unexported symbols that should be public; rename and add docs. [f0320eb]
- [x] Task: Create `pkg/analyzer/doc.go` with package-level documentation. [f0320eb]
- [x] Task: Create `pkg/gitdiff/doc.go` with package-level documentation. [f0320eb]
- [x] Task: Conductor - User Manual Verification 'Phase 2' (Protocol in workflow.md) [3412c3e]

## Phase 3: Documentation & Examples [checkpoint: b17a80c]
- [x] Task: Create `examples/basic_usage/main.go` demonstrating library usage. [048f3e5]
- [x] Task: Update `README.md` with "Library Usage" section and snippets from the example. [048f3e5]
- [x] Task: Ensure all exported functions/types have GoDoc comments. [048f3e5]
- [x] Task: Conductor - User Manual Verification 'Phase 3' (Protocol in workflow.md) [b17a80c]