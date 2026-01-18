# Product Definition - Git Dual-Context Analysis

## Initial Concept
Automated bug diagnosis using Dual-Context Diff Analysis and LLMs.

## Target Users
- **Software Developers and Tech Leads:** Identifying the root cause of complex regressions in large codebases.
- **SREs and DevOps Engineers:** Automating bug triaging within CI/CD pipelines to maintain system stability.
- **Security Researchers:** Analyzing commits for introduced vulnerabilities or "sleeper" bugs that manifest over time.

## Problem Statement
Git Dual-Context Analysis addresses the limitations of traditional debugging by solving:
- **Detection of "Sleeper Bugs":** Identifying issues caused by the interaction of a specific commit with subsequent code evolution (refactors, new features).
- **High Manual Debugging Effort:** Reducing the time spent manually triaging commits by automatically prioritizing likely culprits.
- **Contextual Blindness:** Providing an evolution-aware context for code changes that standard `git diff` cannot capture.

## Key Features
- **Dual-Context Analysis:** Simultaneously analyzes the "Standard Diff" (Micro) for developer intent and the "Full Comparison Diff" (Macro) for evolutionary context.
- **3-Tier Classification System:** Classifies probability as **High** (Smoking Gun), **Medium** (Ambiguous/Suspicious), or **Low** (Unrelated) to reduce false positives.
- **Chain-of-Thought Reasoning:** Forces the LLM to reason step-by-step (Micro -> Macro -> Classification) before delivering a verdict, improving diagnosis accuracy.
- **LLM Reasoning Engine:** Integrates with Google Gemini Pro to act as the core logic engine for complex code reasoning and diagnosis.
- **Smart Logic Filtering:** Automatically focuses on functional code changes while excluding irrelevant artifacts like lock files and documentation.
- **Unified JSON Stream:** Consolidates all tool output—including progress logs and analysis results—into a single, machine-readable NDJSON stream on `stdout`, allowing for complete automation and easy filtering with tools like `jq`.

## Success Criteria & Goals
- **Refined Accuracy:** Continuous improvement of analysis accuracy through optimized LLM prompting and model refinement.
- **Provider Flexibility:** Expansion of support for additional LLM providers and alternative version control systems.
- **Enhanced CLI Experience:** Delivery of robust reporting, unified NDJSON output for seamless automation, and powerful integration hooks for modern developer workflows.
