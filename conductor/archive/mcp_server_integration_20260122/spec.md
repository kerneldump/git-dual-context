# Specification - MCP Server Integration

## Overview
This track involves integrating the Proof-of-Concept (POC) MCP Server into the repository as a supported command. The goal is to make the `git-dual-context` library accessible via the Model Context Protocol (MCP) for AI agents (Gemini, Claude, etc.), while cleaning up the project structure.

## Functional Requirements
- **Directory Structure:** Move `poc` code to `cmd/mcp-server` to follow Go project standards.
- **Dependency Management:** Integrate `github.com/modelcontextprotocol/go-sdk` into the root `go.mod` (removing the need for a separate `poc/go.mod`).
- **Compatibility:** Ensure the server code works with the latest version of the MCP SDK (v1.2.0), specifically adapting the `ToolHandler` signature and `StdioTransport` initialization.
- **Documentation:** Provide clear build and usage instructions for the MCP server in `cmd/mcp-server/README.md`.
- **Security:** Ensure no API keys are hardcoded in the source; rely strictly on environment variables (`GEMINI_API_KEY`).

## Acceptance Criteria
- [x] `poc` directory is removed.
- [x] `cmd/mcp-server/main.go` compiles and runs.
- [x] Root `go.mod` includes MCP SDK dependencies.
- [x] No references to external repos (e.g. `velocity-global`) exist in the codebase.
- [x] `.gitignore` includes the `mcp-server` binary.
- [x] Server connects successfully to an MCP client (verified with `gemini mcp list`).
