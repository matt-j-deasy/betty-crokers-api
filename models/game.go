package models

import (
	"time"

	"gorm.io/gorm"
)

// Game can belong to a Season (league game) or be standalone (exhibition).
type Game struct {
	ID int64 `gorm:"primaryKey"`

	// Nullable: if NULL, this is an exhibition game.
	SeasonID *int64 `gorm:"index;constraint:OnDelete:SET NULL,OnUpdate:CASCADE"`

	// "teams" or "players" — both sides must be the same kind; enforce in service.
	MatchType string `gorm:"type:varchar(16);not null;default:players;index"`

	// Scoring
	TargetPoints int     `gorm:"not null;default:100"`                              // first to 100
	Status       string  `gorm:"type:varchar(16);not null;default:scheduled;index"` // scheduled|in_progress|completed|canceled
	WinnerSide   *string `gorm:"type:char(1)"`                                      // "A" or "B" when completed

	// Scheduling
	ScheduledAt *time.Time
	StartedAt   *time.Time
	EndedAt     *time.Time
	Timezone    string `gorm:"not null;default:America/New_York"`

	// Optional metadata
	Location    *string
	Description *string

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
