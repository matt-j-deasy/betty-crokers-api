package repositories

import (
	"gorm.io/gorm"
)

func InitializeRepositories(db *gorm.DB) (*RepositoriesCollection, error) {

	return &RepositoriesCollection{
		UserRepo:       NewUserRepository(db),
		PlayerRepo:     NewPlayerRepository(db),
		LeagueRepo:     NewLeagueRepository(db),
		SeasonRepo:     NewSeasonRepository(db),
		TeamRepo:       NewTeamRepository(db),
		TeamSeasonRepo: NewTeamSeasonRepository(db),
		GameRepo:       NewGameRepository(db),
		GameSideRepo:   NewGameSideRepository(db),
	}, nil
}

type RepositoriesCollection struct {
	UserRepo       *UserRepository
	PlayerRepo     *PlayerRepository
	LeagueRepo     *LeagueRepository
	SeasonRepo     *SeasonRepository
	TeamRepo       *TeamRepository
	TeamSeasonRepo *TeamSeasonRepository
	GameRepo       *GameRepository
	GameSideRepo   *GameSideRepository
}
