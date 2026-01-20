# Implementation Plan - Library Refactor

## Phase 1: Structure Migration & Refactoring
- [x] Task: Move `internal/analyzer` to `pkg/analyzer` and `internal/gitdiff` to `pkg/gitdiff`. [3aef27a]
- [x] Task: Update import paths in `cmd/git-commit-analysis/main.go` and across `pkg/` packages. [f67332b]
- [x] Task: Update `engine_test.go` and `diff_test.go` to use the new package paths. [f67332b]
- [x] Task: Verify successful build of CLI and passing tests (`go test ./...`). [f67332b]
- [ ] Task: Conductor - User Manual Verification 'Phase 1' (Protocol in workflow.md)

## Phase 2: API Surface Audit & Polish
- [ ] Task: Audit `pkg/analyzer` for unexported symbols that should be public; rename and add docs.
- [ ] Task: Audit `pkg/gitdiff` for unexported symbols that should be public; rename and add docs.
- [ ] Task: Create `pkg/analyzer/doc.go` with package-level documentation.
- [ ] Task: Create `pkg/gitdiff/doc.go` with package-level documentation.
- [ ] Task: Conductor - User Manual Verification 'Phase 2' (Protocol in workflow.md)

## Phase 3: Documentation & Examples
- [ ] Task: Create `examples/basic_usage/main.go` demonstrating library usage.
- [ ] Task: Update `README.md` with "Library Usage" section and snippets from the example.
- [ ] Task: Ensure all exported functions/types have GoDoc comments.
- [ ] Task: Conductor - User Manual Verification 'Phase 3' (Protocol in workflow.md)