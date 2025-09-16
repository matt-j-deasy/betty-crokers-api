package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/matt-j-deasy/betty-crokers-api/handlers"
)

// Public player routes (no auth)
func RegisterPlayerPublicRoutes(rg *gin.RouterGroup, h *handlers.PlayerHandler) {
	g := rg.Group("/players")
	g.GET("", h.List)
	g.GET("/:id", h.Get)
}

// Protected player routes (auth required)
func RegisterPlayerProtectedRoutes(rg *gin.RouterGroup, h *handlers.PlayerHandler) {
	g := rg.Group("/players")
	g.POST("", h.Create)
	g.PUT("/:id", h.Update)
	g.DELETE("/:id", h.Delete)
}
