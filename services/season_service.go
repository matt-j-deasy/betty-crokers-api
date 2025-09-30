package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/matt-j-deasy/betty-crokers-api/models"
	"github.com/matt-j-deasy/betty-crokers-api/repositories"
)

type SeasonService struct {
	repo *repositories.SeasonRepository
}

func NewSeasonService(repos *repositories.RepositoriesCollection) *SeasonService {
	return &SeasonService{repo: repos.SeasonRepo}
}

// -------- Inputs / Outputs

type CreateSeasonInput struct {
	LeagueID    int64   `json:"leagueId"`
	Name        string  `json:"name"`
	StartsOn    string  `json:"startsOn"`           // "YYYY-MM-DD"
	EndsOn      string  `json:"endsOn"`             // "YYYY-MM-DD"
	Timezone    *string `json:"timezone,omitempty"` // IANA; default from model if nil/empty
	Description *string `json:"description,omitempty"`
}

type UpdateSeasonInput struct {
	LeagueID    *int64  `json:"leagueId,omitempty"`
	Name        *string `json:"name,omitempty"`
	StartsOn    *string `json:"startsOn,omitempty"` // "YYYY-MM-DD"
	EndsOn      *string `json:"endsOn,omitempty"`   // "YYYY-MM-DD"
	Timezone    *string `json:"timezone,omitempty"` // IANA
	Description *string `json:"description,omitempty"`
}

type ListSeasonsOptions struct {
	Search   string
	LeagueID *int64
	Page     int
	Size     int
}

type PagedSeasons struct {
	Data  []models.Season `json:"data"`
	Total int64           `json:"total"`
	Page  int             `json:"page"`
	Size  int             `json:"size"`
}

// -------- Public API

func (s *SeasonService) Create(ctx context.Context, in CreateSeasonInput) (*models.Season, error) {
	if in.LeagueID <= 0 {
		return nil, errors.New("leagueId is required")
	}
	if in.Name == "" {
		return nil, errors.New("name is required")
	}

	// Starts/Ends dates are optional
	start := time.Time{}
	if strings.TrimSpace(in.StartsOn) != "" {
		s, err := parseYMD(in.StartsOn)
		if err != nil {
			return nil, errors.New("startsOn must be YYYY-MM-DD")
		}
		start = s
	}

	end := time.Time{}
	if strings.TrimSpace(in.EndsOn) != "" {
		e, err := parseYMD(in.EndsOn)
		if err != nil {
			return nil, errors.New("endsOn must be YYYY-MM-DD")
		}
		end = e
	}

	if !start.IsZero() && !end.IsZero() && end.Before(start) {
		return nil, errors.New("endsOn must be on or after startsOn")
	}

	tz := "America/New_York"
	if in.Timezone != nil && *in.Timezone != "" {
		if _, err := time.LoadLocation(*in.Timezone); err != nil {
			return nil, errors.New("timezone must be a valid IANA timezone")
		}
		tz = *in.Timezone
	}

	season := &models.Season{
		LeagueID:    in.LeagueID,
		Name:        in.Name,
		StartsOn:    start,
		EndsOn:      end,
		Timezone:    tz,
		Description: in.Description,
	}
	if err := s.repo.Create(ctx, season); err != nil {
		return nil, err
	}
	return season, nil
}

