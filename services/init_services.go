package services

import (
	"github.com/matt-j-deasy/betty-crokers-api/config"
	"github.com/matt-j-deasy/betty-crokers-api/repositories"
)

func InitializeServices(
	repos *repositories.RepositoriesCollection,
	cfg config.Environment,
) (*ServicesCollection, error) {
	return &ServicesCollection{
		AuthService:   NewAuthService(repos, cfg),
		UserService:   NewUserService(repos),
		PlayerService: NewPlayerService(repos),
	}, nil
}

type ServicesCollection struct {
	AuthService   *AuthService
	UserService   *UserService
	PlayerService *PlayerService
}
