# Implementation Plan - Refine LLM Reasoning Core

## Phase 1: Analysis & Prompt Engineering [checkpoint: efbc22e]
- [x] Task: Audit current prompts in `internal/analyzer` and identify areas for improvement. bd2f800
- [x] Task: Research and implement "Chain-of-Thought" prompting techniques for better code reasoning. a60a696
- [x] Task: Create a set of "Micro" and "Macro" diff examples to use for iterative prompt testing. 678cd92
- [ ] Task: Conductor - User Manual Verification 'Phase 1: Analysis & Prompt Engineering' (Protocol in workflow.md)

## Phase 2: Core Engine Refinement (3-Tier Classification) [checkpoint: 9d45ba2]
- [x] Task: Update Data Structures to support "Medium" probability. d67287a
    - [ ] Modify `AnalysisResult` (or equivalent) to use an Enum or normalized string for High/Medium/Low.
- [x] Task: Update LLM System Prompt. b753d06
    - [ ] Modify prompt template to explicitly define criteria for High, Medium, and Low.
    - [ ] Instruction: Use "Medium" if commit touches relevant files/vars but lacks obvious flaws.
- [x] Task: Update Parsing Logic. 84777f9
    - [ ] Update regex/parsing to detect "Prob: Medium" (case-insensitive).
    - [ ] Ensure fallback to "Low" on parsing failure.
- [x] Task: Update CLI Output. 18856c6
    - [ ] Format output labels (e.g., `[MED]` or `Prob: Medium`).
    - [ ] Apply color coding (Red=High, Yellow=Medium, Green/Gray=Low).
    - [ ] Ensure sorting order is High -> Medium -> Low.
- [ ] Task: Conductor - User Manual Verification 'Phase 2: Core Engine Refinement' (Protocol in workflow.md)

## Phase 3: Verification & Polish
- [x] Task: Run the refined analyzer against previous known bug scenarios to verify accuracy improvement. 292e079
- [x] Task: Polish the CLI output for reasoning to match the new "Technical & Precise" product guidelines. 18856c6
- [ ] Task: Conductor - User Manual Verification 'Phase 3: Verification & Polish' (Protocol in workflow.md)
