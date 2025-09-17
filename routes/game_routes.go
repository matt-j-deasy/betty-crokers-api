package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/matt-j-deasy/betty-crokers-api/handlers"
)

// Public Game routes (no auth)
func RegisterGamePublicRoutes(rg *gin.RouterGroup, h *handlers.GameHandler) {
	g := rg.Group("/games")
	g.GET("", h.List)                        // GET /api/v1/games
	g.GET("/:id", h.Get)                     // GET /api/v1/games/:id
	g.GET("/:id/with-sides", h.GetWithSides) // GET /api/v1/games/:id/with-sides
}

// Protected Game routes (auth required)
func RegisterGameProtectedRoutes(rg *gin.RouterGroup, h *handlers.GameHandler) {
	g := rg.Group("/games")
	g.POST("", h.Create)    // POST /api/v1/games
	g.PUT("/:id", h.Update) // PUT /api/v1/games/:id
	g.DELETE("/:id", h.Delete)
	g.POST("/:id/complete", h.Complete) // POST /api/v1/games/:id/complete
}
