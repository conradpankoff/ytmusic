package rest

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	
	"fknsrs.biz/p/ytmusic/internal/api"
)

// Handler wraps the API service for REST endpoints
type Handler struct {
	service *api.Service
}

// NewHandler creates a new REST handler
func NewHandler(service *api.Service) *Handler {
	return &Handler{service: service}
}

// GetChannels handles GET /api/rest/channels
func (h *Handler) GetChannels(w http.ResponseWriter, r *http.Request) {
	page, limit, err := api.GetPaginationParams(r)
	if err != nil {
		api.WriteErrorResponse(w, http.StatusBadRequest, "Invalid pagination parameters", err.Error())
		return
	}
	
	// Check for search query
	query := r.URL.Query().Get("q")
	var channels []api.APIChannel
	var total int
	
	if query != "" {
		channels, total, err = h.service.SearchChannels(r.Context(), query, page, limit)
	} else {
		channels, total, err = h.service.GetChannels(r.Context(), page, limit)
	}
	
	if err != nil {
		api.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get channels", err.Error())
		return
	}
	
	hasMore := page*limit < total
	var nextPage *int
	if hasMore {
		next := page + 1
		nextPage = &next
	}
	
	response := api.PaginatedResponse{
		Data:     channels,
		Page:     page,
		Limit:    limit,
		Total:    total,
		HasMore:  hasMore,
		NextPage: nextPage,
	}
	
	api.WriteResponse(w, r, http.StatusOK, response)
}

// GetChannel handles GET /api/rest/channels/{id}
func (h *Handler) GetChannel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	
	id, err := strconv.Atoi(idStr)
	if err != nil {
		api.WriteErrorResponse(w, http.StatusBadRequest, "Invalid channel ID", err.Error())
		return
	}
	
	channel, err := h.service.GetChannel(r.Context(), id)
	if err != nil {
		if err.Error() == "channel not found" {
			api.WriteErrorResponse(w, http.StatusNotFound, "Channel not found", "")
			return
		}
		api.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get channel", err.Error())
		return
	}
	
	api.WriteResponse(w, r, http.StatusOK, channel)
}

// GetPlaylists handles GET /api/rest/playlists
func (h *Handler) GetPlaylists(w http.ResponseWriter, r *http.Request) {
	page, limit, err := api.GetPaginationParams(r)
	if err != nil {
		api.WriteErrorResponse(w, http.StatusBadRequest, "Invalid pagination parameters", err.Error())
		return
	}
	
	var channelID *int
	if channelIDStr := r.URL.Query().Get("channel_id"); channelIDStr != "" {
		id, err := strconv.Atoi(channelIDStr)
		if err != nil {
			api.WriteErrorResponse(w, http.StatusBadRequest, "Invalid channel_id parameter", err.Error())
			return
		}
		channelID = &id
	}
	
	playlists, total, err := h.service.GetPlaylists(r.Context(), page, limit, channelID)
	if err != nil {
		api.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get playlists", err.Error())
		return
	}
	
	hasMore := page*limit < total
	var nextPage *int
	if hasMore {
		next := page + 1
		nextPage = &next
	}
	
	response := api.PaginatedResponse{
		Data:     playlists,
		Page:     page,
		Limit:    limit,
		Total:    total,
		HasMore:  hasMore,
		NextPage: nextPage,
	}
	
	api.WriteResponse(w, r, http.StatusOK, response)
}

// GetPlaylist handles GET /api/rest/playlists/{id}
func (h *Handler) GetPlaylist(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	
	id, err := strconv.Atoi(idStr)
	if err != nil {
		api.WriteErrorResponse(w, http.StatusBadRequest, "Invalid playlist ID", err.Error())
		return
	}
	
	playlist, err := h.service.GetPlaylist(r.Context(), id)
	if err != nil {
		if err.Error() == "playlist not found" {
			api.WriteErrorResponse(w, http.StatusNotFound, "Playlist not found", "")
			return
		}
		api.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get playlist", err.Error())
		return
	}
	
	api.WriteResponse(w, r, http.StatusOK, playlist)
}

// GetVideos handles GET /api/rest/videos
func (h *Handler) GetVideos(w http.ResponseWriter, r *http.Request) {
	page, limit, err := api.GetPaginationParams(r)
	if err != nil {
		api.WriteErrorResponse(w, http.StatusBadRequest, "Invalid pagination parameters", err.Error())
		return
	}
	
	var channelID, playlistID *int
	
	if channelIDStr := r.URL.Query().Get("channel_id"); channelIDStr != "" {
		id, err := strconv.Atoi(channelIDStr)
		if err != nil {
			api.WriteErrorResponse(w, http.StatusBadRequest, "Invalid channel_id parameter", err.Error())
			return
		}
		channelID = &id
	}
	
	if playlistIDStr := r.URL.Query().Get("playlist_id"); playlistIDStr != "" {
		id, err := strconv.Atoi(playlistIDStr)
		if err != nil {
			api.WriteErrorResponse(w, http.StatusBadRequest, "Invalid playlist_id parameter", err.Error())
			return
		}
		playlistID = &id
	}
	
	videos, total, err := h.service.GetVideos(r.Context(), page, limit, channelID, playlistID)
	if err != nil {
		api.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get videos", err.Error())
		return
	}
	
	hasMore := page*limit < total
	var nextPage *int
	if hasMore {
		next := page + 1
		nextPage = &next
	}
	
	response := api.PaginatedResponse{
		Data:     videos,
		Page:     page,
		Limit:    limit,
		Total:    total,
		HasMore:  hasMore,
		NextPage: nextPage,
	}
	
	api.WriteResponse(w, r, http.StatusOK, response)
}

// GetVideo handles GET /api/rest/videos/{id}
func (h *Handler) GetVideo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	
	id, err := strconv.Atoi(idStr)
	if err != nil {
		api.WriteErrorResponse(w, http.StatusBadRequest, "Invalid video ID", err.Error())
		return
	}
	
	video, err := h.service.GetVideo(r.Context(), id)
	if err != nil {
		if err.Error() == "video not found" {
			api.WriteErrorResponse(w, http.StatusNotFound, "Video not found", "")
			return
		}
		api.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get video", err.Error())
		return
	}
	
	api.WriteResponse(w, r, http.StatusOK, video)
}

// RegisterRoutes registers REST API routes
func RegisterRoutes(router *mux.Router, service *api.Service, authConfig api.AuthConfig) {
	handler := NewHandler(service)
	
	// Create API subrouter
	apiRouter := router.PathPrefix("/api/rest").Subrouter()
	
	// Add middleware
	corsConfig := api.DefaultCORSConfig()
	apiRouter.Use(api.CORSMiddleware(corsConfig))
	apiRouter.Use(api.AuthMiddleware(authConfig))
	
	// Register routes
	apiRouter.HandleFunc("/channels", handler.GetChannels).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/channels/{id:[0-9]+}", handler.GetChannel).Methods("GET", "OPTIONS")
	
	apiRouter.HandleFunc("/playlists", handler.GetPlaylists).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/playlists/{id:[0-9]+}", handler.GetPlaylist).Methods("GET", "OPTIONS")
	
	apiRouter.HandleFunc("/videos", handler.GetVideos).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/videos/{id:[0-9]+}", handler.GetVideo).Methods("GET", "OPTIONS")
}