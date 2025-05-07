package jobqueue

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"fknsrs.biz/p/sorm"
	"github.com/sirupsen/logrus"

	"fknsrs.biz/p/ytmusic/internal/catchpanic"
	"fknsrs.biz/p/ytmusic/internal/ctxdb"
	"fknsrs.biz/p/ytmusic/internal/ctxlogger"
)

// worker

var (
	ErrWorkerExists       = fmt.Errorf("worker already exists")
	ErrWorkerDoesNotExist = fmt.Errorf("worker does not exist")
	ErrNoPendingJobs      = fmt.Errorf("no pending jobs")
)

type WorkerFunction func(ctx context.Context, w *Worker, j *Job) (string, error)

type Worker struct {
	l  sync.RWMutex
	ch chan struct{}
	m  map[string]WorkerFunction
}

func NewWorker(workerFunctions map[string]WorkerFunction) *Worker {
	if workerFunctions == nil {
		workerFunctions = make(map[string]WorkerFunction)
	}

	return &Worker{
		ch: make(chan struct{}, 100),
		m:  workerFunctions,
	}
}

func (w *Worker) failIfAnyDoNotExist(queueNames []string) error {
	var a []string

	for _, queueName := range queueNames {
		if _, ok := w.m[queueName]; !ok {
			a = append(a, queueName)
		}
	}

	if len(a) > 0 {
		return fmt.Errorf("jobqueue.Worker.failIfAnyDoNotExist: worker(s) do not exist: %v: %w", a, ErrWorkerDoesNotExist)
	}

	return nil
}

func (w *Worker) failIfAnyExist(queueNames []string) error {
	var a []string

	for _, queueName := range queueNames {
		if _, ok := w.m[queueName]; ok {
			a = append(a, queueName)
		}
	}

	if len(a) > 0 {
		return fmt.Errorf("jobqueue.Worker.failIfAnyExist: worker(s) already exist: %v: %w", a, ErrWorkerExists)
	}

	return nil
}

func (w *Worker) Add(ctx context.Context, tx *sql.Tx, job *Job) error {
	w.l.RLock()
	if err := w.failIfAnyDoNotExist([]string{job.QueueName}); err != nil {
		w.l.RUnlock()
		return fmt.Errorf("jobqueue.Worker.Add: %w", err)
	}
	w.l.RUnlock()

	if job.CreatedAt.IsZero() {
		job.CreatedAt = time.Now()
	}
	if job.RunAfter.IsZero() {
		job.RunAfter = time.Now()
	}
	if job.FailureDelay == 0 {
		job.FailureDelay = DefaultFailureDelay
	}
	if job.AttemptsRemaining == 0 {
		job.AttemptsRemaining = 5
	}

	if err := sorm.CreateRecord(ctx, tx, job); err != nil {
		return fmt.Errorf("jobqueue.Worker.Add: could not create job record: %w", err)
	}

	select {
	case w.ch <- struct{}{}:
		// channel already full
	default:
		// nothing
	}

	return nil
}

func (w *Worker) Trigger(ctx context.Context) {
	w.ch <- struct{}{}
}

func (w *Worker) Register(queueName string, workerFunction WorkerFunction) error {
	w.l.RLock()
	if err := w.failIfAnyExist([]string{queueName}); err != nil {
		w.l.RUnlock()
		return fmt.Errorf("jobqueue.Worker.Register: %w", err)
	}
	w.l.RUnlock()

	w.l.Lock()
	defer w.l.Unlock()

	if err := w.failIfAnyExist([]string{queueName}); err != nil {
		return fmt.Errorf("jobqueue.Worker.Register: %w", err)
	}

	w.m[queueName] = workerFunction

	return nil
}

