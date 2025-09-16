package repositories

import (
	"context"
	"strings"

	"gorm.io/gorm"

	"github.com/matt-j-deasy/betty-crokers-api/models"
)

type PlayerRepository struct {
	db *gorm.DB
}

func NewPlayerRepository(db *gorm.DB) *PlayerRepository {
	return &PlayerRepository{db: db}
}

type ListPlayersFilter struct {
	Search string
	Offset int
	Limit  int
}

func (r *PlayerRepository) Create(ctx context.Context, p *models.Player) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *PlayerRepository) GetByID(ctx context.Context, id int64) (*models.Player, error) {
	var p models.Player
	if err := r.db.WithContext(ctx).First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PlayerRepository) UpdateFields(ctx context.Context, id int64, fields map[string]any) (*models.Player, error) {
	var p models.Player
	tx := r.db.WithContext(ctx).Model(&models.Player{}).Where("id = ?", id).Updates(fields)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if err := r.db.WithContext(ctx).First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PlayerRepository) DeleteByID(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&models.Player{}, id).Error
}

func (r *PlayerRepository) List(ctx context.Context, f ListPlayersFilter) ([]models.Player, int64, error) {
	var (
		items []models.Player
		total int64
	)

	q := r.db.WithContext(ctx).Model(&models.Player{})
	// Soft-delete friendly count
	q.Where("deleted_at IS NULL")

	if s := strings.TrimSpace(f.Search); s != "" {
		// Case-insensitive match on nickname OR first/last name
		ilike := "%" + strings.ToLower(s) + "%"
		q = q.Where(
			"LOWER(nickname) LIKE ? OR LOWER(COALESCE(first_name,'')) LIKE ? OR LOWER(COALESCE(last_name,'')) LIKE ?",
			ilike, ilike, ilike,
		)
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
