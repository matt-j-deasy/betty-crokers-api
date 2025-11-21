// services/season_stats_service.go
package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/matt-j-deasy/betty-crokers-api/repositories"
)

// Public DTOs for API / UI (decoupled from repo structs)

type PlayerStats struct {
	PlayerID int64 `json:"playerId"`

	Games  int64   `json:"games"`
	Wins   int64   `json:"wins"`
	Losses int64   `json:"losses"`
	WinPct float64 `json:"winPct"`

	WhiteWins   int64 `json:"whiteWins"`
	BlackWins   int64 `json:"blackWins"`
	NaturalWins int64 `json:"naturalWins"`

	WhiteGames   int64 `json:"whiteGames"`
	BlackGames   int64 `json:"blackGames"`
	NaturalGames int64 `json:"naturalGames"`
}

type TeamStats struct {
	TeamID int64 `json:"teamId"`

	Games  int64   `json:"games"`
	Wins   int64   `json:"wins"`
	Losses int64   `json:"losses"`
	WinPct float64 `json:"winPct"`

	WhiteWins   int64 `json:"whiteWins"`
	BlackWins   int64 `json:"blackWins"`
	NaturalWins int64 `json:"naturalWins"`

	WhiteGames   int64 `json:"whiteGames"`
	BlackGames   int64 `json:"blackGames"`
	NaturalGames int64 `json:"naturalGames"`

	BestLocation     *string `json:"bestLocation"`
	BestLocationWins int64   `json:"bestLocationWins"`
}

type SeasonStatsService struct {
	repositories *repositories.RepositoriesCollection
}

func NewSeasonStatsService(repos *repositories.RepositoriesCollection) *SeasonStatsService {
	return &SeasonStatsService{
		repositories: repos,
	}
}

// Optionally ensure the season exists before querying stats.
// If you already have SeasonService with a GetSeason method, you can inject that instead.
func (s *SeasonStatsService) validateSeasonExists(ctx context.Context, seasonID int64) error {
	season, err := s.repositories.SeasonRepo.GetByID(ctx, seasonID) // adjust name if needed
	if err != nil {
		return fmt.Errorf("lookup season: %w", err)
	}
	if season == nil {
		return fmt.Errorf("season not found")
	}
	return nil
}

func (s *SeasonStatsService) ListPlayerStats(
	ctx context.Context,
	seasonID int64,
) ([]PlayerStats, error) {
	if err := s.validateSeasonExists(ctx, seasonID); err != nil {
		slog.Error("season validation failed", "seasonID", seasonID, "error", err)
		return nil, err
	}

	rows, err := s.repositories.SeasonRepo.ListPlayerStats(ctx, seasonID)
	if err != nil {
		slog.Error("failed to list player stats from repo", "seasonID", seasonID, "error", err)
		return nil, err
	}

	out := make([]PlayerStats, 0, len(rows))
	for _, r := range rows {
		out = append(out, PlayerStats{
			PlayerID:     r.PlayerID,
			Games:        r.Games,
			Wins:         r.Wins,
			Losses:       r.Losses,
			WinPct:       r.WinPct,
			WhiteWins:    r.WhiteWins,
			BlackWins:    r.BlackWins,
			NaturalWins:  r.NaturalWins,
			WhiteGames:   r.WhiteGames,
			BlackGames:   r.BlackGames,
			NaturalGames: r.NaturalGames,
		})
	}

	slog.Info("fetched player stats", "seasonID", seasonID, "count", len(out))
	return out, nil
}

func (s *SeasonStatsService) ListTeamStats(
	ctx context.Context,
	seasonID int64,
) ([]TeamStats, error) {
	if err := s.validateSeasonExists(ctx, seasonID); err != nil {
		slog.Error("season validation failed", "seasonID", seasonID, "error", err)
		return nil, err
	}

	rows, err := s.repositories.SeasonRepo.ListTeamStats(ctx, seasonID)
	if err != nil {
		slog.Error("failed to list team stats from repo", "seasonID", seasonID, "error", err)
		return nil, err
	}

	out := make([]TeamStats, 0, len(rows))
	for _, r := range rows {
		out = append(out, TeamStats{
			TeamID:           r.TeamID,
			Games:            r.Games,
			Wins:             r.Wins,
			Losses:           r.Losses,
			WinPct:           r.WinPct,
			WhiteWins:        r.WhiteWins,
			BlackWins:        r.BlackWins,
			NaturalWins:      r.NaturalWins,
			WhiteGames:       r.WhiteGames,
			BlackGames:       r.BlackGames,
			NaturalGames:     r.NaturalGames,
			BestLocation:     r.BestLocation,
			BestLocationWins: r.BestLocationWins,
		})
	}

	slog.Info("fetched team stats", "seasonID", seasonID, "count", len(out))
	return out, nil
}
