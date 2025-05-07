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
	ChannelID                  *int `sql:",table:playlist_search"`
	ChannelCreatedAt           *time.Time
	ChannelExternalID          string
	ChannelTitle               string
	ChannelMetadataUpdatedAt   *time.Time
	ChannelThumbnailUpdatedAt  *time.Time
	PlaylistID                 int
	PlaylistCreatedAt          time.Time
	PlaylistExternalID         string
	PlaylistTitle              string
	PlaylistMetadataUpdatedAt  *time.Time
	PlaylistThumbnailUpdatedAt *time.Time
}

func (s *PlaylistSearch) OverrideScan(names []string, scanners []sql.Scanner) error {
	for i, name := range names {
		switch name {
		case "ChannelCreatedAt":
			scanners[i] = &sqltypes.TimePointerScanner{Value: &s.ChannelCreatedAt}
		case "ChannelMetadataUpdatedAt":
			scanners[i] = &sqltypes.TimePointerScanner{Value: &s.ChannelMetadataUpdatedAt}
		case "ChannelThumbnailUpdatedAt":
			scanners[i] = &sqltypes.TimePointerScanner{Value: &s.ChannelThumbnailUpdatedAt}
		case "PlaylistCreatedAt":
			scanners[i] = &sqltypes.TimeScanner{Value: &s.PlaylistCreatedAt}
		case "PlaylistMetadataUpdatedAt":
			scanners[i] = &sqltypes.TimePointerScanner{Value: &s.PlaylistMetadataUpdatedAt}
		case "PlaylistThumbnailUpdatedAt":
			scanners[i] = &sqltypes.TimePointerScanner{Value: &s.PlaylistThumbnailUpdatedAt}
		}
	}

	return nil
}
