# Audit of LLM Prompts and Analysis Logic

**Date:** 2026-01-17
**File Audited:** `internal/analyzer/engine.go`

## Current State
- **Prompt Construction:** Hardcoded `fmt.Sprintf` template within `AnalyzeCommit`.
- **Classification:** Supports "High", "Low", "Unknown".
- **Instructions:** Basic instructions to analyze Standard and Full diffs.
- **Output:** JSON `{ "probability": "...", "reasoning": "..." }`.

## Identified Areas for Improvement

### 1. Classification System
- **Current:** Binary-ish (High/Low/Unknown).
- **Goal:** Explicit 3-Tier (High, Medium, Low).
- **Missing:** The prompt lacks definitions for what constitutes High vs Low, leading to ambiguity.

### 2. Prompt Engineering
- **Context:** The "System" persona is minimal ("expert software debugger").
- **Reasoning:** "Concise reasoning" instruction may limit depth.
- **Chain of Thought:** No instruction to think step-by-step or weigh evidence before deciding.
- **Missing Definitions:**
    - **High:** Needs "Smoking Gun" / Direct contradiction definition.
    - **Medium:** Needs "Relevant Subsystem" / Ambiguous logic definition.
    - **Low:** Needs "Unrelated" definition.

### 3. Parsing & Robustness
- **JSON Parsing:** Uses `json.Unmarshal` with simple markdown stripping. This is fragile if the LLM includes preamble text.
- **Validation:** No validation that the returned probability matches the allowed enum values.

### 4. Recommendations for Implementation
- Refactor the prompt into a dedicated constant or template function.
- Add explicit "Definitions" section to the prompt.
- Update `AnalysisResult` to normalize probability strings (e.g., uppercase).
- Implement a more robust response parser that can handle "Reasoning first, JSON second" or loose JSON.
