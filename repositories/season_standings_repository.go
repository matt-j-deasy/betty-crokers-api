// repositories/season_standings.go
package repositories

import (
	"context"
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
