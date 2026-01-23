package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Verify LLM defaults
	if cfg.LLM.Provider != "gemini" {
		t.Errorf("Expected default provider 'gemini', got %s", cfg.LLM.Provider)
	}
	if cfg.LLM.Temperature != 0.1 {
		t.Errorf("Expected default temperature 0.1, got %f", cfg.LLM.Temperature)
	}

	// Verify Analysis defaults
	if cfg.Analysis.DefaultCommits != 5 {
		t.Errorf("Expected default commits 5, got %d", cfg.Analysis.DefaultCommits)
	}
	if !cfg.Analysis.SkipMergeCommits {
		t.Error("Expected SkipMergeCommits to be true by default")
	}

	// Verify Performance defaults
	if cfg.Performance.Workers != 3 {
		t.Errorf("Expected default workers 3, got %d", cfg.Performance.Workers)
	}

	// Verify Output defaults
	if cfg.Output.Format != "json" {
		t.Errorf("Expected default format 'json', got %s", cfg.Output.Format)
	}
}

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")

	yamlContent := `
llm:
  provider: openai
  model: gpt-4
  temperature: 0.2
  timeout: 10m

analysis:
  default_commits: 10
  max_diff_size: 100000
  skip_merge_commits: false

performance:
  workers: 5
  max_retries: 5
  retry_base_delay: 2s
  retry_max_delay: 60s

output:
  format: markdown
  verbose: true
  commit_message_max_length: 100
`

	if err := os.WriteFile(cfgPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Load config
	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify values
	if cfg.LLM.Provider != "openai" {
		t.Errorf("Expected provider 'openai', got %s", cfg.LLM.Provider)
	}
	if cfg.LLM.Model != "gpt-4" {
		t.Errorf("Expected model 'gpt-4', got %s", cfg.LLM.Model)
	}
	if cfg.Analysis.DefaultCommits != 10 {
		t.Errorf("Expected commits 10, got %d", cfg.Analysis.DefaultCommits)
	}
	if cfg.Performance.Workers != 5 {
		t.Errorf("Expected workers 5, got %d", cfg.Performance.Workers)
	}
	if cfg.Output.Format != "markdown" {
		t.Errorf("Expected format 'markdown', got %s", cfg.Output.Format)
	}
}

func TestLoadConfigNonexistent(t *testing.T) {
	// Loading nonexistent file should return defaults, no error
	cfg, err := LoadConfig("/nonexistent/config.yaml")
	if err != nil {
		t.Errorf("LoadConfig should return defaults for nonexistent file, got error: %v", err)
	}
	if cfg.LLM.Provider != "gemini" {
		t.Error("Should return default config for nonexistent file")
	}
}

func TestLoadConfigEmpty(t *testing.T) {
	// Empty path should return defaults
	cfg, err := LoadConfig("")
	if err != nil {
		t.Errorf("LoadConfig with empty path should not error: %v", err)
	}
	if cfg.LLM.Provider != "gemini" {
		t.Error("Should return default config for empty path")
	}
}

func TestSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "test-config.yaml")

	cfg := DefaultConfig()
	cfg.LLM.Model = "custom-model"
	cfg.Analysis.DefaultCommits = 20

	// Save config
	if err := SaveConfig(cfg, cfgPath); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Load it back
	loaded, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig after save failed: %v", err)
	}

	// Verify
	if loaded.LLM.Model != "custom-model" {
		t.Errorf("Expected model 'custom-model', got %s", loaded.LLM.Model)
	}
	if loaded.Analysis.DefaultCommits != 20 {
		t.Errorf("Expected commits 20, got %d", loaded.Analysis.DefaultCommits)
	}
}

func TestMergeWithFlags(t *testing.T) {
	cfg := DefaultConfig()

	// Set up flag values
	model := "gpt-4"
	commits := 15
	workers := 10
	timeout := 10 * time.Minute
	verbose := true

	// Merge
	cfg.MergeWithFlags(&model, &commits, &workers, &timeout, &verbose)

	// Verify flags override config
	if cfg.LLM.Model != "gpt-4" {
		t.Errorf("Expected model 'gpt-4', got %s", cfg.LLM.Model)
	}
	if cfg.Analysis.DefaultCommits != 15 {
		t.Errorf("Expected commits 15, got %d", cfg.Analysis.DefaultCommits)
	}
	if cfg.Performance.Workers != 10 {
		t.Errorf("Expected workers 10, got %d", cfg.Performance.Workers)
	}
	if cfg.LLM.Timeout != 10*time.Minute {
		t.Errorf("Expected timeout 10m, got %v", cfg.LLM.Timeout)
	}
	if !cfg.Output.Verbose {
		t.Error("Expected verbose to be true")
	}
}

func TestMergeWithFlagsNil(t *testing.T) {
	cfg := DefaultConfig()
	original := cfg.LLM.Model

	// Merge with nil flags shouldn't change anything
	cfg.MergeWithFlags(nil, nil, nil, nil, nil)

	if cfg.LLM.Model != original {
		t.Error("Merging with nil flags should not change config")
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*Config)
		wantErr bool
	}{
		{
			name:    "valid default config",
			setup:   func(c *Config) {},
			wantErr: false,
		},
		{
			name: "empty provider",
			setup: func(c *Config) {
				c.LLM.Provider = ""
			},
			wantErr: true,
		},
		{
			name: "empty model",
			setup: func(c *Config) {
				c.LLM.Model = ""
			},
			wantErr: true,
		},
		{
			name: "invalid temperature low",
			setup: func(c *Config) {
				c.LLM.Temperature = -0.1
			},
			wantErr: true,
		},
		{
			name: "invalid temperature high",
			setup: func(c *Config) {
				c.LLM.Temperature = 1.5
			},
			wantErr: true,
		},
		{
			name: "zero timeout",
			setup: func(c *Config) {
				c.LLM.Timeout = 0
			},
			wantErr: true,
		},
		{
			name: "negative commits",
			setup: func(c *Config) {
				c.Analysis.DefaultCommits = -1
			},
			wantErr: true,
		},
		{
			name: "zero workers",
			setup: func(c *Config) {
				c.Performance.Workers = 0
			},
			wantErr: true,
		},
		{
			name: "invalid output format",
			setup: func(c *Config) {
				c.Output.Format = "invalid"
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			tt.setup(cfg)

			err := cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFindConfigFile(t *testing.T) {
	// Create a temporary config file in current directory
	tmpFile := ".git-dual-context.yaml"
	if err := os.WriteFile(tmpFile, []byte("llm:\n  provider: test\n"), 0644); err != nil {
		t.Fatalf("Failed to create temp config: %v", err)
	}
	defer os.Remove(tmpFile)

	// Should find the file
	found := FindConfigFile()
	if found == "" {
		t.Error("FindConfigFile should find .git-dual-context.yaml in current directory")
	}
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "invalid.yaml")

	// Write invalid YAML
	invalidYAML := `
llm:
  provider: test
  invalid yaml here: [
`
	if err := os.WriteFile(cfgPath, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("Failed to write invalid config: %v", err)
	}

	// Should return error
	_, err := LoadConfig(cfgPath)
	if err == nil {
		t.Error("LoadConfig should return error for invalid YAML")
	}
}
