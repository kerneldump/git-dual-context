package tools

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/kerneldump/git-dual-context/pkg/analyzer"
	"github.com/kerneldump/git-dual-context/pkg/validator"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// AnalyzeInput represents the input parameters for the analyze_root_cause tool
type AnalyzeInput struct {
	RepoPath     string `json:"repo_path" required:"true" description:"Path to local git repository"`
	ErrorMessage string `json:"error_message" required:"true" description:"Bug description or error message to diagnose"`
	NumCommits   int    `json:"num_commits,omitempty" description:"Number of recent commits to analyze (default: 5)"`
	Branch       string `json:"branch,omitempty" description:"Branch to analyze (default: current HEAD)"`
	Concurrency  int    `json:"concurrency,omitempty" description:"Number of concurrent workers (default: 3)"`
}

// CommitResult represents the analysis result for a single commit
type CommitResult struct {
	Hash        string `json:"hash"`
	Message     string `json:"message"`
	Probability string `json:"probability"`
	Reasoning   string `json:"reasoning"`
}

// AnalyzeSummary represents the summary of the analysis
type AnalyzeSummary struct {
	Total   int `json:"total"`
	High    int `json:"high"`
	Medium  int `json:"medium"`
	Low     int `json:"low"`
	Skipped int `json:"skipped"`
	Errors  int `json:"errors"`
}

// AnalyzeOutput represents the output of the analyze_root_cause tool
type AnalyzeOutput struct {
	Results []CommitResult `json:"results"`
	Summary AnalyzeSummary `json:"summary"`
}

// commitWork holds the work item for concurrent processing
type commitWork struct {
	index  int
	commit *object.Commit
}

// commitResultInternal holds the internal result for ordering
type commitResultInternal struct {
	index  int
	result *analyzer.AnalysisResult
	commit *object.Commit
	err    error
}

