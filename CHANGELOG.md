# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Input validation for all user inputs (commit count, workers, branch names, repo paths)
- Comprehensive test suite for retry logic with exponential backoff
- Shared utility function `TruncateCommitMessage` to eliminate code duplication
- Validation package (`pkg/validator`) for security and input sanitization
- Tests for utility functions in analyzer package
- Comprehensive concurrency documentation in `docs/CONCURRENCY.md`
- CONTRIBUTING.md with development guidelines
- CHANGELOG.md for tracking version changes

### Changed
- Improved error handling for JSON encoding operations in CLI
- Enhanced defensive programming with nil checks in gitdiff package
- Refactored message truncation logic to use shared utility

### Fixed
- Fixed ignored error returns from JSON encoding calls
- Added defensive nil check for parent tree in diff generation
- Improved error messages for validation failures

### Security
- Added maximum limits for commits (1000) and workers (50) to prevent resource exhaustion
- Added path validation to prevent directory traversal attacks
- Added branch name sanitization to prevent command injection
- Added protection against analyzing sensitive system directories

## [0.1.0] - 2025-01-XX

### Added
- Initial release of git-dual-context tool
- Dual-context diff analysis (standard diff + full comparison diff)
- LLM-powered bug diagnosis using Google Gemini
- CLI tool for analyzing git repositories
- MCP (Model Context Protocol) server for AI agent integration
- Smart file filtering (excludes lock files, tests, vendor directories)
- Ordered streaming output in NDJSON format
- Retry logic with exponential backoff for API failures
- Graceful shutdown handling
- Support for local and remote repositories
- Configurable concurrency with worker pools
- Comprehensive README with usage examples
- Theoretical framework documentation in `docs/GitCommitAnalysis.md`

### Features
- **Automated Hypothesis Testing**: Calculates probability that each commit caused a bug
- **Standard Diff Analysis**: Analyzes immediate changes (micro-context)
- **Full Comparison Diff**: Analyzes evolution to HEAD (macro-context)
- **Skeptical Reasoning**: LLM instructed to disprove culpability by default
- **Parallel Processing**: Concurrent analysis of multiple commits (CLI)
- **Sequential Processing**: Safe processing for long-lived server (MCP)
- **File Filtering**: Automatic exclusion of irrelevant files
- **Result Streaming**: Results output as they complete, in commit order

### CLI Options
- `-repo`: Repository path or URL
- `-branch`: Branch to analyze
- `-error`: Bug description
- `-n`: Number of commits to analyze
- `-j`: Number of concurrent workers
- `-model`: Gemini model selection
- `-timeout`: Per-commit timeout
- `-o`: Output file path
- `-apikey`: API key (prefer env var)
- `-v`: Verbose debug output

### MCP Tool
- `analyze_root_cause`: Root cause analysis tool for AI agents
  - Input: `repo_path`, `error_message`, `num_commits`, `branch`, `concurrency`
  - Output: Structured results with probability and reasoning

### Documentation
- Comprehensive README with examples
- Theoretical paper on dual-context analysis
- MCP server setup guide
- Package documentation (doc.go files)
- Usage examples in `examples/basic_usage/`

### Package Structure
- `pkg/analyzer`: Core reasoning engine and LLM integration
- `pkg/gitdiff`: Diff extraction and filtering
- `cmd/git-commit-analysis`: CLI tool
- `cmd/mcp-server`: MCP server for AI agents
- `examples/`: Usage examples

### Testing
- Unit tests for analyzer package
- Unit tests for gitdiff package
- Table-driven tests for edge cases
- Test coverage for JSON parsing, truncation, file filtering

## Release Notes

### Version 0.1.0 - Initial Release

This is the first public release of git-dual-context, implementing the theoretical framework described in "Enhanced Bug Diagnosis via Dual-Context Diff Analysis."

**Key Highlights:**
- Novel approach combining standard diffs with evolutionary diffs
- LLM-powered probabilistic reasoning for bug localization
- Production-ready CLI and MCP server implementations
- Comprehensive documentation and examples

**Use Cases:**
- Debugging complex bugs with unknown root causes
- Analyzing "sleeper bugs" that emerge over time
- Understanding how code changes interact across commits
- Automated root cause analysis in CI/CD pipelines

**Known Limitations:**
- Requires Gemini API key (OpenAI support planned)
- Large commits may be truncated (50KB limit)
- MCP server uses sequential processing for stability

**Future Roadmap:**
- Multi-LLM support (OpenAI, Claude, local models)
- Configurable prompt templates
- Enhanced visualization of analysis results
- Integration with issue tracking systems
- Support for analyzing across multiple branches

---

## Version Format

- **MAJOR**: Incompatible API changes
- **MINOR**: New features (backwards-compatible)
- **PATCH**: Bug fixes (backwards-compatible)

## Categories

- **Added**: New features
- **Changed**: Changes to existing functionality
- **Deprecated**: Soon-to-be removed features
- **Removed**: Removed features
- **Fixed**: Bug fixes
- **Security**: Security improvements

[Unreleased]: https://github.com/kerneldump/git-dual-context/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/kerneldump/git-dual-context/releases/tag/v0.1.0
