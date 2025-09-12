package repositories

import (
	"gorm.io/gorm"
)

func InitializeRepositories(db *gorm.DB) (*RepositoriesCollection, error) {

	return &RepositoriesCollection{}, nil
}

type RepositoriesCollection struct {
}
