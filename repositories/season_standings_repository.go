// repositories/season_standings.go
package repositories

import (
	"context"
	"fmt"
)

type SeasonStandingsRow struct {
	TeamID        int64   `json:"teamId" gorm:"column:team_id"`
	TeamName      string  `json:"teamName" gorm:"column:team_name"`
	Games         int     `json:"games" gorm:"column:games"`
	Wins          int     `json:"wins" gorm:"column:wins"`
	Losses        int     `json:"losses" gorm:"column:losses"`
	Ties          int     `json:"ties" gorm:"column:ties"`
	PointsFor     int     `json:"pointsFor" gorm:"column:pf"`
	PointsAgainst int     `json:"pointsAgainst" gorm:"column:pa"`
	PointDiff     int     `json:"pointDiff" gorm:"column:pd"`
	WinPct        float64 `json:"winPct" gorm:"column:win_pct"`
}

func (r *SeasonRepository) GetStandings(ctx context.Context, seasonID int64) ([]SeasonStandingsRow, error) {
	var rows []SeasonStandingsRow

	// Schema assumptions:
	// - games(id, season_id, match_type, status)
	// - game_sides(id, game_id, side, team_id, points)
	// - teams(id, name)
	// Only includes completed team-vs-team games for the given season.
	// Handles ties as 0.5 win in win%.
	sql := `
WITH per_team AS (
  SELECT
    g.id                                        AS game_id,
    g.season_id                                 AS season_id,
    gs1.team_id                                 AS team_id,
    t.name                                      AS team_name,
    COALESCE(gs1.points, 0)                     AS pf,
    COALESCE(gs2.points, 0)                     AS pa,
    CASE WHEN COALESCE(gs1.points,0) > COALESCE(gs2.points,0) THEN 1 ELSE 0 END AS win,
    CASE WHEN COALESCE(gs1.points,0) < COALESCE(gs2.points,0) THEN 1 ELSE 0 END AS loss,
    CASE WHEN COALESCE(gs1.points,0) = COALESCE(gs2.points,0) THEN 1 ELSE 0 END AS tie
  FROM games g
  JOIN game_sides gs1 ON gs1.game_id = g.id
  JOIN game_sides gs2 ON gs2.game_id = g.id AND gs2.side <> gs1.side
  JOIN teams t        ON t.id = gs1.team_id
  WHERE
    g.status = 'completed'
    AND g.match_type = 'teams'
    AND g.season_id = @seasonID
    AND gs1.team_id IS NOT NULL
)
SELECT
  team_id,
  team_name,
  COUNT(*)                                        AS games,
  SUM(win)                                        AS wins,
  SUM(loss)                                       AS losses,
  SUM(tie)                                        AS ties,
  SUM(pf)                                         AS pf,
  SUM(pa)                                         AS pa,
  SUM(pf) - SUM(pa)                               AS pd,
  CASE WHEN COUNT(*) = 0
       THEN 0
       ELSE ROUND( (SUM(win)::decimal + 0.5 * SUM(tie)::decimal) / COUNT(*), 4)
  END                                             AS win_pct
FROM per_team
GROUP BY team_id, team_name
ORDER BY
  wins DESC,
  pd DESC,
  pf DESC,
  team_name ASC;
`
	if err := r.db.WithContext(ctx).Raw(sql, map[string]any{"seasonID": seasonID}).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

type PlayerStandingsRow struct {
	PlayerID      int64
	Games         int64
	Wins          int64
	Losses        int64
	PointsFor     int64
	PointsAgainst int64
	PointDiff     int64
	WinPct        float64
}

type PlayerStandingsCursor struct {
	Wins      int64 `json:"wins"`
	PointDiff int64 `json:"point_diff"`
	PlayerID  int64 `json:"player_id"`
}

type ListPlayerStandingsQuery struct {
	SeasonID int64
	Limit    int
	Cursor   *PlayerStandingsCursor
}

func (r *SeasonRepository) ListPlayerStandings(ctx context.Context, q ListPlayerStandingsQuery) ([]PlayerStandingsRow, *PlayerStandingsCursor, error) {
	if q.Limit <= 0 {
		q.Limit = 50
	}
	if q.Limit > 200 {
		q.Limit = 200
	}

	// Supports player-vs-player AND team-vs-team games.
	// For team games, we "expand" each side into two rows (PlayerAID, PlayerBID).
	// Roster = active memberships UNION players who actually appeared this season.
	base := `
WITH roster AS (
  SELECT DISTINCT ptm.player_id
  FROM player_team_memberships ptm
  WHERE ptm.season_id = ? AND ptm.is_active = TRUE

  UNION

  SELECT DISTINCT p.player_id
  FROM (
    -- players games, take player_id directly
    SELECT gs.player_id
    FROM game_sides gs
    JOIN games g ON g.id = gs.game_id
    WHERE g.season_id = ?
      AND g.status = 'completed'
      AND g.match_type = 'players'
      AND gs.player_id IS NOT NULL

    UNION ALL

    -- team games, expand Team.PlayerAID
    SELECT t.player_a_id AS player_id
    FROM game_sides gs
    JOIN games g ON g.id = gs.game_id
    JOIN teams t ON t.id = gs.team_id
    WHERE g.season_id = ?
      AND g.status = 'completed'
      AND g.match_type = 'teams'
      AND gs.team_id IS NOT NULL

    UNION ALL

    -- team games, expand Team.PlayerBID
    SELECT t.player_b_id AS player_id
    FROM game_sides gs
    JOIN games g ON g.id = gs.game_id
    JOIN teams t ON t.id = gs.team_id
    WHERE g.season_id = ?
      AND g.status = 'completed'
      AND g.match_type = 'teams'
      AND gs.team_id IS NOT NULL
  ) p
),
-- Expand game_sides to per-player rows for ALL completed games
expanded AS (
  -- players format
  SELECT
    g.id           AS game_id,
    gs.side        AS side,
    gs.points      AS points_for,
    NULL           AS team_id,
    gs.player_id   AS player_id,
    g.winner_side  AS winner_side
  FROM game_sides gs
  JOIN games g ON g.id = gs.game_id
  WHERE g.season_id = ?
    AND g.status = 'completed'
    AND g.match_type = 'players'
    AND gs.player_id IS NOT NULL

  UNION ALL

  -- teams format -> PlayerA
  SELECT
    g.id           AS game_id,
    gs.side        AS side,
    gs.points      AS points_for,
    gs.team_id     AS team_id,
    t.player_a_id  AS player_id,
    g.winner_side  AS winner_side
  FROM game_sides gs
  JOIN games g ON g.id = gs.game_id
  JOIN teams t ON t.id = gs.team_id
  WHERE g.season_id = ?
    AND g.status = 'completed'
    AND g.match_type = 'teams'
    AND gs.team_id IS NOT NULL

  UNION ALL

  -- teams format -> PlayerB
  SELECT
    g.id           AS game_id,
    gs.side        AS side,
    gs.points      AS points_for,
    gs.team_id     AS team_id,
    t.player_b_id  AS player_id,
    g.winner_side  AS winner_side
  FROM game_sides gs
  JOIN games g ON g.id = gs.game_id
  JOIN teams t ON t.id = gs.team_id
  WHERE g.season_id = ?
    AND g.status = 'completed'
    AND g.match_type = 'teams'
    AND gs.team_id IS NOT NULL
),
-- Join each per-player row with their opponent row in the same game to get points_against
paired AS (
  SELECT DISTINCT ON (a.player_id, a.game_id)
    a.player_id,
    a.game_id,
    a.side,
    a.points_for,
    b.points_for AS points_against,
    a.winner_side
  FROM expanded a
  JOIN expanded b ON b.game_id = a.game_id AND b.side <> a.side
  ORDER BY a.player_id, a.game_id
),
-- Aggregate per player
agg AS (
  SELECT
    player_id,
    COUNT(*) AS games,
    SUM(CASE WHEN winner_side = side THEN 1 ELSE 0 END) AS wins,
    SUM(CASE WHEN winner_side <> side THEN 1 ELSE 0 END) AS losses,
    SUM(points_for)      AS points_for,
    SUM(points_against)  AS points_against
  FROM paired
  GROUP BY player_id
),
-- Left join roster to include zero-game players
standings AS (
  SELECT
    r.player_id,
    COALESCE(a.games, 0)            AS games,
    COALESCE(a.wins, 0)             AS wins,
    COALESCE(a.losses, 0)           AS losses,
    COALESCE(a.points_for, 0)       AS points_for,
    COALESCE(a.points_against, 0)   AS points_against,
    (COALESCE(a.points_for, 0) - COALESCE(a.points_against, 0)) AS point_diff,
    CASE
      WHEN COALESCE(a.games, 0) = 0 THEN 0.0
      ELSE (CAST(COALESCE(a.wins, 0) AS FLOAT) / CAST(COALESCE(a.games, 0) AS FLOAT))
    END AS win_pct
  FROM roster r
  LEFT JOIN agg a ON a.player_id = r.player_id
)
SELECT
  player_id,
  games,
  wins,
  losses,
  points_for      AS points_for,
  points_against  AS points_against,
  point_diff,
  win_pct
FROM standings
`

	// Keyset pagination (wins DESC, point_diff DESC, player_id ASC)
	pred := ""
	args := []any{
		q.SeasonID, q.SeasonID, q.SeasonID, q.SeasonID, // roster params
		q.SeasonID, q.SeasonID, q.SeasonID, // expanded params
	}
	if q.Cursor != nil {
		pred = `
WHERE (wins < ?)
   OR (wins = ? AND point_diff < ?)
   OR (wins = ? AND point_diff = ? AND player_id > ?)
`
		args = append(args,
			q.Cursor.Wins,
			q.Cursor.Wins, q.Cursor.PointDiff,
			q.Cursor.Wins, q.Cursor.PointDiff, q.Cursor.PlayerID,
		)
	}

	order := `
ORDER BY wins DESC, point_diff DESC, player_id ASC
`
	sql := base + pred + order + fmt.Sprintf("\nLIMIT %d", q.Limit+1)

	type row struct {
		PlayerID      int64   `gorm:"column:player_id"`
		Games         int64   `gorm:"column:games"`
		Wins          int64   `gorm:"column:wins"`
		Losses        int64   `gorm:"column:losses"`
		PointsFor     int64   `gorm:"column:points_for"`
		PointsAgainst int64   `gorm:"column:points_against"`
		PointDiff     int64   `gorm:"column:point_diff"`
		WinPct        float64 `gorm:"column:win_pct"`
	}

	var rows []row
	if err := r.db.WithContext(ctx).Raw(sql, args...).Scan(&rows).Error; err != nil {
		return nil, nil, err
	}

	var next *PlayerStandingsCursor
	if len(rows) > q.Limit {
		last := rows[q.Limit-1]
		next = &PlayerStandingsCursor{
			Wins:      last.Wins,
			PointDiff: last.PointDiff,
			PlayerID:  last.PlayerID,
		}
		rows = rows[:q.Limit]
	}

	out := make([]PlayerStandingsRow, 0, len(rows))
	for _, x := range rows {
		out = append(out, PlayerStandingsRow{
			PlayerID:      x.PlayerID,
			Games:         x.Games,
			Wins:          x.Wins,
			Losses:        x.Losses,
			PointsFor:     x.PointsFor,
			PointsAgainst: x.PointsAgainst,
			PointDiff:     x.PointDiff,
			WinPct:        x.WinPct,
		})
	}
	return out, next, nil
}
