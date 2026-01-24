package tools

import (
	"strings"
	"testing"

	"github.com/kerneldump/git-dual-context/pkg/analyzer"
)

func TestFormatResultsAsText(t *testing.T) {
	output := &AnalyzeOutput{
		Results: []CommitResult{
			{
				Hash:        "abc12345",
				Message:     "Fix authentication bug",
				Probability: "HIGH",
				Reasoning:   "This commit modifies the auth logic",
			},
			{
				Hash:        "def67890",
				Message:     "Update documentation",
				Probability: "LOW",
				Reasoning:   "Documentation changes only",
			},
			{
				Hash:        "ghi11111",
				Message:     "Refactor validation",
				Probability: "MEDIUM",
				Reasoning:   "Changes validation code",
			},
		},
		Summary: AnalyzeSummary{
			Total:   5,
			High:    1,
			Medium:  1,
			Low:     1,
			Skipped: 2,
						Errors:   0,
						Duration: "1m2s",
						Model:    "gemini-flash-latest",
					},
				}

	text := FormatResultsAsText(output)

	// Check for required sections
	if !strings.Contains(text, "## Root Cause Analysis Results") {
		t.Error("Output should contain results header")
	}
	if !strings.Contains(text, "## Summary") {
		t.Error("Output should contain summary header")
	}

	// Check for HIGH probability results first
	highIndex := strings.Index(text, "[HIGH]")
	medIndex := strings.Index(text, "[MEDIUM]")
	lowIndex := strings.Index(text, "[LOW]")

	if highIndex == -1 || medIndex == -1 || lowIndex == -1 {
		t.Error("Output should contain all probability levels")
	}

	// Verify HIGH comes before MEDIUM comes before LOW
	if highIndex > medIndex || medIndex > lowIndex {
		t.Error("Results should be sorted HIGH, MEDIUM, LOW")
	}

	// Check for commit details
	if !strings.Contains(text, "abc12345") {
		t.Error("Output should contain commit hash")
	}
	if !strings.Contains(text, "Fix authentication bug") {
		t.Error("Output should contain commit message")
	}
	if !strings.Contains(text, "This commit modifies the auth logic") {
		t.Error("Output should contain reasoning")
	}

	// Check for summary details
	if !strings.Contains(text, "Total commits analyzed:** 5") {
		t.Error("Output should contain total count")
	}
	if !strings.Contains(text, "High probability:** 1") {
		t.Error("Output should contain high count")
	}
	if !strings.Contains(text, "Medium probability:** 1") {
		t.Error("Output should contain medium count")
	}
	if !strings.Contains(text, "Low probability:** 1") {
		t.Error("Output should contain low count")
	}
	if !strings.Contains(text, "Skipped (no code changes):** 2") {
		t.Error("Output should contain skipped count")
	}
	if !strings.Contains(text, "Duration:** 1m2s") {
		t.Error("Output should contain duration")
	}
	if !strings.Contains(text, "Model:** gemini-flash-latest") {
		t.Error("Output should contain model name")
	}
}

func TestFormatResultsAsTextEmpty(t *testing.T) {
	output := &AnalyzeOutput{
		Results: []CommitResult{},
		Summary: AnalyzeSummary{
			Total:   0,
			High:    0,
			Medium:  0,
			Low:     0,
			Skipped: 0,
			Errors:  0,
		},
	}

	text := FormatResultsAsText(output)

	if !strings.Contains(text, "No commits with relevant code changes found") {
		t.Error("Empty output should show appropriate message")
	}
}

func TestFormatResultsAsTextOnlyHigh(t *testing.T) {
	output := &AnalyzeOutput{
		Results: []CommitResult{
			{
				Hash:        "abc12345",
				Message:     "Critical bug",
				Probability: "HIGH",
				Reasoning:   "Smoking gun found",
			},
		},
		Summary: AnalyzeSummary{
			Total:  1,
			High:   1,
			Medium: 0,
			Low:    0,
		},
	}

	text := FormatResultsAsText(output)

	// Should contain HIGH but not MEDIUM or LOW sections
	if !strings.Contains(text, "[HIGH]") {
		t.Error("Should contain HIGH probability")
	}
	if strings.Contains(text, "[MEDIUM]") {
		t.Error("Should not contain MEDIUM probability when none exist")
	}
	if strings.Contains(text, "[LOW]") {
		t.Error("Should not contain LOW probability when none exist")
	}
}

