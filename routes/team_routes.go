package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/matt-j-deasy/betty-crokers-api/handlers"
)

// Public Team routes (no auth): GET collection + GET by id
func RegisterTeamPublicRoutes(rg *gin.RouterGroup, h *handlers.TeamHandler) {
	g := rg.Group("/teams")
	g.GET("", h.List)    // GET /api/v1/teams?q=&page=&size=&playerId=&seasonId=&onlyActive=
	g.GET("/:id", h.Get) // GET /api/v1/teams/:id
}

// Protected Team routes (auth required): create/update/delete
func RegisterTeamProtectedRoutes(rg *gin.RouterGroup, h *handlers.TeamHandler) {
	g := rg.Group("/teams")
	g.POST("", h.Create)    // POST /api/v1/teams
	g.PUT("/:id", h.Update) // PUT /api/v1/teams/:id
	g.DELETE("/:id", h.Delete)
}
