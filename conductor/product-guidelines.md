# Product Guidelines - Git Dual-Context Analysis

## Editorial & Prose Style
- **Technical & Precise:** User-facing documentation and CLI output must prioritize accuracy and technical depth. Use industry-standard Git and LLM terminology.
- **Clarity Above All:** Complex theoretical concepts (like Dual-Context Diff) should be explained with precision but made accessible through clear, structured language.

## Visual & CLI Design
- **Clarity & Hierarchy:** CLI output must utilize clear headings, consistent indentation, and logical spacing to ensure complex analysis results are easily scannable and digestible.
- **Visual Cues:** Use standard symbols and potentially ANSI colors to highlight critical findings (e.g., high probability scores) and tool status without over-complicating the interface.

## Engineering Principles
- **Modularity & Extensibility:** The architecture must allow for easy swapping of core components, such as LLM providers (Gemini, OpenAI, Anthropic) and Diff engines, to future-proof the tool.
- **Performance Efficiency:** Prioritize efficient resource management and low latency, particularly when interacting with large Git repositories and performing intensive LLM calls.
- **Robustness & Error Handling:** Implement comprehensive error handling that provides informative, actionable messages to the user, especially when dealing with external API failures or unexpected repository states.

## User Experience (UX) & Interaction
- **Transparency & Justification:** Every analysis result must be accompanied by the tool's reasoning. The user should understand *why* a specific commit was flagged with a certain probability.
- **User-Centric Control:** Provide clear hooks and flags for users to customize the analysis scope (e.g., depth of history, excluded file patterns).
- **Managing Uncertainty:** Clearly communicate the difference between high-confidence findings and suggestive signals. Avoid presenting probabilistic results as absolute certainties.
