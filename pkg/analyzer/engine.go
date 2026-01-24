package analyzer

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/kerneldump/git-dual-context/pkg/gitdiff"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/google/generative-ai-go/genai"
)

//go:embed prompts/analysis.txt
var analysisPromptTemplate string

// LLMModel is an interface for LLM interaction, allowing for mocking in tests
// and abstracting different provider-specific implementations.
type LLMModel interface {
	GenerateContent(ctx context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error)
}

// Probability represents the likelihood of a commit causing a bug
type Probability string

const (
	// ProbHigh indicates a high probability that the commit caused the bug.
	ProbHigh Probability = "HIGH"
	// ProbMedium indicates a medium probability that warrants manual review.
	ProbMedium Probability = "MEDIUM"
	// ProbLow indicates a low probability with no clear link to the bug.
	ProbLow Probability = "LOW"
)

// UnmarshalJSON customizes the unmarshaling of Probability from JSON.
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
	Type        string      `json:"type"`
	Hash        string      `json:"hash"`
	Message     string      `json:"message,omitempty"`
	Probability Probability `json:"probability"`
	Reasoning   string      `json:"reasoning"`
}

// Summary represents the final analysis summary
type Summary struct {
	Type    string `json:"type"`
	Total   int    `json:"total"`
	High    int    `json:"high"`
	Medium  int    `json:"medium"`
	Low     int    `json:"low"`
	Skipped int    `json:"skipped"`
		Errors   int    `json:"errors"`
		Duration string `json:"duration"`
		Model    string `json:"model"`
	}

// LogEntry represents a structured log message
type LogEntry struct {
	Type      string `json:"type"`
	Level     string `json:"level"`
	Msg       string `json:"msg"`
	Timestamp string `json:"timestamp"`
}