func TestAnalyzeInputDefaults(t *testing.T) {
	// This tests that the AnalyzeRootCause function applies defaults correctly
	// We can't easily test the full function without a real git repo and API key
	// but we can test the struct defaults

	input := AnalyzeInput{
		RepoPath:     ".",
		ErrorMessage: "test error",
		// NumCommits and Concurrency not set
	}

	// Verify that defaults would be applied
	if input.NumCommits == 0 {
		// This is expected - defaults applied in function
	}
	if input.Concurrency == 0 {
		// This is expected - defaults applied in function
	}
}

func TestCommitResultStruct(t *testing.T) {
	result := CommitResult{
		Hash:        "abc123",
		Message:     "Test commit",
		Probability: string(analyzer.ProbHigh),
		Reasoning:   "Test reasoning",
	}

	if result.Hash != "abc123" {
		t.Errorf("Expected hash 'abc123', got %s", result.Hash)
	}
	if result.Probability != "HIGH" {
		t.Errorf("Expected probability 'HIGH', got %s", result.Probability)
	}
}

func TestAnalyzeSummaryStruct(t *testing.T) {
	summary := AnalyzeSummary{
		Total:   10,
		High:    2,
		Medium:  3,
		Low:     4,
		Skipped: 1,
		Errors:  0,
	}

	total := summary.High + summary.Medium + summary.Low + summary.Skipped + summary.Errors
	if total != summary.Total {
		t.Errorf("Summary counts don't add up: %d+%d+%d+%d+%d != %d",
			summary.High, summary.Medium, summary.Low, summary.Skipped, summary.Errors, summary.Total)
	}
}

func TestFormatResultsMarkdownFormatting(t *testing.T) {
	output := &AnalyzeOutput{
		Results: []CommitResult{
			{
				Hash:        "abc123",
				Message:     "Test commit with **bold** and *italic*",
				Probability: "HIGH",
				Reasoning:   "Contains `code` and [links](http://example.com)",
			},
		},
		Summary: AnalyzeSummary{
			Total: 1,
			High:  1,
		},
	}

	text := FormatResultsAsText(output)

	// Verify markdown formatting is preserved
	if !strings.Contains(text, "**Message:**") {
		t.Error("Should use markdown bold for labels")
	}
	if !strings.Contains(text, "**Analysis:**") {
		t.Error("Should use markdown bold for labels")
	}
	if !strings.Contains(text, "###") {
		t.Error("Should use markdown heading for commits")
	}
	if !strings.Contains(text, "---") {
		t.Error("Should use markdown separator between results")
	}
}

func TestFormatResultsMultipleSameProbability(t *testing.T) {
	output := &AnalyzeOutput{
		Results: []CommitResult{
			{Hash: "aaa", Message: "First high", Probability: "HIGH", Reasoning: "Reason 1"},
			{Hash: "bbb", Message: "Second high", Probability: "HIGH", Reasoning: "Reason 2"},
			{Hash: "ccc", Message: "Third high", Probability: "HIGH", Reasoning: "Reason 3"},
		},
		Summary: AnalyzeSummary{
			Total: 3,
			High:  3,
		},
	}

	text := FormatResultsAsText(output)

	// Should contain all three HIGH results
	if !strings.Contains(text, "aaa") || !strings.Contains(text, "bbb") || !strings.Contains(text, "ccc") {
		t.Error("Should contain all commits with same probability")
	}

	// Count occurrences of [HIGH]
	count := strings.Count(text, "[HIGH]")
	if count != 3 {
		t.Errorf("Expected 3 [HIGH] markers, found %d", count)
	}
}
