// handlers/season_stats_handler.go
package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/matt-j-deasy/betty-crokers-api/services"
)

type SeasonStatsHandler struct {
	services *services.ServicesCollection
}

func NewSeasonStatsHandler(services *services.ServicesCollection) *SeasonStatsHandler {
	return &SeasonStatsHandler{
		services: services,
	}
}

// parseSeasonIDParam is a small helper to DRY param parsing.
func parseSeasonIDParam(c *gin.Context) (int64, bool) {
	raw := c.Param("seasonId")
	if raw == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "missing seasonId",
		})
		return 0, false
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid seasonId",
		})
		return 0, false
	}
	return id, true
}

// GET /seasons/:seasonId/stats/players
func (h *SeasonStatsHandler) ListSeasonPlayerStats(c *gin.Context) {
	seasonID, ok := parseSeasonIDParam(c)
	if !ok {
		slog.Error("invalid season ID param")
		return
	}

	stats, err := h.services.SeasonStatsService.ListPlayerStats(c.Request.Context(), seasonID)
	if err != nil {
		slog.Error("failed to list season player stats", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch player stats." + err.Error(),
		})
		return
	}

	slog.Info("fetched season player stats", "seasonID", seasonID, "count", len(stats))
	c.JSON(http.StatusOK, stats)
}

// GET /seasons/:seasonId/stats/teams
func (h *SeasonStatsHandler) ListSeasonTeamStats(c *gin.Context) {
	seasonID, ok := parseSeasonIDParam(c)
	if !ok {
		return
	}

	stats, err := h.services.SeasonStatsService.ListTeamStats(c.Request.Context(), seasonID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch team stats." + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}
