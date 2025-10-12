package handlers

import (
	"fmt"
	"net/http"

	"fknsrs.biz/p/sorm"

	"fknsrs.biz/p/ytmusic/internal/ctxdb"
	"fknsrs.biz/p/ytmusic/internal/ctxtemplate"
	"fknsrs.biz/p/ytmusic/internal/jobqueue"
)

func Jobs(rw http.ResponseWriter, r *http.Request) {
	show := r.URL.Query().Get("show")
	if show == "" {
		show = "running"
	}

	where := ""

	switch show {
	case "all":
		where = ""
	case "running":
		where = "where finished_at is null"
	default:
		panic(fmt.Errorf("invalid value for `show' parameter"))
	}

	var jobs []jobqueue.Job
	if err := sorm.FindWhere(r.Context(), ctxdb.GetDB(r.Context()), &jobs, where+" order by id desc limit 1500"); err != nil {
		panic(err)
	}

	if err := ctxtemplate.ExecuteTemplateIntoResponse(r, rw, "page_jobs", map[string]interface{}{
		"TemplateName": "page_jobs",
		"Jobs":         jobs,
		"Show":         show,
		"Details":      r.URL.Query().Get("details") != "",
	}); err != nil {
		panic(err)
	}
}
