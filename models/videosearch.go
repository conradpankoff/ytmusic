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
	ChannelID                 *int `sql:",table:video_search"`
	ChannelCreatedAt          *time.Time
	ChannelExternalID         string
	ChannelTitle              string
	ChannelMetadataUpdatedAt  *time.Time
	ChannelThumbnailUpdatedAt *time.Time
	VideoID                   int
	VideoCreatedAt            time.Time
	VideoExternalID           string
	VideoTitle                string
	VideoDescription          string
	VideoMetadataUpdatedAt    *time.Time
	VideoThumbnailUpdatedAt   *time.Time
	VideoDownloadedAt         *time.Time
	VideoTranscoded360At      *time.Time `sql:"video_transcoded_360_at"`
	VideoTranscoded720At      *time.Time `sql:"video_transcoded_720_at"`
	VideoAudioExtractedAt     *time.Time
}

func (s *VideoSearch) OverrideScan(names []string, scanners []sql.Scanner) error {
	for i, name := range names {
		switch name {
		case "ChannelCreatedAt":
			scanners[i] = &sqltypes.TimePointerScanner{Value: &s.ChannelCreatedAt}
		case "ChannelMetadataUpdatedAt":
			scanners[i] = &sqltypes.TimePointerScanner{Value: &s.ChannelMetadataUpdatedAt}
		case "ChannelThumbnailUpdatedAt":
			scanners[i] = &sqltypes.TimePointerScanner{Value: &s.ChannelThumbnailUpdatedAt}
		case "VideoCreatedAt":
			scanners[i] = &sqltypes.TimeScanner{Value: &s.VideoCreatedAt}
		case "VideoMetadataUpdatedAt":
			scanners[i] = &sqltypes.TimePointerScanner{Value: &s.VideoMetadataUpdatedAt}
		case "VideoThumbnailUpdatedAt":
			scanners[i] = &sqltypes.TimePointerScanner{Value: &s.VideoThumbnailUpdatedAt}
		case "VideoDownloadedAt":
			scanners[i] = &sqltypes.TimePointerScanner{Value: &s.VideoDownloadedAt}
		case "VideoTranscoded360At":
			scanners[i] = &sqltypes.TimePointerScanner{Value: &s.VideoTranscoded360At}
		case "VideoTranscoded720At":
			scanners[i] = &sqltypes.TimePointerScanner{Value: &s.VideoTranscoded720At}
		case "VideoAudioExtractedAt":
			scanners[i] = &sqltypes.TimePointerScanner{Value: &s.VideoAudioExtractedAt}
		}
	}

	return nil
}
