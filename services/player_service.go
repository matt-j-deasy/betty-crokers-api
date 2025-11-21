package services

import (
	"context"
	"errors"
	"strings"

	"github.com/matt-j-deasy/betty-crokers-api/models"
	"github.com/matt-j-deasy/betty-crokers-api/repositories"
)

type PlayerService struct {
	repo *repositories.PlayerRepository
}

func NewPlayerService(repos *repositories.RepositoriesCollection) *PlayerService {
	return &PlayerService{
		repo: repos.PlayerRepo,
	}
}

type CreatePlayerInput struct {
	UserID    *int64  `json:"userId"`
	Nickname  string  `json:"nickname"`
	FirstName *string `json:"firstName"`
	LastName  *string `json:"lastName"`
}

type UpdatePlayerInput struct {
	Nickname  *string `json:"nickname"`
	FirstName *string `json:"firstName"`
	LastName  *string `json:"lastName"`
	UserID    *int64  `json:"userId"` // nullable; can set or clear by passing null
}

type ListPlayersOptions struct {
	Search string
	Page   int
	Size   int
}

type PagedPlayers struct {
	Data  []models.Player `json:"data"`
	Total int64           `json:"total"`
	Page  int             `json:"page"`
	Size  int             `json:"size"`
}

func (s *PlayerService) Create(ctx context.Context, in CreatePlayerInput) (*models.Player, error) {
	if strings.TrimSpace(in.Nickname) == "" {
		return nil, errors.New("nickname is required")
	}
	p := &models.Player{
		UserID:    in.UserID,
		Nickname:  strings.TrimSpace(in.Nickname),
		FirstName: in.FirstName,
		LastName:  in.LastName,
	}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *PlayerService) GetByID(ctx context.Context, id int64) (*models.Player, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *PlayerService) Update(ctx context.Context, id int64, in UpdatePlayerInput) (*models.Player, error) {
	fields := map[string]any{}
	if in.Nickname != nil {
		n := strings.TrimSpace(*in.Nickname)
		if n == "" {
			return nil, errors.New("nickname cannot be empty")
		}
		fields["nickname"] = n
	}
	if in.FirstName != nil {
		fields["first_name"] = *in.FirstName
	}
	if in.LastName != nil {
		fields["last_name"] = *in.LastName
	}
	// To clear the user link, client can send "userId": null
	if in.UserID != nil || (in.UserID == nil) {
		fields["user_id"] = in.UserID
	}
	if len(fields) == 0 {
		// no-op; fetch and return current
		return s.repo.GetByID(ctx, id)
	}
	return s.repo.UpdateFields(ctx, id, fields)
}

func (s *PlayerService) Delete(ctx context.Context, id int64) error {
	return s.repo.DeleteByID(ctx, id)
}

func (s *PlayerService) List(ctx context.Context, opts ListPlayersOptions) (*PagedPlayers, error) {
	page := opts.Page
	size := opts.Size
	if page < 1 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 25
	}
	items, total, err := s.repo.List(ctx, repositories.ListPlayersFilter{
		Search: opts.Search,
		Offset: (page - 1) * size,
		Limit:  size,
	})
	if err != nil {
		return nil, err
	}
	return &PagedPlayers{
		Data:  items,
		Total: total,
		Page:  page,
		Size:  size,
	}, nil
}

type PlayerDuplicateGameRow struct {
	GameID     int64   `json:"gameId"`
	SeasonID   *int64  `json:"seasonId"`
	MatchType  string  `json:"matchType"`
	Status     string  `json:"status"`
	WinnerSide *string `json:"winnerSide"`

	Side     string `json:"side"`
	Color    string `json:"color"`
	RowCount int64  `json:"rowCount"` // number of per-player rows for (player, game)
}

func (s *PlayerService) ListPlayerDuplicateGames(
	ctx context.Context,
	playerID int64,
) ([]PlayerDuplicateGameRow, error) {
	rows, err := s.repo.ListPlayerDuplicateGames(ctx, playerID)
	if err != nil {
		return nil, err
	}

	out := make([]PlayerDuplicateGameRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, PlayerDuplicateGameRow{
			GameID:     r.GameID,
			SeasonID:   r.SeasonID,
			MatchType:  r.MatchType,
			Status:     r.Status,
			WinnerSide: r.WinnerSide,
			Side:       r.Side,
			Color:      r.Color,
			RowCount:   r.RowCount,
		})
	}
	return out, nil
}