func (w *Worker) RegisterAll(workers map[string]WorkerFunction) error {
	var queueNames []string
	for queueName := range workers {
		queueNames = append(queueNames, queueName)
	}

	w.l.RLock()
	if err := w.failIfAnyExist(queueNames); err != nil {
		w.l.RUnlock()
		return fmt.Errorf("jobqueue.Worker.RegisterAll: %w", err)
	}
	w.l.RUnlock()

	w.l.Lock()
	defer w.l.Unlock()

	if err := w.failIfAnyExist(queueNames); err != nil {
		return fmt.Errorf("jobqueue.Worker.RegisterAll: %w", err)
	}

	for queueName, workerFunc := range workers {
		w.m[queueName] = workerFunc
	}

	return nil
}

func (w *Worker) GetQueueNames() []string {
	w.l.RLock()
	defer w.l.RUnlock()

	var queueNames []string

	for k := range w.m {
		queueNames = append(queueNames, k)
	}

	return queueNames
}

func (w *Worker) RunOnce(ctx context.Context) (bool, error) {
	db := ctxdb.GetDB(ctx)

	tx1, err := db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("jobqueue.Worker.RunOnce: could not open transaction to find/reserve: %w", err)
	}
	defer tx1.Rollback()

	attempts := 25
again:
	attempts--
	job, err := findNextAndReserve(ctx, tx1, w.GetQueueNames(), time.Now(), time.Minute*5)
	if err != nil {
		if strings.Contains(err.Error(), "database is locked") && attempts > 0 {
			time.Sleep(time.Duration(rand.Int63n(int64(time.Millisecond) * 500)))
			goto again
		}
		return false, fmt.Errorf("jobqueue.Worker.RunOnce: could not find/reserve job: %w", err)
	}

	if job == nil {
		return false, ErrNoPendingJobs
	}

	l := ctxlogger.GetLogger(ctx).WithFields(logrus.Fields{
		"job_queue_name": job.QueueName,
		"job_id":         job.ID,
	})

	if err := tx1.Commit(); err != nil {
		return false, fmt.Errorf("jobqueue.Worker.RunOnce: could not commit transaction to find/reserve: %w", err)
	}

	l.Info("found pending job, running function")

	workerFunction, ok := w.m[job.QueueName]
	if !ok {
		return false, fmt.Errorf("jobqueue.Worker.RunOnce: worker function not set for queue: %s", job.QueueName)
	}

	var errorMessage string
	outputMessage, err := catchpanic.CatchErr1(func() (string, error) { return workerFunction(ctx, w, job) })
	if err != nil {
		errorMessage = err.Error()
	}

	l.WithFields(logrus.Fields{"error_message": errorMessage, "output_message": outputMessage}).Info("finished job")

	tx2, err := db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("jobqueue.Worker.RunOnce: could not open transaction to finish: %w", err)
	}
	defer tx2.Rollback()

	if err := finish(ctx, tx2, job, time.Now(), errorMessage, outputMessage); err != nil {
		return false, fmt.Errorf("jobqueue.Worker.RunOnce: could not finish job: %w", err)
	}

	if err := tx2.Commit(); err != nil {
		return false, fmt.Errorf("jobqueue.Worker.RunOnce: could not commit transaction to finish: %w", err)
	}

	return true, nil
}

func (w *Worker) Run(ctx context.Context) error {
	delay := time.Second * 5

	w.Trigger(ctx)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			if didRunJob, err := w.RunOnce(ctx); err != nil && err != ErrNoPendingJobs {
				ctxlogger.GetLogger(ctx).WithError(err).Error("could not run job")
				delay = time.Second * 30
			} else if didRunJob {
				delay = 0
			} else {
				delay = time.Second * 30
			}
		case <-w.ch:
			if didRunJob, err := w.RunOnce(ctx); err != nil && err != ErrNoPendingJobs {
				ctxlogger.GetLogger(ctx).WithError(err).Error("could not run job")
				delay = time.Second * 30
			} else if didRunJob {
				delay = 0
			} else {
				delay = time.Second * 30
			}
		}
	}
}
