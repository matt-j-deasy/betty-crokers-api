package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/matt-j-deasy/betty-crokers-api/services"
)

type LeagueHandler struct {
	services *services.ServicesCollection
}

func NewLeagueHandler(svcs *services.ServicesCollection) *LeagueHandler {
	return &LeagueHandler{services: svcs}
}

type createLeagueReq struct {
	Name string `json:"name" binding:"required,min=1,max=200"`
}

func (h *LeagueHandler) Create(c *gin.Context) {
	var req createLeagueReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	l, err := h.services.LeagueService.Create(c, services.CreateLeagueInput{
		Name: req.Name,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, l)
}

func (h *LeagueHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid league ID"})
		return
	}
	l, err := h.services.LeagueService.GetByID(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "league not found"})
		return
	}
	c.JSON(http.StatusOK, l)
}

type updateLeagueReq struct {
	Name *string `json:"name" binding:"omitempty,min=1,max=200"`
}

func (h *LeagueHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid league ID"})
		return
	}
	var req updateLeagueReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	l, err := h.services.LeagueService.Update(c, id, services.UpdateLeagueInput{
		Name: req.Name,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, l)
}

func (h *LeagueHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid league ID"})
		return
	}
	if err := h.services.LeagueService.Delete(c, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *LeagueHandler) List(c *gin.Context) {
	page := parseIntDefault(c.Query("page"), 1)
	size := parseIntDefault(c.Query("size"), 25)
	search := c.DefaultQuery("q", "")

	out, err := h.services.LeagueService.List(c, services.ListLeaguesOptions{
		Search: search,
		Page:   page,
		Size:   size,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list leagues"})
		return
	}
	c.JSON(http.StatusOK, out)
}

// small helper (avoid duplicate strconv code)
func parseIntDefault(s string, def int) int {
	if s == "" {
		return def
	}
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return def
}
