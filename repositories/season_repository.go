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

func (r *SeasonRepository) ListPlayerStats(
	ctx context.Context,
	seasonID int64,
) ([]models.PlayerStatsRow, error) {
	sql := `
WITH per_player AS (
  -- Direct player-vs-player games
  SELECT
    gs.player_id                       AS player_id,
    g.id                               AS game_id,
    gs.side                            AS side,
    gs.color                           AS color,
    g.winner_side                      AS winner_side
  FROM games g
  JOIN game_sides gs ON gs.game_id = g.id
  WHERE
    g.status = 'completed'
    AND g.match_type = 'players'
    AND g.season_id = ?
    AND gs.player_id IS NOT NULL

  UNION ALL

  -- Team games, expand to PlayerA
  SELECT
    t.player_a_id                      AS player_id,
    g.id                               AS game_id,
    gs.side                            AS side,
    gs.color                           AS color,
    g.winner_side                      AS winner_side
  FROM games g
  JOIN game_sides gs ON gs.game_id = g.id
  JOIN teams t       ON t.id = gs.team_id
  WHERE
    g.status = 'completed'
    AND g.match_type = 'teams'
    AND g.season_id = ?
    AND gs.team_id IS NOT NULL

  UNION ALL

  -- Team games, expand to PlayerB
  SELECT
    t.player_b_id                      AS player_id,
    g.id                               AS game_id,
    gs.side                            AS side,
    gs.color                           AS color,
    g.winner_side                      AS winner_side
  FROM games g
  JOIN game_sides gs ON gs.game_id = g.id
  JOIN teams t       ON t.id = gs.team_id
  WHERE
    g.status = 'completed'
    AND g.match_type = 'teams'
    AND g.season_id = ?
    AND gs.team_id IS NOT NULL
),
agg AS (
  SELECT
    player_id,
    COUNT(*) AS games,
    SUM(CASE WHEN winner_side = side THEN 1 ELSE 0 END) AS wins,
    SUM(CASE WHEN winner_side <> side THEN 1 ELSE 0 END) AS losses,

    -- Wins by color
    SUM(CASE WHEN winner_side = side AND color = 'white'   THEN 1 ELSE 0 END) AS white_wins,
    SUM(CASE WHEN winner_side = side AND color = 'black'   THEN 1 ELSE 0 END) AS black_wins,
    SUM(CASE WHEN winner_side = side AND color = 'natural' THEN 1 ELSE 0 END) AS natural_wins,

    -- Games by color
    SUM(CASE WHEN color = 'white'   THEN 1 ELSE 0 END) AS white_games,
    SUM(CASE WHEN color = 'black'   THEN 1 ELSE 0 END) AS black_games,
    SUM(CASE WHEN color = 'natural' THEN 1 ELSE 0 END) AS natural_games
  FROM per_player
  GROUP BY player_id
)
SELECT
  player_id,
  games,
  wins,
  losses,
  white_wins,
  black_wins,
  natural_wins,
  white_games,
  black_games,
  natural_games,
  CASE WHEN games = 0 THEN 0.0
       ELSE wins::float / games::float
  END AS win_pct
FROM agg
ORDER BY win_pct DESC, games DESC, player_id ASC;
`

	type row struct {
		PlayerID     int64   `gorm:"column:player_id"`
		Games        int64   `gorm:"column:games"`
		Wins         int64   `gorm:"column:wins"`
		Losses       int64   `gorm:"column:losses"`
		WhiteWins    int64   `gorm:"column:white_wins"`
		BlackWins    int64   `gorm:"column:black_wins"`
		NaturalWins  int64   `gorm:"column:natural_wins"`
		WhiteGames   int64   `gorm:"column:white_games"`
		BlackGames   int64   `gorm:"column:black_games"`
		NaturalGames int64   `gorm:"column:natural_games"`
		WinPct       float64 `gorm:"column:win_pct"`
	}

	var rows []row
	if err := r.db.WithContext(ctx).
		Raw(sql, seasonID, seasonID, seasonID).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	out := make([]models.PlayerStatsRow, 0, len(rows))
	for _, x := range rows {
		out = append(out, models.PlayerStatsRow{
			PlayerID:     x.PlayerID,
			Games:        x.Games,
			Wins:         x.Wins,
			Losses:       x.Losses,
			WinPct:       x.WinPct,
			WhiteWins:    x.WhiteWins,
			BlackWins:    x.BlackWins,
			NaturalWins:  x.NaturalWins,
			WhiteGames:   x.WhiteGames,
			BlackGames:   x.BlackGames,
			NaturalGames: x.NaturalGames,
		})
	}
	return out, nil
}

