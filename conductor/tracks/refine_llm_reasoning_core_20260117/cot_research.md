# Chain-of-Thought (CoT) Prompting Strategy

**Date:** 2026-01-17
**Goal:** Improve bug diagnosis accuracy by forcing the LLM to reason step-by-step.

## Strategy
We will implement a "Zero-Shot CoT" or "Few-Shot CoT" approach (depending on token limits, starting with Zero-Shot instructions).
The prompt will explicitly instruct the model to follow a reasoning framework *before* outputting the JSON decision.

## Proposed Reasoning Framework

1.  **Micro-Analysis (Standard Diff):**
    *   What exactly changed?
    *   Does this change look suspicious in isolation?

2.  **Macro-Analysis (Full Comparison):**
    *   How did this file evolve?
    *   Was this change reverted?
    *   Does it conflict with the current state (HEAD)?

3.  **Synthesis & Classification:**
    *   Does the change *directly* cause the error? -> **HIGH**
    *   Is it related but ambiguous? -> **MEDIUM**
    *   Is it unrelated? -> **LOW**

## Draft Prompt Template

```text
You are an expert software debugger. Your task is to analyze a specific commit to determine if it introduced the bug described below.

BUG DESCRIPTION:
{{.ErrorMsg}}

COMMIT CONTEXT:
Hash: {{.CommitHash}}
Message: {{.CommitMessage}}

---
INPUT DATA:

1. STANDARD DIFF (The immediate changes in this commit):
{{.StandardDiff}}

2. FULL COMPARISON DIFF (Evolution from this commit to HEAD):
{{.FullDiff}}

---
INSTRUCTIONS:

Use the following "Chain of Thought" process to analyze the data. You must output your reasoning for each step.

STEP 1: MICRO-ANALYSIS
Analyze the Standard Diff. What logic changed? Does it look risky?

STEP 2: MACRO-ANALYSIS
Analyze the Full Comparison Diff. Does the code from this commit still exist in HEAD? Was it refactored? Does it conflict with the current system state?

STEP 3: CLASSIFICATION
Classify the probability based on these strict definitions:
- HIGH: The commit contains logic that DIRECTLY contradicts the error message or introduces the specific bug (a "smoking gun").
- MEDIUM: The commit modifies the relevant subsystem or variables, but the logic is not clearly broken. Warrants manual review.
- LOW: The commit is unrelated (docs, assets, different subsystem, safe refactor).

---
OUTPUT FORMAT:

Reasoning: <Your Step-by-Step Chain of Thought Analysis>
Classification: <HIGH|MEDIUM|LOW>

Finally, return the result in this JSON format (do not use markdown blocks):
{
  "probability": "HIGH|MEDIUM|LOW",
  "reasoning": "A concise summary of your analysis."
}
```
