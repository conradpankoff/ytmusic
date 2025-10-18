package graphql

import (
	"github.com/gorilla/mux"
	"github.com/graphql-go/handler"
	
	"fknsrs.biz/p/ytmusic/internal/api"
)

// RegisterRoutes registers GraphQL API routes
func RegisterRoutes(router *mux.Router, service *api.Service, authConfig api.AuthConfig) error {
	schema, err := CreateSchema(service)
	if err != nil {
		return err
	}
	
	// Create GraphQL handler
	h := handler.New(&handler.Config{
		Schema:     &schema,
		Pretty:     true,
		GraphiQL:   true,
		Playground: true,
	})
	
	// Create API subrouter
	apiRouter := router.PathPrefix("/api/graphql").Subrouter()
	
	// Add middleware
	corsConfig := api.DefaultCORSConfig()
	apiRouter.Use(api.CORSMiddleware(corsConfig))
	apiRouter.Use(api.AuthMiddleware(authConfig))
	
	// Register GraphQL endpoint
	apiRouter.Handle("", h).Methods("GET", "POST", "OPTIONS")
	apiRouter.Handle("/", h).Methods("GET", "POST", "OPTIONS")
	
	return nil
}