// AnalyzeRootCause performs dual-context analysis on a git repository
func AnalyzeRootCause(ctx context.Context, input AnalyzeInput, progress func(string)) (*AnalyzeOutput, error) {
	// Apply defaults
	if input.NumCommits <= 0 {
		input.NumCommits = 5
	}
	if input.Concurrency <= 0 {
		input.Concurrency = 3
	}

	// Validate inputs
	if err := validator.ValidateErrorMessage(input.ErrorMessage); err != nil {
		return nil, fmt.Errorf("invalid error message: %w", err)
	}
	if err := validator.ValidateNumCommits(input.NumCommits); err != nil {
		return nil, fmt.Errorf("invalid number of commits: %w", err)
	}
	if err := validator.ValidateNumWorkers(input.Concurrency); err != nil {
		return nil, fmt.Errorf("invalid concurrency value: %w", err)
	}
	if err := validator.ValidateBranchName(input.Branch); err != nil {
		return nil, fmt.Errorf("invalid branch name: %w", err)
	}
	if err := validator.ValidateRepoPath(input.RepoPath); err != nil {
		return nil, fmt.Errorf("invalid repository path: %w", err)
	}

	// Get API key from environment
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable is required")
	}

	// Get model from environment or use default
	modelName := os.Getenv("GEMINI_MODEL")
	if modelName == "" {
		modelName = "gemini-3-pro-preview"
	}

	// Open the repository
	repo, err := git.PlainOpen(input.RepoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository at %s: %w", input.RepoPath, err)
	}

	// Get HEAD reference (or specified branch)
	var headRef *plumbing.Reference
	if input.Branch != "" {
		refName := plumbing.NewBranchReferenceName(input.Branch)
		headRef, err = repo.Reference(refName, true)
		if err != nil {
			return nil, fmt.Errorf("failed to find branch %s: %w", input.Branch, err)
		}
	} else {
		headRef, err = repo.Head()
		if err != nil {
			return nil, fmt.Errorf("failed to get HEAD: %w", err)
		}
	}

	// Get HEAD commit for comparison
	headCommit, err := repo.CommitObject(headRef.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD commit: %w", err)
	}

	// Initialize Gemini client
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}
	defer client.Close()

	model := client.GenerativeModel(modelName)
	model.SetTemperature(0.1)

	// Collect commits
	cIter, err := repo.Log(&git.LogOptions{From: headRef.Hash()})
	if err != nil {
		return nil, fmt.Errorf("failed to get commit log: %w", err)
	}

	var commits []*object.Commit
	count := 0
	for count < input.NumCommits {
		c, err := cIter.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error iterating commits: %w", err)
		}

		// Skip merge commits
		if len(c.ParentHashes) > 1 {
			continue
		}

		commits = append(commits, c)
		count++
	}

	if len(commits) == 0 {
		return &AnalyzeOutput{
			Results: []CommitResult{},
			Summary: AnalyzeSummary{Total: 0},
		}, nil
	}

	// Process commits sequentially to avoid go-git race conditions
	// Note: go-git's ObjectStorage is not thread-safe for concurrent access
	results := make([]*commitResultInternal, len(commits))

	log.Printf("Starting sequential analysis of %d commits", len(commits))

	for i, c := range commits {
		msg := fmt.Sprintf("Analyzing commit %d/%d: %s", i+1, len(commits), c.Hash.String()[:8])
		log.Println(msg)
		if progress != nil {
			progress(msg)
		}

		// Check for cancellation
		select {
		case <-ctx.Done():
			log.Printf("Commit %s: cancelled", c.Hash.String()[:8])
			results[i] = &commitResultInternal{
				index:  i,
				commit: c,
				err:    ctx.Err(),
			}
			continue
		default:
		}

		// Perform analysis with retry
		var res *analyzer.AnalysisResult
		err := analyzer.WithRetry(ctx, analyzer.DefaultRetryConfig(), func() error {
			var analyzeErr error
			res, analyzeErr = analyzer.AnalyzeCommit(ctx, repo, c, headCommit, input.ErrorMessage, model)
			return analyzeErr
		})

		if err != nil {
			log.Printf("Commit %s: ERROR - %v", c.Hash.String()[:8], err)
		} else if res != nil && res.Skipped {
			log.Printf("Commit %s: SKIPPED (no relevant changes)", c.Hash.String()[:8])
		} else if res != nil {
			msg := fmt.Sprintf("Commit %s: %s probability", c.Hash.String()[:8], res.Probability)
			log.Println(msg)
			if progress != nil {
				progress(msg)
			}
		}

		results[i] = &commitResultInternal{
			index:  i,
			commit: c,
			result: res,
			err:    err,
		}
	}

	log.Printf("All commits analyzed")

	// Build output
	output := &AnalyzeOutput{
		Results: make([]CommitResult, 0, len(commits)),
		Summary: AnalyzeSummary{Total: len(commits)},
	}

	for _, r := range results {
		if r.err != nil {
			output.Summary.Errors++
			continue
		}
		if r.result == nil {
			output.Summary.Errors++
			continue
		}
		if r.result.Skipped {
			output.Summary.Skipped++
			continue
		}

		// Count by probability
		switch r.result.Probability {
		case analyzer.ProbHigh:
			output.Summary.High++
		case analyzer.ProbMedium:
			output.Summary.Medium++
		case analyzer.ProbLow:
			output.Summary.Low++
		}

		output.Results = append(output.Results, CommitResult{
			Hash:        r.commit.Hash.String()[:8],
			Message:     analyzer.TruncateCommitMessage(r.commit.Message, 80),
			Probability: string(r.result.Probability),
			Reasoning:   r.result.Reasoning,
		})
	}

	return output, nil
}

// FormatResultsAsText formats the analysis results as human-readable text
func FormatResultsAsText(output *AnalyzeOutput) string {
	var sb strings.Builder

	sb.WriteString("## Root Cause Analysis Results\n\n")

	if len(output.Results) == 0 {
		sb.WriteString("No commits with relevant code changes found.\n\n")
	} else {
		// Sort by probability (HIGH first)
		for _, prob := range []string{"HIGH", "MEDIUM", "LOW"} {
			for _, r := range output.Results {
				if r.Probability == prob {
					sb.WriteString(fmt.Sprintf("### [%s] Commit %s\n", r.Probability, r.Hash))
					sb.WriteString(fmt.Sprintf("**Message:** %s\n\n", r.Message))
					sb.WriteString(fmt.Sprintf("**Analysis:** %s\n\n", r.Reasoning))
					sb.WriteString("---\n\n")
				}
			}
		}
	}

	sb.WriteString("## Summary\n\n")
	sb.WriteString(fmt.Sprintf("- **Total commits analyzed:** %d\n", output.Summary.Total))
	sb.WriteString(fmt.Sprintf("- **High probability:** %d\n", output.Summary.High))
	sb.WriteString(fmt.Sprintf("- **Medium probability:** %d\n", output.Summary.Medium))
	sb.WriteString(fmt.Sprintf("- **Low probability:** %d\n", output.Summary.Low))
	sb.WriteString(fmt.Sprintf("- **Skipped (no code changes):** %d\n", output.Summary.Skipped))
	sb.WriteString(fmt.Sprintf("- **Errors:** %d\n", output.Summary.Errors))

	return sb.String()
}
