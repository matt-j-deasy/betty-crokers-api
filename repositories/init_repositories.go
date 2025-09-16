package repositories

import (
	"gorm.io/gorm"
)

func InitializeRepositories(db *gorm.DB) (*RepositoriesCollection, error) {

	return &RepositoriesCollection{
		UserRepo:   NewUserRepository(db),
		PlayerRepo: NewPlayerRepository(db),
		LeagueRepo: NewLeagueRepository(db),
	}, nil
}

type RepositoriesCollection struct {
	UserRepo   *UserRepository
	PlayerRepo *PlayerRepository
	LeagueRepo *LeagueRepository
}
