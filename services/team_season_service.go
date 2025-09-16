package services

import (
	"context"
	"errors"

	"github.com/matt-j-deasy/betty-crokers-api/models"
	"github.com/matt-j-deasy/betty-crokers-api/repositories"
)

type TeamSeasonService struct {
	repos *repositories.RepositoriesCollection
}

func NewTeamSeasonService(repos *repositories.RepositoriesCollection) *TeamSeasonService {
	return &TeamSeasonService{repos: repos}
}

// ---------- Inputs / Outputs

type LinkTeamSeasonInput struct {
	TeamID   int64 `json:"teamId"`
	SeasonID int64 `json:"seasonId"`
	// default true if omitted by caller; service treats zero-value as true at Create
	IsActive *bool `json:"isActive,omitempty"`
}

type SetTeamSeasonActiveInput struct {
	TeamID   int64 `json:"teamId"`
	SeasonID int64 `json:"seasonId"`
	IsActive bool  `json:"isActive"`
}

type ListTeamSeasonsOptions struct {
	TeamID     *int64
	SeasonID   *int64
	OnlyActive *bool
	Page       int
	Size       int
}

type PagedTeamSeasons struct {
	Data  []models.TeamSeason `json:"data"`
	Total int64               `json:"total"`
	Page  int                 `json:"page"`
	Size  int                 `json:"size"`
}

// ---------- Public API

// Link (idempotent). If a soft-deleted link exists, it is revived; otherwise created.
func (s *TeamSeasonService) Link(ctx context.Context, in LinkTeamSeasonInput) (*models.TeamSeason, error) {
	if in.TeamID <= 0 || in.SeasonID <= 0 {
		return nil, errors.New("teamId and seasonId are required")
	}
	// Validate existence
	if _, err := s.repos.TeamRepo.GetByID(ctx, in.TeamID); err != nil {
		return nil, errors.New("team not found")
	}
	if _, err := s.repos.SeasonRepo.GetByID(ctx, in.SeasonID); err != nil {
		return nil, errors.New("season not found")
	}
	active := true
	if in.IsActive != nil {
		active = *in.IsActive
	}
	return s.repos.TeamSeasonRepo.UpsertLink(ctx, in.TeamID, in.SeasonID, active)
}

func (s *TeamSeasonService) SetActive(ctx context.Context, in SetTeamSeasonActiveInput) (*models.TeamSeason, error) {
	if in.TeamID <= 0 || in.SeasonID <= 0 {
		return nil, errors.New("teamId and seasonId are required")
	}
	// Ensure link exists
	if _, err := s.repos.TeamSeasonRepo.Get(ctx, in.TeamID, in.SeasonID); err != nil {
		return nil, err
	}
	return s.repos.TeamSeasonRepo.SetActive(ctx, in.TeamID, in.SeasonID, in.IsActive)
}

func (s *TeamSeasonService) Unlink(ctx context.Context, teamID, seasonID int64) error {
	if teamID <= 0 || seasonID <= 0 {
		return errors.New("teamId and seasonId are required")
	}
	// Idempotent: deleting a non-existent row is fine (no-op) with soft deletes,
	// but the repo will return nil only if the Delete query itself succeeds.
	return s.repos.TeamSeasonRepo.Unlink(ctx, teamID, seasonID)
}

func (s *TeamSeasonService) List(ctx context.Context, opts ListTeamSeasonsOptions) (*PagedTeamSeasons, error) {
	page := opts.Page
	size := opts.Size
	if page < 1 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 25
	}
	items, total, err := s.repos.TeamSeasonRepo.List(ctx, repositories.ListTeamSeasonsFilter{
		TeamID:     opts.TeamID,
		SeasonID:   opts.SeasonID,
		OnlyActive: opts.OnlyActive,
		Offset:     (page - 1) * size,
		Limit:      size,
	})
	if err != nil {
		return nil, err
	}
	return &PagedTeamSeasons{
		Data:  items,
		Total: total,
		Page:  page,
		Size:  size,
	}, nil
}

// Convenience passthroughs

func (s *TeamSeasonService) ListSeasonsForTeam(ctx context.Context, teamID int64, onlyActive *bool) ([]models.Season, error) {
	if teamID <= 0 {
		return nil, errors.New("teamId must be > 0")
	}
	return s.repos.TeamSeasonRepo.ListSeasonsForTeam(ctx, teamID, onlyActive)
}

func (s *TeamSeasonService) ListTeamsForSeason(ctx context.Context, seasonID int64, onlyActive *bool) ([]models.Team, error) {
	if seasonID <= 0 {
		return nil, errors.New("seasonId must be > 0")
	}
	return s.repos.TeamSeasonRepo.ListTeamsForSeason(ctx, seasonID, onlyActive)
}
