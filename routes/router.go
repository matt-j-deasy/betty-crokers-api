package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/matt-j-deasy/betty-crokers-api/config"
	"github.com/matt-j-deasy/betty-crokers-api/handlers"
	"github.com/matt-j-deasy/betty-crokers-api/middleware"
)

// SetupRouter initializes all routes and returns a configured Gin engine.
func SetupRouter(handlers *handlers.HandlersCollection, cfg config.Environment) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())                 // Default Gin recovery middleware
	router.Use(middleware.CORSMiddleware(cfg)) // Custom CORS middleware with config
	router.Use(middleware.SlogMiddleware())    // Custom logging middleware

	// Base API group with version prefix
	apiV1 := router.Group("/api/v1")

	// Register all routes
	registerRoutes(apiV1, handlers)

	return router
}

// registerRoutes groups all the route registrations in one place.
func registerRoutes(apiV1 *gin.RouterGroup, handlers *handlers.HandlersCollection) {

	// Public routes
	apiV1.GET("/health", handlers.HealthCheckHandler.HealthCheck)
}
