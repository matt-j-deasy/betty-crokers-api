package services

import (
	"github.com/matt-j-deasy/betty-crokers-api/models"
	"github.com/matt-j-deasy/betty-crokers-api/repositories"
)

type UserService struct {
	users *repositories.UserRepository
}

func NewUserService(repos *repositories.RepositoriesCollection) *UserService {
	return &UserService{
		users: repos.UserRepo,
	}
}

func (s *UserService) GetByID(id uint) (*models.User, error) {
	return s.users.GetByID(id)
}

func (s *UserService) UpdateUserRole(userID uint, newRole string) error {
	return s.users.UpdateUserRole(userID, newRole)
}

func (s *UserService) UpdateUserName(userID uint, newName string) error {
	return s.users.UpdateUserName(userID, newName)
}
