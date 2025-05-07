package models

import (
	"time"

	"fknsrs.biz/p/ytmusic/internal/sqlbuilderutil"
)

var (
	VideoTable *sqlbuilderutil.Table
)

func init() {
	VideoTable = sqlbuilderutil.MustMakeTable(Video{})
}

type Video struct {
	ID                int `sql:",table:videos"`
	CreatedAt         time.Time
	ExternalID        string
	ChannelID         *int
	ChannelExternalID string
	Title             string
	Description       string
	PublishDate       *time.Time
	UploadDate        *time.Time

	MetadataUpdatedAt  *time.Time
	ThumbnailUpdatedAt *time.Time
	DownloadedAt       *time.Time
	Transcoded360At    *time.Time `sql:"transcoded_360_at"`
	Transcoded720At    *time.Time `sql:"transcoded_720_at"`
	AudioExtractedAt   *time.Time
}
