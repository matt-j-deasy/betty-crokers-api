package models

import (
	"time"

	"gorm.io/gorm"
)

// Season belongs to one League and will own many Games.
type Season struct {
	ID       int64  `gorm:"primaryKey"`
	LeagueID int64  `gorm:"not null;index;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	Name     string `gorm:"not null;index;"`

	StartsOn time.Time `gorm:"type:date"`
	EndsOn   time.Time `gorm:"type:date"`

	// IANA TZ for scheduling (e.g., "America/New_York"). Defaults to America/New_York.
	Timezone string `gorm:"not null;default:America/New_York"`

	// Human-readable context.
	Description *string

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
