# Specification - Refine LLM Reasoning Core

## Overview
This track focuses on improving the accuracy and depth of the automated bug diagnosis by refining the reasoning engine. This involves optimizing the prompts sent to the Gemini model and adjusting the probabilistic framework that calculates the likelihood of a commit being the root cause of a bug.

## Objectives
- Improve the precision of the LLM's identification of "sleeper bugs."
- Enhance the clarity and depth of the reasoning provided by the LLM for its findings.
- Ensure the probabilistic model correctly weighs signals from both the Micro (Standard Diff) and Macro (Full Comparison Diff) contexts.

## Scope
- `internal/analyzer/engine.go`: Refine the logic for prompt construction and response parsing.
- Prompts: Iterate on the system and user prompts to provide better context and guidance to the LLM.
- Probability Calculation: Adjust the weights and logic used to derive $P(H_k|E)$.

## Success Criteria
- Higher confidence scores for known bug-causing commits in test datasets.
- More detailed and actionable "Reasoning" outputs from the tool.
- Reduced false-positive rate for commits that are logically sound but evolutionary complex.
