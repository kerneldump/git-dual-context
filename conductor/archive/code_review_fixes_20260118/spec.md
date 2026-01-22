# Specification: Code Review Fixes

## Background
A comprehensive code review was conducted on the git-dual-context project, identifying issues across multiple categories: critical bugs, high-priority issues, medium-priority improvements, low-priority enhancements, testing gaps, security considerations, and performance optimizations.

## Goals
1. Fix all critical (P0) bugs that could cause incorrect analysis results
2. Address high-priority (P1) issues affecting reliability and UX
3. Implement medium-priority (P2) improvements for configurability and robustness
4. Add low-priority (P3) enhancements for better usability
5. Improve test coverage
6. Address security and performance concerns

## Requirements

### P0 - Critical Bug Fixes
1. **Parent Retrieval Error Handling**: The `c.Parent(0)` call must properly handle errors instead of silently ignoring them
2. **Response Parsing Safety**: Must detect when LLM response contains no text parts and return appropriate error

### P1 - High Priority
1. **Deterministic Output Order**: Results must print in commit order, not completion order
2. **Diff Context Lines**: Include context lines (unchanged code) in diffs for better LLM understanding
3. **Retry Logic**: Implement exponential backoff for rate limits, timeouts, and transient network errors

### P2 - Medium Priority
1. **Configurable Model/Timeout**: Add CLI flags for `-model` and `-timeout`
2. **Graceful Shutdown**: Handle SIGINT/SIGTERM and clean up properly
3. **Expanded File Filtering**: Add more lock files, test patterns, CI/CD files, build directories
4. **Diff Size Limits**: Truncate large diffs to prevent context window overflow

### P3 - Low Priority Enhancements
1. **Output File Option**: Add `-o` flag for file output
2. **Branch Selection**: Add `-branch` flag to analyze specific branches
3. **Summary Output**: Print HIGH/MEDIUM/LOW/Skipped/Error counts at end
4. **Verbose Mode**: Add `-v` flag for debug output
5. **Commit Message in Output**: Include truncated commit message in JSON results
6. **HEAD Commit Optimization**: Fetch HEAD once, pass to all goroutines

### Testing
1. Add tests for `ShouldIgnoreFile` function
2. Add tests for `TruncateDiff` function

### Security
1. Warn when API key passed via command line
2. Clean up temp directory on fatal exit

### Performance
1. Pre-allocate StringBuilder for diffs
2. Pre-size maps when length is known

## Success Criteria
- All tests pass
- Binary builds successfully
- New CLI options work as documented
- Output streams in order as results become available
