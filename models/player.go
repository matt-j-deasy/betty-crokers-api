package models

import (
	"time"

	"gorm.io/gorm"
)

type Player struct {
	ID        int64  `gorm:"primaryKey"`
	UserID    *int64 `gorm:"index"`
	Nickname  string `gorm:"not null;index"`
	FirstName *string
	LastName  *string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type PlayerTeamMembership struct {
	ID           int64 `gorm:"primaryKey"`
	PlayerID     int64 `gorm:"index"`
	TeamID       int64 `gorm:"index"`
	SeasonID     int64 `gorm:"index"`
	JerseyNumber *int
	IsActive     bool `gorm:"not null;default:true"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}
