# Implementation Plan - MCP Server Integration

## Phase 1: Restructuring & Migration [checkpoint: verified]
- [x] Task: Analyze `poc` structure and dependencies.
- [x] Task: Create `cmd/mcp-server` directory structure.
- [x] Task: Migrate `poc/cmd/main.go` to `cmd/mcp-server/main.go`.
- [x] Task: Migrate `poc/internal/tools` to `cmd/mcp-server/internal/tools`.
- [x] Task: Update imports to use `github.com/kerneldump/git-dual-context` paths.
- [x] Task: Remove legacy `poc` directory.

## Phase 2: Dependency & Build [checkpoint: verified]
- [x] Task: Add `github.com/modelcontextprotocol/go-sdk` to root `go.mod`.
- [x] Task: Update `main.go` to use `&mcp.StdioTransport{}` (SDK v1.2.0 compatibility).
- [x] Task: Update `ToolHandler` signature in `main.go` (SDK v1.2.0 compatibility).
- [x] Task: Clean up unused `go.mod` imports (`go mod tidy`).
- [x] Task: Add `/mcp-server` binary to `.gitignore`.

## Phase 3: Verification & Docs [checkpoint: verified]
- [x] Task: Update `cmd/mcp-server/README.md` with build and usage instructions.
- [x] Task: Perform security audit (grep for keys/secrets).
- [x] Task: Verify build (`go build ./cmd/mcp-server`).
- [x] Task: Verify runtime (`./mcp-server --help` and MCP client connection).
