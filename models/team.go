package models

import (
	"time"

	"gorm.io/gorm"
)

// Team is a fixed 2-player pair. Enforce PlayerAID < PlayerBID in the service layer
type Team struct {
	ID          int64   `gorm:"primaryKey"`
	Name        string  `gorm:"not null"`
	Description *string `gorm:"type:text"`

	// Two distinct players
	PlayerAID int64 `gorm:"not null;index;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE"`
	PlayerBID int64 `gorm:"not null;index;constraint:OnDelete:RESTRICT,OnUpdate:CASCADE"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
