# Specification - Library Refactor (internal to pkg)

## Overview
This track involves refactoring the project structure to expose the core logic (`analyzer` and `gitdiff`) as a public library. Currently, these packages reside in `internal/`, which restricts their use to the local module. Moving them to `pkg/` follows standard Go conventions for reusable library code.

## Functional Requirements
- **Directory Migration:** Move `internal/analyzer` to `pkg/analyzer` and `internal/gitdiff` to `pkg/gitdiff`.
- **Import Resolution:** Update all internal and external import paths to reflect the new `pkg/` structure.
- **API Surface Audit:** Review all unexported symbols in the migrated packages and export those necessary for idiomatic library usage.
- **Documentation:** 
    - Add `doc.go` to both `pkg/analyzer` and `pkg/gitdiff`.
    - Update the main `README.md` with a "Library Usage" section and code examples.
- **Examples:** Create an `examples/` directory containing a standalone Go program that demonstrates how to import and use the library.

## Non-Functional Requirements
- **Backward Compatibility:** Maintain existing CLI functionality in `cmd/git-commit-analysis/`.
- **Code Quality:** Ensure all exported symbols have proper GoDoc comments.
- **Go Conventions:** Adhere to `pkg/` directory conventions and Go library best practices.

## Acceptance Criteria
- [ ] `internal/` directory is removed.
- [ ] CLI tool builds and runs correctly with the new package paths.
- [ ] `go test ./...` passes for all packages.
- [ ] `examples/` code compiles and runs successfully.
- [ ] README contains clear library documentation.
- [ ] Exported symbols have complete documentation.

## Out of Scope
- Tagging a new release version (e.g., `v0.1.0`).
- Adding new analytical features or changing the core reasoning logic.