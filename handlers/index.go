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

	var channelsCondition sb.AsExpr
	channelsOrder := []sb.AsOrderingTerm{sb.OrderDesc(models.ChannelSearchTable.C("ChannelCreatedAt"))}
	var playlistsCondition sb.AsExpr
	playlistsOrder := []sb.AsOrderingTerm{sb.OrderDesc(models.PlaylistSearchTable.C("PlaylistCreatedAt"))}
	var videosCondition sb.AsExpr
	videosOrder := []sb.AsOrderingTerm{sb.OrderDesc(models.VideoSearchTable.C("VideoCreatedAt"))}

	if q != "" {
		channelsCondition = sb.BinaryOperator("match", sb.Literal("channel_search"), sb.Bind(q))
		channelsOrder = []sb.AsOrderingTerm{sb.OrderDesc(sb.Literal("rank"))}
		playlistsCondition = sb.BinaryOperator("match", sb.Literal("playlist_search"), sb.Bind(q))
		playlistsOrder = []sb.AsOrderingTerm{sb.OrderDesc(sb.Literal("rank"))}
		videosCondition = sb.BinaryOperator("match", sb.Literal("video_search"), sb.Bind(q))
		videosOrder = []sb.AsOrderingTerm{sb.OrderDesc(sb.Literal("rank"))}
	}

	var channels []models.ChannelSearch
	if err := qsorm.FindWhere(
		r.Context(),
		ctxdb.GetDB(r.Context()),
		&channels,
		channelsCondition,
		channelsOrder,
		sb.OffsetLimit(nil, sb.Literal("50")),
	); err != nil {
		panic(err)
	}

	var playlists []models.PlaylistSearch
	if err := qsorm.FindWhere(
		r.Context(),
		ctxdb.GetDB(r.Context()),
		&playlists,
		playlistsCondition,
		playlistsOrder,
		sb.OffsetLimit(nil, sb.Literal("50")),
	); err != nil {
		panic(err)
	}

	var videos []models.VideoSearch
	if err := qsorm.FindWhere(
		r.Context(),
		ctxdb.GetDB(r.Context()),
		&videos,
		videosCondition,
		videosOrder,
		sb.OffsetLimit(nil, sb.Literal("1000")),
	); err != nil {
		panic(err)
	}

	if err := ctxtemplate.ExecuteTemplateIntoResponse(r, rw, "page_index", map[string]interface{}{
		"Q":         q,
		"Channels":  channels,
		"Playlists": playlists,
		"Videos":    videos,
	}); err != nil {
		panic(err)
	}
}
