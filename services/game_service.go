package services

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/matt-j-deasy/betty-crokers-api/models"
	"github.com/matt-j-deasy/betty-crokers-api/repositories"
)

type GameService struct {
	repos *repositories.RepositoriesCollection
}

func NewGameService(repos *repositories.RepositoriesCollection) *GameService {
	return &GameService{repos: repos}
}

/* =========================
   DTOs
========================= */

type GameParticipantInput struct {
	TeamID   *int64            `json:"teamId,omitempty"`
	PlayerID *int64            `json:"playerId,omitempty"`
	Color    *models.DiscColor `json:"color,omitempty"` // "white" | "black" | "natural"
}

type CreateGameInput struct {
	SeasonID     *int64               `json:"seasonId,omitempty"`     // nil => exhibition
	MatchType    string               `json:"matchType"`              // "teams" | "players"
	TargetPoints *int                 `json:"targetPoints,omitempty"` // default 100
	ScheduledAt  *string              `json:"scheduledAt,omitempty"`  // RFC3339
	Timezone     *string              `json:"timezone,omitempty"`     // default from season or America/New_York
	Location     *string              `json:"location,omitempty"`
	Description  *string              `json:"description,omitempty"`
	SideA        GameParticipantInput `json:"sideA"`
	SideB        GameParticipantInput `json:"sideB"`
}

type UpdateGameInput struct {
	SeasonID     *int64  `json:"seasonId,omitempty"`
	TargetPoints *int    `json:"targetPoints,omitempty"`
	ScheduledAt  *string `json:"scheduledAt,omitempty"` // RFC3339 or "" to clear
	Timezone     *string `json:"timezone,omitempty"`
	Location     *string `json:"location,omitempty"`    // can be null via handler->fields map if you want clearing
	Description  *string `json:"description,omitempty"` // can be null via handler->fields map if you want clearing
	Status       *string `json:"status,omitempty"`      // "scheduled"|"in_progress"|"completed"|"canceled"
	// Note: winner is computed; do not set directly
	SideAColor *models.DiscColor `json:"sideAColor,omitempty"` // "white" | "black" | "natural"
	SideBColor *models.DiscColor `json:"sideBColor,omitempty"` // "white" | "black" | "natural"
}

type ListGamesOptions struct {
	SeasonID       *int64
	ExhibitionOnly *bool
	Status         []string
	MatchType      *string
	ScheduledFrom  *time.Time
	ScheduledTo    *time.Time
	TeamID         *int64
	PlayerID       *int64
	Page           int
	Size           int
	OrderBy        string // e.g. "scheduled_at desc"
}

type PagedGames struct {
	Data  []models.Game `json:"data"`
	Total int64         `json:"total"`
	Page  int           `json:"page"`
	Size  int           `json:"size"`
}

/* =========================
   Core
========================= */