// NewLogEntry creates a new LogEntry with the current timestamp
func NewLogEntry(level, msg string) LogEntry {
	return LogEntry{
		Type:      "log",
		Level:     level,
		Msg:       msg,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

// ToJSONResult converts an internal AnalysisResult to the CLI-friendly JSONResult
func (ar *AnalysisResult) ToJSONResult(hash string, message string) JSONResult {
	return JSONResult{
		Type:        "result",
		Hash:        hash,
		Message:     TruncateCommitMessage(message, DefaultCommitMessageMaxLength),
		Probability: ar.Probability,
		Reasoning:   ar.Reasoning,
	}
}

// AnalyzeCommit performs the dual-context analysis on a single commit.
// The model parameter accepts any LLMModel implementation (including *genai.GenerativeModel).
func AnalyzeCommit(ctx context.Context, r *git.Repository, c, headCommit *object.Commit, errorMsg string, model LLMModel) (*AnalysisResult, error) {
	// 1. Standard Diff (C vs Parent)
	// For the very first commit, parent is empty. Handle gracefully.
	var parent *object.Commit
	if len(c.ParentHashes) > 0 {
		var err error
		parent, err = c.Parent(0)
		if err != nil {
			return nil, fmt.Errorf("getting parent commit for %s: %w", c.Hash.String()[:8], err)
		}
	}

	stdDiff, modifiedFiles, err := gitdiff.GetStandardDiff(c, parent)
	if err != nil {
		return nil, fmt.Errorf("getting standard diff: %w", err)
	}

	if len(modifiedFiles) == 0 {
		return &AnalysisResult{Skipped: true}, nil
	}

	// 2. Full Comparison Diff (C vs HEAD), filtered by modifiedFiles
	fullDiff, err := gitdiff.GetFullDiff(c, headCommit, modifiedFiles)
	if err != nil {
		return nil, fmt.Errorf("getting full diff: %w", err)
	}

	// 3. Construct Prompt
	prompt := BuildPrompt(errorMsg, c, stdDiff, fullDiff)

	// 4. Call Gemini
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("gemini api call: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from gemini for commit %s", c.Hash.String()[:8])
	}

	// Parse Response
	var result AnalysisResult
	found := false

	for _, part := range resp.Candidates[0].Content.Parts {
		if txt, ok := part.(genai.Text); ok {
			found = true
			cleanTxt := FindJSONBlock(string(txt))
			if cleanTxt == "" {
				return nil, fmt.Errorf("no JSON found in response for %s", c.Hash.String()[:8])
			}
			if err := json.Unmarshal([]byte(cleanTxt), &result); err != nil {
				return nil, fmt.Errorf("parsing JSON for %s: %v. Raw: %s", c.Hash.String()[:8], err, string(txt))
			}
			break // Found and parsed, exit loop
		}
	}

	if !found {
		return nil, fmt.Errorf("no text content in gemini response for %s", c.Hash.String()[:8])
	}

	return &result, nil
}

// BuildPrompt constructs the multi-step analytical prompt for the LLM.
// It incorporates the bug description, commit diffs, and the skeptical persona instructions.
// The prompt template is loaded from prompts/analysis.txt via go:embed.
func BuildPrompt(errorMsg string, c *object.Commit, stdDiff, fullDiff string) string {
	return fmt.Sprintf(analysisPromptTemplate, errorMsg, c.Hash.String(), c.Message, stdDiff, fullDiff)
}

// CommitDiffContext holds pre-extracted diff data for a commit.
// This allows separating git operations (not thread-safe) from LLM calls (thread-safe).
type CommitDiffContext struct {
	Commit        *object.Commit
	StandardDiff  string
	FullDiff      string
	ModifiedFiles []string
	Skipped       bool // true if no relevant files were modified
}

// ExtractDiffs extracts the dual-context diffs from a commit.
// This function performs git operations and is NOT thread-safe with go-git.
// Call this sequentially, then use AnalyzeWithDiffs for parallel LLM calls.
func ExtractDiffs(r *git.Repository, c, headCommit *object.Commit) (*CommitDiffContext, error) {
	ctx := &CommitDiffContext{
		Commit: c,
	}

	// 1. Standard Diff (C vs Parent)
	var parent *object.Commit
	if len(c.ParentHashes) > 0 {
		var err error
		parent, err = c.Parent(0)
		if err != nil {
			return nil, fmt.Errorf("getting parent commit for %s: %w", c.Hash.String()[:8], err)
		}
	}

	stdDiff, modifiedFiles, err := gitdiff.GetStandardDiff(c, parent)
	if err != nil {
		return nil, fmt.Errorf("getting standard diff: %w", err)
	}

	if len(modifiedFiles) == 0 {
		ctx.Skipped = true
		return ctx, nil
	}

	ctx.StandardDiff = stdDiff
	ctx.ModifiedFiles = modifiedFiles

	// 2. Full Comparison Diff (C vs HEAD)
	fullDiff, err := gitdiff.GetFullDiff(c, headCommit, modifiedFiles)
	if err != nil {
		return nil, fmt.Errorf("getting full diff: %w", err)
	}
	ctx.FullDiff = fullDiff

	return ctx, nil
}

// AnalyzeWithDiffs performs LLM analysis using pre-extracted diffs.
// This function is thread-safe and can be called concurrently.
// The model parameter accepts any LLMModel implementation (including *genai.GenerativeModel).
func AnalyzeWithDiffs(ctx context.Context, diffCtx *CommitDiffContext, errorMsg string, model LLMModel) (*AnalysisResult, error) {
	if diffCtx.Skipped {
		return &AnalysisResult{Skipped: true}, nil
	}

	// Build prompt with pre-extracted diffs
	prompt := BuildPrompt(errorMsg, diffCtx.Commit, diffCtx.StandardDiff, diffCtx.FullDiff)

	// Call Gemini (thread-safe)
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("gemini api call: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from gemini for commit %s", diffCtx.Commit.Hash.String()[:8])
	}

	// Parse Response
	var result AnalysisResult
	found := false

	for _, part := range resp.Candidates[0].Content.Parts {
		if txt, ok := part.(genai.Text); ok {
			found = true
			cleanTxt := FindJSONBlock(string(txt))
			if cleanTxt == "" {
				return nil, fmt.Errorf("no JSON found in response for %s", diffCtx.Commit.Hash.String()[:8])
			}
			if err := json.Unmarshal([]byte(cleanTxt), &result); err != nil {
				return nil, fmt.Errorf("parsing JSON for %s: %v. Raw: %s", diffCtx.Commit.Hash.String()[:8], err, string(txt))
			}
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("no text content in gemini response for %s", diffCtx.Commit.Hash.String()[:8])
	}

	return &result, nil
}

// jsonFallbackRegex is used as a fallback for extracting JSON when brace matching fails.
// Compiled once at package initialization for efficiency.
var jsonFallbackRegex = regexp.MustCompile(`(?s)\{[^{}]*"probability"\s*:\s*"[^"]*"[^{}]*\}`)

// FindJSONBlock attempts to find the largest valid JSON object in the text.
// It uses a two-strategy approach:
// 1. First tries scanning from the last '}' backwards to find matching '{'
// 2. Falls back to regex matching if the first strategy fails
func FindJSONBlock(text string) string {
	// Strategy 1: Brace matching from end backwards
	end := strings.LastIndex(text, "}")
	if end != -1 {
		for start := strings.LastIndex(text[:end], "{"); start != -1; start = strings.LastIndex(text[:start], "{") {
			candidate := text[start : end+1]
			// Fast check: does it look like our schema?
			if strings.Contains(candidate, "\"probability\"") {
				var js map[string]interface{}
				if json.Unmarshal([]byte(candidate), &js) == nil {
					return candidate
				}
			}
		}
	}

	// Strategy 2: Regex fallback for edge cases
	// Handles cases where there might be trailing text after the last '}'
	// or multiple JSON blocks where the last one is malformed.
	matches := jsonFallbackRegex.FindAllString(text, -1)
	// Try matches from last to first
	for i := len(matches) - 1; i >= 0; i-- {
		var js map[string]interface{}
		if json.Unmarshal([]byte(matches[i]), &js) == nil {
			return matches[i]
		}
	}

	return ""
}
