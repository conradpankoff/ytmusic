package jobqueue

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"math/rand"
	"net/url"
	"strings"
	"time"

	"fknsrs.biz/p/sorm"
	"fknsrs.biz/p/ytmusic/internal/sqltypes"
)

func ParsePayload(s string) (string, url.Values, error) {
	if !strings.Contains(s, "?") {
		return s, url.Values{}, nil
	}

	a := strings.SplitN(s, "?", 2)

	m, err := url.ParseQuery(a[1])
	if err != nil {
		return a[0], url.Values{}, err
	}

	return a[0], m, nil
}

func FormatPayload(s string, m url.Values) string {
	if m == nil {
		return s
	}

	return s + "?" + m.Encode()
}

const (
	DefaultFailureDelay = time.Second * 5
)

// RetryConfig holds configuration for database retry operations
type RetryConfig struct {
	MaxAttempts     int
	BaseDelay       time.Duration
	MaxDelay        time.Duration
	BackoffFactor   float64
	JitterFactor    float64
}

// DefaultRetryConfig provides sensible defaults for database retries
var DefaultRetryConfig = RetryConfig{
	MaxAttempts:     10,
	BaseDelay:       50 * time.Millisecond,
	MaxDelay:        5 * time.Second,
	BackoffFactor:   2.0,
	JitterFactor:    0.1,
}

// retryWithExponentialBackoff executes a function with exponential backoff retry logic
func retryWithExponentialBackoff(ctx context.Context, config RetryConfig, operation func() error) error {
	var lastErr error
	
	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		lastErr = operation()
		
		if lastErr == nil {
			return nil
		}
		
		// Check if this is a database lock error that we should retry
		if !isDatabaseLockError(lastErr) {
			return lastErr
		}
		
		// Don't sleep on the last attempt
		if attempt == config.MaxAttempts-1 {
			break
		}
		
		// Calculate delay with exponential backoff and jitter
		delay := calculateBackoffDelay(config, attempt)
		
		// Check for context cancellation during sleep
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}
	
	return lastErr
}

// isDatabaseLockError checks if an error is related to database locking
func isDatabaseLockError(err error) bool {
	if err == nil {
		return false
	}
	
	errStr := err.Error()
	return strings.Contains(errStr, "database is locked") ||
		   strings.Contains(errStr, "database table is locked") ||
		   strings.Contains(errStr, "SQLITE_BUSY")
}

// calculateBackoffDelay calculates the delay for exponential backoff with jitter
func calculateBackoffDelay(config RetryConfig, attempt int) time.Duration {
	// Calculate exponential backoff delay
	delay := float64(config.BaseDelay) * math.Pow(config.BackoffFactor, float64(attempt))
	
	// Apply jitter to reduce thundering herd effect
	jitter := delay * config.JitterFactor * (rand.Float64()*2 - 1) // Random value between -jitterFactor and +jitterFactor
	delay += jitter
	
	// Ensure delay doesn't exceed max delay
	if delay > float64(config.MaxDelay) {
		delay = float64(config.MaxDelay)
	}
	
	// Ensure delay is not negative
	if delay < 0 {
		delay = float64(config.BaseDelay)
	}
	
	return time.Duration(delay)
}

// beginTransactionWithRetry creates a database transaction with retry logic and appropriate settings
func beginTransactionWithRetry(ctx context.Context, db *sql.DB) (*sql.Tx, error) {
	var tx *sql.Tx
	
	err := retryWithExponentialBackoff(ctx, DefaultRetryConfig, func() error {
		var err error
		// Use appropriate isolation level for SQLite with WAL mode
		tx, err = db.BeginTx(ctx, &sql.TxOptions{
			Isolation: sql.LevelReadCommitted,
			ReadOnly:  false,
		})
		return err
	})
	
	return tx, err
}

// commitTransactionWithRetry commits a transaction with retry logic
func commitTransactionWithRetry(ctx context.Context, tx *sql.Tx) error {
	return retryWithExponentialBackoff(ctx, DefaultRetryConfig, func() error {
		return tx.Commit()
	})
}

// job definition

type Job struct {
	ID                int `sql:",table:jobs"`
	CreatedAt         time.Time
	QueueName         string
	Payload           string
	RunAfter          time.Time
	FailureDelay      time.Duration
	AttemptsRemaining int
	ReservedAt        *time.Time
	ReservedUntil     *time.Time
	FinishedAt        *time.Time
	Progress          *int // Progress percentage (0-100) for long-running jobs
	ErrorMessages     sqltypes.JSONStringSlice
	OutputMessages    sqltypes.JSONStringSlice
}

