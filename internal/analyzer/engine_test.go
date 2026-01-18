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
	cleanTxt := findJSONBlock(input)
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
		Type:        "result",
		Hash:        "12345678",
		Probability: ProbHigh,
		Reasoning:   "Testing serialization",
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal JSONResult: %v", err)
	}

	expected := `{"type":"result","hash":"12345678","probability":"HIGH","reasoning":"Testing serialization"}`
	if string(data) != expected {
		t.Errorf("expected %s, got %s", expected, string(data))
	}
}

func TestToJSONResult(t *testing.T) {
	ar := &AnalysisResult{
		Probability: ProbLow,
		Reasoning:   "It is fine",
	}
	hash := "abc1234"

	jr := ar.ToJSONResult(hash)

	if jr.Hash != hash {
		t.Errorf("expected hash %s, got %s", hash, jr.Hash)
	}
	if jr.Probability != ProbLow {
		t.Errorf("expected probability %s, got %s", ProbLow, jr.Probability)
	}
	if jr.Reasoning != "It is fine" {
		t.Errorf("expected reasoning 'It is fine', got %s", jr.Reasoning)
	}
}

func TestLogEntrySerialization(t *testing.T) {
	entry := LogEntry{
		Type:      "log",
		Level:     "INFO",
		Msg:       "Started analysis",
		Timestamp: "2026-01-17T17:00:00Z",
	}

	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("failed to marshal LogEntry: %v", err)
	}

	expected := `{"type":"log","level":"INFO","msg":"Started analysis","timestamp":"2026-01-17T17:00:00Z"}`
	if string(data) != expected {
		t.Errorf("expected %s, got %s", expected, string(data))
	}
}

func TestStructuredLogger(t *testing.T) {
	msg := "Test message"
	level := "INFO"
	entry := NewLogEntry(level, msg)

	if entry.Type != "log" {
		t.Errorf("expected type 'log', got %s", entry.Type)
	}
	if entry.Level != level {
		t.Errorf("expected level %s, got %s", level, entry.Level)
	}
	if entry.Msg != msg {
		t.Errorf("expected msg %s, got %s", msg, entry.Msg)
	}
	if entry.Timestamp == "" {
		t.Error("expected non-empty timestamp")
	}
}

func TestMaliciousJSONParsing(t *testing.T) {
	input := `
Reasoning: The code contains a function { return true; } which is fine.
Also checking for edge cases like {}.

{
  "probability": "LOW",
  "reasoning": "The commit is safe."
}
`
	// We want the extraction logic to be smart enough to find the *actual* JSON object
	// For now, let's verify if the current regex finds it.
	cleanTxt := findJSONBlock(input)
	
	var res AnalysisResult
	if err := json.Unmarshal([]byte(cleanTxt), &res); err != nil {
		t.Fatalf("failed to unmarshal malicious input: %v. Extracted text: %s", err, cleanTxt)
	}
	if res.Probability != ProbLow {
		t.Errorf("expected LOW, got %v", res.Probability)
	}
}