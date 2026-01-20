package analyzer

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"google.golang.org/api/googleapi"
)

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
}

// DefaultRetryConfig returns sensible defaults
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries: 3,
		BaseDelay:  1 * time.Second,
		MaxDelay:   30 * time.Second,
	}
}

// IsRetryable determines if an error is worth retrying
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Context deadline exceeded (timeout)
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	// Google API errors
	var apiErr *googleapi.Error
	if errors.As(err, &apiErr) {
		switch apiErr.Code {
		case 429: // Rate limited
			return true
		case 500, 502, 503, 504: // Server errors
			return true
		}
	}

	// Network-related error messages
	errStr := strings.ToLower(err.Error())
	retryableMessages := []string{
		"connection reset",
		"connection refused",
		"no such host",
		"timeout",
		"temporary failure",
	}
	for _, msg := range retryableMessages {
		if strings.Contains(errStr, msg) {
			return true
		}
	}

	return false
}

// WithRetry executes a function with exponential backoff
func WithRetry(ctx context.Context, cfg RetryConfig, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		if !IsRetryable(lastErr) {
			return lastErr
		}

		if attempt == cfg.MaxRetries {
			break
		}

		// Exponential backoff: 1s, 2s, 4s, ...
		delay := cfg.BaseDelay * time.Duration(1<<attempt)
		if delay > cfg.MaxDelay {
			delay = cfg.MaxDelay
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	return fmt.Errorf("max retries (%d) exceeded: %w", cfg.MaxRetries, lastErr)
}
