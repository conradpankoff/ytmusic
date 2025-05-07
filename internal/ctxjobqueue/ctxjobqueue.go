package ctxjobqueue

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"fknsrs.biz/p/ytmusic/internal/jobqueue"
)

// context registration

var workerKey int

func WithWorker(ctx context.Context, w *jobqueue.Worker) context.Context {
	return context.WithValue(ctx, &workerKey, w)
}

func GetWorker(ctx context.Context) *jobqueue.Worker {
	if v := ctx.Value(&workerKey); v != nil {
		return v.(*jobqueue.Worker)
	}

	return nil
}

// middleware

func Register(w *jobqueue.Worker) func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	return func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		next(rw, r.WithContext(WithWorker(r.Context(), w)))
	}
}

// main interface

var (
	ErrNoWorker = fmt.Errorf("no worker found in context")
)

func Add(ctx context.Context, tx *sql.Tx, job *jobqueue.Job) error {
	w := GetWorker(ctx)
	if w == nil {
		return ErrNoWorker
	}

	if err := w.Add(ctx, tx, job); err != nil {
		return fmt.Errorf("ctxjobqueue.Add: %w", err)
	}

	return nil
}
