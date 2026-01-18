package analyzer

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func TestAnalysisResultParsing(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Probability
	}{
		{"High", `{"probability": "HIGH", "reasoning": "test"}`, ProbHigh},
		{"Medium", `{"probability": "MEDIUM", "reasoning": "test"}`, ProbMedium},
		{"Low", `{"probability": "LOW", "reasoning": "test"}`, ProbLow},
		{"Lowercase", `{"probability": "medium", "reasoning": "test"}`, ProbMedium},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var res AnalysisResult
			if err := json.Unmarshal([]byte(tt.input), &res); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}
			if res.Probability != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, res.Probability)
			}
		})
	}
}

func TestBuildPrompt(t *testing.T) {
	c := &object.Commit{
		Hash:    plumbing.NewHash("a1b2c3d4"),
		Message: "test message",
	}
	errorMsg := "panic in main"
	stdDiff := "std diff content"
	fullDiff := "full diff content"

	prompt := buildPrompt(errorMsg, c, stdDiff, fullDiff)

	expectedSections := []string{
		"BUG DESCRIPTION",
		"COMMIT CONTEXT",
		"INPUT DATA",
		"INSTRUCTIONS",
		"STEP 1: MICRO-ANALYSIS",
		"STEP 2: MACRO-ANALYSIS",
		"STEP 3: CLASSIFICATION",
		"HIGH|MEDIUM|LOW",
		"OUTPUT FORMAT",
	}

	for _, section := range expectedSections {
		if !strings.Contains(prompt, section) {
			t.Errorf("prompt missing section: %s", section)
		}
	}
}

func TestNoisyJSONParsing(t *testing.T) {
	input := `
Some reasoning steps here.
STEP 1: ...
STEP 2: ...
{
  "probability": "MEDIUM",
  "reasoning": "noisy response test"
}
`
	cleanTxt := jsonRegex.FindString(input)
	var res AnalysisResult
	if err := json.Unmarshal([]byte(cleanTxt), &res); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if res.Probability != ProbMedium {
		t.Errorf("expected MEDIUM, got %v", res.Probability)
	}
	if res.Reasoning != "noisy response test" {
		t.Errorf("expected 'noisy response test', got %v", res.Reasoning)
	}
}

func TestJSONResultSerialization(t *testing.T) {
	result := JSONResult{
		Hash:        "12345678",
		Probability: ProbHigh,
		Reasoning:   "Testing serialization",
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal JSONResult: %v", err)
	}

	expected := `{"hash":"12345678","probability":"HIGH","reasoning":"Testing serialization"}`
	if string(data) != expected {
		t.Errorf("expected %s, got %s", expected, string(data))
	}
}