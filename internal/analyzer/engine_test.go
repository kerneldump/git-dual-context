package analyzer

import (
	"encoding/json"
	"testing"
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