func findNext(ctx context.Context, db sorm.Querier, queueNames []string, now time.Time) (*Job, error) {
	var parameters []interface{}
	var placeholders []string

	for i := range queueNames {
		parameters = append(parameters, queueNames[i])
		placeholders = append(placeholders, fmt.Sprintf("?%d", i+1))
	}

	parameters = append(parameters, now)

	query := fmt.Sprintf(
		"where queue_name in (%s) and run_after < ?%d and (reserved_until is null or reserved_until < ?%d) and finished_at is null order by run_after asc",
		strings.Join(placeholders, ", "),
		len(parameters),
		len(parameters),
	)

	var job Job
	if err := sorm.FindFirstWhere(ctx, db, &job, query, parameters...); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, fmt.Errorf("jobqueue.findNext: could not find pending job record: %w", err)
	}

	return &job, nil
}

func reserve(ctx context.Context, tx *sql.Tx, job *Job, now time.Time, reserveDuration time.Duration) error {
	if job.ReservedUntil != nil && job.ReservedUntil.After(now) {
		return fmt.Errorf("jobqueue.reserve: can't reserve a job with a non-expired reservation")
	}
	if job.FinishedAt != nil {
		return fmt.Errorf("jobqueue.reserve: can't reserve a job that has already finished")
	}

	if reserveDuration == 0 {
		reserveDuration = time.Minute * 5
	}

	reservedUntil := now.Add(reserveDuration)
	job.ReservedAt = &now
	job.ReservedUntil = &reservedUntil
	job.Progress = nil

	// Use retry mechanism for save operation
	err := retryWithExponentialBackoff(ctx, DefaultRetryConfig, func() error {
		return sorm.SaveRecord(ctx, tx, job)
	})
	
	if err != nil {
		return fmt.Errorf("jobqueue.reserve: could not save job record: %w", err)
	}

	return nil
}

func findNextAndReserve(ctx context.Context, tx *sql.Tx, queueNames []string, now time.Time, reserveDuration time.Duration) (*Job, error) {
	j, err := findNext(ctx, tx, queueNames, now)
	if err != nil {
		return nil, fmt.Errorf("jobqueue.findNextAndReserve: could not find next job: %w", err)
	}

	if j == nil {
		return nil, nil
	}

	if err := reserve(ctx, tx, j, now, reserveDuration); err != nil {
		return nil, fmt.Errorf("jobqueue.findNextAndReserve: could not reserve job: %w", err)
	}

	return j, nil
}

func finish(ctx context.Context, tx *sql.Tx, job *Job, now time.Time, errorMessage, outputMessage string) error {
	if job.FinishedAt != nil {
		return fmt.Errorf("jobqueue.finish: can't finish a job that has already finished")
	}

	job.FinishedAt = &now
	job.ErrorMessages = append(job.ErrorMessages, errorMessage)
	job.OutputMessages = append(job.OutputMessages, outputMessage)

	if errorMessage != "" && job.AttemptsRemaining > 0 {
		job.AttemptsRemaining--
		job.RunAfter = now.Add(job.FailureDelay)
		job.ReservedAt = nil
		job.ReservedUntil = nil
		job.FinishedAt = nil
	}

	// Use retry mechanism for save operation
	err := retryWithExponentialBackoff(ctx, DefaultRetryConfig, func() error {
		return sorm.SaveRecord(ctx, tx, job)
	})
	
	if err != nil {
		return fmt.Errorf("jobqueue.finish: could not save job record: %w", err)
	}

	return nil
}

func updateProgress(ctx context.Context, tx *sql.Tx, job *Job, progress int) error {
	if progress < 0 || progress > 100 {
		return fmt.Errorf("jobqueue.updateProgress: progress must be between 0 and 100")
	}

	// Prevent progress from going backwards
	if job.Progress != nil && progress < *job.Progress {
		return nil // Silently ignore backwards progress updates
	}

	job.Progress = &progress

	// Use retry mechanism for save operation
	err := retryWithExponentialBackoff(ctx, DefaultRetryConfig, func() error {
		return sorm.SaveRecord(ctx, tx, job)
	})
	
	if err != nil {
		return fmt.Errorf("jobqueue.updateProgress: could not save job record: %w", err)
	}

	return nil
}
