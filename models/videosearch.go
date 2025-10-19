package models

import (
	"database/sql"
	"time"

	"fknsrs.biz/p/ytmusic/internal/sqlbuilderutil"
	"fknsrs.biz/p/ytmusic/internal/sqltypes"
)

var (
	VideoSearchTable *sqlbuilderutil.Table
)

func init() {
	VideoSearchTable = sqlbuilderutil.MustMakeTable(VideoSearch{})
}

type VideoSearch struct {
	ChannelID                 sql.NullInt32 `sql:",table:video_search"`
	ChannelCreatedAt          sql.NullTime
	ChannelExternalID         string
	ChannelTitle              string
	ChannelMetadataUpdatedAt  sql.NullTime
	ChannelThumbnailUpdatedAt sql.NullTime
	VideoID                   int
	VideoCreatedAt            time.Time
	VideoExternalID           string
	VideoTitle                string
	VideoDescription          string
	VideoMetadataUpdatedAt    sql.NullTime
	VideoThumbnailUpdatedAt   sql.NullTime
	VideoDownloadedAt         sql.NullTime
	VideoTranscoded360At      sql.NullTime `sql:"video_transcoded_360_at"`
	VideoTranscoded720At      sql.NullTime `sql:"video_transcoded_720_at"`
	VideoAudioExtractedAt     sql.NullTime
}

func (s *VideoSearch) OverrideScan(names []string, scanners []sql.Scanner) error {
	for i, name := range names {
		switch name {
		case "ChannelCreatedAt":
			scanners[i] = &sqltypes.NullTimeScanner{Value: &s.ChannelCreatedAt}
		case "ChannelMetadataUpdatedAt":
			scanners[i] = &sqltypes.NullTimeScanner{Value: &s.ChannelMetadataUpdatedAt}
		case "ChannelThumbnailUpdatedAt":
			scanners[i] = &sqltypes.NullTimeScanner{Value: &s.ChannelThumbnailUpdatedAt}
		case "VideoCreatedAt":
			scanners[i] = &sqltypes.TimeScanner{Value: &s.VideoCreatedAt}
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
