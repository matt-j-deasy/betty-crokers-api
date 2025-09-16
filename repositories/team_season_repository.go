package repositories

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/matt-j-deasy/betty-crokers-api/models"
)

type TeamSeasonRepository struct {
	db *gorm.DB
}

func NewTeamSeasonRepository(db *gorm.DB) *TeamSeasonRepository {
	return &TeamSeasonRepository{db: db}
}

type ListTeamSeasonsFilter struct {
	TeamID     *int64
	SeasonID   *int64
	OnlyActive *bool
	Offset     int
	Limit      int
}

func (r *TeamSeasonRepository) Link(ctx context.Context, teamID, seasonID int64, isActive bool) (*models.TeamSeason, error) {
	link := &models.TeamSeason{
		TeamID:   teamID,
		SeasonID: seasonID,
		IsActive: isActive,
	}
	if err := r.db.WithContext(ctx).Create(link).Error; err != nil {
		return nil, err
	}
	return link, nil
}

// UpsertLink: if a row exists (even soft-deleted), restore/update it; otherwise create one.
func (r *TeamSeasonRepository) UpsertLink(ctx context.Context, teamID, seasonID int64, isActive bool) (*models.TeamSeason, error) {
	var existing models.TeamSeason
	tx := r.db.WithContext(ctx).Unscoped().Where("team_id = ? AND season_id = ?", teamID, seasonID).First(&existing)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return r.Link(ctx, teamID, seasonID, isActive)
		}
		return nil, tx.Error
	}

	// If soft-deleted, revive it
	if existing.DeletedAt.Valid {
		if err := r.db.WithContext(ctx).Unscoped().
			Model(&existing).
			Updates(map[string]any{
				"deleted_at": nil,
				"is_active":  isActive,
			}).Error; err != nil {
			return nil, err
		}
		return &existing, nil
	}

	// Update in place
	if err := r.db.WithContext(ctx).
		Model(&existing).
		Update("is_active", isActive).Error; err != nil {
		return nil, err
	}
	return &existing, nil
}

func (r *TeamSeasonRepository) SetActive(ctx context.Context, teamID, seasonID int64, isActive bool) (*models.TeamSeason, error) {
	var link models.TeamSeason
	if err := r.db.WithContext(ctx).
		Where("team_id = ? AND season_id = ?", teamID, seasonID).
		First(&link).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).
		Model(&link).
		Update("is_active", isActive).Error; err != nil {
		return nil, err
	}
	return &link, nil
}

// Unlink = soft delete the association
func (r *TeamSeasonRepository) Unlink(ctx context.Context, teamID, seasonID int64) error {
	return r.db.WithContext(ctx).
		Where("team_id = ? AND season_id = ?", teamID, seasonID).
		Delete(&models.TeamSeason{}).Error
}

func (r *TeamSeasonRepository) Get(ctx context.Context, teamID, seasonID int64) (*models.TeamSeason, error) {
	var link models.TeamSeason
	if err := r.db.WithContext(ctx).
		Where("team_id = ? AND season_id = ?", teamID, seasonID).
		First(&link).Error; err != nil {
		return nil, err
	}
	return &link, nil
}

func (r *TeamSeasonRepository) List(ctx context.Context, f ListTeamSeasonsFilter) ([]models.TeamSeason, int64, error) {
	var (
		items []models.TeamSeason
		total int64
	)

	q := r.db.WithContext(ctx).Model(&models.TeamSeason{}).Where("deleted_at IS NULL")

	if f.TeamID != nil && *f.TeamID > 0 {
		q = q.Where("team_id = ?", *f.TeamID)
	}
	if f.SeasonID != nil && *f.SeasonID > 0 {
		q = q.Where("season_id = ?", *f.SeasonID)
	}
	if f.OnlyActive != nil {
		q = q.Where("is_active = ?", *f.OnlyActive)
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

// Convenience: list Seasons linked to a Team (optionally active only)
func (r *TeamSeasonRepository) ListSeasonsForTeam(ctx context.Context, teamID int64, onlyActive *bool) ([]models.Season, error) {
	var seasons []models.Season
	q := r.db.WithContext(ctx).
		Table("seasons").
		Joins("JOIN team_seasons ts ON ts.season_id = seasons.id").
		Where("ts.team_id = ? AND seasons.deleted_at IS NULL AND ts.deleted_at IS NULL", teamID)
	if onlyActive != nil {
		q = q.Where("ts.is_active = ?", *onlyActive)
	}
	if err := q.Order("seasons.id DESC").Select("seasons.*").Find(&seasons).Error; err != nil {
		return nil, err
	}
	return seasons, nil
}

// Convenience: list Teams linked to a Season (optionally active only)
func (r *TeamSeasonRepository) ListTeamsForSeason(ctx context.Context, seasonID int64, onlyActive *bool) ([]models.Team, error) {
	var teams []models.Team
	q := r.db.WithContext(ctx).
		Table("teams").
		Joins("JOIN team_seasons ts ON ts.team_id = teams.id").
		Where("ts.season_id = ? AND teams.deleted_at IS NULL AND ts.deleted_at IS NULL", seasonID)
	if onlyActive != nil {
		q = q.Where("ts.is_active = ?", *onlyActive)
	}
	if err := q.Order("teams.id DESC").Select("teams.*").Find(&teams).Error; err != nil {
		return nil, err
	}
	return teams, nil
}
