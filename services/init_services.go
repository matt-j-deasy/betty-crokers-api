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
		AuthService:       NewAuthService(repos, cfg),
		UserService:       NewUserService(repos),
		PlayerService:     NewPlayerService(repos),
		LeagueService:     NewLeagueService(repos),
		SeasonService:     NewSeasonService(repos),
		TeamService:       NewTeamService(repos),
		TeamSeasonService: NewTeamSeasonService(repos),
		GameService:       NewGameService(repos),
		GameSideService:   NewGameSideService(repos),
	}, nil
}

type ServicesCollection struct {
	AuthService       *AuthService
	UserService       *UserService
	PlayerService     *PlayerService
	LeagueService     *LeagueService
	SeasonService     *SeasonService
	TeamService       *TeamService
	TeamSeasonService *TeamSeasonService
	GameService       *GameService
	GameSideService   *GameSideService
}
