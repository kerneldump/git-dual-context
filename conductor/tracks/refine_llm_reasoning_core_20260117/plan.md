# Implementation Plan - Refine LLM Reasoning Core

## Phase 1: Analysis & Prompt Engineering
- [ ] Task: Audit current prompts in `internal/analyzer` and identify areas for improvement.
- [ ] Task: Research and implement "Chain-of-Thought" prompting techniques for better code reasoning.
- [ ] Task: Create a set of "Micro" and "Macro" diff examples to use for iterative prompt testing.
- [ ] Task: Conductor - User Manual Verification 'Phase 1: Analysis & Prompt Engineering' (Protocol in workflow.md)

## Phase 2: Core Engine Refinement
- [ ] Task: Implement improved prompt construction logic in `internal/analyzer/engine.go`.
    - [ ] Write unit tests for the new prompt generator.
    - [ ] Update `engine.go` with refined prompt templates.
- [ ] Task: Refine the probabilistic calculation logic.
    - [ ] Write tests for the weight distribution between Micro and Macro signals.
    - [ ] Adjust the mathematical model in the analyzer.
- [ ] Task: Conductor - User Manual Verification 'Phase 2: Core Engine Refinement' (Protocol in workflow.md)

## Phase 3: Verification & Polish
- [ ] Task: Run the refined analyzer against previous known bug scenarios to verify accuracy improvement.
- [ ] Task: Polish the CLI output for reasoning to match the new "Technical & Precise" product guidelines.
- [ ] Task: Conductor - User Manual Verification 'Phase 3: Verification & Polish' (Protocol in workflow.md)
