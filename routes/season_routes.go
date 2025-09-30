package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/matt-j-deasy/betty-crokers-api/handlers"
)

// Public Season routes (no auth)
func RegisterSeasonPublicRoutes(rg *gin.RouterGroup, h *handlers.SeasonHandler) {
	g := rg.Group("/seasons")
	g.GET("", h.List)    // GET /api/v1/seasons?q=&page=&size=&leagueId=
	g.GET("/:id", h.Get) // GET /api/v1/seasons/:id
	g.GET("/:id/standings", h.Standings)
}

// Protected Season routes (auth required)
func RegisterSeasonProtectedRoutes(rg *gin.RouterGroup, h *handlers.SeasonHandler) {
	g := rg.Group("/seasons")
	g.POST("", h.Create)    // POST /api/v1/seasons
	g.PUT("/:id", h.Update) // PUT /api/v1/seasons/:id
	g.DELETE("/:id", h.Delete)
}
