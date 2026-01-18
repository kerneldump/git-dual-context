# Technology Stack - Git Dual-Context Analysis

## Core Technologies
- **Programming Language:** Go (Golang)
    - Version: 1.25.5 (as specified in `go.mod`)
    - Rationale: High performance, excellent concurrency support for repository scanning, and strong standard library.

## Key Libraries & Frameworks
- **Git Operations:** `github.com/go-git/go-git/v5`
    - Purpose: Provides a pure Go implementation of Git, enabling deep repository analysis without requiring a local Git binary in all environments.
- **AI Integration:** `github.com/google/generative-ai-go`
    - Purpose: Official SDK for interacting with Google's Gemini models, used for the reasoning and analysis core.
- **API Connectivity:** `google.golang.org/api`
    - Purpose: Support library for Google Cloud and API authentication.

## Key Features & Standards
- **Output Format:** Newline Delimited JSON (NDJSON)
    - Rationale: Enables high-performance streaming of both logs and results in a unified, machine-readable format that is easy to process with tools like `jq`.

## Architecture
- **Structure:** Standard Go Project Layout
    - `cmd/`: Contains the main entry points for tool executables.
    - `internal/`: Houses private packages for core logic, including `analyzer` (the reasoning engine) and `gitdiff` (context extraction).
- **Pattern:** Modular component-based design for AI providers and data extraction.
