package tests

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"

	"fknsrs.biz/p/ytmusic/internal/api"
	"fknsrs.biz/p/ytmusic/internal/api/rest"
	"fknsrs.biz/p/ytmusic/internal/api/graphql"
	"fknsrs.biz/p/ytmusic/internal/api/odata"
	"fknsrs.biz/p/ytmusic/internal/api/openapi"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	
	// Create tables
	schema := `
	CREATE TABLE channels (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		created_at DATETIME NOT NULL,
		external_id TEXT NOT NULL UNIQUE,
		title TEXT NOT NULL,
		metadata_updated_at DATETIME,
		thumbnail_updated_at DATETIME,
		playlists_updated_at DATETIME,
		videos_updated_at DATETIME
	);
	
	CREATE TABLE playlists (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		created_at DATETIME NOT NULL,
		external_id TEXT NOT NULL UNIQUE,
		channel_id INTEGER,
		channel_external_id TEXT,
		title TEXT NOT NULL,
		metadata_updated_at DATETIME,
		thumbnail_updated_at DATETIME,
		FOREIGN KEY(channel_id) REFERENCES channels(id)
	);
	
	CREATE TABLE videos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		created_at DATETIME NOT NULL,
		external_id TEXT NOT NULL UNIQUE,
		channel_id INTEGER,
		channel_external_id TEXT,
		title TEXT NOT NULL,
		description TEXT,
		publish_date DATETIME,
		upload_date DATETIME,
		metadata_updated_at DATETIME,
		thumbnail_updated_at DATETIME,
		downloaded_at DATETIME,
		transcoded_360_at DATETIME,
		transcoded_720_at DATETIME,
		audio_extracted_at DATETIME,
		FOREIGN KEY(channel_id) REFERENCES channels(id)
	);`
	
	_, err = db.Exec(schema)
	require.NoError(t, err)
	
	// Insert test data
	now := time.Now()
	
	_, err = db.Exec(`INSERT INTO channels (id, created_at, external_id, title, metadata_updated_at) VALUES 
		(1, ?, 'UCtest1', 'Test Channel 1', ?),
		(2, ?, 'UCtest2', 'Test Channel 2', ?)`,
		now, now, now.Add(-time.Hour), now.Add(-time.Hour))
	require.NoError(t, err)
	
	_, err = db.Exec(`INSERT INTO playlists (id, created_at, external_id, channel_id, title, metadata_updated_at) VALUES 
		(1, ?, 'PLtest1', 1, 'Test Playlist 1', ?),
		(2, ?, 'PLtest2', 2, 'Test Playlist 2', ?)`,
		now, now, now.Add(-time.Hour), now.Add(-time.Hour))
	require.NoError(t, err)
	
	_, err = db.Exec(`INSERT INTO videos (id, created_at, external_id, channel_id, title, description, metadata_updated_at) VALUES 
		(1, ?, 'VIDtest1', 1, 'Test Video 1', 'Description 1', ?),
		(2, ?, 'VIDtest2', 2, 'Test Video 2', 'Description 2', ?)`,
		now, now, now.Add(-time.Hour), now.Add(-time.Hour))
	require.NoError(t, err)
	
	return db
}

func setupTestRouter(t *testing.T, db *sql.DB) *mux.Router {
	service := api.NewService(db)
	authConfig := api.AuthConfig{
		Enabled: false, // Disable auth for tests
	}
	
	router := mux.NewRouter()
	
	// Register all API routes
	rest.RegisterRoutes(router, service, authConfig)
	
	err := graphql.RegisterRoutes(router, service, authConfig)
	require.NoError(t, err)
	
	odata.RegisterRoutes(router, service, authConfig)
	openapi.RegisterRoutes(router)
	
	return router
}

func TestRESTAPIChannels(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	router := setupTestRouter(t, db)
	
	t.Run("GET /api/rest/channels", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/rest/channels", nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response api.PaginatedResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Equal(t, 1, response.Page)
		assert.Equal(t, 50, response.Limit)
		assert.Equal(t, 2, response.Total)
		assert.False(t, response.HasMore)
		
		channels, ok := response.Data.([]interface{})
		require.True(t, ok)
		assert.Len(t, channels, 2)
	})
	
	t.Run("GET /api/rest/channels/1", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/rest/channels/1", nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var channel api.APIChannel
		err := json.Unmarshal(w.Body.Bytes(), &channel)
		require.NoError(t, err)
		
		assert.Equal(t, 1, channel.ID)
		assert.Equal(t, "UCtest1", channel.ExternalID)
		assert.Equal(t, "Test Channel 1", channel.Title)
	})
	
	t.Run("GET /api/rest/channels/999 (not found)", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/rest/channels/999", nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
	
	t.Run("GET /api/rest/channels with search", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/rest/channels?q=Channel+1", nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response api.PaginatedResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		channels, ok := response.Data.([]interface{})
		require.True(t, ok)
		assert.Len(t, channels, 1)
	})
}

