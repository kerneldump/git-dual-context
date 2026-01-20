# Git Dual-Context Analysis

> **Automated Bug Diagnosis via Dual-Context Diff Analysis & LLMs**

`git-dual-context` is a proof-of-concept tool that implements the **Dual-Context Diff Analysis** theoretical framework. It leverages Large Language Models (LLMs) to diagnose complex software bugs by analyzing commits through two distinct lenses:
1.  **Standard Diff (Micro):** The immediate changes introduced by a commit.
2.  **Full Comparison Diff (Macro):** The evolutionary changes of those files from the commit time to the current `HEAD`.

By synthesizing these two signals, the tool can identify "sleeper bugs"—issues that arise not from the immediate change, but from how that change interacts with future code evolution (refactors, new feature interactions, etc.).

---

## The Theory

This tool is the reference implementation for the paper:  
**[Enhanced Bug Diagnosis via Dual-Context Diff Analysis](docs/GitCommitAnalysis.md)**

> *Traditionally, debugging focuses on "What did this commit change?". Use this tool when you need to answer: "How does this commit interact with the current state of the world?"*

## Features

-   **Automated Hypothesis Testing:** Automatically scans the last `N` commits to calculate the probability ($P(H_k|E)$) that a specific commit caused a given bug.
-   **Dual-Context Analysis:**
    -   Generates **Standard Diffs** (with context lines) to understand developer intent.
    -   Generates **Full Comparison Diffs** to understand evolutionary context.
-   **LLM Integration:** Uses Google's Gemini models with configurable model selection.
-   **Smart Filtering:** Automatically excludes lock files, vendor directories, test files, CI/CD configs, and build artifacts to focus on logic changes and conserve tokens.
-   **Ordered Streaming Output:** Results stream in commit order as they become available—no waiting for all analyses to complete.
-   **Retry Logic:** Automatic exponential backoff for rate limits and transient failures.
-   **Graceful Shutdown:** Clean handling of Ctrl+C with proper cleanup.

## Usage

### Prerequisites

