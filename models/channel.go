package models

import (
	"database/sql"
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

	MetadataUpdatedAt  sql.NullTime
	ThumbnailUpdatedAt sql.NullTime
	PlaylistsUpdatedAt sql.NullTime
	VideosUpdatedAt    sql.NullTime
}
