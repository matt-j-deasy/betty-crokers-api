package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/matt-j-deasy/betty-crokers-api/config"
	"github.com/matt-j-deasy/betty-crokers-api/services"
)

// HealthCheckHandler handles health check-related routes
type HealthCheckHandler struct {
	cfg      config.Environment
	services services.ServicesCollection
}

// NewHealthCheckHandler creates a new instance of HealthCheckHandler
func NewHealthCheckHandler(services services.ServicesCollection, cfg config.Environment) *HealthCheckHandler {
	return &HealthCheckHandler{
		cfg:      cfg,
		services: services,
	}
}

// HealthCheck is a simple health check endpoint
func (h *HealthCheckHandler) HealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ok",
	})
}
