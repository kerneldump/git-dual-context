// Package analyzer orchestrator provides shared analysis orchestration logic.
package analyzer

import (
	"context"
	"fmt"
	"io"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/google/generative-ai-go/genai"
)

// AnalysisOptions configures the analysis orchestration.
type AnalysisOptions struct {
	// NumCommits is the number of commits to analyze
	NumCommits int

	// Branch is the branch to analyze (empty for HEAD)
	Branch string

	// ErrorMessage is the bug description to analyze
	ErrorMessage string

	// OnProgress is called with progress messages (optional)
	OnProgress func(msg string)
}

// CommitAnalysisResult represents the result of analyzing a single commit.
type CommitAnalysisResult struct {
	Index   int
	Hash    string
	Message string
	Result  *AnalysisResult
	Error   error
}

// AnalysisSummary represents the summary of an analysis run.
type AnalysisSummary struct {
	Total   int
	High    int
	Medium  int
	Low     int
	Skipped int
	Errors  int
}

// CollectCommits gathers commits from a repository for analysis.
// It skips merge commits and respects the branch and numCommits options.
//
// Two-Phase Analysis Architecture:
// To safely enable parallel LLM calls while respecting go-git's thread-safety
// limitations, use a two-phase approach:
//
//   Phase 1 (Sequential): Extract diffs using ExtractDiffs() - go-git operations
//   Phase 2 (Parallel):   Analyze with AnalyzeWithDiffs() - LLM API calls
//
// This allows maximum parallelism for the expensive LLM calls while keeping
// git operations sequential. See ExtractDiffs and AnalyzeWithDiffs in engine.go.
func CollectCommits(repo *git.Repository, opts AnalysisOptions) ([]*object.Commit, *object.Commit, error) {
	if opts.NumCommits <= 0 {
		opts.NumCommits = DefaultNumCommits
	}

	// Get HEAD reference (or specified branch)
	var headRef *plumbing.Reference
	var err error

	if opts.Branch != "" {
		refName := plumbing.NewBranchReferenceName(opts.Branch)
		headRef, err = repo.Reference(refName, true)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to find branch %s: %w", opts.Branch, err)
		}
	} else {
		headRef, err = repo.Head()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get HEAD: %w", err)
		}
	}

	// Get HEAD commit for comparison
	headCommit, err := repo.CommitObject(headRef.Hash())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get HEAD commit: %w", err)
	}

	// Collect commits
	cIter, err := repo.Log(&git.LogOptions{From: headRef.Hash()})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get commit log: %w", err)
	}

	var commits []*object.Commit
	count := 0

	for count < opts.NumCommits {
		c, err := cIter.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, fmt.Errorf("error iterating commits: %w", err)
		}

		// Skip merge commits
		if len(c.ParentHashes) > 1 {
			continue
		}

		commits = append(commits, c)
		count++
	}

	return commits, headCommit, nil
}

// AnalyzeCommitSequential analyzes a single commit with retry logic.
// This is the recommended approach for maximum reliability.
func AnalyzeCommitSequential(
	ctx context.Context,
	repo *git.Repository,
	commit *object.Commit,
	headCommit *object.Commit,
	errorMessage string,
	model *genai.GenerativeModel,
) (*AnalysisResult, error) {
	var res *AnalysisResult
	err := WithRetry(ctx, DefaultRetryConfig(), func() error {
		var analyzeErr error
		res, analyzeErr = AnalyzeCommit(ctx, repo, commit, headCommit, errorMessage, model)
		return analyzeErr
	})
	return res, err
}

// CalculateSummary computes summary statistics from analysis results.
func CalculateSummary(results []CommitAnalysisResult) AnalysisSummary {
	summary := AnalysisSummary{
		Total: len(results),
	}

	for _, r := range results {
		if r.Error != nil {
			summary.Errors++
			continue
		}
		if r.Result == nil {
			summary.Errors++
			continue
		}
		if r.Result.Skipped {
			summary.Skipped++
			continue
		}

		switch r.Result.Probability {
		case ProbHigh:
			summary.High++
		case ProbMedium:
			summary.Medium++
		case ProbLow:
			summary.Low++
		}
	}

	return summary
}
