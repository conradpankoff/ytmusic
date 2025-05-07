package models

import (
	"time"

	"fknsrs.biz/p/ytmusic/internal/sqlbuilderutil"
)

var (
	ChannelTable *sqlbuilderutil.Table
)

func init() {
	ChannelTable = sqlbuilderutil.MustMakeTable(Channel{})
}

type Channel struct {
	ID         int `sql:",table:channels"`
	CreatedAt  time.Time
	ExternalID string
	Title      string

	MetadataUpdatedAt  *time.Time
	ThumbnailUpdatedAt *time.Time
	PlaylistsUpdatedAt *time.Time
	VideosUpdatedAt    *time.Time
}
