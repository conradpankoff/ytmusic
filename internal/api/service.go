package api

import (
	"context"
	"database/sql"
	"fmt"

	"fknsrs.biz/p/sorm"
	sb "fknsrs.biz/p/sqlbuilder"
	"fknsrs.biz/p/sorm/qsorm"
	
	"fknsrs.biz/p/ytmusic/models"
)

// Service provides the core API functionality
type Service struct {
	db *sql.DB
}

// NewService creates a new API service
func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

// GetChannels returns a paginated list of channels
func (s *Service) GetChannels(ctx context.Context, page, limit int) ([]APIChannel, int, error) {
	offset := (page - 1) * limit
	
	var channels []models.Channel
	err := qsorm.FindWhere(
		ctx,
		s.db,
		&channels,
		nil, // no condition - get all
		[]sb.AsOrderingTerm{sb.OrderDesc(models.ChannelTable.C("CreatedAt"))},
		sb.OffsetLimit(sb.Literal(fmt.Sprintf("%d", offset)), sb.Literal(fmt.Sprintf("%d", limit))),
	)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get channels: %w", err)
	}
	
	// Get total count
	var total int
	err = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM channels").Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get channel count: %w", err)
	}
	
	// Convert to API models
	apiChannels := make([]APIChannel, len(channels))
	for i, ch := range channels {
		apiChannels[i] = ChannelToAPI(ch)
	}
	
	return apiChannels, total, nil
}

// GetChannel returns a single channel by ID
func (s *Service) GetChannel(ctx context.Context, id int) (*APIChannel, error) {
	var channel models.Channel
	err := sorm.FindFirstWhere(ctx, s.db, &channel, "where id = ?", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("channel not found")
		}
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}
	
	apiChannel := ChannelToAPI(channel)
	return &apiChannel, nil
}

// GetPlaylists returns a paginated list of playlists
func (s *Service) GetPlaylists(ctx context.Context, page, limit int, channelID *int) ([]APIPlaylist, int, error) {
	offset := (page - 1) * limit
	
	var condition sb.AsExpr
	if channelID != nil {
		condition = sb.BinaryOperator("=", models.PlaylistTable.C("ChannelID"), sb.Bind(*channelID))
	}
	
	var playlists []models.Playlist
	err := qsorm.FindWhere(
		ctx,
		s.db,
		&playlists,
		condition,
		[]sb.AsOrderingTerm{sb.OrderDesc(models.PlaylistTable.C("CreatedAt"))},
		sb.OffsetLimit(sb.Literal(fmt.Sprintf("%d", offset)), sb.Literal(fmt.Sprintf("%d", limit))),
	)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get playlists: %w", err)
	}
	
	// Get total count
	var total int
	var countQuery string
	if channelID != nil {
		countQuery = "SELECT COUNT(*) FROM playlists WHERE channel_id = ?"
		err = s.db.QueryRowContext(ctx, countQuery, *channelID).Scan(&total)
	} else {
		countQuery = "SELECT COUNT(*) FROM playlists"
		err = s.db.QueryRowContext(ctx, countQuery).Scan(&total)
	}
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get playlist count: %w", err)
	}
	
	// Convert to API models
	apiPlaylists := make([]APIPlaylist, len(playlists))
	for i, pl := range playlists {
		apiPlaylists[i] = PlaylistToAPI(pl)
	}
	
	return apiPlaylists, total, nil
}

// GetPlaylist returns a single playlist by ID
func (s *Service) GetPlaylist(ctx context.Context, id int) (*APIPlaylist, error) {
	var playlist models.Playlist
	err := sorm.FindFirstWhere(ctx, s.db, &playlist, "where id = ?", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("playlist not found")
		}
		return nil, fmt.Errorf("failed to get playlist: %w", err)
	}
	
	apiPlaylist := PlaylistToAPI(playlist)
	return &apiPlaylist, nil
}

