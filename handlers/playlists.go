package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

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

func Playlists(rw http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")

	var condition sb.AsExpr
	order := []sb.AsOrderingTerm{sb.OrderDesc(models.PlaylistSearchTable.C("PlaylistCreatedAt"))}

	if q != "" {
		condition = sb.BinaryOperator("match", sb.Literal("playlist_search"), sb.Bind(q))
		order = []sb.AsOrderingTerm{sb.OrderDesc(sb.Literal("rank"))}
	}

	var playlists []models.PlaylistSearch
	if err := qsorm.FindWhere(
		r.Context(),
		ctxdb.GetDB(r.Context()),
		&playlists,
		condition,
		order,
		sb.OffsetLimit(nil, sb.Literal("1000")),
	); err != nil {
		panic(err)
	}

	if err := ctxtemplate.ExecuteTemplateIntoResponse(r, rw, "page_playlists", map[string]interface{}{
		"Q":         q,
		"Playlists": playlists,
	}); err != nil {
		panic(err)
	}
}

func Playlist(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var playlist models.PlaylistSearch
	if err := sorm.FindFirstWhere(r.Context(), ctxdb.GetDB(r.Context()), &playlist, "where playlist_id = ? or playlist_external_id = ?", vars["id"], vars["id"]); err != nil {
		if err == sql.ErrNoRows {
			httputil.NotFound(rw, r)
			return
		}

		panic(err)
	}

	var channel models.ChannelSearch
	if err := sorm.FindFirstWhere(r.Context(), ctxdb.GetDB(r.Context()), &channel, "where channel_id = ? or channel_external_id = ?", playlist.ChannelID, playlist.ChannelExternalID); err != nil {
		if err != sql.ErrNoRows {
			panic(err)
		}
	}

	var videos []models.VideoInPlaylist
	if err := sorm.FindWhere(r.Context(), ctxdb.GetDB(r.Context()), &videos, "where playlist_id = ? order by playlist_video_position asc", playlist.PlaylistID); err != nil {
		panic(err)
	}

	if err := ctxtemplate.ExecuteTemplateIntoResponse(r, rw, "page_playlist", map[string]interface{}{
		"Playlist": playlist,
		"Channel":  channel,
		"Videos":   videos,
	}); err != nil {
		panic(err)
	}
}

func PlaylistAudio(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var playlist models.PlaylistSearch
	if err := sorm.FindFirstWhere(r.Context(), ctxdb.GetDB(r.Context()), &playlist, "where playlist_id = ? or playlist_external_id = ?", vars["id"], vars["id"]); err != nil {
		if err == sql.ErrNoRows {
			httputil.NotFound(rw, r)
			return
		}

		panic(err)
	}

	var channel models.ChannelSearch
	if err := sorm.FindFirstWhere(r.Context(), ctxdb.GetDB(r.Context()), &channel, "where channel_id = ? or channel_external_id = ?", playlist.ChannelID, playlist.ChannelExternalID); err != nil {
		if err != sql.ErrNoRows {
			panic(err)
		}
	}

	var videos []models.VideoInPlaylist
	if err := sorm.FindWhere(r.Context(), ctxdb.GetDB(r.Context()), &videos, "where playlist_id = ? order by playlist_video_position asc", playlist.PlaylistID); err != nil {
		panic(err)
	}

	if err := ctxtemplate.ExecuteTemplateIntoResponse(r, rw, "page_playlist_audio", map[string]interface{}{
		"Playlist": playlist,
		"Channel":  channel,
		"Videos":   videos,
	}); err != nil {
		panic(err)
	}
}

func PlaylistAudioZip(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var playlist models.PlaylistSearch
	if err := sorm.FindFirstWhere(r.Context(), ctxdb.GetDB(r.Context()), &playlist, "where playlist_id = ? or playlist_external_id = ?", vars["id"], vars["id"]); err != nil {
		if err == sql.ErrNoRows {
			httputil.NotFound(rw, r)
			return
		}

		panic(err)
	}

	var videos []models.VideoInPlaylist
	if err := sorm.FindWhere(r.Context(), ctxdb.GetDB(r.Context()), &videos, "where playlist_id = ? order by playlist_video_position asc", playlist.PlaylistID); err != nil {
		panic(err)
	}

	rw.Header().Set("content-type", "application/x-zip")
	rw.Header().Set("content-disposition", "attachment;filename="+playlist.ChannelTitle+" - "+playlist.PlaylistTitle+".zip")
	rw.WriteHeader(http.StatusOK)

	if err := archiver.VideoInPlaylistZipAudio(r.Context(), rw, videos); err != nil {
		panic(err)
	}
}

func PlaylistVideo(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	index, err := strconv.ParseUint(vars["index"], 10, 64)
	if err != nil {
		panic(err)
	}

	var playlist models.PlaylistSearch
	if err := sorm.FindFirstWhere(r.Context(), ctxdb.GetDB(r.Context()), &playlist, "where playlist_id = ? or playlist_external_id = ?", vars["id"], vars["id"]); err != nil {
		if err == sql.ErrNoRows {
			httputil.NotFound(rw, r)
			return
		}

		panic(err)
	}

	var channel models.ChannelSearch
	if err := sorm.FindFirstWhere(r.Context(), ctxdb.GetDB(r.Context()), &channel, "where channel_id = ? or channel_external_id = ?", playlist.ChannelID, playlist.ChannelExternalID); err != nil {
		if err != sql.ErrNoRows {
			panic(err)
		}
	}

	var videos []models.VideoInPlaylist
	if err := sorm.FindWhere(r.Context(), ctxdb.GetDB(r.Context()), &videos, "where playlist_id = ? order by playlist_video_position asc", playlist.PlaylistID); err != nil {
		panic(err)
	}

	var video *models.VideoInPlaylist
	if int(index) <= len(videos) {
		video = &videos[index]
	}

	var nextVideo *models.VideoInPlaylist
	if int(index)+1 < len(videos) {
		nextVideo = &videos[int(index)+1]
	}

	if err := ctxtemplate.ExecuteTemplateIntoResponse(r, rw, "page_playlist_video", map[string]interface{}{
		"Playlist":  playlist,
		"Channel":   channel,
		"Videos":    videos,
		"Video":     video,
		"NextVideo": nextVideo,
	}); err != nil {
		panic(err)
	}
}
