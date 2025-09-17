package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/matt-j-deasy/betty-crokers-api/models"
	"github.com/matt-j-deasy/betty-crokers-api/services"
)

type GameSideHandler struct {
	services *services.ServicesCollection
}

func NewGameSideHandler(svcs *services.ServicesCollection) *GameSideHandler {
	return &GameSideHandler{services: svcs}
}

/* ===== Requests ===== */

type setSideColorReq struct {
	Color string `json:"color" binding:"required"` // white|black|natural
}

type addPointsReq struct {
	Delta *int `json:"delta" binding:"required,gte=0"`
}

type setPointsReq struct {
	Points *int `json:"points" binding:"required,gte=0"`
}

/* ===== Handlers ===== */

// GET /api/v1/games/:id/sides
func (h *GameSideHandler) ListByGame(c *gin.Context) {
	gameID, ok := parseID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game id"})
		return
	}
	out, err := h.services.GameSideService.ListByGame(c, gameID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list sides"})
		return
	}
	c.JSON(http.StatusOK, out)
}

// PUT /api/v1/games/:id/sides/:side/color
func (h *GameSideHandler) SetColor(c *gin.Context) {
	gameID, ok := parseID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game id"})
		return
	}
	side := strings.ToUpper(strings.TrimSpace(c.Param("side")))
	var req setSideColorReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	color := models.DiscColor(strings.ToLower(req.Color))
	out, err := h.services.GameSideService.SetColor(c, services.SetSideColorInput{
		GameID: gameID,
		Side:   side,
		Color:  color,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, out)
}

// POST /api/v1/games/:id/sides/:side/points/add
func (h *GameSideHandler) AddPoints(c *gin.Context) {
	gameID, ok := parseID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game id"})
		return
	}
	side := strings.ToUpper(strings.TrimSpace(c.Param("side")))
	var req addPointsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	game, sides, err := h.services.GameSideService.AddPoints(c, services.AddPointsInput{
		GameID: gameID,
		Side:   side,
		Delta:  *req.Delta,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"game": game, "sides": sides})
}

// PUT /api/v1/games/:id/sides/:side/points
func (h *GameSideHandler) SetPoints(c *gin.Context) {
	gameID, ok := parseID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game id"})
		return
	}
	side := strings.ToUpper(strings.TrimSpace(c.Param("side")))
	var req setPointsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	game, sides, err := h.services.GameSideService.SetPoints(c, services.SetPointsInput{
		GameID: gameID,
		Side:   side,
		Points: *req.Points,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"game": game, "sides": sides})
}
