package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/matt-j-deasy/betty-crokers-api/handlers"
)

// Public TeamSeason routes (no auth): list links + cross-lookups
func RegisterTeamSeasonPublicRoutes(rg *gin.RouterGroup, h *handlers.TeamSeasonHandler) {
	// List link rows
	rg.GET("/team-seasons", h.List) // GET /api/v1/team-seasons?teamId=&seasonId=&onlyActive=&page=&size=

	// Cross lookups (use :id to match existing /teams/:id and /seasons/:id)
	rg.GET("/teams/:id/seasons", h.ListSeasonsForTeam)       // GET /api/v1/teams/:id/seasons?onlyActive=
	rg.GET("/seasons/:seasonId/teams", h.ListTeamsForSeason) // GET /api/v1/seasons/:seasonId/teams?onlyActive=
}

// Protected TeamSeason routes (auth required): link/unlink/toggle
func RegisterTeamSeasonProtectedRoutes(rg *gin.RouterGroup, h *handlers.TeamSeasonHandler) {
	rg.POST("/team-seasons", h.Link)                              // POST /api/v1/team-seasons
	rg.PUT("/team-seasons/:teamId/:seasonId/active", h.SetActive) // PUT /api/v1/team-seasons/:teamId/:seasonId/active
	rg.DELETE("/team-seasons/:teamId/:seasonId", h.Unlink)        // DELETE /api/v1/team-seasons/:teamId/:seasonId
}
