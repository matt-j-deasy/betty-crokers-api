package handlers

import (
	"github.com/matt-j-deasy/betty-crokers-api/config"
	"github.com/matt-j-deasy/betty-crokers-api/services"
)

// InitializeHandlers initializes all the handlers
func InitializeHandlers(services *services.ServicesCollection, cfg config.Environment) (*HandlersCollection, error) {
	return &HandlersCollection{
		HealthCheckHandler: NewHealthCheckHandler(*services, cfg),
		AuthHandler:        NewAuthHandler(services),
		UserHandler:        NewUserHandler(services),
		PlayerHandler:      NewPlayerHandler(services),
	}, nil
}

// HandlersCollection contains all the handlers
type HandlersCollection struct {
	HealthCheckHandler *HealthCheckHandler
	AuthHandler        *AuthHandler
	UserHandler        *UserHandler
	PlayerHandler      *PlayerHandler
}
