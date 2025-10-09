package ctxdb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"fknsrs.biz/p/ytmusic/internal/dbsavepoint"
)

var (
	ErrNoDB = fmt.Errorf("ctxdb: no db found in context")
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

// context registration

var dbKey int

func WithDB(ctx context.Context, db *sql.DB) context.Context {
	return context.WithValue(ctx, &dbKey, db)
}

func GetDB(ctx context.Context) *sql.DB {
	if v := ctx.Value(&dbKey); v != nil {
		return v.(*sql.DB)
	}

	return nil
}

var savepointKey int

func WithSavepoint(ctx context.Context, sp *dbsavepoint.Savepoint) context.Context {
	return context.WithValue(ctx, &savepointKey, sp)
}

func GetSavepoint(ctx context.Context) *dbsavepoint.Savepoint {
	if v := ctx.Value(&savepointKey); v != nil {
		return v.(*dbsavepoint.Savepoint)
	}

	return nil
}

func createSavepoint(ctx context.Context, name string) (context.Context, *dbsavepoint.Savepoint, error) {
	if parent := GetSavepoint(ctx); parent != nil {
		sp, err := dbsavepoint.CreateFromParent(ctx, parent, name)
		if err != nil {
			return ctx, nil, err
		}

		return WithSavepoint(ctx, sp), sp, nil
	}

	if db := GetDB(ctx); db != nil {
		sp, err := dbsavepoint.CreateFromDB(ctx, db, name)
		if err != nil {
			return ctx, nil, err
		}

		return WithSavepoint(ctx, sp), sp, nil
	}

	return ctx, nil, ErrNoDB
}

type SavepointFunc func(ctx context.Context, sp *dbsavepoint.Savepoint) error

func UsingSavepoint(ctx context.Context, name string, fn SavepointFunc) error {
	ctx2, sp, err := createSavepoint(ctx, name)
	if err != nil {
		return err
	}

	if err := fn(ctx2, sp); err != nil {
		if err2 := sp.Rollback(ctx); err2 != nil {
			return errors.Join(err, err2)
		}

		return err
	}

	if err := sp.Release(ctx); err != nil {
		return err
	}

	return nil
}

type TxFunc func(ctx context.Context, tx *sql.Tx) error

func UsingTx(ctx context.Context, opts *sql.TxOptions, fn TxFunc) error {
	db := GetDB(ctx)
	if db == nil {
		return ErrNoDB
	}

	// Use retry mechanism for the entire transaction
	return retryWithExponentialBackoff(ctx, DefaultRetryConfig, func() error {
		// Set appropriate transaction options if none provided
		if opts == nil {
			opts = &sql.TxOptions{
				Isolation: sql.LevelReadCommitted,
				ReadOnly:  false,
			}
		}

		tx, err := db.BeginTx(ctx, opts)
		if err != nil {
			return err
		}
		defer tx.Rollback()

		if err := fn(ctx, tx); err != nil {
			return err
		}

		if err := tx.Commit(); err != nil {
			return err
		}

		return nil
	})
}

// UsingTxWithRetry is an alias for UsingTx with explicit retry configuration
func UsingTxWithRetry(ctx context.Context, opts *sql.TxOptions, config RetryConfig, fn TxFunc) error {
	db := GetDB(ctx)
	if db == nil {
		return ErrNoDB
	}

	return retryWithExponentialBackoff(ctx, config, func() error {
		if opts == nil {
			opts = &sql.TxOptions{
				Isolation: sql.LevelReadCommitted,
				ReadOnly:  false,
			}
		}

		tx, err := db.BeginTx(ctx, opts)
		if err != nil {
			return err
		}
		defer tx.Rollback()

		if err := fn(ctx, tx); err != nil {
			return err
		}

		if err := tx.Commit(); err != nil {
			return err
		}

		return nil
	})
}

// middleware

func Register(db *sql.DB) func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	return func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		next(rw, r.WithContext(WithDB(r.Context(), db)))
	}
}
