package odata

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	
	"fknsrs.biz/p/ytmusic/internal/api"
)

// Handler wraps the API service for OData endpoints
type Handler struct {
	service *api.Service
}

// NewHandler creates a new OData handler
func NewHandler(service *api.Service) *Handler {
	return &Handler{service: service}
}

// ODataResponse represents an OData response envelope
type ODataResponse struct {
	XMLName xml.Name    `xml:"feed"`
	Xmlns   string      `xml:"xmlns,attr"`
	XmlnsD  string      `xml:"xmlns:d,attr"`
	XmlnsM  string      `xml:"xmlns:m,attr"`
	ID      string      `xml:"id"`
	Title   string      `xml:"title"`
	Updated string      `xml:"updated"`
	Entries []ODataEntry `xml:"entry"`
	Count   *int        `xml:"m:count,omitempty"`
}

// ODataEntry represents an individual OData entry
type ODataEntry struct {
	XMLName xml.Name         `xml:"entry"`
	ID      string           `xml:"id"`
	Title   string           `xml:"title"`
	Updated string           `xml:"updated"`
	Content ODataEntryContent `xml:"content"`
}

// ODataEntryContent represents the content of an OData entry
type ODataEntryContent struct {
	XMLName    xml.Name               `xml:"content"`
	Type       string                 `xml:"type,attr"`
	Properties ODataEntryProperties   `xml:"m:properties"`
}

// ODataEntryProperties represents the properties of an OData entry
type ODataEntryProperties struct {
	XMLName xml.Name `xml:"m:properties"`
	Data    interface{} `xml:",innerxml"`
}

