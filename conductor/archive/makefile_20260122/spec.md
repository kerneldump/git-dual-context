# Specification - Makefile Implementation

## Overview
This track involves creating a `Makefile` in the project root to standardize the development lifecycle, including building binaries, running tests, and ensuring code quality.

## Functional Requirements
- **Build Target:** Must compile `cmd/git-commit-analysis` and `cmd/mcp-server` into separate binaries.
- **Test Target:** Must execute all tests in the repository (`go test ./...`).
- **Quality Targets:** 
  - `fmt`: Run `go fmt` on all packages.
  - `vet`: Run `go vet` on all packages.
  - `lint`: Run the project's linter (e.g., `golangci-lint` if available, or a standard recommendation).
- **Cleanup Target:** Remove compiled binaries and temporary build artifacts.
- **Convenience Targets:**
  - `run`: Run the primary `git-commit-analysis` tool (requires sample arguments).
  - `help`: Provide a self-documenting list of available `make` commands.

## Non-Functional Requirements
- **Portability:** The Makefile should work on macOS (user's OS) and Linux environments.
- **Simplicity:** Avoid complex logic or external dependencies outside of standard Go tools and `make`.

## Acceptance Criteria
- Running `make build` successfully creates `git-commit-analysis` and `mcp-server` binaries.
- Running `make test` executes the test suite and reports results.
- Running `make help` displays a clear description of all targets.
- The `Makefile` adheres to standard Go project conventions.

## Out of Scope
- Version/Git hash injection via `ldflags`.
- Custom output directory configuration.
- Docker image build targets.
