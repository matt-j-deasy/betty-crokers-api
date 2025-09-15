package models

import "time"

type User struct {
	ID           uint   `gorm:"primaryKey"`
	Email        string `gorm:"uniqueIndex;not null"`
	Name         string `gorm:"not null"`
	PasswordHash string `gorm:"not null"`
	Role         string `gorm:"not null;default:'user'"` // e.g., 'user', 'admin'
	Image        string `gorm:"default:'default.png'"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
