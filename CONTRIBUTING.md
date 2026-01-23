# Contributing to git-dual-context

Thank you for your interest in contributing to git-dual-context! This document provides guidelines and instructions for contributing.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Testing](#testing)
- [Code Style](#code-style)
- [Submitting Changes](#submitting-changes)
- [Reporting Bugs](#reporting-bugs)
- [Feature Requests](#feature-requests)

## Code of Conduct

This project adheres to a code of conduct that all contributors are expected to follow:

- Be respectful and inclusive
- Welcome newcomers and beginners
- Focus on constructive feedback
- Assume good intentions
- Respect differing viewpoints and experiences

## Getting Started

### Prerequisites

- **Go 1.21+** installed
- **Git** for version control
- **Google Gemini API Key** for testing (get one [here](https://makersuite.google.com/app/apikey))
- Familiarity with Git, Go, and LLMs

### Development Setup

1. **Fork the Repository**
   ```bash
   # Fork on GitHub, then clone your fork
   git clone https://github.com/YOUR-USERNAME/git-dual-context.git
   cd git-dual-context
   ```

2. **Add Upstream Remote**
   ```bash
   git remote add upstream https://github.com/kerneldump/git-dual-context.git
   ```

3. **Install Dependencies**
   ```bash
   go mod download
   ```

4. **Build the Project**
   ```bash
   make build
   ```

5. **Run Tests**
   ```bash
   make test
   ```

6. **Set Up Environment**
   ```bash
   export GEMINI_API_KEY="your-api-key-here"
   ```

## Making Changes

### Branching Strategy

- **main**: Stable, production-ready code
- **feature/**: New features (`feature/add-claude-support`)
- **fix/**: Bug fixes (`fix/race-condition`)
- **docs/**: Documentation updates (`docs/improve-readme`)
- **refactor/**: Code refactoring (`refactor/extract-utils`)

### Workflow

1. **Create a Branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make Your Changes**
   - Write clear, self-documenting code
   - Add comments for complex logic
   - Follow the existing code style
   - Keep changes focused and atomic

3. **Write Tests**
   - Add unit tests for new functionality
   - Ensure existing tests pass
   - Aim for >80% code coverage

4. **Update Documentation**
   - Update README.md if adding features
   - Add/update code comments
   - Update CHANGELOG.md

5. **Commit Your Changes**
   ```bash
   git add .
   git commit -m "feat: add support for Claude API"
   ```

### Commit Message Format

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks
- `perf`: Performance improvements

**Examples:**
```
feat(analyzer): add support for OpenAI GPT-4

- Implement OpenAI client wrapper
- Add model selection flag
- Update documentation

Closes #42
```

```
fix(gitdiff): prevent panic on nil parent tree

Add defensive nil check before calling pTree.Patch()
to handle first commit edge case gracefully.

Fixes #38
```

## Testing

### Running Tests

```bash
# Run all tests
make test

# Run tests with verbose output
go test ./... -v

# Run tests with race detector
go test ./... -race

# Run tests with coverage
go test ./... -cover

# Generate coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Writing Tests

#### Unit Tests
- Place tests in `*_test.go` files
- Use table-driven tests for multiple scenarios
- Test both success and error cases
- Mock external dependencies (LLM API calls)

Example:
```go
func TestValidateNumCommits(t *testing.T) {
    tests := []struct {
        name    string
        input   int
        wantErr bool
    }{
        {"valid", 5, false},
        {"zero", 0, true},
        {"negative", -1, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateNumCommits(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateNumCommits(%d) error = %v, wantErr %v",
                    tt.input, err, tt.wantErr)
            }
        })
    }
}
```

#### Integration Tests
- Test complete workflows end-to-end
- Use test fixtures or temporary repositories
- Clean up resources after tests

### Test Coverage Requirements

- **Core packages** (`pkg/`): >80% coverage
- **Command packages** (`cmd/`): >50% coverage
- **Critical paths**: 100% coverage (validation, error handling)

## Code Style

### Go Style Guide

Follow the [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md) and [Effective Go](https://go.dev/doc/effective_go):

- Use `gofmt` for formatting: `make fmt`
- Use `go vet` for static analysis: `make vet`
- Use `golangci-lint` for linting: `make lint`

### Best Practices

1. **Error Handling**
   ```go
   // Good: wrap errors with context
   if err != nil {
       return fmt.Errorf("failed to analyze commit %s: %w", hash, err)
   }

   // Bad: ignore or swallow errors
   result, _ := DoSomething()
   ```

2. **Variable Naming**
   ```go
   // Good: clear, descriptive names
   commitMessage := "Fix authentication bug"
   analysisResult, err := analyzer.AnalyzeCommit(...)

   // Bad: cryptic abbreviations
   msg := "Fix authentication bug"
   res, e := analyzer.AnalyzeCommit(...)
   ```

3. **Function Size**
   - Keep functions small and focused
   - Extract complex logic into helper functions
   - Aim for <50 lines per function

4. **Comments**
   ```go
   // Good: explain WHY, not WHAT
   // Sequential processing avoids go-git race conditions in long-lived server
   for i, c := range commits {

   // Bad: redundant comment
   // Loop through commits
   for i, c := range commits {
   ```

## Submitting Changes

### Pull Request Process

1. **Update Your Branch**
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Push to Your Fork**
   ```bash
   git push origin feature/your-feature-name
   ```

3. **Create Pull Request**
   - Go to GitHub and create a PR from your fork
   - Fill out the PR template completely
   - Link related issues
   - Request review from maintainers

### Pull Request Template

```markdown
## Description
Brief description of what this PR does.

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing performed

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Comments added for complex logic
- [ ] Documentation updated
- [ ] Tests pass locally
- [ ] No new warnings introduced
```

### Review Process

- Maintainers will review your PR within 7 days
- Address feedback by pushing new commits
- Once approved, maintainers will merge

## Reporting Bugs

### Before Submitting

1. **Search existing issues** to avoid duplicates
2. **Reproduce the bug** consistently
3. **Test on the latest version**

### Bug Report Template

```markdown
## Bug Description
Clear description of the bug.

## Steps to Reproduce
1. Clone repository
2. Run command: `./git-commit-analysis -error="..." -n 10`
3. Observe error

## Expected Behavior
What you expected to happen.

## Actual Behavior
What actually happened.

## Environment
- OS: macOS 13.0
- Go version: 1.21
- Tool version: v0.1.0

## Additional Context
Logs, screenshots, etc.
```

## Feature Requests

We welcome feature requests! Please:

1. **Check existing issues** first
2. **Describe the use case** clearly
3. **Explain the benefits**
4. **Consider implementation complexity**

### Feature Request Template

```markdown
## Feature Description
What feature would you like to see?

## Use Case
What problem does this solve?

## Proposed Solution
How would you implement this?

## Alternatives Considered
What other approaches did you think about?
```

## Development Tips

### Local Testing

```bash
# Test CLI with local repo
./git-commit-analysis -repo="." -error="test error" -n 5 -v

# Test with remote repo
./git-commit-analysis \
  -repo="https://github.com/user/repo.git" \
  -error="connection timeout" \
  -n 10

# Test MCP server
cd cmd/mcp-server
go run main.go
```

### Debugging

```bash
# Enable verbose logging
./git-commit-analysis -v -error="..." -n 5

# Use delve debugger
dlv debug ./cmd/git-commit-analysis -- -error="..." -n 5

# Check for race conditions
go run -race ./cmd/git-commit-analysis -error="..." -j 10
```

### Performance Profiling

```bash
# CPU profiling
go test -cpuprofile=cpu.prof ./pkg/analyzer
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof ./pkg/analyzer
go tool pprof mem.prof
```

## Questions?

- **GitHub Issues**: For bug reports and feature requests
- **GitHub Discussions**: For questions and general discussion
- **Email**: contact@example.com (TODO: update)

Thank you for contributing to git-dual-context! 🎉
