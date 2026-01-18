package analyzer

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"git-commit-analysis/internal/gitdiff"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/google/generative-ai-go/genai"
)

var jsonRegex = regexp.MustCompile(`(?s)\{.*\}`)

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
	Skipped     bool        `json:"-"`
}

// JSONResult represents the final output format for the CLI
type JSONResult struct {
	Hash        string      `json:"hash"`
	Probability Probability `json:"probability"`
	Reasoning   string      `json:"reasoning"`
}

// AnalyzeCommit performs the dual-context analysis on a single commit
func AnalyzeCommit(ctx context.Context, r *git.Repository, c *object.Commit, headHash plumbing.Hash, errorMsg string, model *genai.GenerativeModel) (*AnalysisResult, error) {
	// 1. Standard Diff (C vs Parent)
	// For the very first commit, parent is empty. Handle gracefully.
	var parent *object.Commit
	if len(c.ParentHashes) > 0 {
		parent, _ = c.Parent(0)
	}

	stdDiff, modifiedFiles, err := gitdiff.GetStandardDiff(c, parent)
	if err != nil {
		return nil, fmt.Errorf("getting standard diff: %w", err)
	}

	if len(modifiedFiles) == 0 {
		return &AnalysisResult{Skipped: true}, nil
	}

	// 2. Full Comparison Diff (C vs HEAD), filtered by modifiedFiles
	headCommit, err := r.CommitObject(headHash)
	if err != nil {
		return nil, fmt.Errorf("getting HEAD commit: %w", err)
	}

	fullDiff, err := gitdiff.GetFullDiff(c, headCommit, modifiedFiles)
	if err != nil {
		return nil, fmt.Errorf("getting full diff: %w", err)
	}

	// 3. Construct Prompt
	prompt := buildPrompt(errorMsg, c, stdDiff, fullDiff)

	// 4. Call Gemini
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("gemini api call: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from gemini")
	}

	// Parse Response
	var result AnalysisResult
	for _, part := range resp.Candidates[0].Content.Parts {
		if txt, ok := part.(genai.Text); ok {
			cleanTxt := jsonRegex.FindString(string(txt))
			if cleanTxt == "" {
				// Fallback if no JSON found
				return nil, fmt.Errorf("no JSON found in response for %s", c.Hash.String()[:8])
			}
			if err := json.Unmarshal([]byte(cleanTxt), &result); err != nil {
				return nil, fmt.Errorf("parsing JSON for %s: %v. Raw: %s", c.Hash.String()[:8], err, string(txt))
			}
		}
	}

	return &result, nil
}


func buildPrompt(errorMsg string, c *object.Commit, stdDiff, fullDiff string) string {
	return fmt.Sprintf(`
You are an expert software debugger. Your task is to analyze a specific commit to determine if it introduced the bug described below.

BUG DESCRIPTION:
%s

COMMIT CONTEXT:
Hash: %s
Message: %s

---
INPUT DATA:

1. STANDARD DIFF (The immediate changes in this commit):
%s

2. FULL COMPARISON DIFF (Evolution from this commit to HEAD):
%s

---
INSTRUCTIONS:

Use the following "Chain of Thought" process to analyze the data. You must output your reasoning for each step.

STEP 1: MICRO-ANALYSIS
Analyze the Standard Diff. What logic changed? Does it look risky?

STEP 2: MACRO-ANALYSIS
Analyze the Full Comparison Diff. Does the code from this commit still exist in HEAD? Was it refactored? Does it conflict with the current system state?

STEP 3: CLASSIFICATION
Classify the probability based on these strict definitions:
- HIGH: The commit contains logic that DIRECTLY contradicts the error message or introduces the specific bug (a "smoking gun").
- MEDIUM: The commit modifies the relevant subsystem or variables, but the logic is not clearly broken. Warrants manual review.
- LOW: The commit is unrelated (docs, assets, different subsystem, safe refactor).

---
OUTPUT FORMAT:

Reasoning: <Your Step-by-Step Chain of Thought Analysis>
Classification: <HIGH|MEDIUM|LOW>

Finally, return the result in this JSON format (do not use markdown blocks):
{
  "probability": "HIGH|MEDIUM|LOW",
  "reasoning": "A concise summary of your analysis."
}
`, errorMsg, c.Hash.String(), c.Message, stdDiff, fullDiff)
}

