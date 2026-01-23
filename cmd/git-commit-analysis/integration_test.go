package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/kerneldump/git-dual-context/pkg/analyzer"
)

// TestIntegration_BasicWorkflow tests the complete CLI workflow
// This test requires GEMINI_API_KEY to be set and will make real API calls
func TestIntegration_BasicWorkflow(t *testing.T) {
	if os.Getenv("GEMINI_API_KEY") == "" {
		t.Skip("Skipping integration test: GEMINI_API_KEY not set")
	}

	if os.Getenv("RUN_INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integration test: RUN_INTEGRATION_TESTS not set")
	}

	// Create a temporary test repository
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")

	repo := createTestRepo(t, repoPath)
	if repo == nil {
		t.Fatal("Failed to create test repository")
	}

	// Build the CLI binary
	binaryPath := filepath.Join(tmpDir, "git-commit-analysis")
	buildCmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build CLI: %v", err)
	}

	// Run the CLI
	var stdout bytes.Buffer
	cmd := exec.Command(
		binaryPath,
		"-repo", repoPath,
		"-error", "test error message",
		"-n", "3",
		"-j", "1",
	)
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "GEMINI_API_KEY="+os.Getenv("GEMINI_API_KEY"))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := runWithContext(ctx, cmd); err != nil {
		t.Fatalf("CLI execution failed: %v", err)
	}

	// Parse output
	output := stdout.String()
	lines := strings.Split(output, "\n")

	var results []analyzer.JSONResult
	var summary analyzer.Summary
	var logs []analyzer.LogEntry

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Try to parse as different types
		var generic struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal([]byte(line), &generic); err != nil {
			continue
		}

		switch generic.Type {
		case "result":
			var result analyzer.JSONResult
			if err := json.Unmarshal([]byte(line), &result); err == nil {
				results = append(results, result)
			}
		case "summary":
			if err := json.Unmarshal([]byte(line), &summary); err != nil {
				t.Errorf("Failed to parse summary: %v", err)
			}
		case "log":
			var log analyzer.LogEntry
			if err := json.Unmarshal([]byte(line), &log); err == nil {
				logs = append(logs, log)
			}
		}
	}

	// Verify output structure
	if len(results) == 0 {
		t.Error("Expected at least one result")
	}
	if summary.Total == 0 {
		t.Error("Expected summary with total > 0")
	}

	// Verify results have required fields
	for i, result := range results {
		if result.Type != "result" {
			t.Errorf("Result %d: expected type 'result', got %s", i, result.Type)
		}
		if result.Hash == "" {
			t.Errorf("Result %d: hash is empty", i)
		}
		if result.Probability == "" {
			t.Errorf("Result %d: probability is empty", i)
		}
		// Verify probability is valid
		validProbs := map[analyzer.Probability]bool{
			analyzer.ProbHigh: true, analyzer.ProbMedium: true, analyzer.ProbLow: true,
		}
		if !validProbs[result.Probability] {
			t.Errorf("Result %d: invalid probability %s", i, result.Probability)
		}
	}

	// Verify summary counts
	if summary.Total != len(results)+summary.Skipped+summary.Errors {
		t.Errorf("Summary total (%d) doesn't match results (%d) + skipped (%d) + errors (%d)",
			summary.Total, len(results), summary.Skipped, summary.Errors)
	}
}

