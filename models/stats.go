package models

// Per-player stats for a season.
type PlayerStatsRow struct {
	PlayerID int64 `json:"playerId"`

	Games  int64 `json:"games"`
	Wins   int64 `json:"wins"`
	Losses int64 `json:"losses"`

	WinPct float64 `json:"winPct"`

	WhiteWins   int64 `json:"whiteWins"`
	BlackWins   int64 `json:"blackWins"`
	NaturalWins int64 `json:"naturalWins"`

	WhiteGames   int64 `json:"whiteGames"`
	BlackGames   int64 `json:"blackGames"`
	NaturalGames int64 `json:"naturalGames"`
}

// Per-team stats for a season.
type TeamStatsRow struct {
	TeamID int64 `json:"teamId"`

	Games  int64 `json:"games"`
	Wins   int64 `json:"wins"`
	Losses int64 `json:"losses"`

	WinPct float64 `json:"winPct"`

	WhiteWins   int64 `json:"whiteWins"`
	BlackWins   int64 `json:"blackWins"`
	NaturalWins int64 `json:"naturalWins"`

	WhiteGames   int64 `json:"whiteGames"`
	BlackGames   int64 `json:"blackGames"`
	NaturalGames int64 `json:"naturalGames"`

	BestLocation     *string `json:"bestLocation"`     // where they’ve won the most
	BestLocationWins int64   `json:"bestLocationWins"` // wins at that location
}
