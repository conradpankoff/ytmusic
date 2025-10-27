package openapi

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// OpenAPISpec represents the OpenAPI 3.0 specification
type OpenAPISpec struct {
	OpenAPI      string                 `json:"openapi"`
	Info         OpenAPIInfo            `json:"info"`
	Servers      []OpenAPIServer        `json:"servers,omitempty"`
	Paths        map[string]interface{} `json:"paths"`
	Components   OpenAPIComponents      `json:"components"`
	Security     []map[string][]string  `json:"security,omitempty"`
}

type OpenAPIInfo struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Contact     *OpenAPIContact `json:"contact,omitempty"`
}

type OpenAPIContact struct {
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

type OpenAPIServer struct {
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

type OpenAPIComponents struct {
	Schemas         map[string]interface{} `json:"schemas"`
	SecuritySchemes map[string]interface{} `json:"securitySchemes"`
}

// GenerateOpenAPISpec generates the OpenAPI specification
func GenerateOpenAPISpec() *OpenAPISpec {
	spec := &OpenAPISpec{
		OpenAPI: "3.0.3",
		Info: OpenAPIInfo{
			Title:       "YouTube Music API",
			Description: "API for accessing YouTube Music channels, playlists, and videos with support for REST, GraphQL, and OData endpoints.",
			Version:     "1.0.0",
			Contact: &OpenAPIContact{
				Name: "YTMusic API",
			},
		},
		Servers: []OpenAPIServer{
			{
				URL:         "/api",
				Description: "API Server",
			},
		},
		Paths: map[string]interface{}{
			"/rest/channels": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Get channels",
					"description": "Retrieve a paginated list of YouTube music channels",
					"tags":        []string{"Channels"},
					"parameters": []map[string]interface{}{
						{
							"name":        "page",
							"in":          "query",
							"description": "Page number (default: 1)",
							"required":    false,
							"schema":      map[string]interface{}{"type": "integer", "minimum": 1},
						},
						{
							"name":        "limit",
							"in":          "query",
							"description": "Items per page (default: 50, max: 1000)",
							"required":    false,
							"schema":      map[string]interface{}{"type": "integer", "minimum": 1, "maximum": 1000},
						},
						{
							"name":        "q",
							"in":          "query",
							"description": "Search query for channel titles using full-text search",
							"required":    false,
							"schema":      map[string]interface{}{"type": "string"},
						},
						{
							"name":        "format",
							"in":          "query",
							"description": "Response format (json or xml)",
							"required":    false,
							"schema":      map[string]interface{}{"type": "string", "enum": []string{"json", "xml"}},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Successful response",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{"$ref": "#/components/schemas/PaginatedChannelResponse"},
								},
								"application/xml": map[string]interface{}{
									"schema": map[string]interface{}{"$ref": "#/components/schemas/PaginatedChannelResponse"},
								},
							},
						},
						"400": map[string]interface{}{
							"description": "Bad request",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{"$ref": "#/components/schemas/ErrorResponse"},
								},
							},
						},
					},
				},
			},
			"/rest/channels/{id}": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Get channel by ID",
					"description": "Retrieve a single YouTube music channel by ID",
					"tags":        []string{"Channels"},
					"parameters": []map[string]interface{}{
						{
							"name":        "id",
							"in":          "path",
							"description": "Channel ID",
							"required":    true,
							"schema":      map[string]interface{}{"type": "integer", "minimum": 1},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Successful response",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{"$ref": "#/components/schemas/Channel"},
								},
							},
						},
						"404": map[string]interface{}{
							"description": "Channel not found",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{"$ref": "#/components/schemas/ErrorResponse"},
								},
							},
						},
					},
				},
			},
			"/rest/playlists": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Get playlists",
					"description": "Retrieve a paginated list of YouTube music playlists with optional search and filtering",
					"tags":        []string{"Playlists"},
					"parameters": []map[string]interface{}{
						{
							"name":        "page",
							"in":          "query",
							"description": "Page number (default: 1)",
							"required":    false,
							"schema":      map[string]interface{}{"type": "integer", "minimum": 1},
						},
						{
							"name":        "limit",
							"in":          "query",
							"description": "Items per page (default: 50, max: 1000)",
							"required":    false,
							"schema":      map[string]interface{}{"type": "integer", "minimum": 1, "maximum": 1000},
						},
						{
							"name":        "q",
							"in":          "query",
							"description": "Search query for playlist titles using full-text search",
							"required":    false,
							"schema":      map[string]interface{}{"type": "string"},
						},
						{
							"name":        "channel_id",
							"in":          "query",
							"description": "Filter by channel ID",
							"required":    false,
							"schema":      map[string]interface{}{"type": "integer", "minimum": 1},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Successful response",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{"$ref": "#/components/schemas/PaginatedPlaylistResponse"},
								},
							},
						},
					},
				},
			},
			"/rest/videos": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Get videos",
					"description": "Retrieve a paginated list of YouTube music videos with optional search and filtering",
					"tags":        []string{"Videos"},
					"parameters": []map[string]interface{}{
						{
							"name":        "page",
							"in":          "query",
							"description": "Page number (default: 1)",
							"required":    false,
							"schema":      map[string]interface{}{"type": "integer", "minimum": 1},
						},
						{
							"name":        "limit",
							"in":          "query",
							"description": "Items per page (default: 50, max: 1000)",
							"required":    false,
							"schema":      map[string]interface{}{"type": "integer", "minimum": 1, "maximum": 1000},
						},
						{
							"name":        "q",
							"in":          "query",
							"description": "Search query for video titles and descriptions using full-text search",
							"required":    false,
							"schema":      map[string]interface{}{"type": "string"},
						},
						{
							"name":        "channel_id",
							"in":          "query",
							"description": "Filter by channel ID",
							"required":    false,
							"schema":      map[string]interface{}{"type": "integer", "minimum": 1},
						},
						{
							"name":        "playlist_id",
							"in":          "query",
							"description": "Filter by playlist ID",
							"required":    false,
							"schema":      map[string]interface{}{"type": "integer", "minimum": 1},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Successful response",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{"$ref": "#/components/schemas/PaginatedVideoResponse"},
								},
							},
						},
					},
				},
			},
			"/graphql": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "GraphQL Playground",
					"description": "Interactive GraphQL playground for exploring the API",
					"tags":        []string{"GraphQL"},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "GraphQL Playground",
							"content": map[string]interface{}{
								"text/html": map[string]interface{}{
									"schema": map[string]interface{}{"type": "string"},
								},
							},
						},
					},
				},
				"post": map[string]interface{}{
					"summary":     "Execute GraphQL Query",
					"description": "Execute a GraphQL query or mutation",
					"tags":        []string{"GraphQL"},
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"query": map[string]interface{}{
											"type":        "string",
											"description": "GraphQL query string",
										},
										"variables": map[string]interface{}{
											"type":        "object",
											"description": "Query variables",
										},
									},
									"required": []string{"query"},
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "GraphQL response",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"data":   map[string]interface{}{"type": "object"},
											"errors": map[string]interface{}{"type": "array"},
										},
									},
								},
							},
						},
					},
				},
			},
			"/odata/$metadata": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "OData Metadata",
					"description": "Get OData service metadata document",
					"tags":        []string{"OData"},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "OData metadata document",
							"content": map[string]interface{}{
								"application/xml": map[string]interface{}{
									"schema": map[string]interface{}{"type": "string"},
								},
							},
						},
					},
				},
			},
			"/odata/Channels": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "OData Channels",
					"description": "Get channels using OData protocol",
					"tags":        []string{"OData"},
					"parameters": []map[string]interface{}{
						{
							"name":        "$skip",
							"in":          "query",
							"description": "Number of records to skip",
							"required":    false,
							"schema":      map[string]interface{}{"type": "integer", "minimum": 0},
						},
						{
							"name":        "$top",
							"in":          "query",
							"description": "Maximum number of records to return",
							"required":    false,
							"schema":      map[string]interface{}{"type": "integer", "minimum": 1, "maximum": 1000},
						},
						{
							"name":        "$filter",
							"in":          "query",
							"description": "OData filter expression",
							"required":    false,
							"schema":      map[string]interface{}{"type": "string"},
						},
						{
							"name":        "$orderby",
							"in":          "query",
							"description": "OData orderby expression",
							"required":    false,
							"schema":      map[string]interface{}{"type": "string"},
						},
						{
							"name":        "$inlinecount",
							"in":          "query",
							"description": "Include total count in response",
							"required":    false,
							"schema":      map[string]interface{}{"type": "string", "enum": []string{"allpages", "none"}},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "OData response",
							"content": map[string]interface{}{
								"application/atom+xml": map[string]interface{}{
									"schema": map[string]interface{}{"type": "string"},
								},
							},
						},
					},
				},
			},
		},
		Components: OpenAPIComponents{
			Schemas: map[string]interface{}{
				"Channel": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id": map[string]interface{}{
							"type":        "integer",
							"description": "Unique channel identifier",
							"example":     1,
						},
						"created_at": map[string]interface{}{
							"type":        "string",
							"format":      "date-time",
							"description": "Channel creation timestamp",
						},
						"external_id": map[string]interface{}{
							"type":        "string",
							"description": "YouTube channel ID",
							"example":     "UCpNvmbdtY8WAzhdNUDxbT2g",
						},
						"title": map[string]interface{}{
							"type":        "string",
							"description": "Channel title",
							"example":     "Example Music Channel",
						},
						"metadata_updated_at": map[string]interface{}{
							"type":        "string",
							"format":      "date-time",
							"description": "Last metadata update timestamp",
							"nullable":    true,
						},
					},
					"required": []string{"id", "created_at", "external_id", "title"},
				},
				"Playlist": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id": map[string]interface{}{
							"type":        "integer",
							"description": "Unique playlist identifier",
						},
						"created_at": map[string]interface{}{
							"type":   "string",
							"format": "date-time",
						},
						"external_id": map[string]interface{}{
							"type":        "string",
							"description": "YouTube playlist ID",
						},
						"title": map[string]interface{}{
							"type":        "string",
							"description": "Playlist title",
						},
						"channel_id": map[string]interface{}{
							"type":        "integer",
							"description": "Associated channel ID",
							"nullable":    true,
						},
					},
				},
				"Video": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id": map[string]interface{}{
							"type":        "integer",
							"description": "Unique video identifier",
						},
						"created_at": map[string]interface{}{
							"type":   "string",
							"format": "date-time",
						},
						"external_id": map[string]interface{}{
							"type":        "string",
							"description": "YouTube video ID",
						},
						"title": map[string]interface{}{
							"type":        "string",
							"description": "Video title",
						},
						"description": map[string]interface{}{
							"type":        "string",
							"description": "Video description",
						},
					},
				},
				"PaginatedChannelResponse": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"data": map[string]interface{}{
							"type":  "array",
							"items": map[string]interface{}{"$ref": "#/components/schemas/Channel"},
						},
						"page": map[string]interface{}{
							"type":        "integer",
							"description": "Current page number",
						},
						"limit": map[string]interface{}{
							"type":        "integer",
							"description": "Items per page",
						},
						"total": map[string]interface{}{
							"type":        "integer",
							"description": "Total number of items",
						},
						"has_more": map[string]interface{}{
							"type":        "boolean",
							"description": "Whether there are more pages",
						},
					},
				},
				"PaginatedPlaylistResponse": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"data": map[string]interface{}{
							"type":  "array",
							"items": map[string]interface{}{"$ref": "#/components/schemas/Playlist"},
						},
						"page": map[string]interface{}{
							"type":        "integer",
							"description": "Current page number",
						},
						"limit": map[string]interface{}{
							"type":        "integer",
							"description": "Items per page",
						},
						"total": map[string]interface{}{
							"type":        "integer",
							"description": "Total number of items",
						},
						"has_more": map[string]interface{}{
							"type":        "boolean",
							"description": "Whether there are more pages",
						},
					},
				},
				"PaginatedVideoResponse": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"data": map[string]interface{}{
							"type":  "array",
							"items": map[string]interface{}{"$ref": "#/components/schemas/Video"},
						},
						"page": map[string]interface{}{
							"type":        "integer",
							"description": "Current page number",
						},
						"limit": map[string]interface{}{
							"type":        "integer",
							"description": "Items per page",
						},
						"total": map[string]interface{}{
							"type":        "integer",
							"description": "Total number of items",
						},
						"has_more": map[string]interface{}{
							"type":        "boolean",
							"description": "Whether there are more pages",
						},
					},
				},
				"ErrorResponse": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"error": map[string]interface{}{
							"type":        "string",
							"description": "Error type",
						},
						"message": map[string]interface{}{
							"type":        "string",
							"description": "Error message",
						},
						"code": map[string]interface{}{
							"type":        "integer",
							"description": "HTTP status code",
						},
					},
				},
			},
			SecuritySchemes: map[string]interface{}{
				"ApiKeyAuth": map[string]interface{}{
					"type": "apiKey",
					"in":   "header",
					"name": "X-API-Key",
					"description": "API key authentication. Pass your API key in the X-API-Key header.",
				},
				"BearerAuth": map[string]interface{}{
					"type":   "http",
					"scheme": "bearer",
					"description": "Bearer token authentication. Pass your token in the Authorization header as 'Bearer <token>'.",
				},
				"BasicAuth": map[string]interface{}{
					"type":   "http",
					"scheme": "basic",
					"description": "HTTP Basic authentication with username and password.",
				},
			},
		},
		Security: []map[string][]string{
			{"ApiKeyAuth": {}},
			{"BearerAuth": {}},
			{"BasicAuth": {}},
		},
	}

	return spec
}

// ServeOpenAPISpec serves the OpenAPI specification as JSON
func ServeOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	spec := GenerateOpenAPISpec()
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	encoder.Encode(spec)
}

// ServeSwaggerUI serves a simple Swagger UI
func ServeSwaggerUI(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>YouTube Music API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@3.52.5/swagger-ui.css" />
    <style>
        html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
        *, *:before, *:after { box-sizing: inherit; }
        body { margin:0; background: #fafafa; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@3.52.5/swagger-ui-bundle.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: '/api/openapi.json',
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIBundle.presets.standalone
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout"
            });
        }
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// RegisterRoutes registers OpenAPI documentation routes
func RegisterRoutes(router *mux.Router) {
	apiRouter := router.PathPrefix("/api").Subrouter()
	
	apiRouter.HandleFunc("/openapi.json", ServeOpenAPISpec).Methods("GET")
	apiRouter.HandleFunc("/docs", ServeSwaggerUI).Methods("GET")
}