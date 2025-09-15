package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/matt-j-deasy/betty-crokers-api/services"
)

// UserHandler handles user-related routes
type UserHandler struct {
	services *services.ServicesCollection
}

// NewUserHandler creates a new instance of UserHandler
func NewUserHandler(services *services.ServicesCollection) *UserHandler {
	return &UserHandler{
		services: services,
	}
}

func (h *UserHandler) GetUser(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}
	user, err := h.services.UserService.GetByID(uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

type updateRoleReq struct {
	Role string `json:"role" binding:"required,oneof=admin user guest"`
}

// UpdateUserRole updates a user's role
func (h *UserHandler) UpdateUserRole(c *gin.Context) {
	var req updateRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	if err := h.services.UserService.UpdateUserRole(uint(userID), req.Role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
}

type updateNameReq struct {
	Name string `json:"name" binding:"required,min=2,max=100"`
}

// UpdateUserName updates a user's name
func (h *UserHandler) UpdateUserName(c *gin.Context) {
	var req updateNameReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	if err := h.services.UserService.UpdateUserName(uint(userID), req.Name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
}
