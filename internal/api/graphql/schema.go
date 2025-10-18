package graphql

import (
	"github.com/graphql-go/graphql"
	
	"fknsrs.biz/p/ytmusic/internal/api"
)

// Define GraphQL types
var channelType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Channel",
	Description: "A YouTube music channel",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.NewNonNull(graphql.Int),
			Description: "Channel ID",
		},
		"created_at": &graphql.Field{
			Type: graphql.NewNonNull(graphql.DateTime),
			Description: "Channel creation timestamp",
		},
		"external_id": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
			Description: "External YouTube channel ID",
		},
		"title": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
			Description: "Channel title",
		},
		"metadata_updated_at": &graphql.Field{
			Type: graphql.DateTime,
			Description: "Last metadata update timestamp",
		},
		"thumbnail_updated_at": &graphql.Field{
			Type: graphql.DateTime,
			Description: "Last thumbnail update timestamp",
		},
		"playlists_updated_at": &graphql.Field{
			Type: graphql.DateTime,
			Description: "Last playlists update timestamp",
		},
		"videos_updated_at": &graphql.Field{
			Type: graphql.DateTime,
			Description: "Last videos update timestamp",
		},
	},
})

var playlistType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Playlist",
	Description: "A YouTube music playlist",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.NewNonNull(graphql.Int),
			Description: "Playlist ID",
		},
		"created_at": &graphql.Field{
			Type: graphql.NewNonNull(graphql.DateTime),
			Description: "Playlist creation timestamp",
		},
		"external_id": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
			Description: "External YouTube playlist ID",
		},
		"channel_id": &graphql.Field{
			Type: graphql.Int,
			Description: "Associated channel ID",
		},
		"channel_external_id": &graphql.Field{
			Type: graphql.String,
			Description: "Associated channel external ID",
		},
		"title": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
			Description: "Playlist title",
		},
		"metadata_updated_at": &graphql.Field{
			Type: graphql.DateTime,
			Description: "Last metadata update timestamp",
		},
		"thumbnail_updated_at": &graphql.Field{
			Type: graphql.DateTime,
			Description: "Last thumbnail update timestamp",
		},
	},
})

var videoType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Video",
	Description: "A YouTube music video",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.NewNonNull(graphql.Int),
			Description: "Video ID",
		},
		"created_at": &graphql.Field{
			Type: graphql.NewNonNull(graphql.DateTime),
			Description: "Video creation timestamp",
		},
		"external_id": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
			Description: "External YouTube video ID",
		},
		"channel_id": &graphql.Field{
			Type: graphql.Int,
			Description: "Associated channel ID",
		},
		"channel_external_id": &graphql.Field{
			Type: graphql.String,
			Description: "Associated channel external ID",
		},
		"title": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
			Description: "Video title",
		},
		"description": &graphql.Field{
			Type: graphql.String,
			Description: "Video description",
		},
		"publish_date": &graphql.Field{
			Type: graphql.DateTime,
			Description: "Video publish date",
		},
		"upload_date": &graphql.Field{
			Type: graphql.DateTime,
			Description: "Video upload date",
		},
		"metadata_updated_at": &graphql.Field{
			Type: graphql.DateTime,
			Description: "Last metadata update timestamp",
		},
		"thumbnail_updated_at": &graphql.Field{
			Type: graphql.DateTime,
			Description: "Last thumbnail update timestamp",
		},
		"downloaded_at": &graphql.Field{
			Type: graphql.DateTime,
			Description: "Video download timestamp",
		},
		"transcoded_360_at": &graphql.Field{
			Type: graphql.DateTime,
			Description: "360p transcode timestamp",
		},
		"transcoded_720_at": &graphql.Field{
			Type: graphql.DateTime,
			Description: "720p transcode timestamp",
		},
		"audio_extracted_at": &graphql.Field{
			Type: graphql.DateTime,
			Description: "Audio extraction timestamp",
		},
	},
})