func (s *SeasonService) GetByID(ctx context.Context, id int64) (*models.Season, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *SeasonService) Update(ctx context.Context, id int64, in UpdateSeasonInput) (*models.Season, error) {
	cur, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	fields := map[string]any{}

	if in.LeagueID != nil {
		if *in.LeagueID <= 0 {
			return nil, errors.New("leagueId must be > 0")
		}
		fields["league_id"] = *in.LeagueID
	}

	if in.Name != nil {
		if *in.Name == "" {
			return nil, errors.New("name cannot be empty")
		}
		fields["name"] = *in.Name
	}

	// Handle date invariants with current values as defaults
	newStart := cur.StartsOn
	newEnd := cur.EndsOn

	if in.StartsOn != nil {
		start, err := parseYMD(*in.StartsOn)
		if err != nil {
			return nil, errors.New("startsOn must be YYYY-MM-DD")
		}
		newStart = start
		fields["starts_on"] = start
	}
	if in.EndsOn != nil {
		end, err := parseYMD(*in.EndsOn)
		if err != nil {
			return nil, errors.New("endsOn must be YYYY-MM-DD")
		}
		newEnd = end
		fields["ends_on"] = end
	}
	if newEnd.Before(newStart) {
		return nil, errors.New("endsOn must be on or after startsOn")
	}

	if in.Timezone != nil {
		if *in.Timezone == "" {
			return nil, errors.New("timezone cannot be empty")
		}
		if _, err := time.LoadLocation(*in.Timezone); err != nil {
			return nil, errors.New("timezone must be a valid IANA timezone")
		}
		fields["timezone"] = *in.Timezone
	}

	if in.Description != nil {
		fields["description"] = in.Description // can be nil to clear
	}

	if len(fields) == 0 {
		return cur, nil
	}
	return s.repo.UpdateFields(ctx, id, fields)
}

func (s *SeasonService) Delete(ctx context.Context, id int64) error {
	return s.repo.DeleteByID(ctx, id)
}

func (s *SeasonService) List(ctx context.Context, opts ListSeasonsOptions) (*PagedSeasons, error) {
	page := opts.Page
	size := opts.Size
	if page < 1 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 25
	}
	items, total, err := s.repo.List(ctx, repositories.ListSeasonsFilter{
		Search:   opts.Search,
		LeagueID: opts.LeagueID,
		Offset:   (page - 1) * size,
		Limit:    size,
	})
	if err != nil {
		return nil, err
	}
	return &PagedSeasons{
		Data:  items,
		Total: total,
		Page:  page,
		Size:  size,
	}, nil
}

type SeasonStandings []repositories.SeasonStandingsRow

func (s *SeasonService) GetStandings(ctx context.Context, seasonID int64) (SeasonStandings, error) {
	// Optional: check that season exists to return 404 vs empty list
	if _, err := s.repo.GetByID(ctx, seasonID); err != nil {
		return nil, err
	}
	rows, err := s.repo.GetStandings(ctx, seasonID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

type PlayerStandingDTO struct {
	PlayerID      int64   `json:"playerId"`
	Games         int64   `json:"games"`
	Wins          int64   `json:"wins"`
	Losses        int64   `json:"losses"`
	PointsFor     int64   `json:"pointsFor"`
	PointsAgainst int64   `json:"pointsAgainst"`
	PointDiff     int64   `json:"pointDiff"`
	WinPct        float64 `json:"winPct"`
	Rank          int     `json:"rank"`
}

func (s *SeasonService) ListPlayerStandings(ctx context.Context, seasonID int64, limit int, cursor *repositories.PlayerStandingsCursor) ([]PlayerStandingDTO, *string, error) {
	rows, next, err := s.repo.ListPlayerStandings(ctx, repositories.ListPlayerStandingsQuery{
		SeasonID: seasonID, Limit: limit, Cursor: cursor,
	})
	if err != nil {
		return nil, nil, err
	}

	out := make([]PlayerStandingDTO, 0, len(rows))
	for i, r := range rows {
		out = append(out, PlayerStandingDTO{
			PlayerID: r.PlayerID, Games: r.Games, Wins: r.Wins, Losses: r.Losses,
			PointsFor: r.PointsFor, PointsAgainst: r.PointsAgainst, PointDiff: r.PointDiff,
			WinPct: r.WinPct, Rank: i + 1,
		})
	}

	var nc *string
	if next != nil {
		b, _ := json.Marshal(next)
		enc := base64.StdEncoding.EncodeToString(b)
		nc = &enc
	}
	return out, nc, nil
}

// -------- Helpers

func parseYMD(s string) (time.Time, error) {
	// treat as date-only, location-agnostic; stored as DATE by GORM
	return time.Parse("2006-01-02", s)
}
