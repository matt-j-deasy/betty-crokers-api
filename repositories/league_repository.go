package repositories

import (
	"context"
	"strings"

	"gorm.io/gorm"

	"github.com/matt-j-deasy/betty-crokers-api/models"
)

type LeagueRepository struct {
	db *gorm.DB
}

func NewLeagueRepository(db *gorm.DB) *LeagueRepository {
	return &LeagueRepository{db: db}
}

type ListLeaguesFilter struct {
	Search string
	Offset int
	Limit  int
}

func (r *LeagueRepository) Create(ctx context.Context, l *models.League) error {
	return r.db.WithContext(ctx).Create(l).Error
}

func (r *LeagueRepository) GetByID(ctx context.Context, id int64) (*models.League, error) {
	var l models.League
	if err := r.db.WithContext(ctx).First(&l, id).Error; err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *LeagueRepository) UpdateFields(ctx context.Context, id int64, fields map[string]any) (*models.League, error) {
	if err := r.db.WithContext(ctx).
		Model(&models.League{}).
		Where("id = ?", id).
		Updates(fields).Error; err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

func (r *LeagueRepository) DeleteByID(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&models.League{}, id).Error
}

func (r *LeagueRepository) List(ctx context.Context, f ListLeaguesFilter) ([]models.League, int64, error) {
	var (
		items []models.League
		total int64
	)

	q := r.db.WithContext(ctx).Model(&models.League{}).Where("deleted_at IS NULL")

	if s := strings.TrimSpace(f.Search); s != "" {
		ilike := "%" + strings.ToLower(s) + "%"
		q = q.Where("LOWER(name) LIKE ?", ilike)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 25
	}
	if f.Offset < 0 {
		f.Offset = 0
	}

	if err := q.Order("id DESC").Limit(f.Limit).Offset(f.Offset).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}
