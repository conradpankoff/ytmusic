package handlers

import (
	"net/http"

	"fknsrs.biz/p/sorm/qsorm"
	sb "fknsrs.biz/p/sqlbuilder"

	"fknsrs.biz/p/ytmusic/internal/ctxdb"
	"fknsrs.biz/p/ytmusic/internal/ctxtemplate"
	"fknsrs.biz/p/ytmusic/models"
)

func Index(rw http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")

	// Temporarily disable search functionality due to FTS5 compatibility issues
	// TODO: Restore FTS5 search with compatible SQLite library
	
	var channels []models.ChannelSearch
	if err := qsorm.FindWhere(
		r.Context(),
		ctxdb.GetDB(r.Context()),
		&channels,
		nil, // No search condition for now
		[]sb.AsOrderingTerm{sb.OrderDesc(models.ChannelSearchTable.C("ChannelCreatedAt"))},
		sb.OffsetLimit(nil, sb.Literal("50")),
	); err != nil {
		// If ChannelSearch table doesn't exist, create empty slice
		channels = []models.ChannelSearch{}
	}

	var playlists []models.PlaylistSearch
	if err := qsorm.FindWhere(
		r.Context(),
		ctxdb.GetDB(r.Context()),
		&playlists,
		nil, // No search condition for now
		[]sb.AsOrderingTerm{sb.OrderDesc(models.PlaylistSearchTable.C("PlaylistCreatedAt"))},
		sb.OffsetLimit(nil, sb.Literal("50")),
	); err != nil {
		// If PlaylistSearch table doesn't exist, create empty slice
		playlists = []models.PlaylistSearch{}
	}

	var videos []models.VideoSearch
	if err := qsorm.FindWhere(
		r.Context(),
		ctxdb.GetDB(r.Context()),
		&videos,
		nil, // No search condition for now
		[]sb.AsOrderingTerm{sb.OrderDesc(models.VideoSearchTable.C("VideoCreatedAt"))},
		sb.OffsetLimit(nil, sb.Literal("1000")),
	); err != nil {
		// If VideoSearch table doesn't exist, create empty slice
		videos = []models.VideoSearch{}
	}

	if err := ctxtemplate.ExecuteTemplateIntoResponse(r, rw, "page_index", map[string]interface{}{
		"TemplateName": "page_index",
		"Q":            q,
		"Channels":     channels,
		"Playlists":    playlists,
		"Videos":       videos,
	}); err != nil {
		panic(err)
	}
}
