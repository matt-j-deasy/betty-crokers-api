package handlers

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/matt-j-deasy/betty-crokers-api/repositories"
	"github.com/matt-j-deasy/betty-crokers-api/services"
)

type SeasonHandler struct {
	services *services.ServicesCollection
}

func NewSeasonHandler(svcs *services.ServicesCollection) *SeasonHandler {
	return &SeasonHandler{services: svcs}
}

type createSeasonReq struct {
	LeagueID    int64   `json:"leagueId" binding:"required"`
	Name        string  `json:"name" binding:"required,min=1,max=200"`
	StartsOn    string  `json:"startsOn"` // "YYYY-MM-DD"
	EndsOn      string  `json:"endsOn"`   // "YYYY-MM-DD"
	Timezone    *string `json:"timezone"` // IANA
	Description *string `json:"description"`
}

func (h *SeasonHandler) Create(c *gin.Context) {
	var req createSeasonReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	out, err := h.services.SeasonService.Create(c, services.CreateSeasonInput{
		LeagueID:    req.LeagueID,
		Name:        req.Name,
		StartsOn:    req.StartsOn,
		EndsOn:      req.EndsOn,
		Timezone:    req.Timezone,
		Description: req.Description,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, out)
}

func (h *SeasonHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid season ID"})
		return
	}
	out, err := h.services.SeasonService.GetByID(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "season not found"})
		return
	}
	c.JSON(http.StatusOK, out)
}

type updateSeasonReq struct {
	LeagueID    *int64  `json:"leagueId"`
	Name        *string `json:"name" binding:"omitempty,min=1,max=200"`
	StartsOn    *string `json:"startsOn"` // "YYYY-MM-DD"
	EndsOn      *string `json:"endsOn"`   // "YYYY-MM-DD"
	Timezone    *string `json:"timezone"` // IANA
	Description *string `json:"description"`
}

func (h *SeasonHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid season ID"})
		return
	}
	var req updateSeasonReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	out, err := h.services.SeasonService.Update(c, id, services.UpdateSeasonInput{
		LeagueID:    req.LeagueID,
		Name:        req.Name,
		StartsOn:    req.StartsOn,
		EndsOn:      req.EndsOn,
		Timezone:    req.Timezone,
		Description: req.Description,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *SeasonHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid season ID"})
		return
	}
	if err := h.services.SeasonService.Delete(c, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *SeasonHandler) List(c *gin.Context) {
	page := parseIntDefault(c.Query("page"), 1)
	size := parseIntDefault(c.Query("size"), 25)
	search := c.DefaultQuery("q", "")

	var leagueIDPtr *int64
	if leagueStr := c.Query("leagueId"); leagueStr != "" {
		if v, err := strconv.ParseInt(leagueStr, 10, 64); err == nil && v > 0 {
			leagueIDPtr = &v
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid leagueId"})
			return
		}
	}

	out, err := h.services.SeasonService.List(c, services.ListSeasonsOptions{
		Search:   search,
		LeagueID: leagueIDPtr,
		Page:     page,
		Size:     size,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list seasons"})
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *SeasonHandler) Standings(c *gin.Context) {
	seasonID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || seasonID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid season ID"})
		return
	}
	rows, err := h.services.SeasonService.GetStandings(c, seasonID)
	if err != nil {
		// If season missing, your GetByID returns an error → 404
		c.JSON(http.StatusNotFound, gin.H{"error": "season not found"})
		return
	}
	c.JSON(http.StatusOK, rows)
}

func (h *SeasonHandler) ListPlayerStandings(c *gin.Context) {
	seasonID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || seasonID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid season ID"})
		return
	}

	limit := 50
	if s := c.Query("limit"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			if v < 1 {
				v = 1
			}
			if v > 200 {
				v = 200
			}
			limit = v
		}
	}

	var cur *repositories.PlayerStandingsCursor
	if s := c.Query("cursor"); s != "" {
		dec, err := base64.StdEncoding.DecodeString(s)
		if err == nil {
			var tmp repositories.PlayerStandingsCursor
			if json.Unmarshal(dec, &tmp) == nil {
				cur = &tmp
			}
		}
	}

	rows, next, err := h.services.SeasonService.ListPlayerStandings(c.Request.Context(), seasonID, limit, cur)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to compute standings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"seasonId":    seasonID,
		"standings":   rows,
		"next_cursor": next,
	})
}
