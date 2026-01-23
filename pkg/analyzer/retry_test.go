package analyzer

import (
	"context"
	"errors"
	"testing"
	"time"

	"google.golang.org/api/googleapi"
)

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "context deadline exceeded",
			err:      context.DeadlineExceeded,
			expected: true,
		},
		{
			name:     "rate limit error (429)",
			err:      &googleapi.Error{Code: 429},
			expected: true,
		},
		{
			name:     "internal server error (500)",
			err:      &googleapi.Error{Code: 500},
			expected: true,
		},
		{
			name:     "bad gateway (502)",
			err:      &googleapi.Error{Code: 502},
			expected: true,
		},
		{
			name:     "service unavailable (503)",
			err:      &googleapi.Error{Code: 503},
			expected: true,
		},
		{
			name:     "gateway timeout (504)",
			err:      &googleapi.Error{Code: 504},
			expected: true,
		},
		{
			name:     "client error (400)",
			err:      &googleapi.Error{Code: 400},
			expected: false,
		},
		{
			name:     "not found (404)",
			err:      &googleapi.Error{Code: 404},
			expected: false,
		},
		{
			name:     "connection reset",
			err:      errors.New("connection reset by peer"),
			expected: true,
		},
		{
			name:     "connection refused",
			err:      errors.New("connection refused"),
			expected: true,
		},
		{
			name:     "no such host",
			err:      errors.New("no such host"),
			expected: true,
		},
		{
			name:     "timeout",
			err:      errors.New("request timeout"),
			expected: true,
		},
		{
			name:     "temporary failure",
			err:      errors.New("temporary failure in name resolution"),
			expected: true,
		},
		{
			name:     "generic error",
			err:      errors.New("some other error"),
			expected: false,
		},
		{
			name:     "case insensitive match",
			err:      errors.New("Connection Reset by peer"),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRetryable(tt.err)
			if result != tt.expected {
				t.Errorf("IsRetryable(%v) = %v, expected %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestWithRetry_Success(t *testing.T) {
	ctx := context.Background()
	cfg := RetryConfig{
		MaxRetries: 3,
		BaseDelay:  10 * time.Millisecond,
		MaxDelay:   100 * time.Millisecond,
	}

	callCount := 0
	fn := func() error {
		callCount++
		return nil // Success on first try
	}

	err := WithRetry(ctx, cfg, fn)
	if err != nil {
		t.Errorf("WithRetry() returned error: %v", err)
	}
	if callCount != 1 {
		t.Errorf("Expected function to be called once, called %d times", callCount)
	}
}

func TestWithRetry_SuccessAfterRetries(t *testing.T) {
	ctx := context.Background()
	cfg := RetryConfig{
		MaxRetries: 3,
		BaseDelay:  10 * time.Millisecond,
		MaxDelay:   100 * time.Millisecond,
	}

	callCount := 0
	fn := func() error {
		callCount++
		if callCount < 3 {
			return &googleapi.Error{Code: 503} // Retryable error
		}
		return nil // Success on third try
	}

	err := WithRetry(ctx, cfg, fn)
	if err != nil {
		t.Errorf("WithRetry() returned error: %v", err)
	}
	if callCount != 3 {
		t.Errorf("Expected function to be called 3 times, called %d times", callCount)
	}
}

func TestWithRetry_NonRetryableError(t *testing.T) {
	ctx := context.Background()
	cfg := RetryConfig{
		MaxRetries: 3,
		BaseDelay:  10 * time.Millisecond,
		MaxDelay:   100 * time.Millisecond,
	}

	callCount := 0
	expectedErr := &googleapi.Error{Code: 400} // Non-retryable
	fn := func() error {
		callCount++
		return expectedErr
	}

	err := WithRetry(ctx, cfg, fn)
	if err != expectedErr {
		t.Errorf("WithRetry() returned error: %v, expected %v", err, expectedErr)
	}
	if callCount != 1 {
		t.Errorf("Expected function to be called once (non-retryable), called %d times", callCount)
	}
}

func TestWithRetry_MaxRetriesExceeded(t *testing.T) {
	ctx := context.Background()
	cfg := RetryConfig{
		MaxRetries: 3,
		BaseDelay:  10 * time.Millisecond,
		MaxDelay:   100 * time.Millisecond,
	}

	callCount := 0
	retryableErr := &googleapi.Error{Code: 503}
	fn := func() error {
		callCount++
		return retryableErr
	}

	err := WithRetry(ctx, cfg, fn)
	if err == nil {
		t.Error("WithRetry() should return error when max retries exceeded")
	}
	// Should be called: initial + 3 retries = 4 times
	if callCount != 4 {
		t.Errorf("Expected function to be called 4 times (initial + 3 retries), called %d times", callCount)
	}
}

func TestWithRetry_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cfg := RetryConfig{
		MaxRetries: 5,
		BaseDelay:  50 * time.Millisecond,
		MaxDelay:   200 * time.Millisecond,
	}

	callCount := 0
	fn := func() error {
		callCount++
		if callCount == 2 {
			cancel() // Cancel after second call
		}
		return &googleapi.Error{Code: 503} // Always retryable
	}

	err := WithRetry(ctx, cfg, fn)
	if err != context.Canceled {
		t.Errorf("WithRetry() should return context.Canceled, got: %v", err)
	}
	// Should be called at least twice before cancellation
	if callCount < 2 {
		t.Errorf("Expected function to be called at least 2 times, called %d times", callCount)
	}
}

func TestWithRetry_ExponentialBackoff(t *testing.T) {
	ctx := context.Background()
	cfg := RetryConfig{
		MaxRetries: 3,
		BaseDelay:  50 * time.Millisecond,
		MaxDelay:   500 * time.Millisecond,
	}

	callTimes := []time.Time{}
	fn := func() error {
		callTimes = append(callTimes, time.Now())
		return &googleapi.Error{Code: 503} // Always retryable
	}

	start := time.Now()
	WithRetry(ctx, cfg, fn)
	duration := time.Since(start)

	// Should have 4 calls (initial + 3 retries)
	if len(callTimes) != 4 {
		t.Errorf("Expected 4 calls, got %d", len(callTimes))
	}

	// Check that delays are approximately exponential
	// First retry: ~50ms, Second: ~100ms, Third: ~200ms
	// Total should be at least 350ms
	if duration < 350*time.Millisecond {
		t.Errorf("Expected total duration >= 350ms with exponential backoff, got %v", duration)
	}

	// Verify delays between calls
	for i := 1; i < len(callTimes); i++ {
		delay := callTimes[i].Sub(callTimes[i-1])
		expectedMin := cfg.BaseDelay * time.Duration(1<<uint(i-1))
		if delay < expectedMin {
			t.Errorf("Delay %d: %v is less than expected minimum %v", i, delay, expectedMin)
		}
	}
}

func TestWithRetry_MaxDelayEnforced(t *testing.T) {
	ctx := context.Background()
	cfg := RetryConfig{
		MaxRetries: 10,
		BaseDelay:  100 * time.Millisecond,
		MaxDelay:   300 * time.Millisecond, // Cap at 300ms
	}

	callTimes := []time.Time{}
	fn := func() error {
		callTimes = append(callTimes, time.Now())
		if len(callTimes) <= 5 {
			return &googleapi.Error{Code: 503} // Retryable
		}
		return nil // Success after 5 attempts
	}

	WithRetry(ctx, cfg, fn)

	// Check that no delay exceeds MaxDelay
	for i := 1; i < len(callTimes); i++ {
		delay := callTimes[i].Sub(callTimes[i-1])
		if delay > cfg.MaxDelay+50*time.Millisecond { // Add small buffer for timing variance
			t.Errorf("Delay %d: %v exceeds MaxDelay %v", i, delay, cfg.MaxDelay)
		}
	}
}

func TestDefaultRetryConfig(t *testing.T) {
	cfg := DefaultRetryConfig()

	if cfg.MaxRetries != 3 {
		t.Errorf("Expected MaxRetries=3, got %d", cfg.MaxRetries)
	}
	if cfg.BaseDelay != 1*time.Second {
		t.Errorf("Expected BaseDelay=1s, got %v", cfg.BaseDelay)
	}
	if cfg.MaxDelay != 30*time.Second {
		t.Errorf("Expected MaxDelay=30s, got %v", cfg.MaxDelay)
	}
}
