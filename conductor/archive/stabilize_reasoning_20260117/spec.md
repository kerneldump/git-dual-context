# Specification - Stabilize & Refine LLM Reasoning

## Overview
This track addresses the inconsistency and depth of automated bug diagnosis by refining the LLM configuration and the reasoning prompt. The goal is to move from "impressionistic" risk analysis to "deterministic" data flow tracing.

## Functional Requirements
- **Deterministic Configuration:**
    - Set model `Temperature` to `0.1`.
- **Reasoning Persona:**
    - Instruct the LLM to adopt a "Skeptic" persona: it must actively attempt to disprove that the commit introduced the bug unless clear evidence (a "smoking gun") exists.
- **Enhanced Prompting Logic:**
    - **Global Instruction:** Instruct the model to trace the specific values mentioned in the bug description (e.g., `-2`) through the provided diffs.
    - **Pre-Analysis Step (Hypothesis Generation):** Before analyzing the specific commit, the model must output a list of potential technical causes for the reported error based on the description alone.
    - **Step-by-Step Chain of Thought:** Maintain the 3-step process (Micro, Macro, Classification) but integrate the tracing and skepticism into each.

## Technical Changes
- `cmd/git-commit-analysis/main.go`: Configure `genai.GenerationConfig` with `Temperature: 0.1`.
- `internal/analyzer/engine.go`: Update `buildPrompt` to incorporate the Skeptic persona, Trace instruction, and Hypothesis step.

## Acceptance Criteria
- Running the tool multiple times on the same commit yields identical or highly similar reasoning and probability.
- The LLM's reasoning explicitly mentions how error-related values (like `-2`) interact with the code changes.
- The model correctly identifies "unmasked bugs" (where a commit makes a bad value reach a validation point) as HIGH or MEDIUM probability.

## Out of Scope
- Adding new LLM providers.
- Changing the 3-tier classification schema (High/Medium/Low).
