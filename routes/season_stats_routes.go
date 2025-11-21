package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/matt-j-deasy/betty-crokers-api/handlers"
)

// Public Season Stats routes (no auth)
func RegisterSeasonStatsPublicRoutes(rg *gin.RouterGroup, h *handlers.SeasonStatsHandler) {
	g := rg.Group("/seasons")

	// GET /api/v1/seasons/:seasonId/stats/players
	g.GET("/:seasonId/stats/players", h.ListSeasonPlayerStats)

	// GET /api/v1/seasons/:seasonId/stats/teams
	g.GET("/:seasonId/stats/teams", h.ListSeasonTeamStats)
}
