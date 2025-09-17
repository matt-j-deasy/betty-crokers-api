package repositories

import (
	"context"

	"gorm.io/gorm"

	"github.com/matt-j-deasy/betty-crokers-api/models"
)

type GameSideRepository struct {
	db *gorm.DB
}

func NewGameSideRepository(db *gorm.DB) *GameSideRepository {
	return &GameSideRepository{db: db}
}

type ListGameSidesFilter struct {
	GameID   *int64
	TeamID   *int64
	PlayerID *int64
	Color    *models.DiscColor
	Offset   int
	Limit    int
}

func (r *GameSideRepository) Create(ctx context.Context, s *models.GameSide) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *GameSideRepository) BulkCreate(ctx context.Context, sides []models.GameSide) error {
	if len(sides) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&sides).Error
}

func (r *GameSideRepository) GetByID(ctx context.Context, id int64) (*models.GameSide, error) {
	var s models.GameSide
	if err := r.db.WithContext(ctx).First(&s, id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *GameSideRepository) GetByGameAndSide(ctx context.Context, gameID int64, side string) (*models.GameSide, error) {
	var s models.GameSide
	if err := r.db.WithContext(ctx).
		Where("game_id = ? AND side = ?", gameID, side).
		First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *GameSideRepository) ListByGame(ctx context.Context, gameID int64) ([]models.GameSide, error) {
	var sides []models.GameSide
	if err := r.db.WithContext(ctx).
		Where("game_id = ? AND deleted_at IS NULL", gameID).
		Order("side asc").
		Find(&sides).Error; err != nil {
		return nil, err
	}
	return sides, nil
}

func (r *GameSideRepository) UpdateFields(ctx context.Context, id int64, fields map[string]any) (*models.GameSide, error) {
	if err := r.db.WithContext(ctx).
		Model(&models.GameSide{}).
		Where("id = ?", id).
		Updates(fields).Error; err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

func (r *GameSideRepository) UpdateFieldsByGameAndSide(ctx context.Context, gameID int64, side string, fields map[string]any) (*models.GameSide, error) {
	if err := r.db.WithContext(ctx).
		Model(&models.GameSide{}).
		Where("game_id = ? AND side = ?", gameID, side).
		Updates(fields).Error; err != nil {
		return nil, err
	}
	return r.GetByGameAndSide(ctx, gameID, side)
}

func (r *GameSideRepository) DeleteByID(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&models.GameSide{}, id).Error
}

func (r *GameSideRepository) DeleteByGame(ctx context.Context, gameID int64) error {
	return r.db.WithContext(ctx).
		Where("game_id = ?", gameID).
		Delete(&models.GameSide{}).Error
}

func (r *GameSideRepository) List(ctx context.Context, f ListGameSidesFilter) ([]models.GameSide, int64, error) {
	var (
		items []models.GameSide
		total int64
	)

	q := r.db.WithContext(ctx).Model(&models.GameSide{}).Where("deleted_at IS NULL")

	if f.GameID != nil && *f.GameID > 0 {
		q = q.Where("game_id = ?", *f.GameID)
	}
	if f.TeamID != nil && *f.TeamID > 0 {
		q = q.Where("team_id = ?", *f.TeamID)
	}
	if f.PlayerID != nil && *f.PlayerID > 0 {
		q = q.Where("player_id = ?", *f.PlayerID)
	}
	if f.Color != nil && *f.Color != "" {
		q = q.Where("color = ?", *f.Color)
	}

	// count
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := f.Limit
	if limit <= 0 || limit > 100 {
		limit = 25
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}

	if err := q.Order("game_id desc, side asc").
		Limit(limit).
		Offset(offset).
		Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}
