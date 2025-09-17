package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/matt-j-deasy/betty-crokers-api/models"
	"github.com/matt-j-deasy/betty-crokers-api/repositories"
)

type GameSideService struct {
	repos *repositories.RepositoriesCollection
}

func NewGameSideService(repos *repositories.RepositoriesCollection) *GameSideService {
	return &GameSideService{repos: repos}
}

/* =========================
   DTOs
========================= */

type SetSideColorInput struct {
	GameID int64            `json:"gameId"`
	Side   string           `json:"side"`  // "A"|"B"
	Color  models.DiscColor `json:"color"` // "white"|"black"|"natural"
}

type AddPointsInput struct {
	GameID int64  `json:"gameId"`
	Side   string `json:"side"`  // "A"|"B"
	Delta  int    `json:"delta"` // >= 0
}

type SetPointsInput struct {
	GameID int64  `json:"gameId"`
	Side   string `json:"side"`   // "A"|"B"
	Points int    `json:"points"` // >= 0
}

/* =========================
   Operations
========================= */

func (s *GameSideService) ListByGame(ctx context.Context, gameID int64) ([]models.GameSide, error) {
	return s.repos.GameSideRepo.ListByGame(ctx, gameID)
}

func (s *GameSideService) SetColor(ctx context.Context, in SetSideColorInput) (*models.GameSide, error) {
	side := normalizeSide(in.Side)
	if side == "" {
		return nil, errors.New("side must be 'A' or 'B'")
	}
	if err := validateColor(in.Color); err != nil {
		return nil, err
	}
	game, err := s.repos.GameRepo.GetByID(ctx, in.GameID)
	if err != nil {
		return nil, err
	}
	if game.Status == "completed" || game.Status == "canceled" {
		return nil, errors.New("cannot change color for completed/canceled game")
	}
	return s.repos.GameSideRepo.UpdateFieldsByGameAndSide(ctx, in.GameID, side, map[string]any{
		"color": in.Color,
	})
}

func (s *GameSideService) AddPoints(ctx context.Context, in AddPointsInput) (*models.Game, []models.GameSide, error) {
	if in.Delta < 0 {
		return nil, nil, errors.New("delta must be >= 0")
	}
	return s.adjustPoints(ctx, in.GameID, normalizeSide(in.Side), nil, &in.Delta)
}

func (s *GameSideService) SetPoints(ctx context.Context, in SetPointsInput) (*models.Game, []models.GameSide, error) {
	if in.Points < 0 {
		return nil, nil, errors.New("points must be >= 0")
	}
	return s.adjustPoints(ctx, in.GameID, normalizeSide(in.Side), &in.Points, nil)
}

/* =========================
   Internal
========================= */

func (s *GameSideService) adjustPoints(ctx context.Context, gameID int64, side string, absolute *int, delta *int) (*models.Game, []models.GameSide, error) {
	if side == "" {
		return nil, nil, errors.New("side must be 'A' or 'B'")
	}

	game, err := s.repos.GameRepo.GetByID(ctx, gameID)
	if err != nil {
		return nil, nil, err
	}
	switch game.Status {
	case "scheduled":
		// first score starts the game
		now := time.Now().UTC()
		if _, err := s.repos.GameRepo.UpdateFields(ctx, gameID, map[string]any{
			"status":     "in_progress",
			"started_at": &now,
		}); err != nil {
			return nil, nil, err
		}
		game.Status = "in_progress"
	case "in_progress":
		// ok
	case "completed", "canceled":
		return nil, nil, errors.New("cannot change points for completed/canceled game")
	default:
		return nil, nil, errors.New("invalid game status")
	}

	// get current side + compute new points
	sd, err := s.repos.GameSideRepo.GetByGameAndSide(ctx, gameID, side)
	if err != nil {
		return nil, nil, err
	}
	newPoints := sd.Points
	if absolute != nil {
		newPoints = *absolute
	} else if delta != nil {
		newPoints += *delta
	}
	if newPoints < 0 {
		newPoints = 0
	}

	if _, err := s.repos.GameSideRepo.UpdateFieldsByGameAndSide(ctx, gameID, side, map[string]any{
		"points": newPoints,
	}); err != nil {
		return nil, nil, err
	}

	if newPoints == sd.Points {
		game, err := s.repos.GameRepo.GetByID(ctx, gameID)
		if err != nil {
			return nil, nil, err
		}
		sides, err := s.repos.GameSideRepo.ListByGame(ctx, gameID)
		if err != nil {
			return nil, nil, err
		}
		return game, sides, nil
	}

	// Transition scheduled -> in_progress on first real score change
	switch game.Status {
	case "scheduled":
		now := time.Now().UTC()
		if _, err := s.repos.GameRepo.UpdateFields(ctx, gameID, map[string]any{
			"status":     "in_progress",
			"started_at": &now,
		}); err != nil {
			return nil, nil, err
		}
	case "in_progress":
		// ok
	case "completed", "canceled":
		return nil, nil, errors.New("cannot change points for completed/canceled game")
	default:
		return nil, nil, errors.New("invalid game status")
	}

	// Persist points
	if _, err := s.repos.GameSideRepo.UpdateFieldsByGameAndSide(ctx, gameID, side, map[string]any{
		"points": newPoints,
	}); err != nil {
		return nil, nil, err
	}

	// Return fresh snapshot (no winner/auto-complete)
	game, err = s.repos.GameRepo.GetByID(ctx, gameID)
	if err != nil {
		return nil, nil, err
	}
	sides, err := s.repos.GameSideRepo.ListByGame(ctx, gameID)
	if err != nil {
		return nil, nil, err
	}
	return game, sides, nil
}

func validateColor(c models.DiscColor) error {
	switch c {
	case models.DiscWhite, models.DiscBlack, models.DiscNatural:
		return nil
	default:
		return errors.New("invalid color (allowed: white, black, natural)")
	}
}

func normalizeSide(s string) string {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "A":
		return "A"
	case "B":
		return "B"
	default:
		return ""
	}
}