func TestRESTAPIPlaylists(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	router := setupTestRouter(t, db)
	
	t.Run("GET /api/rest/playlists", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/rest/playlists", nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response api.PaginatedResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Equal(t, 2, response.Total)
		
		playlists, ok := response.Data.([]interface{})
		require.True(t, ok)
		assert.Len(t, playlists, 2)
	})
	
	t.Run("GET /api/rest/playlists?channel_id=1", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/rest/playlists?channel_id=1", nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response api.PaginatedResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Equal(t, 1, response.Total)
	})
}

func TestRESTAPIVideos(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	router := setupTestRouter(t, db)
	
	t.Run("GET /api/rest/videos", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/rest/videos", nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response api.PaginatedResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Equal(t, 2, response.Total)
		
		videos, ok := response.Data.([]interface{})
		require.True(t, ok)
		assert.Len(t, videos, 2)
	})
	
	t.Run("GET /api/rest/videos/1", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/rest/videos/1", nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var video api.APIVideo
		err := json.Unmarshal(w.Body.Bytes(), &video)
		require.NoError(t, err)
		
		assert.Equal(t, 1, video.ID)
		assert.Equal(t, "VIDtest1", video.ExternalID)
		assert.Equal(t, "Test Video 1", video.Title)
	})
}

func TestGraphQLAPI(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	router := setupTestRouter(t, db)
	
	t.Run("GraphQL channel query", func(t *testing.T) {
		query := `{"query": "{ channel(id: 1) { id external_id title } }"}`
		req := httptest.NewRequest("POST", "/api/graphql", strings.NewReader(query))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		data, exists := response["data"]
		assert.True(t, exists)
		assert.NotNil(t, data)
	})
	
	t.Run("GraphQL channels query with pagination", func(t *testing.T) {
		query := `{"query": "{ channels(page: 1, limit: 10) { data { id title } pagination { total has_more } } }"}`
		req := httptest.NewRequest("POST", "/api/graphql", strings.NewReader(query))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		data, exists := response["data"]
		assert.True(t, exists)
		assert.NotNil(t, data)
	})
}

func TestODataAPI(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	router := setupTestRouter(t, db)
	
	t.Run("GET /api/odata/$metadata", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/odata/$metadata", nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/xml", w.Header().Get("Content-Type"))
		assert.Contains(t, w.Body.String(), "EntityType")
		assert.Contains(t, w.Body.String(), "Channel")
	})
	
	t.Run("GET /api/odata/Channels", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/odata/Channels", nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/atom+xml", w.Header().Get("Content-Type"))
		assert.Contains(t, w.Body.String(), "feed")
		assert.Contains(t, w.Body.String(), "entry")
	})
	
	t.Run("GET /api/odata/Channels(1)", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/odata/Channels(1)", nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/atom+xml", w.Header().Get("Content-Type"))
		assert.Contains(t, w.Body.String(), "feed")
	})
}

func TestCORS(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	router := setupTestRouter(t, db)
	
	t.Run("OPTIONS request includes CORS headers", func(t *testing.T) {
		req := httptest.NewRequest("OPTIONS", "/api/rest/channels", nil)
		req.Header.Set("Origin", "http://example.com")
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "GET")
	})
}

func TestAuthentication(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	service := api.NewService(db)
	authConfig := api.AuthConfig{
		Enabled:  true,
		APIKey:   "test-api-key",
		Username: "testuser",
		Password: "testpass",
	}
	
	router := mux.NewRouter()
	rest.RegisterRoutes(router, service, authConfig)
	
	t.Run("GET request (read-only) without auth should succeed", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/rest/channels", nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
	})
	
	t.Run("POST request without auth should fail", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/rest/channels", nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
	
	t.Run("POST request with API key should succeed", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/rest/channels", nil)
		req.Header.Set("X-API-Key", "test-api-key")
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		// Note: This will still fail because we don't have a POST handler, 
		// but it shouldn't fail with 401 Unauthorized
		assert.NotEqual(t, http.StatusUnauthorized, w.Code)
	})
}

func TestOpenAPIDocumentation(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	router := setupTestRouter(t, db)
	
	t.Run("GET /api/openapi.json", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/openapi.json", nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		
		var spec map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &spec)
		require.NoError(t, err)
		
		assert.Equal(t, "3.0.3", spec["openapi"])
		assert.Contains(t, spec, "info")
		assert.Contains(t, spec, "paths")
		assert.Contains(t, spec, "components")
	})
	
	t.Run("GET /api/docs", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/docs", nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "text/html", w.Header().Get("Content-Type"))
		assert.Contains(t, w.Body.String(), "swagger-ui")
		assert.Contains(t, w.Body.String(), "YouTube Music API Documentation")
	})
}