-   **Go 1.21+** installed.
-   A **Google Gemini API Key** (Get one [here](https://makersuite.google.com/app/apikey)).

### Installation

```bash
# Clone the repository
git clone https://github.com/your-username/git-dual-context.git
cd git-dual-context

# Build the binary
go build -o git-commit-analysis ./cmd/git-commit-analysis
```

### Running the Tool

1.  **Set your API Key:**

    ```bash
    export GEMINI_API_KEY="your_api_key_here"
    ```

2.  **Run the Analysis:**

    ```bash
    ./git-commit-analysis \
      -repo="https://github.com/kerneldump/signal-sentry.git" \
      -error="interval must be greater than 0, got -2" \
      -n 5
    ```

### Command-Line Options

| Flag | Default | Description |
|------|---------|-------------|
| `-repo` | `.` | Path to git repository or remote URL |
| `-branch` | current HEAD | Branch to analyze |
| `-error` | (required) | The error message or bug description to analyze |
| `-n` | `5` | Number of commits to analyze |
| `-j` | `3` | Number of concurrent workers |
| `-model` | `models/gemini-3-pro-preview` | Gemini model to use |
| `-timeout` | `5m` | Timeout per commit analysis |
| `-o` | stdout | Output file path |
| `-apikey` | env `GEMINI_API_KEY` | Google Gemini API Key |
| `-v` | `false` | Verbose output (debug info) |

### Examples

```bash
# Analyze local repository
./git-commit-analysis -error="panic: index out of bounds" -n 10

# Analyze remote repository with custom model
./git-commit-analysis \
  -repo="https://github.com/user/repo.git" \
  -error="connection timeout" \
  -model="models/gemini-1.5-flash"

# Save output to file
./git-commit-analysis -error="nil pointer" -o results.json

# Analyze specific branch with verbose output
./git-commit-analysis \
  -branch="feature/auth" \
  -error="401 unauthorized" \
  -v

# Use fewer workers to avoid rate limits
./git-commit-analysis -error="timeout" -j 1 -n 20
```

---

## Library Usage

`git-dual-context` can be used as a library in your Go projects.

### Installation

```bash
go get github.com/your-username/git-dual-context
```

### Basic Example

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/your-username/git-dual-context/pkg/analyzer"
	"github.com/go-git/go-git/v5"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

func main() {
	ctx := context.Background()
	apiKey := os.Getenv("GEMINI_API_KEY")

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	model := client.GenerativeModel("models/gemini-1.5-pro")
	repo, _ := git.PlainOpen(".")
	headRef, _ := repo.Head()
	headCommit, _ := repo.CommitObject(headRef.Hash())

	errorMsg := "The system is returning a 500 error on the /login endpoint"
	
	// Perform the dual-context analysis
	result, err := analyzer.AnalyzeCommit(ctx, repo, headCommit, headCommit, errorMsg, model)
	if err != nil {
		log.Fatal(err)
	}

	if !result.Skipped {
		fmt.Printf("Probability: %s\n", result.Probability)
		fmt.Printf("Reasoning: %s\n", result.Reasoning)
	}
}
```

For more details, see [examples/basic_usage/main.go](examples/basic_usage/main.go).

### Core Packages

-   **`pkg/analyzer`:** The reasoning engine. Handles prompt construction, LLM interaction, and response parsing.
-   **`pkg/gitdiff`:** Diff extraction and filtering logic. Handles standard and evolutionary diff generation.

---

## Output Format (NDJSON)

The tool outputs results in **Newline Delimited JSON (NDJSON)** format. Results stream in commit order as they become available. Output types are distinguished by the `type` field:

| Type | Description |
|------|-------------|
| `"result"` | Analysis findings with `hash`, `message`, `probability`, `reasoning` |
| `"log"` | Progress and status updates with `level`, `msg`, `timestamp` |
| `"summary"` | Final summary with `total`, `high`, `medium`, `low`, `skipped`, `errors` |

#### Pro-tip: Filter with `jq`

```bash
# Show only high-probability commits
./git-commit-analysis -error="..." | jq 'select(.type=="result" and .probability=="HIGH")'

# Get clean result stream (no logs)
./git-commit-analysis -error="..." | jq 'select(.type=="result")'

# Show just the summary
./git-commit-analysis -error="..." | jq 'select(.type=="summary")'
```

---

## Example Output

```json
{"type":"log","level":"INFO","msg":"Cloning https://github.com/... into temporary directory...","timestamp":"2026-01-18T10:15:00Z"}
{"type":"log","level":"INFO","msg":"Analyzing last 5 commits for error: \"interval must be greater than 0, got -2\"","timestamp":"2026-01-18T10:15:05Z"}
{"type":"result","hash":"be8f779e","message":"Allow negative durations in TimeFilter","probability":"HIGH","reasoning":"The commit modifies NewTimeFilter to accept negative durations instead of ignoring them, which eventually reaches a ticker validation check."}
{"type":"result","hash":"1c932131","message":"Refactor axis bounds calculation","probability":"MEDIUM","reasoning":"The commit modifies axis bounds calculation, which could potentially result in negative intervals in edge cases."}
{"type":"result","hash":"26cb336c","message":"Update README documentation","probability":"LOW","reasoning":"Documentation only change."}
{"type":"summary","total":5,"high":1,"medium":1,"low":1,"skipped":2,"errors":0}
```

---

## File Filtering

The tool automatically skips files that rarely cause logic bugs:

| Category | Examples |
|----------|----------|
| **Lock files** | `go.sum`, `package-lock.json`, `yarn.lock`, `Cargo.lock`, `poetry.lock` |
| **Test files** | `*_test.go`, `*.test.js`, `*.spec.ts`, `test_*.py` |
| **Vendor/deps** | `vendor/`, `node_modules/` |
| **Build output** | `dist/`, `build/`, `out/` |
| **CI/CD** | `.github/workflows/`, `.gitlab-ci.yml`, `.travis.yml` |
| **IDE config** | `.idea/`, `.vscode/` |
| **Cache** | `__pycache__/`, `.pytest_cache/` |

---

## Limitations & Notes

-   **Token Usage:** Analyzing large commits or many files consumes significant context. The tool filters irrelevant files and truncates large diffs (>50KB) automatically.
-   **Rate Limits:** The tool includes automatic retry with exponential backoff for rate limit errors (429) and transient failures. Reduce `-j` workers if you still hit limits.
-   **API Key Security:** Prefer the `GEMINI_API_KEY` environment variable over `-apikey` flag (command-line args are visible in process lists).

## Development

### Running Tests

```bash
go test ./... -v
```

### Building

```bash
go build -o git-commit-analysis ./cmd/git-commit-analysis
```

## License

[MIT](LICENSE)
