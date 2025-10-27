package api

import (
	"time"

	"fknsrs.biz/p/ytmusic/models"
)

// APIChannel represents a channel in API responses
type APIChannel struct {
	ID         int        `json:"id" xml:"ID"`
	CreatedAt  time.Time  `json:"created_at" xml:"CreatedAt"`
	ExternalID string     `json:"external_id" xml:"ExternalID"`
	Title      string     `json:"title" xml:"Title"`
	MetadataUpdatedAt  *time.Time `json:"metadata_updated_at,omitempty" xml:"MetadataUpdatedAt,omitempty"`
	ThumbnailUpdatedAt *time.Time `json:"thumbnail_updated_at,omitempty" xml:"ThumbnailUpdatedAt,omitempty"`
	PlaylistsUpdatedAt *time.Time `json:"playlists_updated_at,omitempty" xml:"PlaylistsUpdatedAt,omitempty"`
	VideosUpdatedAt    *time.Time `json:"videos_updated_at,omitempty" xml:"VideosUpdatedAt,omitempty"`
}

// APIPlaylist represents a playlist in API responses
type APIPlaylist struct {
	ID                int        `json:"id" xml:"ID"`
	CreatedAt         time.Time  `json:"created_at" xml:"CreatedAt"`
	ExternalID        string     `json:"external_id" xml:"ExternalID"`
	ChannelID         *int       `json:"channel_id,omitempty" xml:"ChannelID,omitempty"`
	ChannelExternalID string     `json:"channel_external_id" xml:"ChannelExternalID"`
	Title             string     `json:"title" xml:"Title"`
	MetadataUpdatedAt  *time.Time `json:"metadata_updated_at,omitempty" xml:"MetadataUpdatedAt,omitempty"`
	ThumbnailUpdatedAt *time.Time `json:"thumbnail_updated_at,omitempty" xml:"ThumbnailUpdatedAt,omitempty"`
}

// APIVideo represents a video in API responses
type APIVideo struct {
	ID                int        `json:"id" xml:"ID"`
	CreatedAt         time.Time  `json:"created_at" xml:"CreatedAt"`
	ExternalID        string     `json:"external_id" xml:"ExternalID"`
	ChannelID         *int       `json:"channel_id,omitempty" xml:"ChannelID,omitempty"`
	ChannelExternalID string     `json:"channel_external_id" xml:"ChannelExternalID"`
	Title             string     `json:"title" xml:"Title"`
	Description       string     `json:"description" xml:"Description"`
	PublishDate       *time.Time `json:"publish_date,omitempty" xml:"PublishDate,omitempty"`
	UploadDate        *time.Time `json:"upload_date,omitempty" xml:"UploadDate,omitempty"`
	MetadataUpdatedAt  *time.Time `json:"metadata_updated_at,omitempty" xml:"MetadataUpdatedAt,omitempty"`
	ThumbnailUpdatedAt *time.Time `json:"thumbnail_updated_at,omitempty" xml:"ThumbnailUpdatedAt,omitempty"`
	DownloadedAt       *time.Time `json:"downloaded_at,omitempty" xml:"DownloadedAt,omitempty"`
	Transcoded360At    *time.Time `json:"transcoded_360_at,omitempty" xml:"Transcoded360At,omitempty"`
	Transcoded720At    *time.Time `json:"transcoded_720_at,omitempty" xml:"Transcoded720At,omitempty"`
	AudioExtractedAt   *time.Time `json:"audio_extracted_at,omitempty" xml:"AudioExtractedAt,omitempty"`
}

// APIError represents an error response
type APIError struct {
	Error   string `json:"error" xml:"Error"`
	Message string `json:"message,omitempty" xml:"Message,omitempty"`
	Code    int    `json:"code,omitempty" xml:"Code,omitempty"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Data       interface{} `json:"data" xml:"Data"`
	Page       int         `json:"page" xml:"Page"`
	Limit      int         `json:"limit" xml:"Limit"`
	Total      int         `json:"total" xml:"Total"`
	HasMore    bool        `json:"has_more" xml:"HasMore"`
	NextPage   *int        `json:"next_page,omitempty" xml:"NextPage,omitempty"`
}

// Conversion functions
func ChannelToAPI(ch models.Channel) APIChannel {
	return APIChannel{
		ID:                 ch.ID,
		CreatedAt:          ch.CreatedAt,
		ExternalID:         ch.ExternalID,
		Title:              ch.Title,
		MetadataUpdatedAt:  ch.MetadataUpdatedAt,
		ThumbnailUpdatedAt: ch.ThumbnailUpdatedAt,
		PlaylistsUpdatedAt: ch.PlaylistsUpdatedAt,
		VideosUpdatedAt:    ch.VideosUpdatedAt,
	}
}

func PlaylistToAPI(pl models.Playlist) APIPlaylist {
	return APIPlaylist{
		ID:                 pl.ID,
		CreatedAt:          pl.CreatedAt,
		ExternalID:         pl.ExternalID,
		ChannelID:          pl.ChannelID,
		ChannelExternalID:  pl.ChannelExternalID,
		Title:              pl.Title,
		MetadataUpdatedAt:  pl.MetadataUpdatedAt,
		ThumbnailUpdatedAt: pl.ThumbnailUpdatedAt,
	}
}

func VideoToAPI(v models.Video) APIVideo {
	return APIVideo{
		ID:                 v.ID,
		CreatedAt:          v.CreatedAt,
		ExternalID:         v.ExternalID,
		ChannelID:          v.ChannelID,
		ChannelExternalID:  v.ChannelExternalID,
		Title:              v.Title,
		Description:        v.Description,
		PublishDate:        v.PublishDate,
		UploadDate:         v.UploadDate,
		MetadataUpdatedAt:  v.MetadataUpdatedAt,
		ThumbnailUpdatedAt: v.ThumbnailUpdatedAt,
		DownloadedAt:       v.DownloadedAt,
		Transcoded360At:    v.Transcoded360At,
		Transcoded720At:    v.Transcoded720At,
		AudioExtractedAt:   v.AudioExtractedAt,
	}
}