func (s *GameService) Create(ctx context.Context, in CreateGameInput) (*models.Game, []models.GameSide, error) {
	mt := strings.ToLower(strings.TrimSpace(in.MatchType))
	if mt != "teams" && mt != "players" {
		return nil, nil, errors.New("matchType must be 'teams' or 'players'")
	}

	// Season (optional)
	var seasonTZ string = "America/New_York"
	if in.SeasonID != nil {
		if _, err := s.repos.SeasonRepo.GetByID(ctx, *in.SeasonID); err != nil {
			return nil, nil, errors.New("season not found")
		}
		// If you want season default TZ, fetch and use it (uncomment if needed):
		// sz, _ := s.repos.SeasonRepo.GetByID(ctx, *in.SeasonID)
		// seasonTZ = sz.Timezone
	}

	// Validate sides according to match type
	sideA, err := s.buildSide(ctx, "A", mt, in.SideA)
	if err != nil {
		return nil, nil, err
	}
	sideB, err := s.buildSide(ctx, "B", mt, in.SideB)
	if err != nil {
		return nil, nil, err
	}
	// Distinct participants
	if mt == "teams" {
		if sideA.TeamID != nil && sideB.TeamID != nil && *sideA.TeamID == *sideB.TeamID {
			return nil, nil, errors.New("team A and team B cannot be the same")
		}
	} else {
		if sideA.PlayerID != nil && sideB.PlayerID != nil && *sideA.PlayerID == *sideB.PlayerID {
			return nil, nil, errors.New("player A and player B cannot be the same")
		}
	}

	// Target points
	target := 100
	if in.TargetPoints != nil && *in.TargetPoints > 0 {
		target = *in.TargetPoints
	}

	// Timezone
	tz := seasonTZ
	if in.Timezone != nil && *in.Timezone != "" {
		if _, err := time.LoadLocation(*in.Timezone); err != nil {
			return nil, nil, errors.New("invalid timezone")
		}
		tz = *in.Timezone
	}

	// ScheduledAt
	var scheduledAt *time.Time
	if in.ScheduledAt != nil && strings.TrimSpace(*in.ScheduledAt) != "" {
		t, err := time.Parse(time.RFC3339, *in.ScheduledAt)
		if err != nil {
			return nil, nil, errors.New("scheduledAt must be RFC3339")
		}
		scheduledAt = &t
	}

	game := &models.Game{
		SeasonID:     in.SeasonID,
		MatchType:    mt,
		TargetPoints: target,
		Status:       "scheduled",
		ScheduledAt:  scheduledAt,
		Timezone:     tz,
		Location:     in.Location,
		Description:  in.Description,
	}

	// Defaults
	if sideA.Color == "" {
		sideA.Color = models.DiscNatural
	}
	if sideB.Color == "" {
		sideB.Color = models.DiscNatural
	}

	// Persist in one TX
	if err := s.repos.GameRepo.CreateWithSides(ctx, game, []models.GameSide{sideA, sideB}); err != nil {
		return nil, nil, err
	}
	return game, []models.GameSide{sideA, sideB}, nil
}

func (s *GameService) GetByID(ctx context.Context, id int64) (*models.Game, error) {
	return s.repos.GameRepo.GetByID(ctx, id)
}

func (s *GameService) GetWithSides(ctx context.Context, id int64) (*models.Game, []models.GameSide, error) {
	return s.repos.GameRepo.GetWithSides(ctx, id)
}

func (s *GameService) Update(ctx context.Context, id int64, in UpdateGameInput) (*models.Game, error) {
	cur, err := s.repos.GameRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	fields := map[string]any{}

	if in.SeasonID != nil {
		if *in.SeasonID == 0 {
			// allow clearing to exhibition
			fields["season_id"] = nil
		} else {
			if _, err := s.repos.SeasonRepo.GetByID(ctx, *in.SeasonID); err != nil {
				return nil, errors.New("season not found")
			}
			fields["season_id"] = *in.SeasonID
		}
	}

	if in.TargetPoints != nil {
		if *in.TargetPoints <= 0 {
			return nil, errors.New("targetPoints must be > 0")
		}
		fields["target_points"] = *in.TargetPoints
	}

	if in.ScheduledAt != nil {
		if strings.TrimSpace(*in.ScheduledAt) == "" {
			fields["scheduled_at"] = nil
		} else {
			t, err := time.Parse(time.RFC3339, *in.ScheduledAt)
			if err != nil {
				return nil, errors.New("scheduledAt must be RFC3339")
			}
			fields["scheduled_at"] = t
		}
	}

	if in.Timezone != nil {
		if *in.Timezone == "" {
			return nil, errors.New("timezone cannot be empty")
		}
		if _, err := time.LoadLocation(*in.Timezone); err != nil {
			return nil, errors.New("invalid timezone")
		}
		fields["timezone"] = *in.Timezone
	}

	if in.Location != nil {
		// nil clears
		fields["location"] = in.Location
	}
	if in.Description != nil {
		// nil clears
		fields["description"] = in.Description
	}

	if in.Status != nil {
		ns := strings.ToLower(*in.Status)
		switch ns {
		case "scheduled":
			fields["status"] = ns
			fields["started_at"] = nil
			fields["ended_at"] = nil
			fields["winner_side"] = nil
		case "in_progress":
			fields["status"] = ns
			if cur.StartedAt == nil {
				now := time.Now().UTC()
				fields["started_at"] = &now
			}
		case "completed":
			fields["status"] = ns
			if cur.EndedAt == nil {
				now := time.Now().UTC()
				fields["ended_at"] = &now
			}
			// winner is optional here; use the explicit Complete route to set it
		case "canceled":
			fields["status"] = ns
			now := time.Now().UTC()
			fields["ended_at"] = &now
		default:
			return nil, errors.New("invalid status")
		}
	}

	// First update the game row (if there are any game fields)
	var updated *models.Game
	if len(fields) == 0 {
		updated = cur
	} else {
		updated, err = s.repos.GameRepo.UpdateFields(ctx, id, fields)
		if err != nil {
			return nil, err
		}
	}

	// Then update side colors on game_sides, if requested
	if in.SideAColor != nil {
		if err := s.repos.GameRepo.UpdateSideColor(ctx, id, "A", *in.SideAColor); err != nil {
			return nil, err
		}
	}
	if in.SideBColor != nil {
		if err := s.repos.GameRepo.UpdateSideColor(ctx, id, "B", *in.SideBColor); err != nil {
			return nil, err
		}
	}

	return updated, nil
}

