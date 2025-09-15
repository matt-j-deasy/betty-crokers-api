package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/matt-j-deasy/betty-crokers-api/handlers"
)

func RegisterUserRoutes(rg *gin.RouterGroup, h *handlers.UserHandler) {
	user := rg.Group("/users")
	user.GET("/:id", h.GetUser)
	user.PUT("/:id/role", h.UpdateUserRole)
	user.PUT("/:id/name", h.UpdateUserName)
}
