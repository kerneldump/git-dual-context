# Implementation Plan - JSON Output by Default

## Phase 1: Engine & Data Structure Refinement [checkpoint: 847a621]
- [x] Task: Define a dedicated `JSONResult` struct in `internal/analyzer` to ensure consistent serialization. 554b8f8
- [x] Task: Update `AnalyzeResult` or create a conversion helper to match the required output schema (`hash`, `probability`, `reasoning`). bbc26ba
- [x] Task: Conductor - User Manual Verification 'Phase 1: Engine & Data Structure Refinement' (Protocol in workflow.md)

## Phase 2: CLI Output & Logging Redirection [checkpoint: 7f3c4fe]
- [x] Task: Update `cmd/git-commit-analysis/main.go` to redirect all `fmt.Printf` and `log` calls (except results) to `os.Stderr`. 6ef1811
- [x] Task: Replace the color-coded text printing logic in the results goroutine with `json.Encoder` writing to `os.Stdout`. 3b34aeb
- [x] Task: Ensure each JSON object is followed by a newline (NDJSON format). 3b34aeb
- [x] Task: Conductor - User Manual Verification 'Phase 2: CLI Output & Logging Redirection' (Protocol in workflow.md)

## Phase 3: Verification & Polish [checkpoint: cec9034]
- [x] Task: Verify that `stdout` remains valid NDJSON even when errors occur (errors should go to `stderr`). 94896cb
- [x] Task: Test the output by piping to `jq` to ensure compatibility. 0715b68
- [x] Task: Conductor - User Manual Verification 'Phase 3: Verification & Polish' (Protocol in workflow.md)