// TestIntegration_InvalidInput tests CLI validation
func TestIntegration_InvalidInput(t *testing.T) {
	// This test doesn't need API key, just tests validation

	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "git-commit-analysis")

	// Build the binary
	buildCmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build CLI: %v", err)
	}

	tests := []struct {
		name string
		args []string
		want string // Expected error message substring
	}{
		{
			name: "no error message",
			args: []string{"-repo", ".", "-n", "5"},
			want: "error message cannot be empty",
		},
		{
			name: "invalid num commits",
			args: []string{"-error", "test", "-n", "0"},
			want: "number of commits must be positive",
		},
		{
			name: "too many commits",
			args: []string{"-error", "test", "-n", "9999"},
			want: "exceeds maximum",
		},
		{
			name: "invalid workers",
			args: []string{"-error", "test", "-j", "0"},
			want: "number of workers must be positive",
		},
		{
			name: "invalid branch name",
			args: []string{"-error", "test", "-branch", "../../etc/passwd"},
			want: "Invalid branch name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			cmd := exec.Command(binaryPath, tt.args...)
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			cmd.Env = append(os.Environ(), "GEMINI_API_KEY=test")

			// Should fail
			if err := cmd.Run(); err == nil {
				t.Error("Expected command to fail but it succeeded")
			}

			// Check both stdout and stderr (CLI logs errors to stdout as JSON)
			output := stdout.String() + stderr.String()
			if !strings.Contains(output, tt.want) {
				t.Errorf("Expected error containing %q, got stdout: %s, stderr: %s", tt.want, stdout.String(), stderr.String())
			}
		})
	}
}

// TestIntegration_OutputFile tests writing to a file
func TestIntegration_OutputFile(t *testing.T) {
	if os.Getenv("GEMINI_API_KEY") == "" {
		t.Skip("Skipping integration test: GEMINI_API_KEY not set")
	}

	if os.Getenv("RUN_INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integration test: RUN_INTEGRATION_TESTS not set")
	}

	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")
	outputPath := filepath.Join(tmpDir, "output.json")

	repo := createTestRepo(t, repoPath)
	if repo == nil {
		t.Fatal("Failed to create test repository")
	}

	// Build binary
	binaryPath := filepath.Join(tmpDir, "git-commit-analysis")
	buildCmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build CLI: %v", err)
	}

	// Run with output file
	cmd := exec.Command(
		binaryPath,
		"-repo", repoPath,
		"-error", "test error",
		"-n", "2",
		"-j", "1",
		"-o", outputPath,
	)
	cmd.Env = append(os.Environ(), "GEMINI_API_KEY="+os.Getenv("GEMINI_API_KEY"))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := runWithContext(ctx, cmd); err != nil {
		t.Fatalf("CLI execution failed: %v", err)
	}

	// Verify output file exists and contains valid JSON
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if len(content) == 0 {
		t.Fatal("Output file is empty")
	}

	// Parse each line as JSON
	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var js map[string]interface{}
		if err := json.Unmarshal([]byte(line), &js); err != nil {
			t.Errorf("Line %d is not valid JSON: %v", i, err)
		}
	}
}

// Helper functions

func createTestRepo(t *testing.T, path string) *git.Repository {
	t.Helper()

	// Initialize repository
	repo, err := git.PlainInit(path, false)
	if err != nil {
		t.Fatalf("Failed to init repo: %v", err)
	}

	w, err := repo.Worktree()
	if err != nil {
		t.Fatalf("Failed to get worktree: %v", err)
	}

	// Create some test commits
	commits := []struct {
		file    string
		content string
		message string
	}{
		{"README.md", "# Test Repo\n", "Initial commit"},
		{"main.go", "package main\n\nfunc main() {}\n", "Add main.go"},
		{"main.go", "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"Hello\")\n}\n", "Add hello world"},
	}

	for _, c := range commits {
		filePath := filepath.Join(path, c.file)
		if err := os.WriteFile(filePath, []byte(c.content), 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}

		if _, err := w.Add(c.file); err != nil {
			t.Fatalf("Failed to add file: %v", err)
		}

		if _, err := w.Commit(c.message, &git.CommitOptions{
			Author: &object.Signature{
				Name:  "Test User",
				Email: "test@example.com",
				When:  time.Now(),
			},
		}); err != nil {
			t.Fatalf("Failed to commit: %v", err)
		}
	}

	return repo
}

func runWithContext(ctx context.Context, cmd *exec.Cmd) error {
	if err := cmd.Start(); err != nil {
		return err
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-ctx.Done():
		cmd.Process.Kill()
		return ctx.Err()
	case err := <-done:
		return err
	}
}
