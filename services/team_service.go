package services

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	"github.com/matt-j-deasy/betty-crokers-api/models"
	"github.com/matt-j-deasy/betty-crokers-api/repositories"
)

type TeamService struct {
	repos *repositories.RepositoriesCollection
}

func NewTeamService(repos *repositories.RepositoriesCollection) *TeamService {
	return &TeamService{repos: repos}
}

// ---------- Inputs / Outputs

type CreateTeamInput struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	PlayerAID   int64   `json:"playerAId"`
	PlayerBID   int64   `json:"playerBId"`
}

type UpdateTeamInput struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	PlayerAID   *int64  `json:"playerAId"`
	PlayerBID   *int64  `json:"playerBId"`
}

type ListTeamsOptions struct {
	Search     string
	PlayerID   *int64
	SeasonID   *int64
	OnlyActive *bool
	Page       int
	Size       int
}

type PagedTeams struct {
	Data  []models.Team `json:"data"`
	Total int64         `json:"total"`
	Page  int           `json:"page"`
	Size  int           `json:"size"`
}

// ---------- Public API

func (s *TeamService) Create(ctx context.Context, in CreateTeamInput) (*models.Team, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, errors.New("name is required")
	}
	if in.PlayerAID <= 0 || in.PlayerBID <= 0 {
		return nil, errors.New("playerAId and playerBId are required")
	}
	if in.PlayerAID == in.PlayerBID {
		return nil, errors.New("players must be distinct")
	}

	// Validate players exist (and not soft-deleted)
	if _, err := s.repos.PlayerRepo.GetByID(ctx, in.PlayerAID); err != nil {
		return nil, errors.New("playerAId not found")
	}
	if _, err := s.repos.PlayerRepo.GetByID(ctx, in.PlayerBID); err != nil {
		return nil, errors.New("playerBId not found")
	}

	// Canonicalize pair and check uniqueness
	a, b := canonicalPair(in.PlayerAID, in.PlayerBID)
	if exists, err := s.repos.TeamRepo.ExistsByPlayers(ctx, a, b); err != nil {
		return nil, err
	} else if exists {
		return nil, errors.New("team for this player pair already exists")
	}

	t := &models.Team{
		Name:        name,
		Description: in.Description,
		PlayerAID:   a,
		PlayerBID:   b,
	}
	if err := s.repos.TeamRepo.Create(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *TeamService) GetByID(ctx context.Context, id int64) (*models.Team, error) {
	return s.repos.TeamRepo.GetByID(ctx, id)
}

func (s *TeamService) Update(ctx context.Context, id int64, in UpdateTeamInput) (*models.Team, error) {
	cur, err := s.repos.TeamRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	fields := map[string]any{}

	if in.Name != nil {
		n := strings.TrimSpace(*in.Name)
		if n == "" {
			return nil, errors.New("name cannot be empty")
		}
		fields["name"] = n
	}
	if in.Description != nil {
		// allow null to clear
		fields["description"] = in.Description
	}

	// If either player id is provided, compute the new canonical pair
	if in.PlayerAID != nil || in.PlayerBID != nil {
		newA := cur.PlayerAID
		newB := cur.PlayerBID
		if in.PlayerAID != nil {
			newA = *in.PlayerAID
		}
		if in.PlayerBID != nil {
			newB = *in.PlayerBID
		}
		if newA <= 0 || newB <= 0 {
			return nil, errors.New("playerAId and playerBId must be > 0")
		}
		if newA == newB {
			return nil, errors.New("players must be distinct")
		}

		// Validate players exist
		if _, err := s.repos.PlayerRepo.GetByID(ctx, newA); err != nil {
			return nil, errors.New("playerAId not found")
		}
		if _, err := s.repos.PlayerRepo.GetByID(ctx, newB); err != nil {
			return nil, errors.New("playerBId not found")
		}

		a, b := canonicalPair(newA, newB)

		// If pair changed, ensure no collision with another team
		if a != cur.PlayerAID || b != cur.PlayerBID {
			if t, err := s.repos.TeamRepo.GetByPlayers(ctx, a, b); err == nil {
				if t.ID != cur.ID {
					return nil, errors.New("another team with this player pair already exists")
				}
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, err
			}
		}

		fields["player_a_id"] = a
		fields["player_b_id"] = b
	}

	if len(fields) == 0 {
		return cur, nil
	}
	return s.repos.TeamRepo.UpdateFields(ctx, id, fields)
}

func (s *TeamService) Delete(ctx context.Context, id int64) error {
	return s.repos.TeamRepo.DeleteByID(ctx, id)
}

func (s *TeamService) List(ctx context.Context, opts ListTeamsOptions) (*PagedTeams, error) {
	page := opts.Page
	size := opts.Size
	if page < 1 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 25
	}

	items, total, err := s.repos.TeamRepo.List(ctx, repositories.ListTeamsFilter{
		Search:     opts.Search,
		PlayerID:   opts.PlayerID,
		SeasonID:   opts.SeasonID,
		OnlyActive: opts.OnlyActive,
		Offset:     (page - 1) * size,
		Limit:      size,
	})
	if err != nil {
		return nil, err
	}
	return &PagedTeams{
		Data:  items,
		Total: total,
		Page:  page,
		Size:  size,
	}, nil
}

// --- helpers ---

func canonicalPair(a, b int64) (int64, int64) {
	if a > b {
		return b, a
	}
	return a, b
}
