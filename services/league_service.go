package services

import (
	"context"
	"errors"
	"strings"

	"github.com/matt-j-deasy/betty-crokers-api/models"
	"github.com/matt-j-deasy/betty-crokers-api/repositories"
)

type LeagueService struct {
	repo *repositories.LeagueRepository
}

func NewLeagueService(repos *repositories.RepositoriesCollection) *LeagueService {
	return &LeagueService{repo: repos.LeagueRepo}
}

type CreateLeagueInput struct {
	Name string `json:"name"`
}

type UpdateLeagueInput struct {
	Name *string `json:"name"`
}

type ListLeaguesOptions struct {
	Search string
	Page   int
	Size   int
}

type PagedLeagues struct {
	Data  []models.League `json:"data"`
	Total int64           `json:"total"`
	Page  int             `json:"page"`
	Size  int             `json:"size"`
}

func (s *LeagueService) Create(ctx context.Context, in CreateLeagueInput) (*models.League, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, errors.New("name is required")
	}
	l := &models.League{Name: name}
	if err := s.repo.Create(ctx, l); err != nil {
		return nil, err
	}
	return l, nil
}

func (s *LeagueService) GetByID(ctx context.Context, id int64) (*models.League, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *LeagueService) Update(ctx context.Context, id int64, in UpdateLeagueInput) (*models.League, error) {
	fields := map[string]any{}

	if in.Name != nil {
		n := strings.TrimSpace(*in.Name)
		if n == "" {
			return nil, errors.New("name cannot be empty")
		}
		fields["name"] = n
	}

	if len(fields) == 0 {
		return s.repo.GetByID(ctx, id)
	}
	return s.repo.UpdateFields(ctx, id, fields)
}

func (s *LeagueService) Delete(ctx context.Context, id int64) error {
	return s.repo.DeleteByID(ctx, id)
}

func (s *LeagueService) List(ctx context.Context, opts ListLeaguesOptions) (*PagedLeagues, error) {
	page := opts.Page
	size := opts.Size
	if page < 1 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 25
	}

	items, total, err := s.repo.List(ctx, repositories.ListLeaguesFilter{
		Search: opts.Search,
		Offset: (page - 1) * size,
		Limit:  size,
	})
	if err != nil {
		return nil, err
	}
	return &PagedLeagues{
		Data:  items,
		Total: total,
		Page:  page,
		Size:  size,
	}, nil
}
