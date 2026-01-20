# Specification - Fix Module Path

## Overview
The current `go.mod` declares the module name as `git-commit-analysis`, which prevents the project from being imported as an external library via `go get`. This track will rename the module to the canonical GitHub path `github.com/kerneldump/git-dual-context` and update all internal import references.

## Functional Requirements
- **Module Renaming:** Change the module path in `go.mod` from `git-commit-analysis` to `github.com/kerneldump/git-dual-context`.
- **Import Updates:** Update all internal `import` statements in the codebase to use the new module path.
    - `cmd/git-commit-analysis/main.go`
    - `pkg/analyzer/engine.go`
    - `examples/basic_usage/main.go`
    - Any test files importing internal packages.

## Non-Functional Requirements
- **No Logical Changes:** The tool's functionality must remain identical.
- **Build Integrity:** The project must build and test successfully after the rename.

## Acceptance Criteria
- [ ] `go.mod` uses `github.com/kerneldump/git-dual-context`.
- [ ] All internal imports use the new path.
- [ ] `go build ./...` succeeds.
- [ ] `go test ./...` passes.
