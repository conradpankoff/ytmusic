package jobqueue

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRetryWithExponentialBackoff(t *testing.T) {
	tests := []struct {
		name          string
		operation     func() error
		config        RetryConfig
		expectedCalls int
		shouldFail    bool
	}{
		{
			name: "success on first try",
			operation: func() error {
				return nil
			},
			config: RetryConfig{
				MaxAttempts:   3,
				BaseDelay:     10 * time.Millisecond,
				MaxDelay:      100 * time.Millisecond,
				BackoffFactor: 2.0,
				JitterFactor:  0.1,
			},
			expectedCalls: 1,
			shouldFail:    false,
		},
		{
			name: "success on third try with database lock",
			operation: func() func() error {
				calls := 0
				return func() error {
					calls++
					if calls < 3 {
						return errors.New("database is locked")
					}
					return nil
				}
			}(),
			config: RetryConfig{
				MaxAttempts:   5,
				BaseDelay:     1 * time.Millisecond,
				MaxDelay:      100 * time.Millisecond,
				BackoffFactor: 2.0,
				JitterFactor:  0.1,
			},
			expectedCalls: 3,
			shouldFail:    false,
		},
		{
			name: "non-retryable error fails immediately",
			operation: func() error {
				return errors.New("some other error")
			},
			config: RetryConfig{
				MaxAttempts:   3,
				BaseDelay:     10 * time.Millisecond,
				MaxDelay:      100 * time.Millisecond,
				BackoffFactor: 2.0,
				JitterFactor:  0.1,
			},
			expectedCalls: 1,
			shouldFail:    true,
		},
		{
			name: "exhausted retries with database lock",
			operation: func() error {
				return errors.New("database is locked")
			},
			config: RetryConfig{
				MaxAttempts:   3,
				BaseDelay:     1 * time.Millisecond,
				MaxDelay:      100 * time.Millisecond,
				BackoffFactor: 2.0,
				JitterFactor:  0.1,
			},
			expectedCalls: 3,
			shouldFail:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			operation := func() error {
				calls++
				return tt.operation()
			}

			ctx := context.Background()
			err := retryWithExponentialBackoff(ctx, tt.config, operation)

			if tt.shouldFail && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.shouldFail && err != nil {
				t.Errorf("expected no error but got: %v", err)
			}
			if calls != tt.expectedCalls {
				t.Errorf("expected %d calls but got %d", tt.expectedCalls, calls)
			}
		})
	}
}

func TestIsDatabaseLockError(t *testing.T) {
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
			name:     "database is locked",
			err:      errors.New("database is locked"),
			expected: true,
		},
		{
			name:     "database table is locked",
			err:      errors.New("database table is locked"),
			expected: true,
		},
		{
			name:     "SQLITE_BUSY error",
			err:      errors.New("SQLITE_BUSY"),
			expected: true,
		},
		{
			name:     "other error",
			err:      errors.New("some other error"),
			expected: false,
		},
		{
			name:     "error containing database is locked",
			err:      errors.New("error: database is locked: details"),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isDatabaseLockError(tt.err)
			if result != tt.expected {
				t.Errorf("expected %v but got %v for error: %v", tt.expected, result, tt.err)
			}
		})
	}
}

func TestCalculateBackoffDelay(t *testing.T) {
	config := RetryConfig{
		BaseDelay:     100 * time.Millisecond,
		MaxDelay:      5 * time.Second,
		BackoffFactor: 2.0,
		JitterFactor:  0.1,
	}

	// Test that delays increase exponentially
	delay0 := calculateBackoffDelay(config, 0)
	delay1 := calculateBackoffDelay(config, 1)
	delay2 := calculateBackoffDelay(config, 2)

	// Due to jitter, exact comparison is difficult, but we can check rough ranges
	if delay0 < 80*time.Millisecond || delay0 > 120*time.Millisecond {
		t.Errorf("unexpected delay for attempt 0: %v", delay0)
	}
	if delay1 < 180*time.Millisecond || delay1 > 220*time.Millisecond {
		t.Errorf("unexpected delay for attempt 1: %v", delay1)
	}
	if delay2 < 360*time.Millisecond || delay2 > 440*time.Millisecond {
		t.Errorf("unexpected delay for attempt 2: %v", delay2)
	}

	// Test max delay cap
	delayHigh := calculateBackoffDelay(config, 10)
	if delayHigh > config.MaxDelay {
		t.Errorf("delay exceeded max delay: %v > %v", delayHigh, config.MaxDelay)
	}
}

func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	
	operation := func() error {
		return errors.New("database is locked")
	}
	
	config := RetryConfig{
		MaxAttempts:   10,
		BaseDelay:     100 * time.Millisecond,
		MaxDelay:      1 * time.Second,
		BackoffFactor: 2.0,
		JitterFactor:  0.1,
	}
	
	// Cancel context after a short delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()
	
	err := retryWithExponentialBackoff(ctx, config, operation)
	
	if err != context.Canceled {
		t.Errorf("expected context.Canceled but got: %v", err)
	}
}