package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/matt-j-deasy/betty-crokers-api/handlers"
)

func RegisterAuthRoutes(rg *gin.RouterGroup, h *handlers.AuthHandler) {
	auth := rg.Group("/auth")
	auth.POST("/register", h.Register)
	auth.POST("/login", h.Login)
}
