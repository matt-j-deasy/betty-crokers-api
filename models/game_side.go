package models

import (
	"time"

	"gorm.io/gorm"
)

// DiscColor enumerates allowed colors for a side.
type DiscColor string

const (
	DiscWhite   DiscColor = "white"
	DiscBlack   DiscColor = "black"
	DiscNatural DiscColor = "natural"
)

type GameSide struct {
	ID int64 `gorm:"primaryKey"`

	GameID int64  `gorm:"not null;index;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;uniqueIndex:uniq_game_side,priority:1"`
	Side   string `gorm:"type:char(1);not null;index;uniqueIndex:uniq_game_side,priority:2"` // "A" | "B"

	TeamID   *int64 `gorm:"index"`
	PlayerID *int64 `gorm:"index"`

	Color  DiscColor `gorm:"type:varchar(16);not null;default:natural;index"`
	Points int       `gorm:"not null;default:0"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
