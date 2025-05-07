package handlers

import (
	"net/http"

	"fknsrs.biz/p/sorm"

	"fknsrs.biz/p/ytmusic/internal/ctxdb"
	"fknsrs.biz/p/ytmusic/internal/ctxtemplate"
	"fknsrs.biz/p/ytmusic/internal/jobqueue"
)

func Jobs(rw http.ResponseWriter, r *http.Request) {
	var jobs []jobqueue.Job
	if err := sorm.FindWhere(r.Context(), ctxdb.GetDB(r.Context()), &jobs, "where finished_at is null order by id desc limit 1500"); err != nil {
		panic(err)
	}

	if err := ctxtemplate.ExecuteTemplateIntoResponse(r, rw, "page_jobs", map[string]interface{}{
		"Jobs": jobs,
	}); err != nil {
		panic(err)
	}
}
