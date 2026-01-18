# Specification - JSON Output by Default

## Overview
This track modifies the `git-commit-analysis` tool to output results in a machine-readable JSON format by default. To preserve the streaming performance of the tool, results will be emitted as Newline Delimited JSON (NDJSON).

## Functional Requirements
- **Default Format:** The standard text-based output (with ANSI colors) will be replaced by NDJSON.
- **Output Schema:** Each line on `stdout` will be a valid JSON object containing:
    - `hash`: The short commit hash (string).
    - `probability`: The analysis probability (HIGH, MEDIUM, or LOW).
    - `reasoning`: The concise summary of the LLM's logic.
- **Streaming Support:** Objects must be printed to `stdout` immediately upon completion of each individual commit analysis.
- **Log Segregation:** All non-result output (e.g., "Cloning repo...", progress updates, and error logs) must be redirected to `stderr` to ensure `stdout` contains only valid JSON.

## Non-Functional Requirements
- **Performance:** Maintain the current concurrent processing and streaming architecture.
- **API Reliability:** Maintain the existing 300s timeout per request.

## Acceptance Criteria
- Running the tool produces one JSON object per line on `stdout`.
- Piping `stdout` to a JSON processor (like `jq`) works without syntax errors.
- Progress messages are visible in the terminal (via `stderr`) but do not interfere with piped `stdout`.

## Out of Scope
- Support for the legacy color-coded text format.
- Multi-line formatted JSON output (pretty-printing).
