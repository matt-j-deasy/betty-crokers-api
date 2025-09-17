package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/matt-j-deasy/betty-crokers-api/models"
	"github.com/matt-j-deasy/betty-crokers-api/services"
)

type GameHandler struct {
	services *services.ServicesCollection
}

func NewGameHandler(svcs *services.ServicesCollection) *GameHandler {
	return &GameHandler{services: svcs}
}

/* ===== Requests ===== */

type gameParticipantReq struct {
	TeamID   *int64  `json:"teamId"`
	PlayerID *int64  `json:"playerId"`
	Color    *string `json:"color"` // "white" | "black" | "natural"
}

type createGameReq struct {
	SeasonID     *int64             `json:"seasonId"`
	MatchType    string             `json:"matchType" binding:"required,oneof=teams players"`
	TargetPoints *int               `json:"targetPoints"`
	ScheduledAt  *string            `json:"scheduledAt"` // RFC3339
	Timezone     *string            `json:"timezone"`    // IANA
	Location     *string            `json:"location"`
	Description  *string            `json:"description"`
	SideA        gameParticipantReq `json:"sideA" binding:"required"`
	SideB        gameParticipantReq `json:"sideB" binding:"required"`
}

type updateGameReq struct {
	SeasonID     *int64  `json:"seasonId"`
	TargetPoints *int    `json:"targetPoints"`
	ScheduledAt  *string `json:"scheduledAt"` // RFC3339 or "" to clear
	Timezone     *string `json:"timezone"`
	Location     *string `json:"location"`    // send null to clear
	Description  *string `json:"description"` // send null to clear
	Status       *string `json:"status"`      // scheduled|in_progress|canceled|completed
}

type completeReq struct {
	WinnerSide string `json:"winnerSide" binding:"required,oneof=A B"`
}

/* ===== Handlers ===== */

func (h *GameHandler) Create(c *gin.Context) {
	var req createGameReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	var colorA, colorB *models.DiscColor
	if req.SideA.Color != nil && *req.SideA.Color != "" {
		v := models.DiscColor(strings.ToLower(*req.SideA.Color))
		colorA = &v
	}
	if req.SideB.Color != nil && *req.SideB.Color != "" {
		v := models.DiscColor(strings.ToLower(*req.SideB.Color))
		colorB = &v
	}

	game, sides, err := h.services.GameService.Create(c, services.CreateGameInput{
		SeasonID:     req.SeasonID,
		MatchType:    req.MatchType,
		TargetPoints: req.TargetPoints,
		ScheduledAt:  req.ScheduledAt,
		Timezone:     req.Timezone,
		Location:     req.Location,
		Description:  req.Description,
		SideA: services.GameParticipantInput{
			TeamID:   req.SideA.TeamID,
			PlayerID: req.SideA.PlayerID,
			Color:    colorA,
		},
		SideB: services.GameParticipantInput{
			TeamID:   req.SideB.TeamID,
			PlayerID: req.SideB.PlayerID,
			Color:    colorB,
		},
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"game": game, "sides": sides})
}

func (h *GameHandler) Get(c *gin.Context) {
	id, ok := parseID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game id"})
		return
	}
	g, err := h.services.GameService.GetByID(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}
	c.JSON(http.StatusOK, g)
}

func (h *GameHandler) GetWithSides(c *gin.Context) {
	id, ok := parseID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game id"})
		return
	}
	g, sides, err := h.services.GameService.GetWithSides(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"game": g, "sides": sides})
}

func (h *GameHandler) Update(c *gin.Context) {
	id, ok := parseID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game id"})
		return
	}
	var req updateGameReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	out, err := h.services.GameService.Update(c, id, services.UpdateGameInput{
		SeasonID:     req.SeasonID,
		TargetPoints: req.TargetPoints,
		ScheduledAt:  req.ScheduledAt,
		Timezone:     req.Timezone,
		Location:     req.Location,
		Description:  req.Description,
		Status:       req.Status,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *GameHandler) Delete(c *gin.Context) {
	id, ok := parseID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game id"})
		return
	}
	if err := h.services.GameService.Delete(c, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete"})
		return
	}
	c.Status(http.StatusNoContent)
}

// GET /api/v1/games?seasonId=&exhibitionOnly=&status=&matchType=&scheduledFrom=&scheduledTo=&teamId=&playerId=&page=&size=&orderBy=
func (h *GameHandler) List(c *gin.Context) {
	page := parseIntDefault(c.Query("page"), 1)
	size := parseIntDefault(c.Query("size"), 25)
	orderBy := strings.TrimSpace(c.DefaultQuery("orderBy", "games.id desc"))

	var seasonIDPtr *int64
	if v := c.Query("seasonId"); v != "" {
		if id, ok := parseID(v); ok {
			seasonIDPtr = &id
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid seasonId"})
			return
		}
	}

	var exhibitionOnlyPtr *bool
	if v := c.Query("exhibitionOnly"); v != "" {
		if b, ok := parseBoolFlexible(v); ok {
			exhibitionOnlyPtr = &b
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid exhibitionOnly"})
			return
		}
	}

	var matchTypePtr *string
	if v := strings.ToLower(strings.TrimSpace(c.Query("matchType"))); v != "" {
		if v != "teams" && v != "players" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "matchType must be 'teams' or 'players'"})
			return
		}
		matchTypePtr = &v
	}

	// Status can be comma-separated: e.g., "scheduled,in_progress"
	var statuses []string
	if raw := c.Query("status"); raw != "" {
		for _, s := range strings.Split(raw, ",") {
			ss := strings.ToLower(strings.TrimSpace(s))
			if ss == "" {
				continue
			}
			switch ss {
			case "scheduled", "in_progress", "completed", "canceled":
				statuses = append(statuses, ss)
			default:
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status: " + ss})
				return
			}
		}
	}

	var scheduledFromPtr, scheduledToPtr *time.Time
	if v := strings.TrimSpace(c.Query("scheduledFrom")); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "scheduledFrom must be RFC3339"})
			return
		}
		scheduledFromPtr = &t
	}
	if v := strings.TrimSpace(c.Query("scheduledTo")); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "scheduledTo must be RFC3339"})
			return
		}
		scheduledToPtr = &t
	}

	var teamIDPtr, playerIDPtr *int64
	if v := c.Query("teamId"); v != "" {
		if id, ok := parseID(v); ok {
			teamIDPtr = &id
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid teamId"})
			return
		}
	}
	if v := c.Query("playerId"); v != "" {
		if id, ok := parseID(v); ok {
			playerIDPtr = &id
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid playerId"})
			return
		}
	}

	out, err := h.services.GameService.List(c, services.ListGamesOptions{
		SeasonID:       seasonIDPtr,
		ExhibitionOnly: exhibitionOnlyPtr,
		Status:         statuses,
		MatchType:      matchTypePtr,
		ScheduledFrom:  scheduledFromPtr,
		ScheduledTo:    scheduledToPtr,
		TeamID:         teamIDPtr,
		PlayerID:       playerIDPtr,
		Page:           page,
		Size:           size,
		OrderBy:        orderBy,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list games"})
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *GameHandler) Complete(c *gin.Context) {
	id, ok := parseID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game id"})
		return
	}
	var req completeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	out, err := h.services.GameService.CompleteWithWinner(c, id, req.WinnerSide)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, out)
}

/* ===== helpers ===== */

func parseID(s string) (int64, bool) {
	id, err := strconv.ParseInt(s, 10, 64)
	return id, err == nil && id > 0
}
