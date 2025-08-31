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
	ID       int  `json:"id"`
	Progress *int `json:"progress"`
	Status   string `json:"status"`
}

// JobsSSE handles Server-Sent Events for real-time job updates
func JobsSSE(rw http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")
	rw.Header().Set("Access-Control-Allow-Origin", "*")

	ctx := r.Context()
	
	// Track last seen progress for each job to detect changes
	lastProgress := make(map[int]*int)
	
	// Send updates every 2 seconds
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Get current job status
			var jobs []jobqueue.Job
			if err := sorm.FindWhere(ctx, ctxdb.GetDB(ctx), &jobs, "where finished_at is null order by id desc limit 100"); err != nil {
				continue // Skip on error
			}

			// Check for progress changes and send updates
			for _, job := range jobs {
				var status string
				if job.FinishedAt != nil {
					status = "finished"
				} else if job.ReservedAt != nil {
					status = "running"
				} else {
					status = "pending"
				}

				// Check if this is the first time we see this job OR if progress changed
				lastProg, exists := lastProgress[job.ID]
				progressChanged := false
				
				if !exists {
					// First time seeing this job
					progressChanged = true
				} else if job.Progress == nil && lastProg != nil {
					// Progress was removed
					progressChanged = true
				} else if job.Progress != nil && lastProg == nil {
					// Progress was added
					progressChanged = true
				} else if job.Progress != nil && lastProg != nil && *job.Progress != *lastProg {
					// Progress value changed
					progressChanged = true
				}
				
				if progressChanged {
					update := JobUpdate{
						ID:       job.ID,
						Progress: job.Progress,
						Status:   status,
					}

					data, err := json.Marshal(update)
					if err != nil {
						continue
					}

					// Send SSE message
					fmt.Fprintf(rw, "data: %s\n\n", data)
					if f, ok := rw.(http.Flusher); ok {
						f.Flush()
					}

					// Update tracking
					lastProgress[job.ID] = job.Progress
				}
			}
		}
	}
}