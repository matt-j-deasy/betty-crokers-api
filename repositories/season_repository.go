package repositories

import (
	"context"
	"strings"

	"gorm.io/gorm"

	"github.com/matt-j-deasy/betty-crokers-api/models"
)

type SeasonRepository struct {
	db *gorm.DB
}

func NewSeasonRepository(db *gorm.DB) *SeasonRepository {
	return &SeasonRepository{db: db}
}

type ListSeasonsFilter struct {
	Search   string
	LeagueID *int64
	Offset   int
	Limit    int
}

func (r *SeasonRepository) Create(ctx context.Context, s *models.Season) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *SeasonRepository) GetByID(ctx context.Context, id int64) (*models.Season, error) {
	var s models.Season
	if err := r.db.WithContext(ctx).First(&s, id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SeasonRepository) UpdateFields(ctx context.Context, id int64, fields map[string]any) (*models.Season, error) {
	if err := r.db.WithContext(ctx).
		Model(&models.Season{}).
		Where("id = ?", id).
		Updates(fields).Error; err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

func (r *SeasonRepository) DeleteByID(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&models.Season{}, id).Error
}

func (r *SeasonRepository) List(ctx context.Context, f ListSeasonsFilter) ([]models.Season, int64, error) {
	var (
		items []models.Season
		total int64
	)

	q := r.db.WithContext(ctx).Model(&models.Season{}).Where("deleted_at IS NULL")

	if f.LeagueID != nil {
		q = q.Where("league_id = ?", *f.LeagueID)
	}

	if s := strings.TrimSpace(f.Search); s != "" {
		ilike := "%" + strings.ToLower(s) + "%"
		q = q.Where("LOWER(name) LIKE ?", ilike)
	}

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

	if err := q.Order("id DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}
