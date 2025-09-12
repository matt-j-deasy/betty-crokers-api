package server

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/matt-j-deasy/betty-crokers-api/config"
	"github.com/matt-j-deasy/betty-crokers-api/handlers"
	"github.com/matt-j-deasy/betty-crokers-api/routes"
	"gorm.io/gorm"
)

type Server struct {
	Config     *config.Environment
	Router     *gin.Engine
	DB         *gorm.DB
	httpServer *http.Server // Add HTTP server instance
}

// CreateServer initializes and returns a configured server instance
func CreateServer(cfg config.Environment, db *gorm.DB, handlers *handlers.HandlersCollection) *Server {
	gin.SetMode(gin.ReleaseMode) // Set Gin to production mode
	router := routes.SetupRouter(handlers, cfg)

	s := &Server{
		Config: &cfg,
		DB:     db,
		Router: router,
		httpServer: &http.Server{
			Addr:    strconv.Itoa(cfg.LocalPort), // Default address; will be overridden by Start if port is provided
			Handler: router,                      // Use Gin router as the handler
		},
	}

	return s
}

// Start runs the server
func (s *Server) Start(port string) error {
	s.httpServer.Addr = port
	slog.Info("🚀 Server running on port", "port", port)
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully stops the server
func (s *Server) Shutdown() error {
	slog.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) // Allow 3 seconds for shutdown
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		slog.Error("Server shutdown failed", "error", err)
		return err
	}
	slog.Info("Server shut down successfully")
	return nil
}
