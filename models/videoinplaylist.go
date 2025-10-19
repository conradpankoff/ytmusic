package models

import (
	"database/sql"
	"time"

	"fknsrs.biz/p/ytmusic/internal/sqlbuilderutil"
	"fknsrs.biz/p/ytmusic/internal/sqltypes"
)

var (
	VideoInPlaylistTable *sqlbuilderutil.Table
)

func init() {
	VideoInPlaylistTable = sqlbuilderutil.MustMakeTable(VideoInPlaylist{})
}

type VideoInPlaylist struct {
	ChannelID                  sql.NullInt32 `sql:",table:video_in_playlist_view"`
	ChannelCreatedAt           sql.NullTime
	ChannelExternalID          string
	ChannelTitle               string
	ChannelMetadataUpdatedAt   sql.NullTime
	ChannelThumbnailUpdatedAt  sql.NullTime
	PlaylistID                 int
	PlaylistCreatedAt          time.Time
	PlaylistExternalID         string
	PlaylistTitle              string
	PlaylistMetadataUpdatedAt  sql.NullTime
	PlaylistThumbnailUpdatedAt sql.NullTime
	PlaylistVideoID            int
	PlaylistVideoCreatedAt     time.Time
	PlaylistVideoPosition      int
	VideoID                    sql.NullInt32
	VideoCreatedAt             sql.NullTime
	VideoExternalID            string
	VideoTitle                 string
	VideoDescription           string
	VideoMetadataUpdatedAt     sql.NullTime
	VideoThumbnailUpdatedAt    sql.NullTime
	VideoDownloadedAt          sql.NullTime
	VideoTranscoded360At       sql.NullTime `sql:"video_transcoded_360_at"`
	VideoTranscoded720At       sql.NullTime `sql:"video_transcoded_720_at"`
	VideoAudioExtractedAt      sql.NullTime
}

func (s *VideoInPlaylist) OverrideScanx(names []string, scanners []sql.Scanner) error {
	for i, name := range names {
		switch name {
		case "ChannelCreatedAt":
			scanners[i] = &sqltypes.NullTimeScanner{Value: &s.ChannelCreatedAt}
		case "ChannelMetadataUpdatedAt":
			scanners[i] = &sqltypes.NullTimeScanner{Value: &s.ChannelMetadataUpdatedAt}
		case "ChannelThumbnailUpdatedAt":
			scanners[i] = &sqltypes.NullTimeScanner{Value: &s.ChannelThumbnailUpdatedAt}
		case "PlaylistVideoCreatedAt":
			scanners[i] = &sqltypes.TimeScanner{Value: &s.PlaylistVideoCreatedAt}
		case "PlaylistCreatedAt":
			scanners[i] = &sqltypes.TimeScanner{Value: &s.PlaylistCreatedAt}
		case "PlaylistMetadataUpdatedAt":
			scanners[i] = &sqltypes.NullTimeScanner{Value: &s.PlaylistMetadataUpdatedAt}
		case "PlaylistThumbnailUpdatedAt":
			scanners[i] = &sqltypes.NullTimeScanner{Value: &s.PlaylistThumbnailUpdatedAt}
		case "VideoCreatedAt":
			scanners[i] = &sqltypes.NullTimeScanner{Value: &s.VideoCreatedAt}
		case "VideoMetadataUpdatedAt":
			scanners[i] = &sqltypes.NullTimeScanner{Value: &s.VideoMetadataUpdatedAt}
		case "VideoThumbnailUpdatedAt":
			scanners[i] = &sqltypes.NullTimeScanner{Value: &s.VideoThumbnailUpdatedAt}
		case "VideoDownloadedAt":
			scanners[i] = &sqltypes.NullTimeScanner{Value: &s.VideoDownloadedAt}
		case "VideoTranscoded360At":
			scanners[i] = &sqltypes.NullTimeScanner{Value: &s.VideoTranscoded360At}
		case "VideoTranscoded720At":
			scanners[i] = &sqltypes.NullTimeScanner{Value: &s.VideoTranscoded720At}
		case "VideoAudioExtractedAt":
			scanners[i] = &sqltypes.NullTimeScanner{Value: &s.VideoAudioExtractedAt}
		}
	}

	return nil
}
