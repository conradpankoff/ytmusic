package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/monoculum/formam"

	"fknsrs.biz/p/ytmusic/internal/ctxdb"
	"fknsrs.biz/p/ytmusic/internal/ctxjobqueue"
	"fknsrs.biz/p/ytmusic/internal/ctxtemplate"
	"fknsrs.biz/p/ytmusic/internal/httputil"
	"fknsrs.biz/p/ytmusic/internal/jobqueue"
	"fknsrs.biz/p/ytmusic/internal/queuenames"
	"fknsrs.biz/p/ytmusic/internal/ytutil"
)

func Add(rw http.ResponseWriter, r *http.Request) {
	if err := ctxtemplate.ExecuteTemplateIntoResponse(r, rw, "page_add", map[string]interface{}{}); err != nil {
		panic(err)
	}
}

func AddAction(rw http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		panic(err)
	}

	var input struct {
		URLsOrIDs string `formam:"urls_or_ids"`
	}

	if err := formam.Decode(r.PostForm, &input); err != nil {
		panic(err)
	}

	ids, err := ytutil.ExtractAndIdentifyIDs(input.URLsOrIDs, false)
	if err != nil {
		httputil.RedirectWithError(rw, r, "/", "Could not extract IDs from input: "+err.Error())
		return
	}

	if len(ids) == 0 {
		httputil.RedirectWithError(rw, r, "/", "No IDs found in input")
		return
	}

	if err := ctxdb.UsingTx(r.Context(), nil, func(ctx context.Context, tx *sql.Tx) error {
		for _, id := range ids {
			var queueName string

			switch id.Type {
			case ytutil.ChannelID:
				queueName = queuenames.ChannelUpdateMetadata
			case ytutil.PlaylistID:
				queueName = queuenames.PlaylistUpdateMetadata
			case ytutil.VideoID:
				queueName = queuenames.VideoUpdateMetadata
			default:
				return fmt.Errorf("could not determine queue name for id type %s", id.Type)
			}

			if err := ctxjobqueue.Add(ctx, tx, &jobqueue.Job{
				QueueName: queueName,
				Payload:   id.Value,
			}); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		panic(err)
	}

	httputil.RedirectWithSuccess(rw, r, "/add", fmt.Sprintf("%d items will be added or updated soon.", len(ids)))
}
