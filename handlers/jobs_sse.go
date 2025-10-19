package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"fknsrs.biz/p/sorm"

	"fknsrs.biz/p/ytmusic/internal/ctxdb"
	"fknsrs.biz/p/ytmusic/internal/jobqueue"
)

// JobUpdate represents a job progress update for SSE
type JobUpdate struct {
	ID       int    `json:"id"`
	Progress *int32 `json:"progress"`
	Status   string `json:"status"`
}

func JobsSSE(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")
	rw.Header().Set("Access-Control-Allow-Origin", "*")

	ctx := r.Context()

	lastProgress := make(map[int]*int32)
	lastStatus := make(map[int]string)

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			var jobs []jobqueue.Job
			if err := sorm.FindWhere(ctx, ctxdb.GetDB(ctx), &jobs, "where finished_at is null order by id desc limit 100"); err != nil {
				continue
			}

			for _, job := range jobs {
				var status string
				if job.FinishedAt.Valid {
					status = "finished"
				} else if job.ReservedAt.Valid {
					status = "running"
				} else {
					status = "pending"
				}

				changed := false

				if v, exists := lastProgress[job.ID]; !exists {
					changed = true
				} else if !job.Progress.Valid && v != nil {
					changed = true
				} else if job.Progress.Valid && v == nil {
					changed = true
				} else if job.Progress.Valid && v != nil && job.Progress.Int32 != *v {
					changed = true
				}

				if v, exists := lastStatus[job.ID]; !exists {
					changed = true
				} else if v != status {
					changed = true
				}

				if changed {
					var progress *int32
					if job.Progress.Valid {
						progress = &job.Progress.Int32
					}
					update := JobUpdate{
						ID:       job.ID,
						Progress: progress,
						Status:   status,
					}

					data, err := json.Marshal(update)
					if err != nil {
						continue
					}

					fmt.Fprintf(rw, "data: %s\n\n", data)
					if f, ok := rw.(http.Flusher); ok {
						f.Flush()
					}

					lastProgress[job.ID] = progress
					lastStatus[job.ID] = status
				}
			}
		}
	}
}
