package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/matt-j-deasy/betty-crokers-api/handlers"
)

// Public GameSide routes (no auth)
func RegisterGameSidePublicRoutes(rg *gin.RouterGroup, h *handlers.GameSideHandler) {
	g := rg.Group("/games")
	g.GET("/:id/sides", h.ListByGame) // GET /api/v1/games/:id/sides
}

// Protected GameSide routes (auth required)
func RegisterGameSideProtectedRoutes(rg *gin.RouterGroup, h *handlers.GameSideHandler) {
	g := rg.Group("/games")
	// :id is game id; :side is "A" or "B"
	g.PUT("/:id/sides/:side/color", h.SetColor)        // PUT /api/v1/games/:id/sides/:side/color
	g.POST("/:id/sides/:side/points/add", h.AddPoints) // POST /api/v1/games/:id/sides/:side/points/add
	g.PUT("/:id/sides/:side/points", h.SetPoints)      // PUT /api/v1/games/:id/sides/:side/points
}
