package repositories

import (
	"gorm.io/gorm"
)

func InitializeRepositories(db *gorm.DB) (*RepositoriesCollection, error) {

	return &RepositoriesCollection{
		UserRepo: NewUserRepository(db),
	}, nil
}

type RepositoriesCollection struct {
	UserRepo *UserRepository
}