var paginationInfoType = graphql.NewObject(graphql.ObjectConfig{
	Name: "PaginationInfo",
	Description: "Pagination information",
	Fields: graphql.Fields{
		"page": &graphql.Field{
			Type: graphql.NewNonNull(graphql.Int),
			Description: "Current page number",
		},
		"limit": &graphql.Field{
			Type: graphql.NewNonNull(graphql.Int),
			Description: "Items per page",
		},
		"total": &graphql.Field{
			Type: graphql.NewNonNull(graphql.Int),
			Description: "Total number of items",
		},
		"has_more": &graphql.Field{
			Type: graphql.NewNonNull(graphql.Boolean),
			Description: "Whether there are more pages",
		},
		"next_page": &graphql.Field{
			Type: graphql.Int,
			Description: "Next page number",
		},
	},
})

var channelConnectionType = graphql.NewObject(graphql.ObjectConfig{
	Name: "ChannelConnection",
	Description: "A paginated list of channels",
	Fields: graphql.Fields{
		"data": &graphql.Field{
			Type: graphql.NewList(channelType),
			Description: "List of channels",
		},
		"pagination": &graphql.Field{
			Type: graphql.NewNonNull(paginationInfoType),
			Description: "Pagination information",
		},
	},
})

var playlistConnectionType = graphql.NewObject(graphql.ObjectConfig{
	Name: "PlaylistConnection",
	Description: "A paginated list of playlists",
	Fields: graphql.Fields{
		"data": &graphql.Field{
			Type: graphql.NewList(playlistType),
			Description: "List of playlists",
		},
		"pagination": &graphql.Field{
			Type: graphql.NewNonNull(paginationInfoType),
			Description: "Pagination information",
		},
	},
})

var videoConnectionType = graphql.NewObject(graphql.ObjectConfig{
	Name: "VideoConnection",
	Description: "A paginated list of videos",
	Fields: graphql.Fields{
		"data": &graphql.Field{
			Type: graphql.NewList(videoType),
			Description: "List of videos",
		},
		"pagination": &graphql.Field{
			Type: graphql.NewNonNull(paginationInfoType),
			Description: "Pagination information",
		},
	},
})

