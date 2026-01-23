// Package config provides configuration file support for git-dual-context
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the complete configuration for git-dual-context
type Config struct {
	// LLM settings
	LLM LLMConfig `yaml:"llm"`

	// Analysis settings
	Analysis AnalysisConfig `yaml:"analysis"`

	// Performance settings
	Performance PerformanceConfig `yaml:"performance"`

	// Output settings
	Output OutputConfig `yaml:"output"`
}

// LLMConfig contains LLM-specific settings
type LLMConfig struct {
	// Provider is the LLM provider (gemini, openai, anthropic)
	Provider string `yaml:"provider"`

	// Model is the specific model to use
	Model string `yaml:"model"`

	// APIKey is the API key (can be overridden by env var)
	APIKey string `yaml:"api_key,omitempty"`

	// Temperature controls randomness (0.0 to 1.0)
	Temperature float32 `yaml:"temperature"`

	// Timeout for each LLM request
	Timeout time.Duration `yaml:"timeout"`
}

// AnalysisConfig contains analysis-specific settings
type AnalysisConfig struct {
	// DefaultCommits is the default number of commits to analyze
	DefaultCommits int `yaml:"default_commits"`

	// MaxDiffSize is the maximum diff size in characters
	MaxDiffSize int `yaml:"max_diff_size"`

	// SkipMergeCommits whether to skip merge commits
	SkipMergeCommits bool `yaml:"skip_merge_commits"`

	// FileFilters contains glob patterns for files to exclude
	FileFilters []string `yaml:"file_filters,omitempty"`
}

// PerformanceConfig contains performance-related settings
type PerformanceConfig struct {
	// Workers is the default number of concurrent workers
	Workers int `yaml:"workers"`

	// MaxRetries for failed API calls
	MaxRetries int `yaml:"max_retries"`

	// RetryBaseDelay is the base delay for exponential backoff
	RetryBaseDelay time.Duration `yaml:"retry_base_delay"`

	// RetryMaxDelay is the maximum retry delay
	RetryMaxDelay time.Duration `yaml:"retry_max_delay"`
}

// OutputConfig contains output formatting settings
type OutputConfig struct {
	// Format is the output format (json, text, markdown)
	Format string `yaml:"format"`

	// Verbose enables verbose logging
	Verbose bool `yaml:"verbose"`

	// CommitMessageMaxLength for truncation
	CommitMessageMaxLength int `yaml:"commit_message_max_length"`
}

// DefaultConfig returns sensible default configuration
func DefaultConfig() *Config {
	return &Config{
		LLM: LLMConfig{
			Provider:    "gemini",
			Model:       "gemini-3-pro-preview",
			Temperature: 0.1,
			Timeout:     5 * time.Minute,
		},
		Analysis: AnalysisConfig{
			DefaultCommits:   5,
			MaxDiffSize:      50000,
			SkipMergeCommits: true,
			FileFilters:      []string{},
		},
		Performance: PerformanceConfig{
			Workers:        3,
			MaxRetries:     3,
			RetryBaseDelay: 1 * time.Second,
			RetryMaxDelay:  30 * time.Second,
		},
		Output: OutputConfig{
			Format:                 "json",
			Verbose:                false,
			CommitMessageMaxLength: 80,
		},
	}
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
	// Start with defaults
	cfg := DefaultConfig()

	// If path is empty, return defaults
	if path == "" {
		return cfg, nil
	}

	// Expand home directory
	if path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		path = filepath.Join(home, path[2:])
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil // File doesn't exist, use defaults
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return cfg, nil
}

// SaveConfig saves configuration to a YAML file
func SaveConfig(cfg *Config, path string) error {
	// Expand home directory
	if path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		path = filepath.Join(home, path[2:])
	}

	// Create directory if needed
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// FindConfigFile searches for a config file in standard locations
func FindConfigFile() string {
	locations := []string{
		".git-dual-context.yaml",
		".git-dual-context.yml",
		"~/.config/git-dual-context/config.yaml",
		"~/.config/git-dual-context/config.yml",
		"~/.git-dual-context.yaml",
		"~/.git-dual-context.yml",
	}

	for _, loc := range locations {
		// Expand home directory
		path := loc
		if len(loc) > 2 && loc[:2] == "~/" {
			home, err := os.UserHomeDir()
			if err != nil {
				continue
			}
			path = filepath.Join(home, loc[2:])
		}

		// Check if file exists
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// MergeWithFlags merges configuration with command-line flags
// Flags take precedence over config file values
func (c *Config) MergeWithFlags(
	model *string,
	numCommits *int,
	numWorkers *int,
	timeout *time.Duration,
	verbose *bool,
) {
	if model != nil && *model != "" {
		c.LLM.Model = *model
	}
	if numCommits != nil && *numCommits > 0 {
		c.Analysis.DefaultCommits = *numCommits
	}
	if numWorkers != nil && *numWorkers > 0 {
		c.Performance.Workers = *numWorkers
	}
	if timeout != nil && *timeout > 0 {
		c.LLM.Timeout = *timeout
	}
	if verbose != nil {
		c.Output.Verbose = *verbose
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate LLM config
	if c.LLM.Provider == "" {
		return fmt.Errorf("llm.provider cannot be empty")
	}
	if c.LLM.Model == "" {
		return fmt.Errorf("llm.model cannot be empty")
	}
	if c.LLM.Temperature < 0 || c.LLM.Temperature > 1 {
		return fmt.Errorf("llm.temperature must be between 0 and 1, got %f", c.LLM.Temperature)
	}
	if c.LLM.Timeout <= 0 {
		return fmt.Errorf("llm.timeout must be positive, got %v", c.LLM.Timeout)
	}

	// Validate Analysis config
	if c.Analysis.DefaultCommits <= 0 {
		return fmt.Errorf("analysis.default_commits must be positive, got %d", c.Analysis.DefaultCommits)
	}
	if c.Analysis.MaxDiffSize <= 0 {
		return fmt.Errorf("analysis.max_diff_size must be positive, got %d", c.Analysis.MaxDiffSize)
	}

	// Validate Performance config
	if c.Performance.Workers <= 0 {
		return fmt.Errorf("performance.workers must be positive, got %d", c.Performance.Workers)
	}
	if c.Performance.MaxRetries < 0 {
		return fmt.Errorf("performance.max_retries cannot be negative, got %d", c.Performance.MaxRetries)
	}

	// Validate Output config
	validFormats := map[string]bool{"json": true, "text": true, "markdown": true}
	if !validFormats[c.Output.Format] {
		return fmt.Errorf("output.format must be json, text, or markdown, got %s", c.Output.Format)
	}

	return nil
}
