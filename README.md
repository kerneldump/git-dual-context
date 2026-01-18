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
    -   Generates **Standard Diffs** to understand developer intent.
    -   Generates **Full Comparison Diffs** to understand evolutionary context.
-   **LLM Integration:** Uses Google's Gemini Pro to act as the reasoning engine for probabilistic inference.
-   **Smart Filtering:** Automatically excludes lock files, documentation, and tests to focus on logic changes and conserve tokens.

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

### Output Format (NDJSON)

The tool outputs results in **Newline Delimited JSON (NDJSON)** format by default on `stdout`. This includes both progress logs and analysis results, which can be distinguished by the `type` field.

-   `"type": "result"`: Contains analysis findings (`hash`, `probability`, `reasoning`).
-   `"type": "log"`: Contains progress and status updates (`level`, `msg`, `timestamp`).

#### Pro-tip: Isolate Results with `jq`

To see only high-probability commits:

```bash
./git-commit-analysis -error="..." | jq 'select(.type=="result" and .probability=="HIGH")'
```

To silence logs and get a clean JSON stream:

```bash
./git-commit-analysis -error="..." | jq 'select(.type=="result")'
```

---

## Example Output

```json
{"type":"log","level":"INFO","msg":"Cloning https://github.com/... into temporary directory...","timestamp":"2026-01-17T17:15:00Z"}
{"type":"log","level":"INFO","msg":"Analyzing last 5 commits for error: \"interval must be greater than 0, got -2\"","timestamp":"2026-01-17T17:15:05Z"}
{"type":"result","hash":"be8f779e","probability":"HIGH","reasoning":"The commit modifies NewTimeFilter to accept negative durations instead of ignoring them, which eventually reaches a ticker validation check."}
{"type":"result","hash":"1c932131","probability":"MEDIUM","reasoning":"The commit modifies axis bounds calculation, which could potentially result in negative intervals in edge cases."}
{"type":"result","hash":"26cb336c","probability":"LOW","reasoning":"Documentation only change."}
```

### Flags

## Limitations & Notes

-   **Token Usage:** Analyzing large commits or many files consumes significant context. The tool attempts to filter irrelevant files (`.lock`, `_test.go`), but be mindful of costs.
-   **Rate Limits:** The tool processes commits concurrently (default: 3 workers). If you hit API rate limits, consider reducing `N` or adding retry logic.

## License

[MIT](LICENSE)