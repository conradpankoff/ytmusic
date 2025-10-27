package api

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	APIKey    string
	Username  string
	Password  string
	Enabled   bool
}

// AuthMiddleware provides HTTP authentication for non-read-only operations
func AuthMiddleware(config AuthConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Allow read operations without authentication if auth is disabled
			if !config.Enabled {
				next.ServeHTTP(w, r)
				return
			}

			// Check if this is a read-only operation
			isReadOnly := r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS"
			
			// Allow read-only operations without authentication
			if isReadOnly {
				next.ServeHTTP(w, r)
				return
			}

			// For non-read-only operations, require authentication
			if !isAuthenticated(r, config) {
				w.Header().Set("WWW-Authenticate", `Basic realm="API"`)
				WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required", "")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// isAuthenticated checks if the request is properly authenticated
func isAuthenticated(r *http.Request, config AuthConfig) bool {
	// Check API key in header
	if config.APIKey != "" {
		apiKey := r.Header.Get("X-API-Key")
		if apiKey != "" {
			return subtle.ConstantTimeCompare([]byte(apiKey), []byte(config.APIKey)) == 1
		}
	}

	// Check API key in query parameter
	if config.APIKey != "" {
		queryAPIKey := r.URL.Query().Get("api_key")
		if queryAPIKey != "" {
			return subtle.ConstantTimeCompare([]byte(queryAPIKey), []byte(config.APIKey)) == 1
		}
	}

	// Check Basic Authentication
	if config.Username != "" && config.Password != "" {
		username, password, ok := r.BasicAuth()
		if ok {
			userMatch := subtle.ConstantTimeCompare([]byte(username), []byte(config.Username)) == 1
			passMatch := subtle.ConstantTimeCompare([]byte(password), []byte(config.Password)) == 1
			return userMatch && passMatch
		}
	}

	// Check Bearer token
	if config.APIKey != "" {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				return subtle.ConstantTimeCompare([]byte(parts[1]), []byte(config.APIKey)) == 1
			}
		}
	}

	return false
}