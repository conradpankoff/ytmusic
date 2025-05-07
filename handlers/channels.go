package handlers

import (
	"database/sql"
	"net/http"

	"fknsrs.biz/p/sorm"
	"fknsrs.biz/p/sorm/qsorm"
	sb "fknsrs.biz/p/sqlbuilder"
	"github.com/gorilla/mux"

	"fknsrs.biz/p/ytmusic/internal/archiver"
	"fknsrs.biz/p/ytmusic/internal/ctxdb"
	"fknsrs.biz/p/ytmusic/internal/ctxtemplate"
	"fknsrs.biz/p/ytmusic/internal/httputil"
	"fknsrs.biz/p/ytmusic/models"
)

func Channels(rw http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")

	var condition sb.AsExpr
	order := []sb.AsOrderingTerm{sb.OrderDesc(models.ChannelSearchTable.C("ChannelCreatedAt"))}

	if q != "" {
		condition = sb.BinaryOperator("match", sb.Literal("channel_search"), sb.Bind(q))
		order = []sb.AsOrderingTerm{sb.OrderDesc(sb.Literal("rank"))}
	}

	var channels []models.ChannelSearch
	if err := qsorm.FindWhere(
		r.Context(),
		ctxdb.GetDB(r.Context()),
		&channels,
		condition,
		order,
		sb.OffsetLimit(nil, sb.Literal("1000")),
	); err != nil {
		panic(err)
	}

	if err := ctxtemplate.ExecuteTemplateIntoResponse(r, rw, "page_channels", map[string]interface{}{
		"Q":        q,
		"Channels": channels,
	}); err != nil {
		panic(err)
	}
}

func Channel(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var channel models.ChannelSearch
	if err := sorm.FindFirstWhere(r.Context(), ctxdb.GetDB(r.Context()), &channel, "where channel_id = ? or channel_external_id = ?", vars["id"], vars["id"]); err != nil {
		if err == sql.ErrNoRows {
			httputil.NotFound(rw, r)
			return
		}

		panic(err)
	}

	var playlists []models.PlaylistSearch
	if err := sorm.FindWhere(r.Context(), ctxdb.GetDB(r.Context()), &playlists, "where channel_id = ? order by playlist_id desc", channel.ChannelID); err != nil {
		panic(err)
	}

	var videos []models.VideoSearch
	if err := sorm.FindWhere(r.Context(), ctxdb.GetDB(r.Context()), &videos, "where channel_id = ? order by video_id desc", channel.ChannelID); err != nil {
		panic(err)
	}

	if err := ctxtemplate.ExecuteTemplateIntoResponse(r, rw, "page_channel", map[string]interface{}{
		"Channel":   channel,
		"Playlists": playlists,
		"Videos":    videos,
	}); err != nil {
		panic(err)
	}
}

func ChannelAudio(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var channel models.ChannelSearch
	if err := sorm.FindFirstWhere(r.Context(), ctxdb.GetDB(r.Context()), &channel, "where channel_id = ? or channel_external_id = ?", vars["id"], vars["id"]); err != nil {
		if err == sql.ErrNoRows {
			httputil.NotFound(rw, r)
			return
		}

		panic(err)
	}

	var videos []models.VideoSearch
	if err := sorm.FindWhere(r.Context(), ctxdb.GetDB(r.Context()), &videos, "where channel_id = ? order by video_id desc", channel.ChannelID); err != nil {
		panic(err)
	}

	if err := ctxtemplate.ExecuteTemplateIntoResponse(r, rw, "page_channel_audio", map[string]interface{}{
		"Channel": channel,
		"Videos":  videos,
	}); err != nil {
		panic(err)
	}
}

func ChannelAudioZip(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var channel models.ChannelSearch
	if err := sorm.FindFirstWhere(r.Context(), ctxdb.GetDB(r.Context()), &channel, "where channel_id = ? or channel_external_id = ?", vars["id"], vars["id"]); err != nil {
		if err == sql.ErrNoRows {
			httputil.NotFound(rw, r)
			return
		}

		panic(err)
	}

	var videos []models.VideoSearch
	if err := sorm.FindWhere(r.Context(), ctxdb.GetDB(r.Context()), &videos, "where channel_id = ? order by video_id desc", channel.ChannelID); err != nil {
		panic(err)
	}

	rw.Header().Set("content-type", "application/x-zip")
	rw.Header().Set("content-disposition", "attachment;filename="+channel.ChannelTitle+".zip")
	rw.WriteHeader(http.StatusOK)

	if err := archiver.VideoSearchZipAudio(r.Context(), rw, videos); err != nil {
		panic(err)
	}
}
