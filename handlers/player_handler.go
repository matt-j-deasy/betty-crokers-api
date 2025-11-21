package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/matt-j-deasy/betty-crokers-api/services"
)

type PlayerHandler struct {
	services *services.ServicesCollection
}

func NewPlayerHandler(svcs *services.ServicesCollection) *PlayerHandler {
	return &PlayerHandler{services: svcs}
}

type createPlayerReq struct {
	UserID    *int64  `json:"userId"`
	Nickname  string  `json:"nickname" binding:"required,min=1,max=120"`
	FirstName *string `json:"firstName"`
	LastName  *string `json:"lastName"`
}

func (h *PlayerHandler) Create(c *gin.Context) {
	var req createPlayerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	p, err := h.services.PlayerService.Create(c, services.CreatePlayerInput{
		UserID:    req.UserID,
		Nickname:  req.Nickname,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, p)
}

func (h *PlayerHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid player ID"})
		return
	}
	p, err := h.services.PlayerService.GetByID(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "player not found"})
		return
	}
	c.JSON(http.StatusOK, p)
}

type updatePlayerReq struct {
	UserID    *int64  `json:"userId"`
	Nickname  *string `json:"nickname" binding:"omitempty,min=1,max=120"`
	FirstName *string `json:"firstName"`
	LastName  *string `json:"lastName"`
}

func (h *PlayerHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid player ID"})
		return
	}
	var req updatePlayerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	p, err := h.services.PlayerService.Update(c, id, services.UpdatePlayerInput{
		UserID:    req.UserID,
		Nickname:  req.Nickname,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *PlayerHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid player ID"})
		return
	}
	if err := h.services.PlayerService.Delete(c, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *PlayerHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "25"))
	search := c.DefaultQuery("q", "")

	out, err := h.services.PlayerService.List(c, services.ListPlayersOptions{
		Search: search,
		Page:   page,
		Size:   size,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list players"})
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *PlayerHandler) ListPlayerDuplicateGames(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid player ID",
		})
		return
	}

	rows, err := h.services.PlayerService.ListPlayerDuplicateGames(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch duplicate games",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"playerId": id,
		"data":     rows,
	})
}
