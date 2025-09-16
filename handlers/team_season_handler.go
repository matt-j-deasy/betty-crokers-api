package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/matt-j-deasy/betty-crokers-api/services"
)

type TeamSeasonHandler struct {
	services *services.ServicesCollection
}

func NewTeamSeasonHandler(svcs *services.ServicesCollection) *TeamSeasonHandler {
	return &TeamSeasonHandler{services: svcs}
}

// POST /api/v1/team-seasons
type linkTeamSeasonReq struct {
	TeamID   int64 `json:"teamId" binding:"required,gt=0"`
	SeasonID int64 `json:"seasonId" binding:"required,gt=0"`
	IsActive *bool `json:"isActive"` // optional; defaults true
}

func (h *TeamSeasonHandler) Link(c *gin.Context) {
	var req linkTeamSeasonReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	out, err := h.services.TeamSeasonService.Link(c, services.LinkTeamSeasonInput{
		TeamID:   req.TeamID,
		SeasonID: req.SeasonID,
		IsActive: req.IsActive,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, out)
}

// PUT /api/v1/team-seasons/:teamId/:seasonId/active
type setActiveReq struct {
	IsActive bool `json:"isActive"`
}

func (h *TeamSeasonHandler) SetActive(c *gin.Context) {
	teamID, ok1 := parseIDParam(c.Param("teamId"))
	seasonID, ok2 := parseIDParam(c.Param("seasonId"))
	if !ok1 || !ok2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid teamId or seasonId"})
		return
	}
	var req setActiveReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	out, err := h.services.TeamSeasonService.SetActive(c, services.SetTeamSeasonActiveInput{
		TeamID:   teamID,
		SeasonID: seasonID,
		IsActive: req.IsActive,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, out)
}

// DELETE /api/v1/team-seasons/:teamId/:seasonId
func (h *TeamSeasonHandler) Unlink(c *gin.Context) {
	teamID, ok1 := parseIDParam(c.Param("teamId"))
	seasonID, ok2 := parseIDParam(c.Param("seasonId"))
	if !ok1 || !ok2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid teamId or seasonId"})
		return
	}
	if err := h.services.TeamSeasonService.Unlink(c, teamID, seasonID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unlink"})
		return
	}
	c.Status(http.StatusNoContent)
}

// GET /api/v1/team-seasons?teamId=&seasonId=&onlyActive=&page=&size=
func (h *TeamSeasonHandler) List(c *gin.Context) {
	page := parseIntDefault(c.Query("page"), 1)
	size := parseIntDefault(c.Query("size"), 25)

	var teamIDPtr *int64
	if v := c.Query("teamId"); v != "" {
		if id, ok := parseIDParam(v); ok {
			teamIDPtr = &id
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid teamId"})
			return
		}
	}

	var seasonIDPtr *int64
	if v := c.Query("seasonId"); v != "" {
		if id, ok := parseIDParam(v); ok {
			seasonIDPtr = &id
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid seasonId"})
			return
		}
	}

	var onlyActivePtr *bool
	if v := c.Query("onlyActive"); v != "" {
		if b, ok := parseBoolFlexible(v); ok {
			onlyActivePtr = &b
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid onlyActive"})
			return
		}
	}

	out, err := h.services.TeamSeasonService.List(c, services.ListTeamSeasonsOptions{
		TeamID:     teamIDPtr,
		SeasonID:   seasonIDPtr,
		OnlyActive: onlyActivePtr,
		Page:       page,
		Size:       size,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list team seasons"})
		return
	}
	c.JSON(http.StatusOK, out)
}

// GET /api/v1/teams/:id/seasons?onlyActive=
func (h *TeamSeasonHandler) ListSeasonsForTeam(c *gin.Context) {
	teamID, ok := parseIDParam(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid teamId"})
		return
	}
	var onlyActivePtr *bool
	if v := c.Query("onlyActive"); v != "" {
		if b, ok := parseBoolFlexible(v); ok {
			onlyActivePtr = &b
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid onlyActive"})
			return
		}
	}
	out, err := h.services.TeamSeasonService.ListSeasonsForTeam(c, teamID, onlyActivePtr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch seasons"})
		return
	}
	c.JSON(http.StatusOK, out)
}

// GET /api/v1/seasons/:seasonId/teams?onlyActive=
func (h *TeamSeasonHandler) ListTeamsForSeason(c *gin.Context) {
	seasonID, ok := parseIDParam(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid seasonId"})
		return
	}
	var onlyActivePtr *bool
	if v := c.Query("onlyActive"); v != "" {
		if b, ok := parseBoolFlexible(v); ok {
			onlyActivePtr = &b
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid onlyActive"})
			return
		}
	}
	out, err := h.services.TeamSeasonService.ListTeamsForSeason(c, seasonID, onlyActivePtr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch teams"})
		return
	}
	c.JSON(http.StatusOK, out)
}
