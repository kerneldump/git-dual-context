# Specification - Unified JSON Output

## Overview
This track consolidates all tool output (analysis results and progress logs) into a single stream on `stdout`, with every line being a valid JSON object. This allows for complete machine readability of the entire tool execution.

## Functional Requirements
- **Unified Stream:** Redirect all output currently on `stderr` to `stdout`.
- **JSON for Everything:** Every line emitted by the tool must be a valid JSON object.
- **Distinction by Type:** Every JSON object must include a `type` field to distinguish the content:
    - `"type": "result"`: Contains analysis findings (the current `JSONResult` schema).
    - `"type": "log"`: Contains progress, status, or error messages.
- **Log Schema:** Log objects (`"type": "log"`) must include:
    - `level`: The severity (e.g., "INFO", "ERROR").
    - `msg`: The descriptive message.
    - `timestamp`: ISO 8601 formatted string of the log event.
- **Result Schema:** Result objects (`"type": "result"`) will continue to include `hash`, `probability`, and `reasoning`.

## Acceptance Criteria
- Running the tool with `2> /dev/null` yields no loss of information on `stdout`.
- Piping the entire output to `jq` works without errors.
- External tools can filter results vs. logs using the `type` field.

## Out of Scope
- Support for any non-JSON output format.
