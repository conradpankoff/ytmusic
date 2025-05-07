package models

import (
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
	ChannelID         *int
	ChannelExternalID string
	Title             string

	MetadataUpdatedAt  *time.Time
	ThumbnailUpdatedAt *time.Time
}
