# Specification - Configuration Defaults & Metrics Update

## Goals
1.  **Optimize Defaults:** Update default settings to reflect current best practices and model availability.
    *   Timeout: Increase to 10 minutes to handle larger diffs/slower responses.
    *   Model: Switch to `gemini-flash-latest` for cost/speed efficiency.
2.  **Enhance Observability:** Add execution duration to the analysis summary to track performance.

## Requirements

### Configuration
- Default timeout must be 10 minutes (up from 5m).
- Default model must be `gemini-flash-latest` (changed from `gemini-3-pro-preview`).
- Changes must apply to:
    - Default constants in code.
    - Configuration file loading logic.
    - CLI flags defaults.
    - MCP server environment defaults.
    - Documentation (`README.md`, `config.example.yaml`).

### Metrics
- `Summary` struct (JSON output) must include a `duration` field.
- CLI output must show duration in the JSON summary.
- MCP server output must show duration in both JSON and text formats.
- Duration must capture the total analysis time (including diff extraction and LLM processing).

## Non-Functional Requirements
- Backward compatibility for existing config files (unless they rely on implicit defaults).
- Thread-safety for timing metrics in concurrent execution.
