package models

import (
	"database/sql"
	"testing"
	"time"
	
	"fknsrs.biz/p/ytmusic/internal/ptr"
)

// Test that our model type changes work correctly
func TestModelTypeConversions(t *testing.T) {
	t.Run("Video model with new types", func(t *testing.T) {
		now := time.Now()
		
		video := Video{
			ID:                1,
			CreatedAt:         now,
			ExternalID:        "test-video",
			ChannelID:         ptr.NullInt32FromInt(42),
			ChannelExternalID: "test-channel",
			Title:             "Test Video",
			Description:       "Test Description",
			PublishDate:       ptr.NullTimeFromPtr(ptr.Time(now.AddDate(0, 0, -1))),
			UploadDate:        sql.NullTime{Valid: false},
			MetadataUpdatedAt: ptr.NullTime(now),
		}
		
		// Test ChannelID
		if !video.ChannelID.Valid {
			t.Error("ChannelID should be valid")
		}
		if video.ChannelID.Int32 != 42 {
			t.Errorf("ChannelID should be 42, got %d", video.ChannelID.Int32)
		}
		
		// Test PublishDate
		if !video.PublishDate.Valid {
			t.Error("PublishDate should be valid")
		}
		
		// Test UploadDate (null)
		if video.UploadDate.Valid {
			t.Error("UploadDate should be null/invalid")
		}
		
		// Test MetadataUpdatedAt
		if !video.MetadataUpdatedAt.Valid {
			t.Error("MetadataUpdatedAt should be valid")
		}
	})
	
	t.Run("Null conversions from nil pointers", func(t *testing.T) {
		var nilInt *int
		var nilTime *time.Time
		
		video := Video{}
		
		// Test nil int conversion
		video.ChannelID = ptr.NullInt32FromIntPtr(nilInt)
		if video.ChannelID.Valid {
			t.Error("ChannelID from nil pointer should be invalid")
		}
		
		// Test nil time conversion  
		video.PublishDate = ptr.NullTimeFromPtr(nilTime)
		if video.PublishDate.Valid {
			t.Error("PublishDate from nil pointer should be invalid")
		}
	})
	
	t.Run("Channel model with new types", func(t *testing.T) {
		now := time.Now()
		
		channel := Channel{
			ID:         1,
			CreatedAt:  now,
			ExternalID: "test-channel",
			Title:      "Test Channel",
			MetadataUpdatedAt:  ptr.NullTime(now),
			ThumbnailUpdatedAt: sql.NullTime{Valid: false},
			PlaylistsUpdatedAt: sql.NullTime{Valid: false},
			VideosUpdatedAt:    sql.NullTime{Valid: false},
		}
		
		if !channel.MetadataUpdatedAt.Valid {
			t.Error("MetadataUpdatedAt should be valid")
		}
		
		if channel.ThumbnailUpdatedAt.Valid {
			t.Error("ThumbnailUpdatedAt should be invalid")
		}
	})
	
	t.Run("PlaylistVideo model with new types", func(t *testing.T) {
		now := time.Now()
		
		playlistVideo := PlaylistVideo{
			ID:                 1,
			CreatedAt:          now,
			PlaylistID:         5,
			PlaylistExternalID: "test-playlist",
			VideoID:            ptr.NullInt32FromInt(10),
			VideoExternalID:    "test-video",
			Position:           0,
		}
		
		if !playlistVideo.VideoID.Valid {
			t.Error("VideoID should be valid")
		}
		
		if playlistVideo.VideoID.Int32 != 10 {
			t.Errorf("VideoID should be 10, got %d", playlistVideo.VideoID.Int32)
		}
		
		// Test setting to null
		playlistVideo.VideoID = sql.NullInt32{Valid: false}
		if playlistVideo.VideoID.Valid {
			t.Error("VideoID should be invalid after setting to null")
		}
	})
}