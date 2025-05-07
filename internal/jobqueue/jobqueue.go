package jobqueue

import (
	"context"
	"database/sql"
	"fmt"
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

	if err := sorm.SaveRecord(ctx, tx, job); err != nil {
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

	if err := sorm.SaveRecord(ctx, tx, job); err != nil {
		return fmt.Errorf("jobqueue.finish: could not save job record: %w", err)
	}

	return nil
}
