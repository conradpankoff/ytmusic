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

func Videos(rw http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")

	var condition sb.AsExpr
	order := []sb.AsOrderingTerm{sb.OrderDesc(models.VideoSearchTable.C("VideoCreatedAt"))}

	if q != "" {
		condition = sb.BinaryOperator("match", sb.Literal("video_search"), sb.Bind(q))
		order = []sb.AsOrderingTerm{sb.OrderDesc(sb.Literal("rank"))}
	}

	var videos []models.VideoSearch
	if err := qsorm.FindWhere(
		r.Context(),
		ctxdb.GetDB(r.Context()),
		&videos,
		condition,
		order,
		sb.OffsetLimit(nil, sb.Literal("1000")),
	); err != nil {
		panic(err)
	}

	if err := ctxtemplate.ExecuteTemplateIntoResponse(r, rw, "page_videos", map[string]interface{}{
		"Q":      q,
		"Videos": videos,
	}); err != nil {
		panic(err)
	}
}

func VideosAudio(rw http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")

	var condition sb.AsExpr
	order := []sb.AsOrderingTerm{sb.OrderDesc(models.VideoSearchTable.C("VideoCreatedAt"))}

	if q != "" {
		condition = sb.BinaryOperator("match", sb.Literal("video_search"), sb.Bind(q))
		order = []sb.AsOrderingTerm{sb.OrderDesc(sb.Literal("rank"))}
	}

	var videos []models.VideoSearch
	if err := qsorm.FindWhere(
		r.Context(),
		ctxdb.GetDB(r.Context()),
		&videos,
		condition,
		order,
		sb.OffsetLimit(nil, sb.Literal("100")),
	); err != nil {
		panic(err)
	}

	if err := ctxtemplate.ExecuteTemplateIntoResponse(r, rw, "page_videos_audio", map[string]interface{}{
		"Q":      q,
		"Videos": videos,
	}); err != nil {
		panic(err)
	}
}

func VideosAudioZip(rw http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")

	var condition sb.AsExpr
	order := []sb.AsOrderingTerm{sb.OrderDesc(models.VideoSearchTable.C("VideoCreatedAt"))}

	if q != "" {
		condition = sb.BinaryOperator("match", sb.Literal("video_search"), sb.Bind(q))
		order = []sb.AsOrderingTerm{sb.OrderDesc(sb.Literal("rank"))}
	}

	var videos []models.VideoSearch
	if err := qsorm.FindWhere(
		r.Context(),
		ctxdb.GetDB(r.Context()),
		&videos,
		condition,
		order,
		sb.OffsetLimit(nil, sb.Literal("100")),
	); err != nil {
		panic(err)
	}

	rw.Header().Set("content-type", "application/x-zip")
	rw.Header().Set("content-disposition", "attachment;filename=Audio.zip")
	rw.WriteHeader(http.StatusOK)

	if err := archiver.VideoSearchZipAudio(r.Context(), rw, videos); err != nil {
		panic(err)
	}
}

func Video(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var video models.VideoSearch
	if err := sorm.FindFirstWhere(r.Context(), ctxdb.GetDB(r.Context()), &video, "where video_id = ? or video_external_id = ?", vars["id"], vars["id"]); err != nil {
		if err == sql.ErrNoRows {
			httputil.NotFound(rw, r)
			return
		}

		panic(err)
	}

	var channel models.ChannelSearch
	if err := sorm.FindFirstWhere(r.Context(), ctxdb.GetDB(r.Context()), &channel, "where channel_id = ? or channel_external_id = ?", video.ChannelID, video.ChannelExternalID); err != nil {
		if err != sql.ErrNoRows {
			panic(err)
		}
	}

	var videoInPlaylists []models.VideoInPlaylist
	if err := sorm.FindWhere(r.Context(), ctxdb.GetDB(r.Context()), &videoInPlaylists, "where video_id = ? or video_external_id = ?", video.VideoID, video.VideoExternalID); err != nil {
		if err != sql.ErrNoRows {
			panic(err)
		}
	}

	if err := ctxtemplate.ExecuteTemplateIntoResponse(r, rw, "page_video", map[string]interface{}{
		"Video":            video,
		"Channel":          channel,
		"VideoInPlaylists": videoInPlaylists,
	}); err != nil {
		panic(err)
	}
}
