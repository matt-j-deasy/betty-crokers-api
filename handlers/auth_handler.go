package handlers

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/matt-j-deasy/betty-crokers-api/services"
)

type AuthHandler struct {
	svcs *services.ServicesCollection
}

func NewAuthHandler(svcs *services.ServicesCollection) *AuthHandler {
	return &AuthHandler{svcs: svcs}
}

type loginReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=128"`
}
type registerReq = loginReq

func (h *AuthHandler) Register(c *gin.Context) {
	slog.Info("Register endpoint called")

	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Warn("invalid register payload", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	u, err := h.svcs.AuthService.Register(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	slog.Info("User registered", slog.String("email", req.Email))
	c.JSON(http.StatusCreated, gin.H{"user": u})
}

func (h *AuthHandler) Login(c *gin.Context) {
	slog.Info("Login endpoint called")

	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	token, exp, u, err := h.svcs.AuthService.Login(req.Email, req.Password)
	if err != nil {
		// don't leak whether email exists
		slog.Debug("Failed login attempt", slog.String("email", req.Email))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	slog.Debug("User logged in", slog.String("email", req.Email))
	c.JSON(http.StatusOK, gin.H{
		"token":     token,
		"expiresAt": exp.UTC().Format(time.RFC3339),
		"user":      u,
	})
}
