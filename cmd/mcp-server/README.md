# Git Dual-Context MCP Server

An MCP (Model Context Protocol) server that exposes the **Dual-Context Diff Analysis** tool for automated bug diagnosis. This tool wraps the [git-dual-context](https://github.com/kerneldump/git-dual-context) library.

## Overview

```
┌──────────────────────────────────────┐
│  MCP Client (Gemini-CLI/Cursor)      │
└────────────┬─────────────────────────┘
             │ JSON-RPC over Stdio
             ▼
┌──────────────────────────────────────┐
│  Go MCP Server                       │
│  ┌────────────────────────────────┐  │
│  │ Tool: analyze_root_cause       │  │
│  │                                │  │
│  │ ┌──────────────────────────┐   │  │
│  │ │ git-dual-context logic   │   │  │
│  │ │ (pkg/analyzer, pkg/diff) │   │  │
│  │ └──────────────────────────┘   │  │
│  │                                │  │
│  │ Input: repo_path, error_msg    │  │
│  │ Output: root cause analysis    │  │
│  └────────────────────────────────┘  │
└──────────────────────────────────────┘
```

## Prerequisites

- **Go 1.25+** installed
- **Google Gemini API Key** ([Get one here](https://aistudio.google.com/app/apikey))
- **Gemini-CLI** installed (`npm install -g @anthropic-ai/gemini-cli` or via homebrew)

## Installation

### Build from Source

```bash
# Clone the repository (if you haven't already)
git clone https://github.com/kerneldump/git-dual-context
cd git-dual-context

# Build the server
go build -o mcp-server ./cmd/mcp-server
```

### Verify Build

```bash
./mcp-server --help
```

## Configuration

### Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `GEMINI_API_KEY` | Yes | - | Google Gemini API key |
| `GEMINI_MODEL` | No | `gemini-flash-latest` | Gemini model to use |

### Running the Server

```bash
export GEMINI_API_KEY="your-api-key"
export GEMINI_MODEL="gemini-flash-latest"  # optional
./mcp-server
```

## Usage with Gemini-CLI

### 1. Add the MCP Server

```bash
# Add the server to Gemini-CLI (assuming you built it in the root of the repo)
gemini mcp add git-context "/path/to/git-dual-context/mcp-server"
```

Or manually edit `~/.gemini/settings.json`:

```json
{
  "mcpServers": {
    "git-context": {
      "command": "/path/to/git-dual-context/mcp-server",
      "env": {
        "GEMINI_API_KEY": "your-api-key-here"
      }
    }
  }
}
```

### 2. Verify Connection

```bash
# List available MCP servers
gemini mcp list
```

### 3. Use the Tool

Start Gemini-CLI and use the tool:

```bash
gemini
```

Then in the chat:

```
Analyze the last 5 commits in /path/to/my-project for the error "panic: index out of bounds"
```

Or more explicitly:

```
Use the analyze_root_cause tool with:
- repo_path: /Users/me/projects/my-app
- error_message: "connection refused on port 5432"
- num_commits: 10
```

## Tool Reference

### `analyze_root_cause`

Diagnose bugs using dual-context diff analysis.

#### Input Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `repo_path` | string | Yes | - | Path to local git repository |
| `error_message` | string | Yes | - | Bug description or error message to diagnose |
| `num_commits` | integer | No | 5 | Number of recent commits to analyze |
| `branch` | string | No | HEAD | Branch to analyze |

> **Note:** Commits are analyzed sequentially due to thread-safety constraints in the underlying git library.

#### Output

```json
{
  "results": [
    {
      "hash": "be8f779e",
      "message": "Allow negative durations in TimeFilter",
      "probability": "HIGH",
      "reasoning": "The commit modifies NewTimeFilter to accept negative durations..."
    },
    {
      "hash": "1c932131",
      "message": "Refactor axis bounds calculation",
      "probability": "MEDIUM",
      "reasoning": "The commit modifies axis bounds calculation..."
    }
  ],
  "summary": {
    "total": 5,
    "high": 1,
    "medium": 1,
    "low": 1,
    "skipped": 2,
    "errors": 0
  }
}
```

#### Probability Levels

| Level | Description |
|-------|-------------|
| **HIGH** | "Smoking gun" found - commit directly contradicts the error or enables the bug |
| **MEDIUM** | Commit modifies relevant subsystems, creates plausible path for bug |
| **LOW** | No direct or plausible link found |