func (r *SeasonRepository) ListTeamStats(
	ctx context.Context,
	seasonID int64,
) ([]models.TeamStatsRow, error) {
	sql := `
WITH per_team AS (
  SELECT
    t.id                                  AS team_id,
    g.id                                  AS game_id,
    gs.side                               AS side,
    gs.color                              AS color,
    COALESCE(g.location, 'Unknown')       AS location,
    g.winner_side                         AS winner_side
  FROM games g
  JOIN game_sides gs ON gs.game_id = g.id
  JOIN teams t       ON t.id = gs.team_id
  WHERE
    g.status = 'completed'
    AND g.match_type = 'teams'
    AND g.season_id = ?
),
agg AS (
  SELECT
    team_id,
    COUNT(*) AS games,
    SUM(CASE WHEN winner_side = side THEN 1 ELSE 0 END) AS wins,
    SUM(CASE WHEN winner_side <> side THEN 1 ELSE 0 END) AS losses,

    -- Wins by color
    SUM(CASE WHEN winner_side = side AND color = 'white'   THEN 1 ELSE 0 END) AS white_wins,
    SUM(CASE WHEN winner_side = side AND color = 'black'   THEN 1 ELSE 0 END) AS black_wins,
    SUM(CASE WHEN winner_side = side AND color = 'natural' THEN 1 ELSE 0 END) AS natural_wins,

    -- Games by color
    SUM(CASE WHEN color = 'white'   THEN 1 ELSE 0 END) AS white_games,
    SUM(CASE WHEN color = 'black'   THEN 1 ELSE 0 END) AS black_games,
    SUM(CASE WHEN color = 'natural' THEN 1 ELSE 0 END) AS natural_games
  FROM per_team
  GROUP BY team_id
),
loc AS (
  SELECT
    team_id,
    location,
    COUNT(*) FILTER (WHERE winner_side = side) AS wins_at_location,
    ROW_NUMBER() OVER (
      PARTITION BY team_id
      ORDER BY COUNT(*) FILTER (WHERE winner_side = side) DESC, location ASC
    ) AS rn
  FROM per_team
  GROUP BY team_id, location
)
SELECT
  a.team_id,
  a.games,
  a.wins,
  a.losses,
  a.white_wins,
  a.black_wins,
  a.natural_wins,
  a.white_games,
  a.black_games,
  a.natural_games,
  CASE WHEN a.games = 0 THEN 0.0
       ELSE a.wins::float / a.games::float
  END AS win_pct,
  l.location       AS best_location,
  COALESCE(l.wins_at_location, 0) AS best_location_wins
FROM agg a
LEFT JOIN loc l
  ON l.team_id = a.team_id
 AND l.rn = 1
ORDER BY win_pct DESC, games DESC, team_id ASC;
`

	type row struct {
		TeamID           int64   `gorm:"column:team_id"`
		Games            int64   `gorm:"column:games"`
		Wins             int64   `gorm:"column:wins"`
		Losses           int64   `gorm:"column:losses"`
		WhiteWins        int64   `gorm:"column:white_wins"`
		BlackWins        int64   `gorm:"column:black_wins"`
		NaturalWins      int64   `gorm:"column:natural_wins"`
		WhiteGames       int64   `gorm:"column:white_games"`
		BlackGames       int64   `gorm:"column:black_games"`
		NaturalGames     int64   `gorm:"column:natural_games"`
		WinPct           float64 `gorm:"column:win_pct"`
		BestLocation     *string `gorm:"column:best_location"`
		BestLocationWins int64   `gorm:"column:best_location_wins"`
	}

	var rows []row
	if err := r.db.WithContext(ctx).
		Raw(sql, seasonID).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	out := make([]models.TeamStatsRow, 0, len(rows))
	for _, x := range rows {
		out = append(out, models.TeamStatsRow{
			TeamID:           x.TeamID,
			Games:            x.Games,
			Wins:             x.Wins,
			Losses:           x.Losses,
			WinPct:           x.WinPct,
			WhiteWins:        x.WhiteWins,
			BlackWins:        x.BlackWins,
			NaturalWins:      x.NaturalWins,
			WhiteGames:       x.WhiteGames,
			BlackGames:       x.BlackGames,
			NaturalGames:     x.NaturalGames,
			BestLocation:     x.BestLocation,
			BestLocationWins: x.BestLocationWins,
		})
	}
	return out, nil
}
