package models

import (
	"time"

	"gorm.io/gorm"
)

// TeamSeason links a Team to a Season (many-to-many).
type TeamSeason struct {
	ID       int64 `gorm:"primaryKey"`
	TeamID   int64 `gorm:"not null;index;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	SeasonID int64 `gorm:"not null;index;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`

	// For simple enable/disable without deleting
	IsActive bool `gorm:"not null;default:true"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
