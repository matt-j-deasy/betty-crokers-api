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

type PlayerDuplicateGameDBRow struct {
	GameID     int64   `gorm:"column:game_id"`
	SeasonID   *int64  `gorm:"column:season_id"`
	MatchType  string  `gorm:"column:match_type"`
	Status     string  `gorm:"column:status"`
	WinnerSide *string `gorm:"column:winner_side"`

	Side     string `gorm:"column:side"`
	Color    string `gorm:"column:color"`
	RowCount int64  `gorm:"column:row_count"`
}

func (r *PlayerRepository) ListPlayerDuplicateGames(
	ctx context.Context,
	playerID int64,
) ([]PlayerDuplicateGameDBRow, error) {

	sql := `
WITH per_player AS (

  -- direct player matchups
  SELECT
    gs.player_id AS player_id,
    g.id         AS game_id,
    g.season_id  AS season_id,
    g.match_type AS match_type,
    g.status     AS status,
    g.winner_side AS winner_side,
    gs.side      AS side,
    gs.color     AS color
  FROM games g
  JOIN game_sides gs ON gs.game_id = g.id
  WHERE g.status = 'completed'
    AND g.match_type = 'players'
    AND gs.player_id IS NOT NULL

  UNION ALL

  -- team matchups -> PlayerA
  SELECT
    t.player_a_id AS player_id,
    g.id          AS game_id,
    g.season_id   AS season_id,
    g.match_type  AS match_type,
    g.status      AS status,
    g.winner_side AS winner_side,
    gs.side       AS side,
    gs.color      AS color
  FROM games g
  JOIN game_sides gs ON gs.game_id = g.id
  JOIN teams t       ON t.id = gs.team_id
  WHERE g.status = 'completed'
    AND g.match_type = 'teams'
    AND gs.team_id IS NOT NULL

  UNION ALL

  -- team matchups -> PlayerB
  SELECT
    t.player_b_id AS player_id,
    g.id          AS game_id,
    g.season_id   AS season_id,
    g.match_type  AS match_type,
    g.status      AS status,
    g.winner_side AS winner_side,
    gs.side       AS side,
    gs.color      AS color
  FROM games g
  JOIN game_sides gs ON gs.game_id = g.id
  JOIN teams t       ON t.id = gs.team_id
  WHERE g.status = 'completed'
    AND g.match_type = 'teams'
    AND gs.team_id IS NOT NULL
),

suspicious AS (
  SELECT
    game_id,
    COUNT(*) AS row_count
  FROM per_player
  WHERE player_id = @playerID
  GROUP BY game_id
  HAVING COUNT(*) > 1
)

SELECT
  p.game_id,
  p.season_id,
  p.match_type,
  p.status,
  p.winner_side,
  p.side,
  p.color,
  s.row_count
FROM per_player p
JOIN suspicious s ON s.game_id = p.game_id
WHERE p.player_id = @playerID
ORDER BY p.game_id, p.side, p.color;
`

	var rows []PlayerDuplicateGameDBRow
	if err := r.db.WithContext(ctx).
		Raw(sql, map[string]any{
			"playerID": playerID,
		}).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}