// Metadata handles GET /$metadata
func (h *Handler) Metadata(w http.ResponseWriter, r *http.Request) {
	metadata := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
  <edmx:DataServices>
    <Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="YTMusic">
      
      <EntityType Name="Channel">
        <Key>
          <PropertyRef Name="ID"/>
        </Key>
        <Property Name="ID" Type="Edm.Int32" Nullable="false"/>
        <Property Name="CreatedAt" Type="Edm.DateTimeOffset" Nullable="false"/>
        <Property Name="ExternalID" Type="Edm.String" Nullable="false"/>
        <Property Name="Title" Type="Edm.String" Nullable="false"/>
        <Property Name="MetadataUpdatedAt" Type="Edm.DateTimeOffset"/>
        <Property Name="ThumbnailUpdatedAt" Type="Edm.DateTimeOffset"/>
        <Property Name="PlaylistsUpdatedAt" Type="Edm.DateTimeOffset"/>
        <Property Name="VideosUpdatedAt" Type="Edm.DateTimeOffset"/>
      </EntityType>

      <EntityType Name="Playlist">
        <Key>
          <PropertyRef Name="ID"/>
        </Key>
        <Property Name="ID" Type="Edm.Int32" Nullable="false"/>
        <Property Name="CreatedAt" Type="Edm.DateTimeOffset" Nullable="false"/>
        <Property Name="ExternalID" Type="Edm.String" Nullable="false"/>
        <Property Name="ChannelID" Type="Edm.Int32"/>
        <Property Name="ChannelExternalID" Type="Edm.String"/>
        <Property Name="Title" Type="Edm.String" Nullable="false"/>
        <Property Name="MetadataUpdatedAt" Type="Edm.DateTimeOffset"/>
        <Property Name="ThumbnailUpdatedAt" Type="Edm.DateTimeOffset"/>
      </EntityType>

      <EntityType Name="Video">
        <Key>
          <PropertyRef Name="ID"/>
        </Key>
        <Property Name="ID" Type="Edm.Int32" Nullable="false"/>
        <Property Name="CreatedAt" Type="Edm.DateTimeOffset" Nullable="false"/>
        <Property Name="ExternalID" Type="Edm.String" Nullable="false"/>
        <Property Name="ChannelID" Type="Edm.Int32"/>
        <Property Name="ChannelExternalID" Type="Edm.String"/>
        <Property Name="Title" Type="Edm.String" Nullable="false"/>
        <Property Name="Description" Type="Edm.String"/>
        <Property Name="PublishDate" Type="Edm.DateTimeOffset"/>
        <Property Name="UploadDate" Type="Edm.DateTimeOffset"/>
        <Property Name="MetadataUpdatedAt" Type="Edm.DateTimeOffset"/>
        <Property Name="ThumbnailUpdatedAt" Type="Edm.DateTimeOffset"/>
        <Property Name="DownloadedAt" Type="Edm.DateTimeOffset"/>
        <Property Name="Transcoded360At" Type="Edm.DateTimeOffset"/>
        <Property Name="Transcoded720At" Type="Edm.DateTimeOffset"/>
        <Property Name="AudioExtractedAt" Type="Edm.DateTimeOffset"/>
      </EntityType>

      <EntityContainer Name="YTMusicContainer">
        <EntitySet Name="Channels" EntityType="YTMusic.Channel"/>
        <EntitySet Name="Playlists" EntityType="YTMusic.Playlist"/>
        <EntitySet Name="Videos" EntityType="YTMusic.Video"/>
      </EntityContainer>
      
    </Schema>
  </edmx:DataServices>
</edmx:Edmx>`

	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(metadata))
}

// GetChannels handles GET /Channels
func (h *Handler) GetChannels(w http.ResponseWriter, r *http.Request) {
	page, limit := h.parseODataParams(r)
	
	channels, total, err := h.service.GetChannels(r.Context(), page, limit)
	if err != nil {
		api.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get channels", err.Error())
		return
	}
	
	h.writeODataResponse(w, r, "Channels", channels, total)
}

// GetChannel handles GET /Channels(id)
func (h *Handler) GetChannel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	
	// Remove parentheses if present
	idStr = strings.Trim(idStr, "()")
	
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
	
	h.writeODataResponse(w, r, "Channels", []api.APIChannel{*channel}, 1)
}

// GetPlaylists handles GET /Playlists
func (h *Handler) GetPlaylists(w http.ResponseWriter, r *http.Request) {
	page, limit := h.parseODataParams(r)
	
	playlists, total, err := h.service.GetPlaylists(r.Context(), page, limit, nil)
	if err != nil {
		api.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get playlists", err.Error())
		return
	}
	
	h.writeODataResponse(w, r, "Playlists", playlists, total)
}

// GetPlaylist handles GET /Playlists(id)
func (h *Handler) GetPlaylist(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	
	// Remove parentheses if present
	idStr = strings.Trim(idStr, "()")
	
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
	
	h.writeODataResponse(w, r, "Playlists", []api.APIPlaylist{*playlist}, 1)
}

// GetVideos handles GET /Videos
func (h *Handler) GetVideos(w http.ResponseWriter, r *http.Request) {
	page, limit := h.parseODataParams(r)
	
	videos, total, err := h.service.GetVideos(r.Context(), page, limit, nil, nil)
	if err != nil {
		api.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get videos", err.Error())
		return
	}
	
	h.writeODataResponse(w, r, "Videos", videos, total)
}

// GetVideo handles GET /Videos(id)
func (h *Handler) GetVideo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	
	// Remove parentheses if present
	idStr = strings.Trim(idStr, "()")
	
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
	
	h.writeODataResponse(w, r, "Videos", []api.APIVideo{*video}, 1)
}

// parseODataParams extracts OData query parameters
func (h *Handler) parseODataParams(r *http.Request) (page, limit int) {
	page = 1
	limit = 50
	
	// Handle $skip and $top for pagination
	if skip := r.URL.Query().Get("$skip"); skip != "" {
		if s, err := strconv.Atoi(skip); err == nil && s >= 0 {
			page = (s / limit) + 1
		}
	}
	
	if top := r.URL.Query().Get("$top"); top != "" {
		if t, err := strconv.Atoi(top); err == nil && t > 0 && t <= 1000 {
			limit = t
		}
	}
	
	return page, limit
}

// writeODataResponse writes data in OData format
func (h *Handler) writeODataResponse(w http.ResponseWriter, r *http.Request, entitySet string, data interface{}, total int) {
	baseURL := fmt.Sprintf("%s://%s%s", "http", r.Host, r.URL.Path)
	
	response := ODataResponse{
		Xmlns:   "http://www.w3.org/2005/Atom",
		XmlnsD:  "http://docs.oasis-open.org/odata/ns/data",
		XmlnsM:  "http://docs.oasis-open.org/odata/ns/metadata",
		ID:      baseURL,
		Title:   entitySet,
		Updated: "2023-01-01T00:00:00Z",
	}
	
	// Add count if requested
	if r.URL.Query().Get("$inlinecount") == "allpages" {
		response.Count = &total
	}
	
	// Convert data to entries (simplified - in a real implementation, you'd properly marshal each entity)
	switch d := data.(type) {
	case []api.APIChannel:
		for i, item := range d {
			entry := ODataEntry{
				ID:      fmt.Sprintf("%s(%d)", baseURL, item.ID),
				Title:   item.Title,
				Updated: item.CreatedAt.Format("2006-01-02T15:04:05Z"),
				Content: ODataEntryContent{
					Type: "application/xml",
					Properties: ODataEntryProperties{
						Data: item,
					},
				},
			}
			response.Entries = append(response.Entries, entry)
			if i >= 100 { // Limit to prevent huge responses
				break
			}
		}
	case []api.APIPlaylist:
		for i, item := range d {
			entry := ODataEntry{
				ID:      fmt.Sprintf("%s(%d)", baseURL, item.ID),
				Title:   item.Title,
				Updated: item.CreatedAt.Format("2006-01-02T15:04:05Z"),
				Content: ODataEntryContent{
					Type: "application/xml",
					Properties: ODataEntryProperties{
						Data: item,
					},
				},
			}
			response.Entries = append(response.Entries, entry)
			if i >= 100 {
				break
			}
		}
	case []api.APIVideo:
		for i, item := range d {
			entry := ODataEntry{
				ID:      fmt.Sprintf("%s(%d)", baseURL, item.ID),
				Title:   item.Title,
				Updated: item.CreatedAt.Format("2006-01-02T15:04:05Z"),
				Content: ODataEntryContent{
					Type: "application/xml",
					Properties: ODataEntryProperties{
						Data: item,
					},
				},
			}
			response.Entries = append(response.Entries, entry)
			if i >= 100 {
				break
			}
		}
	}
	
	w.Header().Set("Content-Type", "application/atom+xml")
	w.WriteHeader(http.StatusOK)
	
	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")
	encoder.Encode(response)
}

// RegisterRoutes registers OData API routes
func RegisterRoutes(router *mux.Router, service *api.Service, authConfig api.AuthConfig) {
	handler := NewHandler(service)
	
	// Create API subrouter
	apiRouter := router.PathPrefix("/api/odata").Subrouter()
	
	// Add middleware
	corsConfig := api.DefaultCORSConfig()
	apiRouter.Use(api.CORSMiddleware(corsConfig))
	apiRouter.Use(api.AuthMiddleware(authConfig))
	
	// Register routes
	apiRouter.HandleFunc("/$metadata", handler.Metadata).Methods("GET", "OPTIONS")
	
	apiRouter.HandleFunc("/Channels", handler.GetChannels).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/Channels({id:[0-9]+})", handler.GetChannel).Methods("GET", "OPTIONS")
	
	apiRouter.HandleFunc("/Playlists", handler.GetPlaylists).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/Playlists({id:[0-9]+})", handler.GetPlaylist).Methods("GET", "OPTIONS")
	
	apiRouter.HandleFunc("/Videos", handler.GetVideos).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/Videos({id:[0-9]+})", handler.GetVideo).Methods("GET", "OPTIONS")
}