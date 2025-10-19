package models

import (
	"database/sql"
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
	ChannelID         sql.NullInt32
	ChannelExternalID string
	Title             string
	Description       string
	PublishDate       sql.NullTime
	UploadDate        sql.NullTime

	MetadataUpdatedAt  sql.NullTime
	ThumbnailUpdatedAt sql.NullTime
	DownloadedAt       sql.NullTime
	Transcoded360At    sql.NullTime `sql:"transcoded_360_at"`
	Transcoded720At    sql.NullTime `sql:"transcoded_720_at"`
	AudioExtractedAt   sql.NullTime
}
