package repositories

import (
	"gorm.io/gorm"

	"github.com/matt-j-deasy/betty-crokers-api/models"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(u *models.User) error {
	return r.db.Create(u).Error
}

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	var u models.User
	if err := r.db.Where("email = ?", email).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) GetByID(id uint) (*models.User, error) {
	var u models.User
	if err := r.db.First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) UpdateUserRole(userID uint, newRole string) error {
	var u models.User
	if err := r.db.First(&u, userID).Error; err != nil {
		return err
	}
	u.Role = newRole
	return r.db.Save(&u).Error
}

func (r *UserRepository) UpdateUserName(userID uint, newName string) error {
	var u models.User
	if err := r.db.First(&u, userID).Error; err != nil {
		return err
	}
	u.Name = newName
	return r.db.Save(&u).Error
}
