package queuenames

const (
	ChannelUpdateMetadata  = "channel_update_metadata"
	ChannelUpdatePlaylists = "channel_update_playlists"
	ChannelUpdateVideos    = "channel_update_videos"
	PlaylistUpdateMetadata = "playlist_update_metadata"
	PlaylistUpdateVideos   = "playlist_update_videos"
	VideoUpdateMetadata    = "video_update_metadata"
	VideoDownload          = "video_download"
	VideoUpdateThumbnail   = "video_update_thumbnail"
	VideoTranscode         = "video_transcode"
	VideoExtractAudio      = "video_extract_audio"
)

var Priority = []string{
	VideoUpdateMetadata,
	ChannelUpdateMetadata,
	PlaylistUpdateMetadata,
	PlaylistUpdateVideos,
	ChannelUpdatePlaylists,
	ChannelUpdateVideos,
	VideoDownload,
	VideoUpdateThumbnail,
	VideoExtractAudio,
	VideoTranscode,
}
