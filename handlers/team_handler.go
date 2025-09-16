package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/matt-j-deasy/betty-crokers-api/services"
)

type TeamHandler struct {
	services *services.ServicesCollection
}

func NewTeamHandler(svcs *services.ServicesCollection) *TeamHandler {
	return &TeamHandler{services: svcs}
}

type createTeamReq struct {
	Name        string  `json:"name" binding:"required,min=1,max=200"`
	Description *string `json:"description"`
	PlayerAID   int64   `json:"playerAId" binding:"required,gt=0"`
	PlayerBID   int64   `json:"playerBId" binding:"required,gt=0"`
}

func (h *TeamHandler) Create(c *gin.Context) {
	var req createTeamReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	out, err := h.services.TeamService.Create(c, services.CreateTeamInput{
		Name:        req.Name,
		Description: req.Description,
		PlayerAID:   req.PlayerAID,
		PlayerBID:   req.PlayerBID,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, out)
}

func (h *TeamHandler) Get(c *gin.Context) {
	id, ok := parseIDParam(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid team ID"})
		return
	}
	out, err := h.services.TeamService.GetByID(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "team not found"})
		return
	}
	c.JSON(http.StatusOK, out)
}

type updateTeamReq struct {
	Name        *string `json:"name" binding:"omitempty,min=1,max=200"`
	Description *string `json:"description"`
	PlayerAID   *int64  `json:"playerAId" binding:"omitempty,gt=0"`
	PlayerBID   *int64  `json:"playerBId" binding:"omitempty,gt=0"`
}

func (h *TeamHandler) Update(c *gin.Context) {
	id, ok := parseIDParam(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid team ID"})
		return
	}
	var req updateTeamReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	out, err := h.services.TeamService.Update(c, id, services.UpdateTeamInput{
		Name:        req.Name,
		Description: req.Description,
		PlayerAID:   req.PlayerAID,
		PlayerBID:   req.PlayerBID,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *TeamHandler) Delete(c *gin.Context) {
	id, ok := parseIDParam(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid team ID"})
		return
	}
	if err := h.services.TeamService.Delete(c, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete"})
		return
	}
	c.Status(http.StatusNoContent)
}

// GET /api/v1/teams?q=&page=&size=&playerId=&seasonId=&onlyActive=
func (h *TeamHandler) List(c *gin.Context) {
	page := parseIntDefault(c.Query("page"), 1)
	size := parseIntDefault(c.Query("size"), 25)
	search := strings.TrimSpace(c.DefaultQuery("q", ""))

	var playerIDPtr *int64
	if v := c.Query("playerId"); v != "" {
		if id, ok := parseIDParam(v); ok {
			playerIDPtr = &id
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid playerId"})
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

	out, err := h.services.TeamService.List(c, services.ListTeamsOptions{
		Search:     search,
		PlayerID:   playerIDPtr,
		SeasonID:   seasonIDPtr,
		OnlyActive: onlyActivePtr,
		Page:       page,
		Size:       size,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list teams"})
		return
	}
	c.JSON(http.StatusOK, out)
}

// ---- helpers ----

func parseIDParam(s string) (int64, bool) {
	id, err := strconv.ParseInt(s, 10, 64)
	return id, err == nil && id > 0
}

func parseBoolFlexible(s string) (bool, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "true", "t", "yes", "y":
		return true, true
	case "0", "false", "f", "no", "n":
		return false, true
	default:
		return false, false
	}
}
