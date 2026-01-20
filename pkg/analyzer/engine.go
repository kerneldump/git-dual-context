package analyzer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"git-commit-analysis/internal/gitdiff"

	"github.com/go-git/go-git/v5"
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
	Errors  int    `json:"errors"`
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
	// Truncate message to first line
	firstLine := strings.Split(message, "\n")[0]
	if len(firstLine) > 80 {
		firstLine = firstLine[:77] + "..."
	}

	return JSONResult{
		Type:        "result",
		Hash:        hash,
		Message:     firstLine,
		Probability: ar.Probability,
		Reasoning:   ar.Reasoning,
	}
}

// AnalyzeCommit performs the dual-context analysis on a single commit
func AnalyzeCommit(ctx context.Context, r *git.Repository, c, headCommit *object.Commit, errorMsg string, model *genai.GenerativeModel) (*AnalysisResult, error) {
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
	prompt := buildPrompt(errorMsg, c, stdDiff, fullDiff)

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
			cleanTxt := findJSONBlock(string(txt))
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


func buildPrompt(errorMsg string, c *object.Commit, stdDiff, fullDiff string) string {
	return fmt.Sprintf(`
You are an expert software debugger and a rigorous technical skeptic. Your goal is to determine if a specific commit introduced the bug described below.

SKEPTIC PERSONA:
You must actively attempt to DISPROVE that this commit caused the bug. Assume the commit is safe unless you find a "smoking gun" (e.g., direct logic contradiction, enabling a path for an invalid value, removing a critical guard). If the evidence is circumstantial or the logic is merely "suspicious" but not demonstrably broken, you must lean towards a LOWER probability.

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

Follow this rigorous analytical process. You must output your reasoning for each step.

GLOBAL INSTRUCTION: Value Tracing
Identify any specific numeric values or state-related terms in the BUG DESCRIPTION (e.g., "-2"). You MUST explicitly trace how these values could originate from or be affected by the logic in the provided diffs.

STEP 0: HYPOTHESIS GENERATION
Based ONLY on the BUG DESCRIPTION, what are the most likely technical causes for this error? List 2-3 potential scenarios where this error could manifest.

STEP 1: MICRO-ANALYSIS (Skeptical Review)
Analyze the Standard Diff. What logic changed? Does it DIRECTLY produce the error? Look for unmasked paths where a previously ignored bad value can now reach a validation point.

STEP 2: MACRO-ANALYSIS (Evolutionary Context)
Analyze the Full Comparison Diff. Does the code from this commit still exist in HEAD? Was it refactored in a way that introduced the bug later? Does it conflict with the current system state?

STEP 3: CLASSIFICATION
Classify the probability based on these strict definitions:
- HIGH: You found a "smoking gun." The commit contains logic that DIRECTLY contradicts the error message or enables the specific bug.
- MEDIUM: The commit modifies relevant subsystems/variables and creates a plausible, though not certain, path for the bug. Warrants manual review.
- LOW: No direct or plausible link found. The change is unrelated or the skeptic's doubts remain unaddressed.

---
OUTPUT FORMAT:

Hypothesis: <Potential causes based on description>
Reasoning: <Step-by-Step Tracing and Analysis>
Classification: <HIGH|MEDIUM|LOW>

Finally, return the result in this JSON format (do not use markdown blocks):
{
  "probability": "HIGH|MEDIUM|LOW",
  "reasoning": "A concise summary of your tracing and verdict."
}
`, errorMsg, c.Hash.String(), c.Message, stdDiff, fullDiff)
}

// findJSONBlock attempts to find the largest valid JSON object in the text.
// It scans from the last '}' backwards to find a matching '{'.
func findJSONBlock(text string) string {
	end := strings.LastIndex(text, "}")
	if end == -1 {
		return ""
	}

	// Simple heuristic: find the last '}' and the first '{' before it
	// But simply finding the first '{' might match too early (e.g. nested braces in reasoning text).
	// So we can try to parse from every '{' found before 'end' until we succeed.
	
	// Optimization: Start searching for '{' from the end backwards.
	for start := strings.LastIndex(text[:end], "{"); start != -1; start = strings.LastIndex(text[:start], "{") {
		candidate := text[start : end+1]
		// Fast check: does it look like our schema?
		if !strings.Contains(candidate, "\"probability\"") {
			continue
		}
		var js map[string]interface{}
		if json.Unmarshal([]byte(candidate), &js) == nil {
			return candidate
		}
	}
	
	return ""
}

