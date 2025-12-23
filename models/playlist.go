package models

import (
	"database/sql"
	"time"

	"fknsrs.biz/p/ytmusic/internal/sqlbuilderutil"
)

var (
	PlaylistTable *sqlbuilderutil.Table
)

func init() {
	PlaylistTable = sqlbuilderutil.MustMakeTable(Playlist{})
}

type Playlist struct {
	ID                int `sql:",table:playlists"`
	CreatedAt         time.Time
	ExternalID        string
	ChannelID         sql.NullInt32
	ChannelExternalID string
	Title             string

	MetadataUpdatedAt  sql.NullTime
	ThumbnailUpdatedAt sql.NullTime
}
