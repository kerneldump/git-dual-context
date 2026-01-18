# Implementation Plan - Unified JSON Output

## Phase 1: Engine & Logger Refinement [checkpoint: c821896]
- [x] Task: Update `JSONResult` to include the `"type": "result"` discriminator. b386887
- [x] Task: Define a `LogEntry` struct in `internal/analyzer` for structured logging. 40e09e2
- [x] Task: Implement a structured logger helper to ensure consistent JSON formatting for logs. fac9ffe
- [x] Task: Conductor - User Manual Verification 'Phase 1: Engine & Logger Refinement' (Protocol in workflow.md)

## Phase 2: CLI Unified Output
- [ ] Task: Update `cmd/git-commit-analysis/main.go` to use the new structured logger for all progress and status messages.
- [ ] Task: Redirect all remaining `stderr` output (cloning progress, etc.) to the structured logger on `stdout`.
- [ ] Task: Ensure the `json.Encoder` is shared or consistently used for all JSON output types.
- [ ] Task: Conductor - User Manual Verification 'Phase 2: CLI Unified Output' (Protocol in workflow.md)

## Phase 3: Verification
- [ ] Task: Verify that all stages of execution (Cloning, Analyzing, Results) appear as valid JSON on `stdout`.
- [ ] Task: Verify that `stderr` is now empty during normal execution.
- [ ] Task: Conductor - User Manual Verification 'Phase 3: Verification' (Protocol in workflow.md)
