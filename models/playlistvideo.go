package models

import (
	"time"

	"fknsrs.biz/p/ytmusic/internal/sqlbuilderutil"
)

var (
	PlaylistVideoTable *sqlbuilderutil.Table
)

func init() {
	PlaylistVideoTable = sqlbuilderutil.MustMakeTable(PlaylistVideo{})
}

type PlaylistVideo struct {
	ID                 int `sql:",table:playlist_videos"`
	CreatedAt          time.Time
	PlaylistID         int
	PlaylistExternalID string
	VideoID            *int
	VideoExternalID    string
	Position           int
}
