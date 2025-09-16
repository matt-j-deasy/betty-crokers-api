package models

import (
	"time"

	"gorm.io/gorm"
)

// League is an organizational concept
type League struct {
	ID        int64  `gorm:"primaryKey"`
	Name      string `gorm:"not null;index;"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
