# Implementation Plan - Stabilize & Refine LLM Reasoning

## Phase 1: Deterministic Model Configuration [checkpoint: 71a0d8c]
- [x] Task: Update `cmd/git-commit-analysis/main.go` to set the LLM `Temperature` to `0.1` using `genai.GenerationConfig`. cd0d768
- [x] Task: Conductor - User Manual Verification 'Phase 1: Deterministic Model Configuration' (Protocol in workflow.md)

## Phase 2: Reasoning Engine & Prompt Refinement [checkpoint: 0da3661]
- [x] Task: Update `buildPrompt` in `internal/analyzer/engine.go` to include the **Skeptic Persona** and the **Global Value Tracing** instruction. 12eaa0b
- [x] Task: Add the **Hypothesis Generation** step to the prompt template, requiring the model to analyze the bug description before the commit. 12eaa0b
- [x] Task: Update unit tests in `internal/analyzer/engine_test.go` to ensure the new prompt sections (Skeptic, Hypothesis, Tracing) are present in the generated string. 12eaa0b
- [x] Task: Conductor - User Manual Verification 'Phase 2: Reasoning Engine & Prompt Refinement' (Protocol in workflow.md)

## Phase 3: Stability & Depth Verification
- [ ] Task: Execute the tool 3 times against the known problematic commits (`be8f779e` and `1c932131`) and verify that results are consistent.
- [ ] Task: Audit the generated reasoning to ensure it explicitly traces the path of the `-2` value through the diffs as per the new instructions.
- [ ] Task: Conductor - User Manual Verification 'Phase 3: Stability & Depth Verification' (Protocol in workflow.md)
