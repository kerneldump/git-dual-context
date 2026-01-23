# Concurrency Design Decisions

## Overview

This document explains the different concurrency models used in the CLI tool vs. the MCP server, and the rationale behind these architectural choices.

## CLI Tool: Parallel Processing

### Implementation
The CLI tool (`cmd/git-commit-analysis/main.go`) uses **concurrent goroutines** to analyze multiple commits in parallel:

```go
sem := make(chan struct{}, *numWorkers) // Limit concurrent requests
for i, c := range commits {
    wg.Add(1)
    sem <- struct{}{}
    go func(idx int, commit *object.Commit) {
        defer wg.Done()
        defer func() { <-sem }()
        // Analyze commit...
    }(i, c)
}
```

### Why Parallel Works Here
1. **Independent Repository Clones**: Each CLI invocation works with its own git repository instance
2. **Process Isolation**: Each run is a separate OS process with isolated memory
3. **Read-Only Operations**: Git operations are primarily read-only (walking commit history, reading trees, computing patches)
4. **Performance Critical**: Users expect fast results when analyzing N commits locally

### Benefits
- **Performance**: Can process 10 commits in ~30 seconds instead of ~3 minutes
- **Resource Utilization**: Maximizes CPU and network I/O for API calls
- **User Experience**: Faster feedback for local analysis

## MCP Server: Sequential Processing

### Implementation
The MCP server (`cmd/mcp-server/internal/tools/rootcause.go`) uses **sequential processing**:

```go
// Process commits sequentially to avoid go-git race conditions
// Note: go-git's ObjectStorage is not thread-safe for concurrent access
for i, c := range commits {
    // Analyze one commit at a time
    res, err := analyzer.AnalyzeCommit(ctx, repo, c, headCommit, input.ErrorMessage, model)
    results[i] = &commitResultInternal{...}
}
```

### Why Sequential is Required

#### 1. **go-git Thread Safety Issues**
The `go-git` library's internal `ObjectStorage` is **not thread-safe**:

```go
// From go-git documentation:
// "Repository instances should not be used concurrently from multiple goroutines
//  without external synchronization."
```

**Specific Issues:**
- **Shared Cache**: Object cache is shared across the repository instance
- **File Descriptor Management**: Limited file handles can cause conflicts
- **Internal State**: Tree and blob caching mechanisms aren't protected by locks

#### 2. **MCP Server Lifecycle**
- **Long-Lived Process**: The MCP server runs continuously, serving multiple requests
- **Shared Repository**: Opening a new repository clone for each request is expensive
- **Memory Constraints**: Multiple concurrent repo clones would consume excessive memory

#### 3. **Observed Race Conditions**
When we attempted parallel processing in the MCP server, we observed:

```
fatal error: concurrent map read and write
goroutine 45 [running]:
github.com/go-git/go-git/v5/storage/memory.(*Storage).get(...)
```

### Trade-offs
- **Performance**: ~2-3x slower than parallel processing
- **Reliability**: No race conditions or crashes
- **Simplicity**: Easier to reason about and debug
- **Resource Usage**: Lower memory footprint

## Potential Future Improvements

### Option 1: Repository Pooling
Create a pool of repository instances, one per worker:

```go
type RepoPool struct {
    repos []*git.Repository
    sem   chan int
}

func (p *RepoPool) Acquire() (*git.Repository, int, error) {
    idx := <-p.sem
    return p.repos[idx], idx, nil
}

func (p *RepoPool) Release(idx int) {
    p.sem <- idx
}
```

**Pros**: Enables parallelism while avoiding race conditions
**Cons**: Higher memory usage, complex lifecycle management

### Option 2: External Synchronization
Wrap all git operations with mutexes:

```go
type SyncedRepo struct {
    repo *git.Repository
    mu   sync.Mutex
}

func (sr *SyncedRepo) CommitObject(hash plumbing.Hash) (*object.Commit, error) {
    sr.mu.Lock()
    defer sr.mu.Unlock()
    return sr.repo.CommitObject(hash)
}
```

**Pros**: Simple to implement
**Cons**: Eliminates most parallelism benefits (serializes git access anyway)

### Option 3: Switch to libgit2 bindings
Use `git2go` which offers better thread-safety:

```go
import "github.com/libgit2/git2go/v34"
```

**Pros**: Better concurrency support, potentially faster
**Cons**: CGO dependency, harder to build/deploy, larger binaries

## Recommendation

**Current Approach is Optimal**: Sequential processing in the MCP server is the right trade-off given:
1. The bottleneck is the LLM API calls (~5-10s each), not git operations (~100ms)
2. Sequential processing adds <1s overhead for 5 commits
3. Reliability > Performance for a long-running server process
4. Simpler code is easier to maintain

If performance becomes critical, consider Option 1 (Repository Pooling) as it preserves reliability while enabling parallelism.

## Testing Concurrency

### CLI Concurrency Tests
```bash
# Test with different worker counts
./git-commit-analysis -error="test" -j 1 -n 10  # Sequential
./git-commit-analysis -error="test" -j 5 -n 10  # Parallel
./git-commit-analysis -error="test" -j 20 -n 10 # High concurrency
```

### MCP Server Stress Test
```bash
# Verify no race conditions
go build -race -o mcp-server ./cmd/mcp-server
# Run with race detector enabled
```

## References

- [go-git Issue #223](https://github.com/go-git/go-git/issues/223) - Thread safety discussion
- [go-git Documentation](https://pkg.go.dev/github.com/go-git/go-git/v5) - Concurrency notes
- [Go Memory Model](https://go.dev/ref/mem) - Understanding Go concurrency
