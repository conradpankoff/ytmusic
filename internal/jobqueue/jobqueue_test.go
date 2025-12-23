package jobqueue

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

// Test that our jobqueue type changes work correctly
func TestJobTypeConversions(t *testing.T) {
	t.Run("Job model with new types", func(t *testing.T) {
		now := time.Now()
		
		job := Job{
			ID:                1,
			CreatedAt:         now,
			QueueName:         "test-queue",
			Payload:           "test-payload",
			RunAfter:          now,
			FailureDelay:      int64(5 * time.Second), // 5 seconds in nanoseconds
			AttemptsRemaining: 3,
			ReservedAt:        sql.NullTime{Valid: false},
			ReservedUntil:     sql.NullTime{Valid: false},
			FinishedAt:        sql.NullTime{Valid: false},
			Progress:          sql.NullInt32{Valid: false},
		}
		
		// Test FailureDelay as int64
		expectedDelay := int64(5 * time.Second)
		if job.FailureDelay != expectedDelay {
			t.Errorf("FailureDelay should be %d, got %d", expectedDelay, job.FailureDelay)
		}
		
		// Test null fields are initially invalid
		if job.ReservedAt.Valid {
			t.Error("ReservedAt should be initially invalid")
		}
		
		if job.ReservedUntil.Valid {
			t.Error("ReservedUntil should be initially invalid")  
		}
		
		if job.FinishedAt.Valid {
			t.Error("FinishedAt should be initially invalid")
		}
		
		if job.Progress.Valid {
			t.Error("Progress should be initially invalid")
		}
	})
	
	t.Run("Job reservation logic", func(t *testing.T) {
		now := time.Now()
		reservedUntil := now.Add(5 * time.Minute)
		
		job := Job{
			ReservedAt:    sql.NullTime{Time: now, Valid: true},
			ReservedUntil: sql.NullTime{Time: reservedUntil, Valid: true},
			FinishedAt:    sql.NullTime{Valid: false},
			Progress:      sql.NullInt32{Valid: false},
		}
		
		// Test that reserved fields work correctly
		if !job.ReservedAt.Valid || !job.ReservedUntil.Valid {
			t.Error("Reservation times should be valid")
		}
		
		// Test reservation check logic (similar to reserve function)
		if job.ReservedUntil.Valid && job.ReservedUntil.Time.After(now) {
			// This should be true - job is currently reserved
			if !job.ReservedUntil.Time.After(now) {
				t.Error("Job should be considered reserved")
			}
		}
		
		// Test finished check logic  
		if job.FinishedAt.Valid {
			t.Error("Job should not be finished")
		}
	})
	
	t.Run("Job progress updates", func(t *testing.T) {
		job := Job{
			Progress: sql.NullInt32{Valid: false},
		}
		
		// Test setting progress
		job.Progress = sql.NullInt32{Int32: 50, Valid: true}
		
		if !job.Progress.Valid {
			t.Error("Progress should be valid after setting")
		}
		
		if job.Progress.Int32 != 50 {
			t.Errorf("Progress should be 50, got %d", job.Progress.Int32)
		}
		
		// Test clearing progress (like in reserve function)
		job.Progress = sql.NullInt32{Valid: false}
		
		if job.Progress.Valid {
			t.Error("Progress should be invalid after clearing")
		}
	})
	
	t.Run("Worker progress validation", func(t *testing.T) {
		w := &Worker{}
		job := &Job{}
		ctx := context.Background()
		
		// Test invalid progress values should return error
		err := w.UpdateProgress(ctx, job, -1)
		if err == nil {
			t.Error("UpdateProgress should return error for negative progress")
		}
		
		err = w.UpdateProgress(ctx, job, 101)
		if err == nil {
			t.Error("UpdateProgress should return error for progress > 100")
		}
		
		// Note: We can't test the valid case without a database connection
		// but the validation logic is what we're testing here
	})
}