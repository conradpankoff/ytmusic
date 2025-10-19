package models

import (
	"database/sql"
	"time"

	"fknsrs.biz/p/ytmusic/internal/sqlbuilderutil"
	"fknsrs.biz/p/ytmusic/internal/sqltypes"
)

var (
	PlaylistSearchTable *sqlbuilderutil.Table
)

func init() {
	PlaylistSearchTable = sqlbuilderutil.MustMakeTable(PlaylistSearch{})
}

type PlaylistSearch struct {
	ChannelID                  sql.NullInt32 `sql:",table:playlist_search"`
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
}

func (s *PlaylistSearch) OverrideScan(names []string, scanners []sql.Scanner) error {
	for i, name := range names {
		switch name {
		case "ChannelCreatedAt":
			scanners[i] = &sqltypes.NullTimeScanner{Value: &s.ChannelCreatedAt}
		case "ChannelMetadataUpdatedAt":
			scanners[i] = &sqltypes.NullTimeScanner{Value: &s.ChannelMetadataUpdatedAt}
		case "ChannelThumbnailUpdatedAt":
			scanners[i] = &sqltypes.NullTimeScanner{Value: &s.ChannelThumbnailUpdatedAt}
		case "PlaylistCreatedAt":
			scanners[i] = &sqltypes.TimeScanner{Value: &s.PlaylistCreatedAt}
		case "PlaylistMetadataUpdatedAt":
			scanners[i] = &sqltypes.NullTimeScanner{Value: &s.PlaylistMetadataUpdatedAt}
		case "PlaylistThumbnailUpdatedAt":
			scanners[i] = &sqltypes.NullTimeScanner{Value: &s.PlaylistThumbnailUpdatedAt}
		}
	}

	return nil
}
