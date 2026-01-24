package analyzer

import (
	"fmt"
	"testing"
)

func TestAnalysisOptionsDefaults(t *testing.T) {
	opts := AnalysisOptions{
		ErrorMessage: "test error",
	}

	// Verify zero values
	if opts.NumCommits != 0 {
		t.Error("NumCommits should be zero by default")
	}
	if opts.Branch != "" {
		t.Error("Branch should be empty by default")
	}
}

func TestCommitAnalysisResultFields(t *testing.T) {
	result := CommitAnalysisResult{
		Index:   0,
		Hash:    "abc12345",
		Message: "Test commit",
		Result: &AnalysisResult{
			Probability: ProbHigh,
			Reasoning:   "Test reasoning",
		},
		Error: nil,
	}

	if result.Index != 0 {
		t.Errorf("Expected index 0, got %d", result.Index)
	}
	if result.Hash != "abc12345" {
		t.Errorf("Expected hash 'abc12345', got %s", result.Hash)
	}
	if result.Result.Probability != ProbHigh {
		t.Errorf("Expected probability HIGH, got %s", result.Result.Probability)
	}
}

func TestCalculateSummary(t *testing.T) {
	results := []CommitAnalysisResult{
		{Result: &AnalysisResult{Probability: ProbHigh}},
		{Result: &AnalysisResult{Probability: ProbHigh}},
		{Result: &AnalysisResult{Probability: ProbMedium}},
		{Result: &AnalysisResult{Probability: ProbLow}},
		{Result: &AnalysisResult{Skipped: true}},
		{Error: fmt.Errorf("test error")},
	}

	summary := CalculateSummary(results)

	if summary.Total != 6 {
		t.Errorf("Expected total 6, got %d", summary.Total)
	}
	if summary.High != 2 {
		t.Errorf("Expected high 2, got %d", summary.High)
	}
	if summary.Medium != 1 {
		t.Errorf("Expected medium 1, got %d", summary.Medium)
	}
	if summary.Low != 1 {
		t.Errorf("Expected low 1, got %d", summary.Low)
	}
	if summary.Skipped != 1 {
		t.Errorf("Expected skipped 1, got %d", summary.Skipped)
	}
	if summary.Errors != 1 {
		t.Errorf("Expected errors 1, got %d", summary.Errors)
	}
}

func TestCalculateSummaryEmpty(t *testing.T) {
	results := []CommitAnalysisResult{}
	summary := CalculateSummary(results)

	if summary.Total != 0 {
		t.Errorf("Expected total 0, got %d", summary.Total)
	}
}

func TestCalculateSummaryNilResult(t *testing.T) {
	results := []CommitAnalysisResult{
		{Result: nil}, // nil result should count as error
	}

	summary := CalculateSummary(results)

	if summary.Errors != 1 {
		t.Errorf("Expected errors 1 for nil result, got %d", summary.Errors)
	}
}
