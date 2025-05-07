package models

import (
	"database/sql"
	"time"

	"fknsrs.biz/p/ytmusic/internal/sqlbuilderutil"
	"fknsrs.biz/p/ytmusic/internal/sqltypes"
)

var (
	ChannelSearchTable *sqlbuilderutil.Table
)

func init() {
	ChannelSearchTable = sqlbuilderutil.MustMakeTable(ChannelSearch{})
}

type ChannelSearch struct {
	ChannelID                 int `sql:",table:channel_search"`
	ChannelCreatedAt          time.Time
	ChannelExternalID         string
	ChannelTitle              string
	ChannelMetadataUpdatedAt  *time.Time
	ChannelThumbnailUpdatedAt *time.Time
}

func (s *ChannelSearch) OverrideScan(names []string, scanners []sql.Scanner) error {
	for i, name := range names {
		switch name {
		case "ChannelCreatedAt":
			scanners[i] = &sqltypes.TimeScanner{Value: &s.ChannelCreatedAt}
		case "ChannelMetadataUpdatedAt":
			scanners[i] = &sqltypes.TimePointerScanner{Value: &s.ChannelMetadataUpdatedAt}
		case "ChannelThumbnailUpdatedAt":
			scanners[i] = &sqltypes.TimePointerScanner{Value: &s.ChannelThumbnailUpdatedAt}
		}
	}

	return nil
}
