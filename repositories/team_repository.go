package repositories

import (
	"context"
	"strings"

	"gorm.io/gorm"

	"github.com/matt-j-deasy/betty-crokers-api/models"
)

type TeamRepository struct {
	db *gorm.DB
}

func NewTeamRepository(db *gorm.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

type ListTeamsFilter struct {
	Search     string // case-insensitive match on name
	PlayerID   *int64 // return teams where this player is A or B
	SeasonID   *int64 // only teams linked to this season (via team_seasons)
	OnlyActive *bool  // if SeasonID is set, optionally filter by link activity
	Offset     int
	Limit      int
}

func (r *TeamRepository) Create(ctx context.Context, t *models.Team) error {
	canonicalizePair(&t.PlayerAID, &t.PlayerBID)
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *TeamRepository) GetByID(ctx context.Context, id int64) (*models.Team, error) {
	var t models.Team
	if err := r.db.WithContext(ctx).First(&t, id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

// GetByPlayers returns the team for the given pair (order-insensitive).
func (r *TeamRepository) GetByPlayers(ctx context.Context, aID, bID int64) (*models.Team, error) {
	canonicalizePair(&aID, &bID)
	var t models.Team
	if err := r.db.WithContext(ctx).
		Where("player_a_id = ? AND player_b_id = ?", aID, bID).
		First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

// ExistsByPlayers returns true if a team already exists for this exact pair (order-insensitive).
func (r *TeamRepository) ExistsByPlayers(ctx context.Context, aID, bID int64) (bool, error) {
	canonicalizePair(&aID, &bID)
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&models.Team{}).
		Where("player_a_id = ? AND player_b_id = ?", aID, bID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *TeamRepository) UpdateFields(ctx context.Context, id int64, fields map[string]any) (*models.Team, error) {
	// If caller tries to update the pair, canonicalize it
	if a, ok := fields["player_a_id"].(int64); ok {
		if b, ok2 := fields["player_b_id"].(int64); ok2 {
			canonicalizePair(&a, &b)
			fields["player_a_id"] = a
			fields["player_b_id"] = b
		}
	}
	if err := r.db.WithContext(ctx).
		Model(&models.Team{}).
		Where("id = ?", id).
		Updates(fields).Error; err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

func (r *TeamRepository) DeleteByID(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&models.Team{}, id).Error
}

func (r *TeamRepository) List(ctx context.Context, f ListTeamsFilter) ([]models.Team, int64, error) {
	var (
		items []models.Team
		total int64
	)

	// base filter builder
	base := r.db.WithContext(ctx).Model(&models.Team{}).Where("teams.deleted_at IS NULL")
	apply := func(q *gorm.DB) *gorm.DB {
		if s := strings.TrimSpace(f.Search); s != "" {
			ilike := "%" + strings.ToLower(s) + "%"
			q = q.Where("LOWER(teams.name) LIKE ?", ilike)
		}
		if f.PlayerID != nil && *f.PlayerID > 0 {
			q = q.Where("(teams.player_a_id = ? OR teams.player_b_id = ?)", *f.PlayerID, *f.PlayerID)
		}
		if f.SeasonID != nil && *f.SeasonID > 0 {
			q = q.Joins(`JOIN team_seasons ts
				ON ts.team_id = teams.id
				AND ts.season_id = ?
				AND ts.deleted_at IS NULL`, *f.SeasonID)
			if f.OnlyActive != nil {
				q = q.Where("ts.is_active = ?", *f.OnlyActive)
			}
		}
		return q
	}

	// total count (distinct ids to de-dupe joins)
	countQ := apply(base.Session(&gorm.Session{}))
	if err := countQ.Distinct("teams.id").Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// pagination guards
	limit := f.Limit
	if limit <= 0 || limit > 100 {
		limit = 25
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}

	// items query (distinct across all selected columns)
	itemsQ := apply(base.Session(&gorm.Session{}))
	if err := itemsQ.
		Select("teams.*").
		Distinct(). // <-- no args, so it's DISTINCT on all selected columns
		Order("teams.id DESC").
		Limit(limit).
		Offset(offset).
		Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// --- helpers ---

func canonicalizePair(a, b *int64) {
	if *a > *b {
		*a, *b = *b, *a
	}
}