// GetVideos returns a paginated list of videos
func (s *Service) GetVideos(ctx context.Context, page, limit int, channelID, playlistID *int) ([]APIVideo, int, error) {
	offset := (page - 1) * limit
	
	// Build query based on filters
	var query string
	var countQuery string
	var args []interface{}
	
	if channelID != nil {
		query = "SELECT id, created_at, external_id, channel_id, channel_external_id, title, description, publish_date, upload_date, metadata_updated_at, thumbnail_updated_at, downloaded_at, transcoded_360_at, transcoded_720_at, audio_extracted_at FROM videos WHERE channel_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?"
		countQuery = "SELECT COUNT(*) FROM videos WHERE channel_id = ?"
		args = []interface{}{*channelID, limit, offset}
	} else if playlistID != nil {
		query = "SELECT v.id, v.created_at, v.external_id, v.channel_id, v.channel_external_id, v.title, v.description, v.publish_date, v.upload_date, v.metadata_updated_at, v.thumbnail_updated_at, v.downloaded_at, v.transcoded_360_at, v.transcoded_720_at, v.audio_extracted_at FROM videos v JOIN playlist_videos pv ON v.id = pv.video_id WHERE pv.playlist_id = ? ORDER BY v.created_at DESC LIMIT ? OFFSET ?"
		countQuery = "SELECT COUNT(*) FROM videos v JOIN playlist_videos pv ON v.id = pv.video_id WHERE pv.playlist_id = ?"
		args = []interface{}{*playlistID, limit, offset}
	} else {
		query = "SELECT id, created_at, external_id, channel_id, channel_external_id, title, description, publish_date, upload_date, metadata_updated_at, thumbnail_updated_at, downloaded_at, transcoded_360_at, transcoded_720_at, audio_extracted_at FROM videos ORDER BY created_at DESC LIMIT ? OFFSET ?"
		countQuery = "SELECT COUNT(*) FROM videos"
		args = []interface{}{limit, offset}
	}
	
	// Execute query
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get videos: %w", err)
	}
	defer rows.Close()
	
	var videos []models.Video
	for rows.Next() {
		var v models.Video
		var channelExternalID sql.NullString
		var description sql.NullString
		
		err := rows.Scan(&v.ID, &v.CreatedAt, &v.ExternalID, &v.ChannelID, &channelExternalID, 
			&v.Title, &description, &v.PublishDate, &v.UploadDate, &v.MetadataUpdatedAt,
			&v.ThumbnailUpdatedAt, &v.DownloadedAt, &v.Transcoded360At, &v.Transcoded720At, &v.AudioExtractedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan video: %w", err)
		}
		
		if channelExternalID.Valid {
			v.ChannelExternalID = channelExternalID.String
		}
		if description.Valid {
			v.Description = description.String
		}
		
		videos = append(videos, v)
	}
	
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error reading video rows: %w", err)
	}
	
	// Get total count
	var total int
	if playlistID != nil {
		err = s.db.QueryRowContext(ctx, countQuery, *playlistID).Scan(&total)
	} else if channelID != nil {
		err = s.db.QueryRowContext(ctx, countQuery, *channelID).Scan(&total)
	} else {
		err = s.db.QueryRowContext(ctx, countQuery).Scan(&total)
	}
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get video count: %w", err)
	}
	
	// Convert to API models
	apiVideos := make([]APIVideo, len(videos))
	for i, v := range videos {
		apiVideos[i] = VideoToAPI(v)
	}
	
	return apiVideos, total, nil
}

// GetVideo returns a single video by ID
func (s *Service) GetVideo(ctx context.Context, id int) (*APIVideo, error) {
	var video models.Video
	err := sorm.FindFirstWhere(ctx, s.db, &video, "where id = ?", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("video not found")
		}
		return nil, fmt.Errorf("failed to get video: %w", err)
	}
	
	apiVideo := VideoToAPI(video)
	return &apiVideo, nil
}

// SearchChannels searches channels by title
func (s *Service) SearchChannels(ctx context.Context, query string, page, limit int) ([]APIChannel, int, error) {
	offset := (page - 1) * limit
	
	condition := sb.BinaryOperator("LIKE", models.ChannelTable.C("Title"), sb.Bind("%"+query+"%"))
	
	var channels []models.Channel
	err := qsorm.FindWhere(
		ctx,
		s.db,
		&channels,
		condition,
		[]sb.AsOrderingTerm{sb.OrderDesc(models.ChannelTable.C("CreatedAt"))},
		sb.OffsetLimit(sb.Literal(fmt.Sprintf("%d", offset)), sb.Literal(fmt.Sprintf("%d", limit))),
	)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search channels: %w", err)
	}
	
	// Get total count
	var total int
	err = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM channels WHERE title LIKE ?", "%"+query+"%").Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get search channel count: %w", err)
	}
	
	// Convert to API models
	apiChannels := make([]APIChannel, len(channels))
	for i, ch := range channels {
		apiChannels[i] = ChannelToAPI(ch)
	}
	
	return apiChannels, total, nil
}