// CreateSchema creates the GraphQL schema
func CreateSchema(service *api.Service) (graphql.Schema, error) {
	rootQuery := graphql.NewObject(graphql.ObjectConfig{
		Name: "RootQuery",
		Description: "Root query for the YouTube Music API",
		Fields: graphql.Fields{
			"channel": &graphql.Field{
				Type: channelType,
				Description: "Get a single channel by ID",
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.Int),
						Description: "Channel ID",
					},
				},
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					id := params.Args["id"].(int)
					return service.GetChannel(params.Context, id)
				},
			},
			"channels": &graphql.Field{
				Type: channelConnectionType,
				Description: "Get a paginated list of channels",
				Args: graphql.FieldConfigArgument{
					"page": &graphql.ArgumentConfig{
						Type: graphql.Int,
						DefaultValue: 1,
						Description: "Page number (default: 1)",
					},
					"limit": &graphql.ArgumentConfig{
						Type: graphql.Int,
						DefaultValue: 50,
						Description: "Items per page (default: 50, max: 1000)",
					},
					"search": &graphql.ArgumentConfig{
						Type: graphql.String,
						Description: "Search query for channel titles",
					},
				},
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					page := params.Args["page"].(int)
					limit := params.Args["limit"].(int)
					search, hasSearch := params.Args["search"].(string)
					
					if limit > 1000 {
						limit = 1000
					}
					
					var channels []api.APIChannel
					var total int
					var err error
					
					searchQuery := ""
					if hasSearch && search != "" {
						searchQuery = search
					}
					
					channels, total, err = service.GetChannels(params.Context, page, limit, searchQuery)
					
					if err != nil {
						return nil, err
					}
					
					hasMore := page*limit < total
					var nextPage *int
					if hasMore {
						next := page + 1
						nextPage = &next
					}
					
					return map[string]interface{}{
						"data": channels,
						"pagination": map[string]interface{}{
							"page":      page,
							"limit":     limit,
							"total":     total,
							"has_more":  hasMore,
							"next_page": nextPage,
						},
					}, nil
				},
			},
			"playlist": &graphql.Field{
				Type: playlistType,
				Description: "Get a single playlist by ID",
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.Int),
						Description: "Playlist ID",
					},
				},
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					id := params.Args["id"].(int)
					return service.GetPlaylist(params.Context, id)
				},
			},
			"playlists": &graphql.Field{
				Type: playlistConnectionType,
				Description: "Get a paginated list of playlists",
				Args: graphql.FieldConfigArgument{
					"page": &graphql.ArgumentConfig{
						Type: graphql.Int,
						DefaultValue: 1,
						Description: "Page number (default: 1)",
					},
					"limit": &graphql.ArgumentConfig{
						Type: graphql.Int,
						DefaultValue: 50,
						Description: "Items per page (default: 50, max: 1000)",
					},
					"channel_id": &graphql.ArgumentConfig{
						Type: graphql.Int,
						Description: "Filter by channel ID",
					},
					"search": &graphql.ArgumentConfig{
						Type: graphql.String,
						Description: "Search query for playlist titles",
					},
				},
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					page := params.Args["page"].(int)
					limit := params.Args["limit"].(int)
					search, hasSearch := params.Args["search"].(string)
					
					if limit > 1000 {
						limit = 1000
					}
					
					var channelID *int
					if cid, ok := params.Args["channel_id"]; ok && cid != nil {
						id := cid.(int)
						channelID = &id
					}
					
					searchQuery := ""
					if hasSearch && search != "" {
						searchQuery = search
					}
					
					playlists, total, err := service.GetPlaylists(params.Context, page, limit, channelID, searchQuery)
					if err != nil {
						return nil, err
					}
					
					hasMore := page*limit < total
					var nextPage *int
					if hasMore {
						next := page + 1
						nextPage = &next
					}
					
					return map[string]interface{}{
						"data": playlists,
						"pagination": map[string]interface{}{
							"page":      page,
							"limit":     limit,
							"total":     total,
							"has_more":  hasMore,
							"next_page": nextPage,
						},
					}, nil
				},
			},
			"video": &graphql.Field{
				Type: videoType,
				Description: "Get a single video by ID",
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.Int),
						Description: "Video ID",
					},
				},
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					id := params.Args["id"].(int)
					return service.GetVideo(params.Context, id)
				},
			},
			"videos": &graphql.Field{
				Type: videoConnectionType,
				Description: "Get a paginated list of videos",
				Args: graphql.FieldConfigArgument{
					"page": &graphql.ArgumentConfig{
						Type: graphql.Int,
						DefaultValue: 1,
						Description: "Page number (default: 1)",
					},
					"limit": &graphql.ArgumentConfig{
						Type: graphql.Int,
						DefaultValue: 50,
						Description: "Items per page (default: 50, max: 1000)",
					},
					"channel_id": &graphql.ArgumentConfig{
						Type: graphql.Int,
						Description: "Filter by channel ID",
					},
					"playlist_id": &graphql.ArgumentConfig{
						Type: graphql.Int,
						Description: "Filter by playlist ID",
					},
					"search": &graphql.ArgumentConfig{
						Type: graphql.String,
						Description: "Search query for video titles and descriptions",
					},
				},
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					page := params.Args["page"].(int)
					limit := params.Args["limit"].(int)
					search, hasSearch := params.Args["search"].(string)
					
					if limit > 1000 {
						limit = 1000
					}
					
					var channelID *int
					if cid, ok := params.Args["channel_id"]; ok && cid != nil {
						id := cid.(int)
						channelID = &id
					}
					
					var playlistID *int
					if pid, ok := params.Args["playlist_id"]; ok && pid != nil {
						id := pid.(int)
						playlistID = &id
					}
					
					searchQuery := ""
					if hasSearch && search != "" {
						searchQuery = search
					}
					
					videos, total, err := service.GetVideos(params.Context, page, limit, channelID, playlistID, searchQuery)
					if err != nil {
						return nil, err
					}
					
					hasMore := page*limit < total
					var nextPage *int
					if hasMore {
						next := page + 1
						nextPage = &next
					}
					
					return map[string]interface{}{
						"data": videos,
						"pagination": map[string]interface{}{
							"page":      page,
							"limit":     limit,
							"total":     total,
							"has_more":  hasMore,
							"next_page": nextPage,
						},
					}, nil
				},
			},
		},
	})
	
	return graphql.NewSchema(graphql.SchemaConfig{
		Query: rootQuery,
	})
}