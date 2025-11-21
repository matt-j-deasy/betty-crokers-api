package repositories

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/matt-j-deasy/betty-crokers-api/models"
)

type GameRepository struct {
	db *gorm.DB
}

func NewGameRepository(db *gorm.DB) *GameRepository {
	return &GameRepository{db: db}
}

type ListGamesFilter struct {
	SeasonID       *int64     // if nil: all; if set: filter league games by season
	ExhibitionOnly *bool      // when true: only games with season_id IS NULL
	Status         []string   // scheduled|in_progress|completed|canceled
	MatchType      *string    // "teams" | "players"
	ScheduledFrom  *time.Time // filter by scheduled_at >=
	ScheduledTo    *time.Time // filter by scheduled_at <=
	TeamID         *int64     // any game where a side has this team_id
	PlayerID       *int64     // any game where a side has this player_id
	Offset         int
	Limit          int
	OrderBy        string // e.g., "scheduled_at desc", defaults to "games.id desc"
}

func (r *GameRepository) Create(ctx context.Context, g *models.Game) error {
	return r.db.WithContext(ctx).Create(g).Error
}

func (r *GameRepository) GetByID(ctx context.Context, id int64) (*models.Game, error) {
	var g models.Game
	if err := r.db.WithContext(ctx).First(&g, id).Error; err != nil {
		return nil, err
	}
	return &g, nil
}

// GetWithSides fetches a game and its two sides.
func (r *GameRepository) GetWithSides(ctx context.Context, id int64) (*models.Game, []models.GameSide, error) {
	g, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	var sides []models.GameSide
	if err := r.db.WithContext(ctx).
		Where("game_id = ? AND deleted_at IS NULL", id).
		Order("side asc").
		Find(&sides).Error; err != nil {
		return nil, nil, err
	}
	return g, sides, nil
}

func (r *GameRepository) UpdateFields(ctx context.Context, id int64, fields map[string]any) (*models.Game, error) {
	if err := r.db.WithContext(ctx).
		Model(&models.Game{}).
		Where("id = ?", id).
		Updates(fields).Error; err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

func (r *GameRepository) DeleteByID(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&models.Game{}, id).Error
}

// CreateWithSides runs in a single transaction: creates the game and two sides.
func (r *GameRepository) CreateWithSides(ctx context.Context, g *models.Game, sides []models.GameSide) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(g).Error; err != nil {
			return err
		}
		for i := range sides {
			sides[i].GameID = g.ID
		}
		if err := tx.Create(&sides).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *GameRepository) List(ctx context.Context, f ListGamesFilter) ([]models.Game, int64, error) {
	var (
		items []models.Game
		total int64
	)

	// Base builder
	base := r.db.WithContext(ctx).Model(&models.Game{}).Where("games.deleted_at IS NULL")

	apply := func(q *gorm.DB) *gorm.DB {
		if f.SeasonID != nil {
			q = q.Where("games.season_id = ?", *f.SeasonID)
		} else if f.ExhibitionOnly != nil && *f.ExhibitionOnly {
			q = q.Where("games.season_id IS NULL")
		}
		if len(f.Status) > 0 {
			q = q.Where("games.status IN ?", f.Status)
		}
		if f.MatchType != nil && *f.MatchType != "" {
			q = q.Where("games.match_type = ?", *f.MatchType)
		}
		if f.ScheduledFrom != nil {
			q = q.Where("games.scheduled_at >= ?", *f.ScheduledFrom)
		}
		if f.ScheduledTo != nil {
			q = q.Where("games.scheduled_at <= ?", *f.ScheduledTo)
		}
		// Participant filters via joins
		if f.TeamID != nil && *f.TeamID > 0 {
			q = q.Joins(`JOIN game_sides gs_t ON gs_t.game_id = games.id AND gs_t.team_id = ? AND gs_t.deleted_at IS NULL`, *f.TeamID)
		}
		if f.PlayerID != nil && *f.PlayerID > 0 {
			q = q.Joins(`JOIN game_sides gs_p ON gs_p.game_id = games.id AND gs_p.player_id = ? AND gs_p.deleted_at IS NULL`, *f.PlayerID)
		}
		return q
	}

	// Count with DISTINCT id (avoid join duplicates)
	countQ := apply(base.Session(&gorm.Session{}))
	if err := countQ.Distinct("games.id").Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Pagination defaults
	limit := f.Limit
	if limit <= 0 || limit > 100 {
		limit = 25
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}
	orderBy := f.OrderBy
	if orderBy == "" {
		orderBy = "games.id desc"
	}

	// Items query (DISTINCT over full row)
	itemsQ := apply(base.Session(&gorm.Session{}))
	if err := itemsQ.
		Select("games.*").
		Distinct().
		Order(orderBy).
		Limit(limit).
		Offset(offset).
		Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *GameRepository) UpdateSideColor(
	ctx context.Context,
	gameID int64,
	side string, // "A" or "B"
	color models.DiscColor,
) error {
	return r.db.WithContext(ctx).
		Model(&models.GameSide{}).
		Where("game_id = ? AND side = ?", gameID, side).
		Updates(map[string]any{
			"color": color,
		}).Error
}
