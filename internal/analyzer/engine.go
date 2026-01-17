package analyzer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"git-commit-analysis/internal/gitdiff"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/google/generative-ai-go/genai"
)

// Probability represents the likelihood of a commit causing a bug
type Probability string

const (
	ProbHigh   Probability = "HIGH"
	ProbMedium Probability = "MEDIUM"
	ProbLow    Probability = "LOW"
)

func (p *Probability) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	switch strings.ToUpper(s) {
	case "HIGH":
		*p = ProbHigh
	case "MEDIUM", "MED":
		*p = ProbMedium
	default:
		*p = ProbLow
	}
	return nil
}

// AnalysisResult represents the JSON output from the LLM
type AnalysisResult struct {
	Probability Probability `json:"probability"`
	Reasoning   string      `json:"reasoning"`
}

// AnalyzeCommit performs the dual-context analysis on a single commit
func AnalyzeCommit(ctx context.Context, r *git.Repository, c *object.Commit, headHash plumbing.Hash, errorMsg string, model *genai.GenerativeModel) (string, error) {
	// 1. Standard Diff (C vs Parent)
	// For the very first commit, parent is empty. Handle gracefully.
	var parent *object.Commit
	if len(c.ParentHashes) > 0 {
		parent, _ = c.Parent(0)
	}

	stdDiff, modifiedFiles, err := gitdiff.GetStandardDiff(c, parent)
	if err != nil {
		return "", fmt.Errorf("getting standard diff: %w", err)
	}

	if len(modifiedFiles) == 0 {
		return fmt.Sprintf("Commit %s: [Skipped - No relevant code changes]\n", c.Hash.String()[:8]), nil
	}

	// 2. Full Comparison Diff (C vs HEAD), filtered by modifiedFiles
	headCommit, err := r.CommitObject(headHash)
	if err != nil {
		return "", fmt.Errorf("getting HEAD commit: %w", err)
	}

	fullDiff, err := gitdiff.GetFullDiff(c, headCommit, modifiedFiles)
	if err != nil {
		return "", fmt.Errorf("getting full diff: %w", err)
	}

	// 3. Construct Prompt
	prompt := fmt.Sprintf(`

You are an expert software debugger performing a "Dual-Context Diff Analysis".
Your goal is to determine the probability that the following commit introduced the bug described below.

BUG DESCRIPTION:
%s

COMMIT INFO:
Hash: %s
Message: %s

ANALYSIS DATA:
1. STANDARD DIFF (Changes made in this commit):
%s

2. FULL COMPARISON DIFF (Evolution of these specific files from this commit to HEAD):
%s

INSTRUCTIONS:
- Analyze the Standard Diff to understand what changed.
- Analyze the Full Comparison Diff to see if those changes were reverted, modified, or if they interact badly with current code.
- Ignore test files unless the bug is in a test.
- Provide a probability (High, Low, Unknown) and a concise reasoning.

Return JSON: { "probability": "High|Low|Unknown", "reasoning": "string" }
`, errorMsg, c.Hash.String(), c.Message, stdDiff, fullDiff)

	// 4. Call Gemini
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("gemini api call: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty response from gemini")
	}

	// Parse Response
	var result AnalysisResult
	for _, part := range resp.Candidates[0].Content.Parts {
		if txt, ok := part.(genai.Text); ok {
			// Basic cleanup if markdown blocks are returned
			cleanTxt := strings.TrimPrefix(strings.TrimSuffix(string(txt), "```"), "```json")
			if err := json.Unmarshal([]byte(cleanTxt), &result); err != nil {
				// Fallback if not pure JSON
				return "", fmt.Errorf("parsing JSON for %s: %v. Raw: %s", c.Hash.String()[:8], err, string(txt))
			}
		}
	}

	// Format Output
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Commit: %s | Prob: %s\n", c.Hash.String()[:8], result.Probability))
	sb.WriteString(fmt.Sprintf("Reason: %s\n", result.Reasoning))
	sb.WriteString("---------------------------------------------------\n")

	return sb.String(), nil
}
