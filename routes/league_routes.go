package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/matt-j-deasy/betty-crokers-api/handlers"
)

// Public League routes (no auth): GET collection + GET by id
func RegisterLeaguePublicRoutes(rg *gin.RouterGroup, h *handlers.LeagueHandler) {
	g := rg.Group("/leagues")
	g.GET("", h.List)
	g.GET("/:id", h.Get)
}

// Protected League routes (auth required): create/update/delete
func RegisterLeagueProtectedRoutes(rg *gin.RouterGroup, h *handlers.LeagueHandler) {
	g := rg.Group("/leagues")
	g.POST("", h.Create)
	g.PUT("/:id", h.Update)
	g.DELETE("/:id", h.Delete)
}