func (s *GameService) Delete(ctx context.Context, id int64) error {
	return s.repos.GameRepo.DeleteByID(ctx, id)
}

func (s *GameService) List(ctx context.Context, opts ListGamesOptions) (*PagedGames, error) {
	page := opts.Page
	size := opts.Size
	if page < 1 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 25
	}
	slog.Debug("Listing games", "options", opts, "page", page, "size", size)
	items, total, err := s.repos.GameRepo.List(ctx, repositories.ListGamesFilter{
		SeasonID:       opts.SeasonID,
		ExhibitionOnly: opts.ExhibitionOnly,
		Status:         opts.Status,
		MatchType:      opts.MatchType,
		ScheduledFrom:  opts.ScheduledFrom,
		ScheduledTo:    opts.ScheduledTo,
		TeamID:         opts.TeamID,
		PlayerID:       opts.PlayerID,
		Offset:         (page - 1) * size,
		Limit:          size,
		OrderBy:        opts.OrderBy,
	})
	if err != nil {
		return nil, err
	}

	slog.Info("Listed games", "returned", len(items), "total", total)
	return &PagedGames{Data: items, Total: total, Page: page, Size: size}, nil
}

/* =========================
   Helpers
========================= */

func (s *GameService) buildSide(ctx context.Context, label string, matchType string, in GameParticipantInput) (models.GameSide, error) {
	var color models.DiscColor = models.DiscNatural
	if in.Color != nil && *in.Color != "" {
		switch *in.Color {
		case models.DiscWhite, models.DiscBlack, models.DiscNatural:
			color = *in.Color
		default:
			return models.GameSide{}, errors.New("invalid color for side " + label)
		}
	}

	gs := models.GameSide{Side: label, Color: color, Points: 0}

	if matchType == "teams" {
		if in.TeamID == nil || *in.TeamID <= 0 {
			return gs, errors.New("side " + label + ": teamId is required for team match")
		}
		if _, err := s.repos.TeamRepo.GetByID(ctx, *in.TeamID); err != nil {
			return gs, errors.New("side " + label + ": team not found")
		}
		gs.TeamID = in.TeamID
		gs.PlayerID = nil
	} else {
		if in.PlayerID == nil || *in.PlayerID <= 0 {
			return gs, errors.New("side " + label + ": playerId is required for player match")
		}
		if _, err := s.repos.PlayerRepo.GetByID(ctx, *in.PlayerID); err != nil {
			return gs, errors.New("side " + label + ": player not found")
		}
		gs.PlayerID = in.PlayerID
		gs.TeamID = nil
	}
	return gs, nil
}

// CompleteWithWinner sets winner, marks completed, stamps ended_at.
// If already completed: idempotently ensure winner_side is set/updated.
func (s *GameService) CompleteWithWinner(ctx context.Context, id int64, winner string) (*models.Game, error) {
	w := strings.ToUpper(strings.TrimSpace(winner))
	if w != "A" && w != "B" {
		return nil, errors.New("winnerSide must be 'A' or 'B'")
	}

	cur, err := s.repos.GameRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if cur.Status == "canceled" {
		return nil, errors.New("cannot complete a canceled game")
	}

	now := time.Now().UTC()
	fields := map[string]any{
		"status":      "completed",
		"winner_side": w,
	}
	if cur.EndedAt == nil {
		fields["ended_at"] = &now
	}
	return s.repos.GameRepo.UpdateFields(ctx, id, fields)
}
